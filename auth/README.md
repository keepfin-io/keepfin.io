# Keepfin authentication

Vue uses the official `better-auth/vue` client. A private Node 26 TypeScript service owns
Google OAuth, Magic Links and opaque cookie sessions. The Go API calls that
service's `/api/auth/get-session` on every protected request, then resolves the
immutable auth user ID to the existing application UUID. Go remains responsible
for account status, financial-record ownership and authorization.

```text
Browser → same-origin frontend proxy
          ├─ /api/auth/* → Better Auth → auth_* PostgreSQL tables
          └─ /api/v1/*   → Go → Better Auth session verification
                             → users.auth_user_id → application UUID
```

## Development

Run `make hot`. This creates a random secret in ignored `.env.auth.local`, runs
additive auth migrations, and starts Vue, Go, auth and PostgreSQL. Open
`http://localhost:8080/login`. `/register` redirects to `/login`.

Email delivery is deliberately local-only: enter an email and open the `url` in
the corresponding ignored `.auth-mail/*.json` file. Nothing is sent externally.
Messages contain credentials: do not publish, log or commit them. There is no
HTTP inbox endpoint. Links expire in five minutes and are atomically single use.
The successful link automatically creates an auth identity and application user.

For Google, put both `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` in
`.env.auth.local`, then recreate the auth service:

```sh
docker compose -p keepnode -f docker-compose.hot.yml up -d --force-recreate auth
```

Configure the OAuth client as a web application with this redirect URI:
`http://localhost:8080/api/auth/callback/google`. The authorized JavaScript origin
is `http://localhost:8080`. Google and Apple buttons are always visible. Google
is disabled with a "Coming soon" label until configured; Apple is a disabled
placeholder and has no OAuth integration yet. Keep secrets out of `VUE_APP_*` variables and the frontend bundle.

### Internal testing admin

The hot-development Compose stack sets `AUTH_INTERNAL_TESTING=true`. After the
Go application migrations and permission setup finish, the one-shot `auth-seed`
job atomically seeds one account. The auth HTTP service cannot access application
user or financial tables:

- Username: `internal-testing` (email alias: `internal-testing@keepfin.test`)
- Password: `12345`
- Application role: `admin`

Open **Internal testing sign-in** on `/login`. This uses Better Auth credential
verification and normal HttpOnly sessions; `/api/v1/user/me` resolves the seeded
application UUID and admin role. Passwords are hashed with Better Auth's own
hasher. Public password sign-up remains disabled and only this test identity can
use password sign-in. Ordinary users continue using Magic Links or configured Google.

The flag defaults off outside the hot stack and is accepted only with
`NODE_ENV=development` or `test`. Production rejects it. Seed runs preserve IDs,
passwords, preferences and financial data; any conflicting identity fails without
automatically linking accounts by email. Email matching never grants admin access.
To turn the development account off, set the Compose flag to `"false"` and
recreate auth. To invalidate existing sessions too, sign out or revoke them.

## Registration

Registration is part of the first successful Google or Magic Link sign-in. Better
Auth creates the verified auth identity; the first protected Go API request then
creates the application user with the `user` role. Concurrent first requests retain
one application UUID. An existing legacy email requires explicit reviewed mapping.

There is no separate registration form: `/register` redirects to `/login`.
Public email/password sign-up is disabled. In production, Google must be configured;
Magic Link delivery is currently available only through the development file adapter,
so production email registration still needs a real delivery integration.

## Existing users

The previous Go password endpoints, legacy demo seed and hardcoded-secret JWT
implementation are removed. The development test identity above is separate. Old `auth_token` cookies no longer authenticate.
The old password column remains for migration compatibility but is unused and
never serialized. Existing user UUIDs, account ownership and preferences remain.

A new verified auth identity matching an existing application email receives
HTTP 409 until an operator explicitly maps it. Email equality alone does not
authorize inheriting historical financial records. Verify ownership through a
trusted migration/support process before running the tool from an operator workstation:

```sh
# Supply a personal database login that is a member of identity_operator.
# Inject IDENTITY_OPERATOR_DATABASE_URL through your secret manager or shell.
npm --prefix auth run build
npm --prefix auth run link-user -- APPLICATION_UUID AUTH_UUID \
  --reason "Verified legacy account migration" \
  --ownership-reference "SUPPORT-123" --confirm
```

The tool does not accept the runtime `DATABASE_URL` as a fallback. It can only call
`keepfin_security.link_identity`; the operator has no direct UPDATE privilege on
`users`. The function checks verified email, email equality, active/non-deleted
status and conflicting mappings under row locks. Mapping and audit insertion are
one transaction. The audit records the authenticated PostgreSQL `session_user`,
target IDs, previous mapping, timestamp, reason and ownership-review reference.
Runtime services and operators cannot alter or delete audit records. Database
administrators remain trusted. A reference documents an **external ownership
review**; entering a ticket number does not itself verify ownership.

See [database security and upgrade instructions](../docs/database-security.md)
for personal operator roles, audit inspection and credential separation.

## TypeScript development

Use Node 26.x and npm 11.x; `packageManager` records npm 11.20.0 as the reference
version. `devEngines` rejects unsupported tooling before install/ci/run, and
`.npmrc` enables strict engine checks, exact dependency versions and disables
automatic install/pre/post scripts. Explicit `npm run`, `npm test` and `npm start`
commands still work. Both Docker stages copy this policy before `npm ci`.
If a future dependency requires an installation script, review it before changing
the policy. Keep the lockfile committed and use `npm ci` for reproducible installs.
The existing `private: true` setting prevents accidental npm publication.

Install dependencies with `npm --prefix auth ci`. All application and test sources
are TypeScript. `npm --prefix auth run typecheck` checks both with strict null,
optional-property and array-index checking. PostgreSQL query results have explicit
row types; external JSON is validated before tests use its fields.

Application type definitions live in `src/types/`, grouped into configuration,
the inferred Better Auth instance, database tables, seed inputs and identity linking.
Import these with `import type` from the relevant file.

Custom database operations use Kysely through `src/queries/`. Table/column names,
insert values and query results are checked against `src/types/database.ts`.
Seed orchestration calls named methods such as `createApplicationUser`; it contains
no SQL strings or manual BEGIN/COMMIT statements. Kysely binds values and controls
the transaction. The identity-link query still calls the protected database function,
so typed queries do not bypass its authorization or audit transaction. PostgreSQL
roles/grants/functions remain explicit deployment SQL under `deploy/postgres/`.

```sh
npm --prefix auth run typecheck
npm --prefix auth run build
npm --prefix auth start
```

Build emits ESM JavaScript into `auth/dist`; `start`, `migrate` and `link-user` use
that output. The production Docker stage type-checks and builds the sources, then
copies only compiled application files and runtime dependencies into its image.

Local development uses `npm --prefix auth run dev` and `migrate:dev` with the
existing environment configuration. Node 26 runs the `.ts` sources directly and
watches them; it does **not** type-check on each restart. Run `typecheck` while
editing and before committing. Source imports use explicit `.ts` extensions,
which the compiler rewrites to `.js` in the build. See
[Node's TypeScript support](https://nodejs.org/api/typescript.html).

Hot Compose selects the development Docker stage. Migrations, permission setup
and test-account seeding are separate one-shot jobs; only the HTTP service watches
source files. See the upgrade steps below before restarting an existing stack.

## Deployment

Copy `.env.example` to `.env`; supply separate random passwords for Go runtime,
Go migration, auth runtime and auth migration, plus the PostgreSQL administrator
password. Runtime containers never receive migration/admin credentials. All roles
use the same dedicated database, with table permissions separating `auth_*` from
application data. Runtime roles cannot perform DDL.

Startup order is `db-roles` → Go/auth migration jobs → `db-security` → services.
Development adds `auth-seed` after permission setup. Go no longer runs migrations
when its HTTP server starts; standalone deployments must run `keepnode migrate`
with migration credentials before starting it with runtime credentials.

For an existing installation, back up PostgreSQL, stop the stack **without deleting
volumes**, fill the new environment variables and recreate it. The Compose network
now uses fixed proxy/API addresses; see [upgrade instructions](../docs/database-security.md).
The role migration changes ownership and grants in place; it retains application
UUIDs, financial records and auth tables.

The frontend listens on loopback port 8080 behind your TLS reverse proxy. Production
does not expose Go, auth or PostgreSQL host ports. Set `TRUSTED_EDGE_PROXY_CIDR` to
the exact ingress address seen by Nginx; that ingress must overwrite X-Forwarded-For.
Nginx and the development proxy overwrite X-Real-IP. Go and auth accept that header
only from their configured proxy peers, and Go forwards the resolved client IP to
auth. Never use `0.0.0.0/0` or `::/0` as trusted proxy networks. Disable callback
query-string logging at the external TLS proxy too.

Default per-minute limits are 600 requests per IP before Go session verification,
300 per authenticated application user, 900 auth session checks per IP, 5 password
or magic-link attempts and 10 social sign-in attempts per IP. Public shared reports
retain their stricter 20/IP limit. Excess requests return 429 with Retry-After;
Go preserves auth's 429 instead of converting it into an authentication outage.
Go limits use per-process memory; auth limits use PostgreSQL and are shared across
replicas. Multiple Go replicas require a shared limiter/ingress quota before treating
the configured budget as a deployment-wide limit. Budgets are configurable through
`.env.example`; IP limits also aggregate IPv6 /64 addresses.

Production email is disabled until a real delivery adapter is implemented; the
development file adapter is rejected in production.

Sessions use HttpOnly, SameSite=Lax cookies and Secure on HTTPS. Cookie caching
is disabled so revocation takes effect at the next Go request. Go forwards
renewed Set-Cookie headers, rejects missing/foreign Origin on mutations, and
returns 503 on auth outages rather than trusting stale browser identity. Auth
tables persist rate limits across processes. OAuth tokens are encrypted at rest.
Keep the encryption secret backed up and manage rotation deliberately.

Automatic linking within Better Auth requires verified local email ownership;
different-email linking is disabled. Password authentication is disabled outside
the explicit development test flow; email changes are disabled. The separate legacy Go identity mapping remains explicit.

## Two-factor follow-up

2FA is **not enabled** in this change. Better Auth's standard 2FA challenge is
applied to credential sign-ins, not automatically to Google or Magic Links.
`allowPasswordless: true` permits enrollment but does not change enforcement.
Before enabling TOTP, implement and test a server-side challenge for both
passwordless methods: no application-usable session before the second factor;
also cover recovery codes, enrollment verification, session revocation and
account recovery. A frontend redirect alone cannot enforce this boundary.

## Validation

```sh
go test ./...
npm --prefix auth test
npm --prefix auth run build
cd web && npm run lint -- --no-fix && npm run build
npm run test:auth
```

The complete role/permission suite starts and removes its own PostgreSQL container
and volume; it tests both a legacy-owner upgrade and an empty installation:

```sh
sh scripts/test-db-security.sh
```

This also exercises real auth with restricted credentials, denied cross-service
reads/DDL, identity ownership/status rejection, operator attribution, audit failure
rollback, runtime profile updates and account/outbox operations.

The other database suites are opt-in and use randomized schemas that they remove:

```sh
AUTH_TEST_DATABASE_URL=postgres://USER:PASSWORD@localhost:5432/DISPOSABLE_DB go test ./... -count=1
AUTH_TEST_DATABASE_URL=postgres://USER:PASSWORD@localhost:5432/DISPOSABLE_DB npm --prefix auth test
```

An optional running-stack check creates then removes one random test user:

```sh
AUTH_SMOKE_DATABASE_URL=postgres://postgres:postgres@localhost:5432/keepnode npm --prefix auth run test:smoke
```

Tests cover token hashing, replay/concurrent consumption, expiry, revocation,
origin/redirect rejection, rate limits, Google authorization URL/state rejection,
Go fail-closed behavior, blocked/deleted users and concurrent UUID provisioning.
Vue DOM tests also cover cached-identity rejection, registration redirects,
passwordless form submission, rate-limit errors, migration/sign-out failures,
callback routing and theme preservation. These use JSDOM and mocked responses;
they are not a real-browser end-to-end test.

Real Google callback completion requires your OAuth credentials and a Google
test account; it is not proven by the mocked authorization-start test.

The existing web dependency graph still has npm audit findings (including Vue
server-renderer and CLI transitive dependencies). This static SPA does not run
Vue SSR, but the remaining dependencies still need a separate upgrade review.
The auth service's installed dependency audit reported zero findings on
2026-10-23. Keep checking both graphs as advisories change.

References: [Node integration](https://better-auth.com/docs/integrations/express),
[Vue client](https://better-auth.com/docs/concepts/client),
[Magic Link](https://better-auth.com/docs/plugins/magic-link),
[email/password options](https://better-auth.com/docs/authentication/email-password),
[sessions](https://better-auth.com/docs/concepts/session-management),
[database migrations](https://better-auth.com/docs/concepts/database),
[security](https://better-auth.com/docs/reference/security),
[2FA limitations](https://better-auth.com/docs/plugins/2fa).
