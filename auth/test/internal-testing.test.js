import { test } from 'node:test'
import assert from 'node:assert/strict'
import { randomUUID } from 'node:crypto'
import pg from 'pg'
import { getMigrations } from 'better-auth/db/migration'
import { createAuth } from '../src/auth.js'
import { internalTestingUser as user, seedInternalTesting } from '../src/internal-testing.js'

const origin = 'http://localhost:8080'
async function fixture(t) {
  const schema = `seed_test_${randomUUID().replaceAll('-', '')}`
  const admin = new pg.Pool({ connectionString: process.env.AUTH_TEST_DATABASE_URL })
  await admin.query(`CREATE SCHEMA "${schema}"`)
  const pool = new pg.Pool({ connectionString: process.env.AUTH_TEST_DATABASE_URL, options: `-c search_path=${schema}` })
  t.after(async () => { await pool.end(); await admin.query(`DROP SCHEMA "${schema}" CASCADE`); await admin.end() })
  const config = { baseURL: origin, secret: randomUUID() + randomUUID(), mailMode: 'disabled', google: false, internalTesting: true }
  const auth = createAuth(config, pool)
  await (await getMigrations(auth.options)).runMigrations()
  await pool.query(`CREATE TABLE users (
    id uuid PRIMARY KEY, hash_id text UNIQUE, is_dummy boolean, email text UNIQUE,
    password_hash text, auth_user_id text UNIQUE, user_status int, user_role text,
    theme text, base_currency text, locale text, created_at timestamptz, updated_at timestamptz, deleted_at timestamptz
  )`)
  return { auth, pool, config }
}
function request(auth, path, { body, cookie, requestOrigin = origin, ip = '192.0.2.1' } = {}) {
  return auth.handler(new Request(new URL(`/api/auth${path}`, origin), {
    method: body ? 'POST' : 'GET',
    headers: { Origin: requestOrigin, 'X-Real-IP': ip, ...(body ? { 'Content-Type': 'application/json' } : {}), ...(cookie ? { Cookie: cookie } : {}) },
    body: body ? JSON.stringify(body) : undefined,
  }))
}
const skip = !process.env.AUTH_TEST_DATABASE_URL

test('internal admin seed is atomic, idempotent and authenticates through Better Auth', { skip }, async t => {
  const { auth, pool, config } = await fixture(t)
  await Promise.all([seedInternalTesting(auth, pool, config), seedInternalTesting(auth, pool, config)])
  const { rows: [appUser] } = await pool.query('SELECT * FROM users')
  assert.equal(appUser.user_role, 'admin')
  assert.equal(appUser.auth_user_id, user.authID)
  assert.equal(appUser.id, user.applicationID)
  assert.equal(appUser.password_hash, '')
  assert.equal(appUser.is_dummy, true)
  const { rows: [credential] } = await pool.query('SELECT password FROM auth_account')
  assert.notEqual(credential.password, '12345')
  await seedInternalTesting(auth, pool, config)
  assert.equal((await pool.query('SELECT password FROM auth_account')).rows[0].password, credential.password)
  assert.equal((await pool.query('SELECT * FROM auth_user')).rows.length, 1)
  assert.equal((await pool.query('SELECT * FROM auth_account')).rows.length, 1)

  const signedIn = await request(auth, '/sign-in/email', { body: { email: user.email, password: '12345' } })
  assert.equal(signedIn.status, 200, await signedIn.clone().text())
  const cookie = signedIn.headers.getSetCookie().map(value => value.split(';')[0]).join('; ')
  assert.match(signedIn.headers.getSetCookie().join(';'), /HttpOnly/i)
  const session = await (await request(auth, '/get-session', { cookie })).json()
  assert.equal(session.user.name, 'internal-testing')
  assert.equal(session.user.emailVerified, true)
  assert.equal(session.user.id, appUser.auth_user_id)
  assert.equal((await request(auth, '/sign-in/email', { body: { email: user.email, password: 'incorrect' } })).status, 401)
  assert.equal((await request(auth, '/sign-in/email', { body: { email: 'someone@example.com', password: '12345' } })).status, 401)
  assert.equal((await request(auth, '/sign-in/email', { requestOrigin: 'https://evil.example', body: { email: user.email, password: '12345' } })).status, 403)
  assert.notEqual((await request(auth, '/sign-up/email', { body: { name: 'Other', email: 'other@example.com', password: '12345' } })).status, 200)
  for (let i = 0; i < 5; i++) assert.equal((await request(auth, '/sign-in/email', { ip: '198.51.100.3', body: { email: user.email, password: 'wrong' } })).status, 401)
  assert.equal((await request(auth, '/sign-in/email', { ip: '198.51.100.3', body: { email: user.email, password: '12345' } })).status, 429)
  const disabled = createAuth({ ...config, internalTesting: false }, pool)
  assert.notEqual((await request(disabled, '/sign-in/email', { ip: '198.51.100.4', body: { email: user.email, password: '12345' } })).status, 200)
  assert.equal(await seedInternalTesting(auth, pool, { ...config, internalTesting: false }), false)
  await assert.rejects(seedInternalTesting(auth, pool, { ...config, production: true }), /development\/test-only/)
})

test('seed refuses email collisions and preserves existing identities and roles', { skip }, async t => {
  const { auth, pool, config } = await fixture(t)
  const existing = randomUUID()
  await pool.query("INSERT INTO users (id,email,user_role) VALUES ($1,$2,'user')", [existing, user.email])
  await assert.rejects(seedInternalTesting(auth, pool, config), /conflicts/)
  assert.equal((await pool.query('SELECT * FROM auth_user')).rows.length, 0)
  assert.equal((await pool.query('SELECT * FROM auth_account')).rows.length, 0)
  const { rows: [retained] } = await pool.query('SELECT * FROM users')
  assert.equal(retained.id, existing)
  assert.equal(retained.user_role, 'user')
  assert.equal(retained.auth_user_id, null)
})
