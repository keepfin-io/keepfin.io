# API responses

Go JSON endpoints under `/api/v1` and `/api/public` use one envelope:

```json
{
  "data": { "currency": "USD", "summary": { "value": 1990 } },
  "meta": {
    "request_id": "a7864281-2f27-4d10-9350-b02339e661a7",
    "timestamp": "2026-11-07T10:00:00Z"
  },
  "error": null
}
```

`data` contains the resource directly. A portfolio response no longer has a
`portfolio` wrapper; accounts, activities, tags, and market search return arrays.
Holdings with valuation summaries return `{ "holdings": [], "summary": {} }`.
Successful operations without a resource may return `data: null` or an operation
result. `error` is always `null` on success.

Errors retain their HTTP status and return `data: null` with
`error: { "code": "not_found", "message": "Account not found" }`. Provider errors
keep their specific `market_*` codes. Unexpected errors are logged on the server
and exposed as `internal_error` without internal details.

Every response has a server-generated `meta.request_id`, also sent as
`X-Request-ID`. Supplied request IDs are not reused. `meta.timestamp` is the UTC
response creation time, **not** the age of market data; use domain fields such as
`price_date` for data freshness.

Paginated holdings searches and consolidated activities include
`meta.pagination: { "page": 1, "page_size": 10, "total": 21, "total_pages": 3 }`.
Pagination is one-based; an empty collection has one empty page. Existing query
parameters (`page`, `per_page`) are unchanged. Holdings search `data` contains
`holdings`, `facets`, and `source`; consolidated activity `data` is an array.

Controllers use `apiresponse.RespondOK(c, data)` or `apiresponse.RespondCreated(c, data)`.
Add `apiresponse.WithPagination(page, pageSize, total)` when needed. Return
`fiber.NewError(...)` for standard HTTP errors or
`apiresponse.NewError(status, code, message)` for an explicit public error code.
`main.go` installs the shared error handler and request identity middleware.

Frontend API services return the envelope. Consume `response.data` and
`response.meta`; the common client throws an `APIError` containing `status`,
`code`, `message`, `requestId`, and `meta`. Public portfolio requests use the same
response reader. Deploy the frontend and backend together because the old
resource wrappers and boolean `error` field are replaced.

Better Auth's `/api/auth/*` endpoints retain their SDK-defined contract.
CSV/JSON downloads remain attachments in their original import/export format;
only download failures use the API error envelope.
