# Application packages

Each layer is grouped by business area. Keep a feature's implementation and its
tests in the same package.

JSON endpoints share the [API response contract](../docs/api-responses.md).
Use `pkg/apiresponse` for success responses and the central Fiber error handler.

Protected controllers receive the mapped application user ID as a `uuid.UUID`
parameter. Register them with `authhandler.WithUser(controller.Method)` under
the existing `middleware.Session` route group. `Session` verifies authentication;
`WithUser` reads its request-local identity and returns an enveloped `401` when
it is missing, has the wrong type, or is the zero UUID. Never derive this identity
from a request header, query, or body. Account/asset ownership checks still belong
in controllers and services. Public portfolio sharing uses its existing token
validation and is registered without `WithUser`.

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

## Holdings search

Run `make search-init` once, then `make hot`. Compose starts Typesense 30.2 on
the private application network. `TYPESENSE_API_KEY` is generated in the ignored
`.env`; never expose it in Vue environment variables. Outside Compose, configure
both `TYPESENSE_URL` and `TYPESENSE_API_KEY` on the Go server.

PostgreSQL remains authoritative. Holding and account triggers update
`holding_search_states` in the same transaction as each change. This covers
ledger rebuilds, imports, position edits, closing positions and account deletion.
After a successful mutation, middleware waits up to 2.5 seconds for the user's
pending revisions to reach Typesense. Each bulk-import acknowledgement is checked.
An account row in the outbox serializes concurrent index writers; document IDs
are the persisted holding UUIDs.

`GET /api/v1/accounts/:id/holdings/search` checks account ownership in PostgreSQL
and supports `q`, `page`, `per_page` (10/25/50/100), `sort`, `direction`, and exact
`currency`, `exchange`, `asset_class`, `sector`, `country` filters. It returns
`holdings`, `total`, `pages`, `facets` and `source`. The displayed list uses this
endpoint; account-wide financial summaries remain independent of pagination.
The market asset discovery endpoint remains separate.

If synchronization or search fails, reads use a fresh, ownership-scoped database
snapshot (`source: database`). Exact filters and pagination remain available;
typo tolerance is temporarily unavailable and the UI shows a notice. A committed
financial write still succeeds with `X-Holdings-Search-Sync: pending`, so clients
must not repeat it merely because indexing failed. Successful synchronization
sets this header to `ready`. A worker wakes on committed PostgreSQL notifications
and scans durable pending revisions on startup. A one-minute fallback handles
missed notifications; acknowledged accounts are reconciled after five minutes. Reads also
try to synchronize before searching. Recreating a lost collection requeues all
accounts. This provides immediate visibility through fallback, not an atomic
transaction across PostgreSQL and Typesense.

Integration tests require `PORTFOLIO_TEST_DATABASE_URL`,
`HOLDINGS_SEARCH_TEST_URL`, and `TYPESENSE_API_KEY`:

```sh
go test -v ./app/services/holding ./pkg/typesense ./pkg/middleware
```

The tests create and clean up unique PostgreSQL schemas and Typesense collections;
they cover persistence, filtering, pagination, ownership, outages, partial import
failure and retry after restart. Existing financial records are not modified.
