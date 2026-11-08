import type { AuthInstance } from './types/auth.ts'

// Better Auth emits X-Retry-After. Also expose the standard HTTP header to clients.
export function authHTTPHandler(auth: Pick<AuthInstance, 'handler'>) {
  return async (request: Request): Promise<Response> => {
    const response = await auth.handler(request)
    if (response.status === 429) {
      const seconds = Number(response.headers.get('X-Retry-After'))
      response.headers.set('Retry-After', String(Number.isInteger(seconds) && seconds > 0 && seconds <= 3600 ? seconds : 60))
      response.headers.set('Cache-Control', 'no-store')
    }
    return response
  }
}
