import { test, type TestContext } from 'node:test'
import assert from 'node:assert/strict'
import { randomUUID } from 'node:crypto'
import pg from 'pg'
import { getMigrations } from 'better-auth/db/migration'
import { createAuth } from '../src/auth.ts'
import type { AuthInstance } from '../src/types/auth.ts'
import { readConfig } from '../src/config.ts'
import { internalTestingUser as user, seedInternalTesting } from '../src/internal-testing.ts'
import { readSession, type AuthRequestOptions } from './helpers.ts'

interface ApplicationUserRow {
  id: string
  user_role: string | null
  user_status: string
  auth_user_id: string | null
  password_hash: string
  is_dummy: boolean
}

const origin = 'http://localhost:8080'
async function fixture(t: TestContext) {
  const schema = `seed_test_${randomUUID().replaceAll('-', '')}`
  const admin = new pg.Pool({ connectionString: process.env.AUTH_TEST_DATABASE_URL })
  await admin.query(`CREATE SCHEMA "${schema}"`)
  const pool = new pg.Pool({ connectionString: process.env.AUTH_TEST_DATABASE_URL, options: `-c search_path=${schema}` })
  t.after(async () => { await pool.end(); await admin.query(`DROP SCHEMA "${schema}" CASCADE`); await admin.end() })
  const config = readConfig({
    NODE_ENV: 'test', APP_ORIGIN: origin, BETTER_AUTH_SECRET: randomUUID() + randomUUID(),
    DATABASE_URL: process.env.AUTH_TEST_DATABASE_URL, AUTH_INTERNAL_TESTING: 'true',
  })
  const auth = createAuth(config, pool)
  await (await getMigrations(auth.options)).runMigrations()
  await pool.query(`CREATE TABLE users (
    id uuid PRIMARY KEY, hash_id text UNIQUE, is_dummy boolean, email text UNIQUE,
    password_hash text, auth_user_id text UNIQUE, user_status bigint, user_role text,
    theme text, base_currency text, locale text, created_at timestamptz, updated_at timestamptz, deleted_at timestamptz
  )`)
  return { auth, pool, config }
}
function request(
  auth: AuthInstance,
  path: string,
  { body, cookie, requestOrigin = origin, ip = '192.0.2.1' }: AuthRequestOptions = {},
): Promise<Response> {
  return auth.handler(new Request(new URL(`/api/auth${path}`, origin), {
    method: body ? 'POST' : 'GET',
    headers: { Origin: requestOrigin, 'X-Real-IP': ip, ...(body ? { 'Content-Type': 'application/json' } : {}), ...(cookie ? { Cookie: cookie } : {}) },
    ...(body ? { body: JSON.stringify(body) } : {}),
  }))
}
const skip = !process.env.AUTH_TEST_DATABASE_URL

test('internal admin seed is atomic, idempotent and authenticates through Better Auth', { skip }, async t => {
  const { auth, pool, config } = await fixture(t)
  await Promise.all([seedInternalTesting(auth, pool, config), seedInternalTesting(auth, pool, config)])
  const { rows: [appUser] } = await pool.query<ApplicationUserRow>('SELECT * FROM users')
  assert.ok(appUser)
  assert.equal(appUser.user_role, 'admin')
  // GORM creates bigint statuses, which pg returns as strings by default.
  assert.equal(appUser.user_status, '1')
  assert.equal(appUser.auth_user_id, user.authID)
  assert.equal(appUser.id, user.applicationID)
  assert.equal(appUser.password_hash, '')
  assert.equal(appUser.is_dummy, true)
  const { rows: [credential] } = await pool.query<{ password: string }>('SELECT password FROM auth_account')
  assert.ok(credential)
  assert.notEqual(credential.password, '12345')
  await seedInternalTesting(auth, pool, config)
  assert.equal((await pool.query<{ password: string }>('SELECT password FROM auth_account')).rows[0]?.password, credential.password)
  assert.equal((await pool.query('SELECT * FROM auth_user')).rows.length, 1)
  assert.equal((await pool.query('SELECT * FROM auth_account')).rows.length, 1)

  const signedIn = await request(auth, '/sign-in/email', { body: { email: user.email, password: '12345' } })
  assert.equal(signedIn.status, 200, await signedIn.clone().text())
  const cookie = signedIn.headers.getSetCookie().map(value => value.split(';')[0]).join('; ')
  assert.match(signedIn.headers.getSetCookie().join(';'), /HttpOnly/i)
  const session = await readSession(await request(auth, '/get-session', { cookie }))
  assert.ok(session)
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
  const { rows: [retained] } = await pool.query<ApplicationUserRow>('SELECT * FROM users')
  assert.ok(retained)
  assert.equal(retained.id, existing)
  assert.equal(retained.user_role, 'user')
  assert.equal(retained.auth_user_id, null)
})

test('seed refuses inactive existing test accounts without changing their status', { skip }, async t => {
  const { auth, pool, config } = await fixture(t)
  await seedInternalTesting(auth, pool, config)
  for (const status of [0, 2]) {
    await pool.query('UPDATE users SET user_status = $1 WHERE id = $2', [status, user.applicationID])
    await assert.rejects(seedInternalTesting(auth, pool, config), /conflicts/)
    const { rows: [retained] } = await pool.query<{ user_status: string }>('SELECT user_status FROM users WHERE id = $1', [user.applicationID])
    assert.ok(retained)
    assert.equal(retained.user_status, String(status))
  }
})

test('concurrent typed seed transactions create one identity and credential', { skip }, async t => {
  const { auth, pool, config } = await fixture(t)
  await Promise.all([seedInternalTesting(auth, pool, config), seedInternalTesting(auth, pool, config)])
  for (const table of ['auth_user', 'auth_account', 'users']) {
    assert.equal((await pool.query(`SELECT count(*)::int AS count FROM ${table}`)).rows[0].count, 1)
  }
})

test('a late credential conflict rolls back the typed seed transaction', { skip }, async t => {
  const { auth, pool, config } = await fixture(t)
  const otherID = randomUUID()
  await pool.query('INSERT INTO auth_user(id,name,email,"emailVerified","createdAt","updatedAt") VALUES($1,$2,$3,true,now(),now())', [otherID, 'Other', 'other@example.test'])
  await pool.query('INSERT INTO auth_account(id,"accountId","providerId","userId",password,"createdAt","updatedAt") VALUES($1,$2,$3,$4,$5,now(),now())', [user.credentialID, otherID, 'credential', otherID, 'existing-hash'])
  await assert.rejects(seedInternalTesting(auth, pool, config), /conflicts/)
  assert.equal((await pool.query('SELECT id FROM auth_user WHERE id=$1', [user.authID])).rowCount, 0)
  assert.equal((await pool.query('SELECT id FROM users')).rowCount, 0)
  assert.equal((await pool.query('SELECT password FROM auth_account WHERE id=$1', [user.credentialID])).rows[0].password, 'existing-hash')
})
