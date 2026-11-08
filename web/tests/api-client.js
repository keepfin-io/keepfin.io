const { test } = require('node:test')
const assert = require('node:assert/strict')
const { readFile } = require('node:fs/promises')
const path = require('node:path')

const client = readFile(path.join(__dirname, '../src/services/api.js'), 'utf8')
  .then(source => import(`data:text/javascript;base64,${Buffer.from(source).toString('base64')}`))
const meta = { request_id: 'test-request', timestamp: '2026-11-07T00:00:00Z' }
const reply = (data, { status = 200, error = null, metadata = meta } = {}) => new Response(
  JSON.stringify({ data, meta: metadata, error }),
  { status, headers: { 'Content-Type': 'application/json', 'X-Request-ID': metadata.request_id } },
)

test('objects, arrays, empty collections and null retain the envelope', async () => {
  const { readAPIResponse } = await client
  for (const data of [{ summary: { value: 1990 } }, [{ id: 'asset' }], [], null, true]) {
    assert.deepEqual(await readAPIResponse(reply(data)), { data, meta, error: null })
  }
  const metadata = { ...meta, pagination: { page: 2, page_size: 10, total: 12, total_pages: 2 } }
  assert.deepEqual((await readAPIResponse(reply([], { metadata }))).meta.pagination, metadata.pagination)
})

test('domain errors retain HTTP status, code, message and request identity', async () => {
  const { readAPIResponse, APIError } = await client
  await assert.rejects(readAPIResponse(reply(null, {
    status: 503,
    error: { code: 'market_not_configured', message: 'Market provider unavailable' },
  })), error => error instanceof APIError && error.status === 503 &&
    error.code === 'market_not_configured' && error.message === 'Market provider unavailable' &&
    error.requestId === 'test-request' && error.meta.timestamp === meta.timestamp)
})

test('invalid successes fail visibly; proxy HTML does not become an error message', async () => {
  const { readAPIResponse } = await client
  for (const body of ['{}', 'null', '{"error":false,"portfolio":{}}', '<html>upstream detail</html>']) {
    await assert.rejects(readAPIResponse(new Response(body)), { code: 'invalid_response' })
  }
  await assert.rejects(readAPIResponse(new Response('<html>private proxy detail</html>', {
    status: 502, headers: { 'X-Request-ID': 'proxy-request' },
  })), { code: 'request_failed', status: 502, message: 'Request failed', requestId: 'proxy-request' })
})

test('fetch preserves options and redirects only on an application 401', async t => {
  const { default: apiFetch } = await client
  const previousWindow = global.window
  const redirects = []
  global.window = { location: { pathname: '/', assign: value => redirects.push(value) } }
  t.after(() => { global.window = previousWindow })
  const requests = []
  let response = () => reply({ id: 'account' }, { status: 201 })
  t.mock.method(global, 'fetch', async (url, options) => {
    requests.push({ url, options })
    return response()
  })
  const controller = new AbortController()
  const result = await apiFetch('/accounts/create', {
    method: 'POST', body: '{}', signal: controller.signal, headers: { 'X-Test': 'kept' },
  })
  assert.equal(result.data.id, 'account')
  assert.equal(requests[0].url, '/api/v1/accounts/create')
  assert.equal(requests[0].options.headers.get('Content-Type'), 'application/json')
  assert.equal(requests[0].options.headers.get('X-Test'), 'kept')
  assert.equal(requests[0].options.credentials, 'include')
  assert.equal(requests[0].options.signal, controller.signal)
  response = () => reply(null, { status: 503, error: { code: 'market_provider_access_denied', message: 'Unavailable' } })
  await assert.rejects(apiFetch('/market/search'), { status: 503 })
  assert.deepEqual(redirects, [])
  response = () => reply(null, { status: 401, error: { code: 'unauthorized', message: 'Sign in' } })
  await assert.rejects(apiFetch('/portfolio'), { status: 401 })
  assert.deepEqual(redirects, ['/login'])
  global.window.location.pathname = '/login'
  await assert.rejects(apiFetch('/user/me'), { status: 401 })
  assert.deepEqual(redirects, ['/login'])
})
