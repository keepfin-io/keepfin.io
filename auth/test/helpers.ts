import assert from 'node:assert/strict'
import type { AuthInstance } from '../src/types/auth.ts'

export interface AuthRequestOptions {
  method?: string
  body?: Record<string, unknown>
  cookie?: string
  requestOrigin?: string
  ip?: string
}

export function object(value: unknown): Record<string, unknown> {
  assert.ok(isObject(value), 'expected a JSON object')
  return value
}

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

export function string(value: unknown): string {
  assert.ok(typeof value === 'string', 'expected a JSON string')
  return value
}

type Session = AuthInstance['$Infer']['Session']
interface SessionResponse {
  user: Pick<Session['user'], 'id' | 'name' | 'email' | 'emailVerified'>
  session: Pick<Session['session'], 'userId'>
}

export async function readSession(response: Response): Promise<SessionResponse | null> {
  const data: unknown = await response.json()
  if (data === null) return null
  const result = object(data)
  const user = object(result.user)
  const session = object(result.session)
  assert.ok(typeof user.emailVerified === 'boolean')
  return {
    user: {
      id: string(user.id), name: string(user.name), email: string(user.email),
      emailVerified: user.emailVerified,
    },
    session: { userId: string(session.userId) },
  }
}
