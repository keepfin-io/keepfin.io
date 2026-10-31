# Application packages

Each layer is grouped by business area. Keep a feature's implementation and its
tests in the same package.

```text
app/
  controllers/
    account/ activity/ holding/ market/ portfolio/ user/
  models/
    account/ activity/ base/ currency/ holding/ market/ platform/ user/
  queries/
    account/ activity/ holding/ user/
  services/
    account/ activity/ holding/ market/ portfolio/ user/
```

- **Controllers** parse HTTP requests and call services. Routes are wired in
  `pkg/routes/routes.go`.
- **Services** coordinate business rules and database transactions. Activity CSV
  import/export lives in `services/activity`; identity mapping lives in
  `services/user`; the EODHD client and shared price/FX book live in
  `services/market`.
- **Queries** contain database reads and writes, including the shared account row
  lock and ordered ledger reads. Call `GetAccountForUpdate` with an active
  transaction so the lock covers all related writes.
- **Models** contain persisted entities and request/response types. Common
  persistence fields live in `models/base`; currency and platform validation
  live in their respective packages. The pure ledger calculations remain in
  `pkg/portfolio`.

Use explicit import aliases when several layers share a name, for example
`accountmodel`, `accountquery`, `accountservice`, and `accountcontroller`.
Keep dependencies directed from controllers to services to queries/models;
models and queries must not import services or controllers.

Run `go test ./...` and `go vet ./...` after package changes. PostgreSQL integration
tests additionally use `AUTH_TEST_DATABASE_URL` and
`PORTFOLIO_TEST_DATABASE_URL`. Point them at a local test database; each suite
creates and removes its own temporary schema.
