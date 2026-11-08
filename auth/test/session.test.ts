import { test } from 'node:test'
import assert from 'node:assert/strict'
import pg from 'pg'
import { mkdtemp, readdir, readFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { randomUUID } from 'node:crypto'
import { getMigrations } from 'better-auth/db/migration'
import { createAuth } from '../src/auth.ts'
import { authHTTPHandler } from '../src/http-handler.ts'
import { readConfig } from '../src/config.ts'
import { object, string, readSession, type AuthRequestOptions } from './helpers.ts'

test('PostgreSQL: passwordless sessions, replay, expiry, revocation, origin and rate limits', {
  skip: !process.env.AUTH_TEST_DATABASE_URL,
}, async (t) => {
  const databaseURL = process.env.AUTH_TEST_DATABASE_URL
  assert.ok(databaseURL)
  const schema = `auth_test_${randomUUID().replaceAll('-', '')}`
  const admin = new pg.Pool({ connectionString: process.env.AUTH_TEST_DATABASE_URL })
  await admin.query(`CREATE SCHEMA "${schema}"`)
  const pool = new pg.Pool({ connectionString: process.env.AUTH_TEST_DATABASE_URL, options: `-c search_path=${schema}` })
  const mailDirectory = await mkdtemp(join(tmpdir(), 'keepfin-auth-test-'))
  t.after(async () => {
    await pool.end()
    await admin.query(`DROP SCHEMA "${schema}" CASCADE`)
    await admin.end()
    await rm(mailDirectory, { recursive: true, force: true })
  })
  const origin = 'http://localhost:8080'
  const auth = createAuth(readConfig({
    NODE_ENV: 'test', APP_ORIGIN: origin, DATABASE_URL: databaseURL,
    BETTER_AUTH_SECRET: randomUUID() + randomUUID(), AUTH_MAIL_MODE: 'file', AUTH_MAIL_DIRECTORY: mailDirectory,
    GOOGLE_CLIENT_ID: 'test-client', GOOGLE_CLIENT_SECRET: 'test-secret',
  }), pool)
  await (await getMigrations(auth.options)).runMigrations()
  let nextIP = 1
  function request(path: string, { method = 'GET', body, cookie, requestOrigin = origin, ip = `192.0.2.${nextIP++}` }: AuthRequestOptions = {}): Promise<Response> {
    return auth.handler(new Request(new URL(path, origin), {
      method, headers: { Origin: requestOrigin, 'X-Real-IP': ip, ...(body ? { 'Content-Type': 'application/json' } : {}), ...(cookie ? { Cookie: cookie } : {}) },
      ...(body ? { body: JSON.stringify(body) } : {}),
    }))
  }
  function cookies(res: Response): string {
    return res.headers.getSetCookie().map(value => value.split(';', 1)[0] ?? '').filter(value => value && !value.endsWith('=')).join('; ')
  }
  async function newLink(email = `${randomUUID()}@example.com`): Promise<string> {
    const before = new Set(await readdir(mailDirectory))
    const response = await request('/api/auth/sign-in/magic-link', { method: 'POST', body: { email, callbackURL: '/auth/callback' } })
    assert.equal(response.status, 200, await response.text())
    const file = (await readdir(mailDirectory)).find(name => !before.has(name))
    assert.ok(file, 'expected a new development email')
    return string(object(JSON.parse(await readFile(join(mailDirectory, file), 'utf8'))).url)
  }

  const link = await newLink()
  const rawToken = new URL(link).searchParams.get('token')
  assert.ok(rawToken)
  const stored = await pool.query('SELECT * FROM auth_verification')
  assert.equal(JSON.stringify(stored.rows).includes(rawToken), false, 'raw token must not be stored')
  const verified = await request(link)
  assert.equal(verified.status, 302)
  assert.equal(verified.headers.get('location'), `${origin}/auth/callback`)
  const cookie = cookies(verified)
  assert.match(cookie, /keepfin.session_token=/)
  assert.match(verified.headers.getSetCookie().join(';'), /HttpOnly/i)
  const session = await readSession(await request('/api/auth/get-session', { cookie }))
  assert.ok(session)
  assert.equal(session.user.emailVerified, true)
  assert.equal(session.session.userId, session.user.id)
  assert.equal(await (await request('/api/auth/get-session', { cookie: `${cookie}tampered` })).json(), null)
  const replay = await request(link)
  assert.equal(cookies(replay).includes('keepfin.session_token='), false)
  assert.match(replay.headers.get('location') ?? '', /error=/)

  const denied = await request('/api/auth/sign-out', { method: 'POST', cookie, body: {}, requestOrigin: 'https://evil.example' })
  assert.equal(denied.status, 403)
  assert.ok((await readSession(await request('/api/auth/get-session', { cookie })))?.session)
  assert.equal((await request('/api/auth/sign-out', { method: 'POST', cookie, body: {} })).status, 200)
  assert.equal(await (await request('/api/auth/get-session', { cookie })).json(), null, 'revoked session must not survive')

  const expired = await newLink()
  await pool.query('UPDATE auth_verification SET "expiresAt" = NOW() - interval \'1 hour\'')
  assert.equal(cookies(await request(expired)).includes('keepfin.session_token='), false)
  const concurrent = await newLink()
  const outcomes = await Promise.all([request(concurrent), request(concurrent)])
  assert.equal(outcomes.filter(res => cookies(res).includes('keepfin.session_token=')).length, 1)

  const invalidRedirect = await request('/api/auth/sign-in/magic-link', { method: 'POST', body: { email: 'user@example.com', callbackURL: 'https://evil.example' } })
  assert.equal(invalidRedirect.status, 403)
  assert.notEqual((await request('/api/auth/sign-in/email', { method: 'POST', body: { email: 'user@example.com', password: 'legacy-password' } })).status, 200)
  for (let i = 0; i < 5; i++) assert.equal((await request('/api/auth/sign-in/magic-link', { method: 'POST', ip: '198.51.100.1', body: { email: 'rate@example.com' } })).status, 200)
  assert.equal((await request('/api/auth/sign-in/magic-link', { method: 'POST', ip: '198.51.100.1', body: { email: 'rate@example.com' } })).status, 429)

  const google = await request('/api/auth/sign-in/social', { method: 'POST', body: { provider: 'google', callbackURL: '/auth/callback' } })
  assert.equal(google.status, 200)
  const googleURL = new URL(string(object(await google.json()).url))
  assert.equal(googleURL.hostname, 'accounts.google.com')
  assert.equal(googleURL.searchParams.get('redirect_uri'), `${origin}/api/auth/callback/google`)
  assert.ok(googleURL.searchParams.get('state'))
  const invalidState = await request('/api/auth/callback/google?state=forged&code=forged')
  assert.equal(cookies(invalidState).includes('keepfin.session_token='), false)

  // Session verification is bounded and callers do not share a global bucket.
  const limited = createAuth({ ...readConfig({
    NODE_ENV: 'test', APP_ORIGIN: origin, DATABASE_URL: databaseURL,
    BETTER_AUTH_SECRET: 'a'.repeat(64), AUTH_SESSION_RATE_LIMIT: '2',
  }) }, pool)
  const limitedHandler = authHTTPHandler(limited)
  const getSession = (ip: string) => limitedHandler(new Request(`${origin}/api/auth/get-session`, { headers: { 'X-Real-IP': ip } }))
  assert.equal((await getSession('198.51.100.91')).status, 200)
  assert.equal((await getSession('198.51.100.91')).status, 200)
  const exhausted = await getSession('198.51.100.91')
  assert.equal(exhausted.status, 429)
  assert.ok(exhausted.headers.get('retry-after'))
  assert.equal((await getSession('198.51.100.92')).status, 200)
})
