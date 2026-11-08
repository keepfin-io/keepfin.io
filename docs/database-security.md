# Database roles, identity linking and request limits

Keepfin uses one **dedicated** PostgreSQL database. Separation is enforced by
database grants, not table naming or TypeScript types. Do not run the provisioning
scripts against a database shared with other applications: they adopt existing
public tables and revoke public access.

| Role | Used by | Access |
| --- | --- | --- |
| `keepfin_runtime` | Go HTTP server/workers | Application data; no auth tables or DDL. Existing users allow profile/soft-delete updates only. |
| `auth_runtime` | Auth HTTP server | CRUD on `auth_*`; no application tables or DDL. |
| `keepfin_migrator` | One-shot Go migration job | Owns application tables; not supplied to the HTTP server. |
| `auth_migrator` | One-shot auth migration job | Owns auth tables; not supplied to the HTTP server. |
| `identity_operator` | Group granted to named human DB logins | Execute the audited link function only; cannot log in itself. |
| `keepfin_security_owner` | Non-login function/table owner | Minimal linking privileges and ownership of the protected audit schema. |
| PostgreSQL administrator | Bootstrap/permission jobs | Trusted provisioning authority; not a runtime credential. |

The one-shot development `auth-seed` job also uses administrator credentials to
create the explicitly enabled test identity. It has no HTTP listener and is absent
from production Compose. Auth runtime does not seed users or run migrations.

## Install or upgrade

Production needs `POSTGRES_PASSWORD`, `KEEPFIN_DB_PASSWORD`,
`KEEPFIN_MIGRATION_DB_PASSWORD`, `AUTH_DB_PASSWORD` and
`AUTH_MIGRATION_DB_PASSWORD`. Generate independent passwords, for example with
`openssl rand -hex 32`. Runtime/migration passwords must be distinct and at least
16 characters. Keep migration/admin secrets out of runtime container environments.
Compose supplies passwords separately from connection URLs, so special characters
do not need URL encoding there.

For an existing installation:

1. Back up the database and arrange a maintenance window. Use the **same Compose
   project name** so the existing database volume is retained.
2. Stop the stack without deleting volumes (`docker compose down`, **never `-v`**).
3. Add the new variables from `.env.example`, including the trusted TLS proxy CIDR.
4. Recreate with `docker compose up --build -d`.
5. Check `db-roles`, `keepnode-migrate`, `auth-migrate` and `db-security` exited with
   status 0 before accepting traffic.

For development, run `make down` followed by `make hot`. The development file has
local-only credentials and retains the existing `keepnode` project/volume names.
Do not use its credentials for a public installation.

`db-roles` creates roles, transfers existing object ownership and configures default
privileges. The two migrators then apply their own migrations. `db-security` applies
permissions to existing/new tables and installs the audited linking function.
The jobs are repeatable. They preserve rows, UUIDs, sessions and financial history;
they do not reset the database. A failed job blocks dependent runtime startup.

The Compose networks reserve `172.30.80.0/24` (production) and `172.30.81.0/24`
(development). The frontend is `.10`, Go is `.11`; other containers get dynamic IPs.
If these subnets conflict with your network, change the subnet, static addresses
and trusted proxy CIDRs together. An old default Docker network must be recreated
with `down`/`up`; volumes survive that operation.

Standalone Go startup is now explicit: run `keepnode migrate` with the migration
role, then launch `keepnode` with the runtime role. Do not reintroduce automatic
HTTP-startup migration or use the administrator as a fallback.

## Link a reviewed legacy identity

The operator must independently verify ownership of the historical application
account. A verified current email and a matching legacy email are necessary but
do not replace that review. Record the reviewed support ticket or migration record.

A database administrator provisions a **personal** login using `psql`:

```sql
CREATE ROLE keepfin_operator_alex LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE;
\password keepfin_operator_alex
GRANT identity_operator TO keepfin_operator_alex;
```

Use your normal privileged-access controls (for example, a restricted network and
time-limited credentials). Never grant membership to an application role. Supply
`IDENTITY_OPERATOR_DATABASE_URL` through the operator's secret environment, then:

```sh
npm --prefix auth run build
npm --prefix auth run link-user -- APPLICATION_UUID AUTH_UUID \
  --reason "Verified legacy account migration" \
  --ownership-reference "SUPPORT-123" --confirm
```

The CLI calls a fixed `SECURITY DEFINER` function with bound parameters. The
function has a fixed safe search path, checks role membership and locks the two
rows before validating verification, email, status and conflicts. Existing mappings
to another identity are rejected. The update and audit insertion commit together;
if auditing fails, linking fails. A successful invocation prints its audit UUID.

Audit attribution comes from PostgreSQL `session_user`, **not** a caller-provided
`--actor`. The record contains target UUIDs, the previous mapping, the new auth ID,
reason, review reference, time and database client address. A retry of an already
matching link creates another audit record with the prior mapping present. On a
connection error, inspect the audit before retrying: a lost response does not prove
that PostgreSQL rolled the transaction back.

Administrators can inspect records and revoke access after the operation:

```sql
SELECT * FROM keepfin_security.identity_link_audit ORDER BY occurred_at DESC;
REVOKE identity_operator FROM keepfin_operator_alex;
ALTER ROLE keepfin_operator_alex NOLOGIN;
```

Runtime roles and operators cannot modify/delete the audit table. Database owners
and administrators remain trusted; this is not an external tamper-proof log. Failed
link attempts return errors without a successful-link audit row. The CLI does not
print SQL, credentials or identity details on failure. Keep PostgreSQL/privileged
access logs according to your operational retention policy.

## Rate limits and trusted proxies

Both services enforce limits. Go has a pre-auth IP budget and a post-auth user
budget. Auth has endpoint-specific login budgets and a session-check budget. Go
forwards only the resolved client IP, so all users do not share the API container's
session quota. Headers from untrusted peers are ignored; IPv6 /64 addresses share
an IP budget. HTTP 429 responses include `Retry-After`.

The production TLS proxy must overwrite `X-Forwarded-For`; configure its exact
address in `TRUSTED_EDGE_PROXY_CIDR`. Nginx overwrites the internal `X-Real-IP` header.
Go trusts only the frontend; auth trusts only the frontend and Go. Do not expose
private services directly or trust arbitrary private ranges/all addresses.

See `.env.example` for budgets. Auth counters persist in PostgreSQL; Go counters
are per-process and reset on restart. For multiple Go replicas, add a shared store
or deployment-wide ingress quota. Limits reduce abuse; they are not a DDoS service.

## Verify

`sh scripts/test-db-security.sh` creates a disposable PostgreSQL container, tests
legacy ownership migration and fresh setup, then runs auth and Go tests against it.
It removes the container and test volume on exit and never uses the application DB.
The suite checks denied cross-service reads/DDL, named-operator attribution,
mapping validation, audit atomicity and normal runtime data operations.
