# Background workers

PostgreSQL stores pending holdings-index revisions and market-data jobs. Workers
scan on startup, then wake through `LISTEN/NOTIFY` when a transaction commits.
Notifications carry only the table's schema name. Job contents and ownership
always come from database queries, not notification payloads.

- Holdings: account/holding triggers update the durable outbox and notify workers.
  Successful batches drain without waiting for another interval. Foreground
  mutations still attempt immediate indexing; reads retain the database fallback.
- Market data: queued/retried jobs wake workers. `FOR UPDATE SKIP LOCKED` protects
  claims across replicas. Future retry timestamps and the three-attempt limit
  remain in effect. Abandoned running jobs are recovered during reconciliation.
- Provider configuration: changes invalidate the in-process cache on each replica.
  A one-minute cache expiry covers missed notifications, including missing config
  rows (which use the configured environment default).

One dedicated PostgreSQL connection per API instance listens for all channels.
It reconnects with a 5–60 second backoff and wakes workers after resubscribing.
Workers also check once per minute while idle, so a lost notification or listener
outage cannot lose durable work. Index drift is reconciled after five minutes;
scheduled market retries may run up to one fallback interval after becoming due.
Shutdown cancels the listener and both workers and waits for them to finish.

Apply the Go migration before starting the updated API to install notification
triggers. Compose's migration service performs this step. The runtime role needs
no new table grants, DDL privileges, or external queue service. With a connection
pooler, the listener requires session pooling or a direct PostgreSQL connection.

Run isolated PostgreSQL/Typesense and restricted-role integration tests with:

```sh
sh scripts/test-db-security.sh --workers
```

The harness creates disposable containers, schemas, and search collections; it
does not target the application's database or search index.
