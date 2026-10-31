# Keepfin authentication

Vue uses the official `better-auth/vue` client. A private Node 24 service owns
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
Go application migrations finish, auth startup atomically seeds one account:

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

## Existing users

The previous Go password endpoints, legacy demo seed and hardcoded-secret JWT
implementation are removed. The development test identity above is separate. Old `auth_token` cookies no longer authenticate.
The old password column remains for migration compatibility but is unused and
never serialized. Existing user UUIDs, account ownership and preferences remain.

A new verified auth identity matching an existing application email receives
HTTP 409 until an operator explicitly maps it. Email equality alone does not
authorize inheriting historical financial records. Verify ownership through a
trusted migration/support process before running the following command with
the two reviewed IDs:

```sh
docker compose -p keepnode -f docker-compose.hot.yml exec auth \
  node src/link-user.js APPLICATION_UUID AUTH_UUID --confirm
```

The tool requires a verified auth email, matching application email, active
non-deleted user and no conflicting mapping. It retains the application UUID.
It refuses replacement of an existing mapping. Review IDs with read-only SQL
against `users (id, email, auth_user_id)` and `auth_user (id, email, emailVerified)`.
Never bulk merge by email without checking historical account ownership.

## Deployment

Copy `.env.example` to `.env`, supply a random auth secret, database credentials,
Google credentials and the exact public HTTPS `APP_ORIGIN` without a trailing
slash. `AUTH_DATABASE_URL` points to the same database as Go, using separate
`auth_*` tables. URL-encode special characters in database credentials.

`docker compose up --build -d` runs auth migration as a separate release step
before auth startup. The frontend listens on loopback port 8080 behind your TLS
reverse proxy. Production does not expose Go, auth or PostgreSQL host ports.
Production email is disabled until a real delivery adapter is implemented;
the development file adapter is rejected in production.

Both proxies overwrite the auth IP header. When adding an external TLS proxy,
configure Nginx `set_real_ip_from` for only that proxy's exact address/CIDR and
`real_ip_header` for the header it overwrites. Otherwise all users share the
proxy's rate-limit bucket. Never trust client-supplied forwarding headers or
expose the private auth service directly. Disable query-string logging for
auth/callback URLs at every external proxy as well.

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
cd web && npm run lint -- --no-fix && npm run build
npm run test:auth
```

The database suites are opt-in and use randomized schemas that they remove:

```sh
AUTH_TEST_DATABASE_URL=postgres://USER:PASSWORD@localhost:5432/DISPOSABLE_DB go test ./... -count=1
AUTH_TEST_DATABASE_URL=postgres://USER:PASSWORD@localhost:5432/DISPOSABLE_DB npm --prefix auth test
```

An optional running-stack check creates then removes one random test user:

```sh
AUTH_SMOKE_DATABASE_URL=postgres://postgres:postgres@localhost:5432/keepnode node auth/test/live-smoke.js
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
