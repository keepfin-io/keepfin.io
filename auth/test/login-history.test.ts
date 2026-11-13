import { test } from 'node:test'
import assert from 'node:assert/strict'
import { randomUUID } from 'node:crypto'
import { mkdtemp, readdir, readFile, rm } from 'node:fs/promises'
import { join } from 'node:path'
import { tmpdir } from 'node:os'
import pg from 'pg'
import { createAuth } from '../src/auth.ts'
import { readConfig } from '../src/config.ts'
import { object, string } from './helpers.ts'

test('login history is atomic, survives logout/expiry, and is protected from runtime roles', {
  skip: !process.env.AUTH_SECURITY_DATABASE_URL,
}, async t => {
  const databaseURL = process.env.AUTH_SECURITY_DATABASE_URL
  assert.ok(databaseURL)
  assert.equal(new URL(databaseURL).pathname, '/keepfin_security_test', 'use the disposable security database')
  const admin = new pg.Pool({ connectionString: databaseURL })
  const runtimeURL = new URL(databaseURL)
  runtimeURL.username = 'auth_runtime'
  runtimeURL.password = 'security-test-auth'
  const runtime = new pg.Pool({ connectionString: runtimeURL.toString() })
  const mailDirectory = await mkdtemp(join(tmpdir(), 'keepfin-login-history-'))
  t.after(async () => {
    await Promise.all([admin.end(), runtime.end()])
    await rm(mailDirectory, { recursive: true, force: true })
  })
  const origin = 'http://localhost:8080'
  const auth = createAuth(readConfig({
    NODE_ENV: 'test', APP_ORIGIN: origin, DATABASE_URL: runtimeURL.toString(),
    BETTER_AUTH_SECRET: randomUUID() + randomUUID(), AUTH_MAIL_MODE: 'file', AUTH_MAIL_DIRECTORY: mailDirectory,
  }), runtime)
  const email = `history-${randomUUID()}@example.test`
  const userAgent = 'Mozilla/5.0 (X11; Linux x86_64) Firefox/144.0'
  let nextIP = 60
  async function signIn(ipAddress: string) {
    const before = new Set(await readdir(mailDirectory))
    const response = await auth.handler(new Request(`${origin}/api/auth/sign-in/magic-link`, {
      method: 'POST', headers: { Origin: origin, 'Content-Type': 'application/json', 'X-Real-IP': `192.0.2.${nextIP++}` },
      body: JSON.stringify({ email, callbackURL: '/profile' }),
    }))
    assert.equal(response.status, 200)
    const file = (await readdir(mailDirectory)).find(name => !before.has(name))
    assert.ok(file)
    const url = string(object(JSON.parse(await readFile(join(mailDirectory, file), 'utf8'))).url)
    return auth.handler(new Request(url, { headers: { 'User-Agent': userAgent, 'X-Real-IP': ipAddress } }))
  }
  const history = (id: string) => admin.query('SELECT * FROM public.auth_login_history WHERE auth_user_id = $1 ORDER BY signed_in_at, session_id', [id])
  const denied = (sql: string) => assert.rejects(runtime.query(sql), (error: unknown) => error instanceof Error && 'code' in error && error.code === '42501')

  // The upgrade preserves old timestamps and repeat provisioning does not duplicate them.
  const legacy = await admin.query("SELECT * FROM public.auth_login_history WHERE session_id = '40000000-0000-4000-8000-000000000001'")
  assert.equal(legacy.rowCount, 1)
  assert.equal(legacy.rows[0].signed_in_at.toISOString(), '2026-01-01T00:00:00.000Z')
  assert.equal(legacy.rows[0].user_agent, 'Legacy test browser')
  assert.equal(legacy.rows[0].ip_address, '192.0.2.40')
  const moved = await admin.query("SELECT * FROM public.auth_login_history WHERE session_id = '40000000-0000-4000-8000-000000000002'")
  assert.equal(moved.rowCount, 1, 'moving the table must retain logged-out sessions')
  assert.equal(moved.rows[0].signed_in_at.toISOString(), '2026-01-02T00:00:00.000Z')
  assert.equal(moved.rows[0].recorded_at.toISOString(), '2026-01-03T00:00:00.000Z')
  assert.equal(moved.rows[0].user_agent, 'Logged-out test browser')
  assert.equal(moved.rows[0].ip_address, '192.0.2.41')
  assert.equal((await admin.query("SELECT to_regclass('keepfin_security.auth_login_history') AS old_table")).rows[0].old_table, null)

  const signedIn = await signIn('203.0.113.60')
  assert.equal(signedIn.status, 302)
  const cookie = signedIn.headers.getSetCookie().map(value => value.split(';')[0]).join('; ')
  const active = await runtime.query('SELECT s.* FROM auth_session s JOIN auth_user u ON u.id = s."userId" WHERE u.email = $1', [email])
  assert.equal(active.rowCount, 1)
  const session = object(active.rows[0])
  const userID = string(session.userId)
  const rows = await history(userID)
  assert.equal(rows.rowCount, 1)
  const saved = object(rows.rows[0])
  assert.equal(saved.session_id, session.id)
  assert.equal(saved.user_agent, userAgent)
  assert.equal(saved.ip_address, '203.0.113.60')
  assert.deepEqual(saved.signed_in_at, session.createdAt)
  assert.ok(saved.recorded_at instanceof Date)
  assert.deepEqual(Object.keys(saved).sort(), ['auth_user_id', 'ip_address', 'recorded_at', 'session_id', 'signed_in_at', 'user_agent'])
  assert.equal(JSON.stringify(saved).includes(string(session.token)), false)

  await runtime.query('UPDATE auth_session SET "updatedAt" = NOW(), "expiresAt" = NOW() + interval \'7 days\', "userAgent" = $1 WHERE id = $2', ['Go-http-client/1.1', session.id])
  assert.deepEqual((await history(userID)).rows, rows.rows, 'session refresh must not alter the original login snapshot')
  const logout = await auth.handler(new Request(`${origin}/api/auth/sign-out`, {
    method: 'POST', headers: { Origin: origin, Cookie: cookie, 'Content-Type': 'application/json' }, body: '{}',
  }))
  assert.equal(logout.status, 200)
  assert.equal((await runtime.query('SELECT id FROM auth_session WHERE id = $1', [session.id])).rowCount, 0)
  assert.deepEqual((await history(userID)).rows, rows.rows, 'logout must retain login history')

  const secondLogin = await signIn('203.0.113.61')
  assert.equal(secondLogin.status, 302)
  assert.equal((await history(userID)).rowCount, 2, 'a new login creates a new history row')
  const secondCookie = secondLogin.headers.getSetCookie().map(value => value.split(';')[0]).join('; ')
  await runtime.query('UPDATE auth_session SET "expiresAt" = NOW() - interval \'1 hour\' WHERE "userId" = $1', [userID])
  const expired = await auth.handler(new Request(`${origin}/api/auth/get-session`, { headers: { Cookie: secondCookie } }))
  assert.equal(await expired.json(), null)
  assert.equal((await runtime.query('SELECT id FROM auth_session WHERE "userId" = $1', [userID])).rowCount, 0)
  assert.equal((await history(userID)).rowCount, 2)

  // Make history insertion fail for this request only; a session must not survive it.
  await admin.query("ALTER TABLE public.auth_login_history ADD CONSTRAINT history_test_reject CHECK (ip_address IS DISTINCT FROM '203.0.113.254')")
  try {
    const failed = await signIn('203.0.113.254')
    assert.equal(failed.headers.getSetCookie().some(value => /^keepfin\.session_token=[^;]/.test(value)), false)
    assert.equal((await runtime.query('SELECT id FROM auth_session WHERE "userId" = $1', [userID])).rowCount, 0)
    assert.equal((await history(userID)).rowCount, 2)
  } finally {
    await admin.query('ALTER TABLE public.auth_login_history DROP CONSTRAINT history_test_reject')
  }

  for (const sql of [
    'SELECT * FROM public.auth_login_history',
    "INSERT INTO public.auth_login_history(session_id,auth_user_id,signed_in_at) VALUES ('forged','forged',NOW())",
    'UPDATE public.auth_login_history SET user_agent=NULL',
    'DELETE FROM public.auth_login_history',
    'TRUNCATE public.auth_login_history',
    'ALTER TABLE auth_session DISABLE TRIGGER auth_session_login_history',
    "SET session_replication_role = 'replica'",
  ]) await denied(sql)
  await runtime.query('DELETE FROM auth_user WHERE id = $1', [userID])
  assert.equal((await history(userID)).rowCount, 2, 'identity cleanup must not cascade into audit history')
})
