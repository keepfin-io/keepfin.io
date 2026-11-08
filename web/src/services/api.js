const API_URL = '/api/v1'

export class APIError extends Error {
  constructor(message, { status, code, meta } = {}) {
    super(message)
    this.name = 'APIError'
    this.status = status
    this.code = code
    this.meta = meta ?? null
    this.requestId = meta?.request_id ?? null
  }
}

/** Read the shared Go API envelope, including errors from public endpoints. */
export async function readAPIResponse(response) {
  let envelope
  try {
    envelope = await response.json()
  } catch {
    throw new APIError(response.ok ? 'Invalid API response' : 'Request failed', {
      status: response.status,
      code: response.ok ? 'invalid_response' : 'request_failed',
      meta: { request_id: response.headers.get('X-Request-ID') },
    })
  }

  const meta = envelope?.meta ?? { request_id: response.headers.get('X-Request-ID') }
  if (!response.ok || envelope?.error) {
    throw new APIError(envelope?.error?.message ?? 'Request failed', {
      status: response.status,
      code: envelope?.error?.code ?? 'request_failed',
      meta,
    })
  }
  if (!envelope || !Object.prototype.hasOwnProperty.call(envelope, 'data') ||
      envelope.error !== null || !envelope.meta?.request_id || !envelope.meta?.timestamp) {
    throw new APIError('Invalid API response', { status: response.status, code: 'invalid_response', meta })
  }
  return envelope
}

async function apiFetch(endpoint, options = {}) {
  const headers = new Headers(options.headers)
  if (!headers.has('Content-Type')) headers.set('Content-Type', 'application/json')
  const response = await fetch(`${API_URL}${endpoint}`, {
    credentials: 'include',
    ...options,
    headers,
  })

  try {
    return await readAPIResponse(response)
  } catch (error) {
    if (response.status === 401 && !['/login', '/register', '/auth/callback'].includes(window.location.pathname)) {
      window.location.assign('/login')
    }
    throw error
  }
}

export default apiFetch
