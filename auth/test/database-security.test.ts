import { test } from 'node:test'
import assert from 'node:assert/strict'
import pg from 'pg'
import { randomUUID } from 'node:crypto'
import { execFile } from 'node:child_process'
import { promisify } from 'node:util'
import { fileURLToPath } from 'node:url'
import { mkdtemp, readdir, readFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { createAuth } from '../src/auth.ts'
import { readConfig } from '../src/config.ts'
import { object, string, readSession } from './helpers.ts'

test('real PostgreSQL roles enforce isolation and atomic, attributable identity linking', {
  skip: !process.env.AUTH_SECURITY_DATABASE_URL,
}, async t => {
  const adminURL = process.env.AUTH_SECURITY_DATABASE_URL
  assert.ok(adminURL)
  assert.equal(new URL(adminURL).pathname, '/keepfin_security_test', 'use the disposable security test database')
  const admin = new pg.Pool({ connectionString: adminURL })
  const urlFor = (user: string, password: string) => {
    const url = new URL(adminURL); url.username = user; url.password = password; return url.toString()
  }
  const app = new pg.Pool({ connectionString: urlFor('keepfin_runtime', 'security-test-runtime') })
  const authPool = new pg.Pool({ connectionString: urlFor('auth_runtime', 'security-test-auth') })
  await admin.query("CREATE ROLE security_test_operator LOGIN PASSWORD 'security-test-operator' NOSUPERUSER NOCREATEDB NOCREATEROLE")
  await admin.query('GRANT identity_operator TO security_test_operator')
  const operatorURL = urlFor('security_test_operator', 'security-test-operator')
  const operator = new pg.Pool({ connectionString: operatorURL })
  const mailDirectory = await mkdtemp(join(tmpdir(), 'keepfin-security-mail-'))
  t.after(async () => { await Promise.all([admin.end(), app.end(), authPool.end(), operator.end()]); await rm(mailDirectory, { recursive: true, force: true }) })
  const denied = async (pool: pg.Pool, sql: string) => assert.rejects(pool.query(sql), (error: unknown) => error instanceof Error && 'code' in error && error.code === '42501')
  await app.query('SELECT id FROM users LIMIT 1')
  await authPool.query('SELECT id FROM auth_user LIMIT 1')
  for (const sql of ['SELECT * FROM users', 'SELECT * FROM accounts', 'CREATE TABLE auth_forbidden(id int)', 'ALTER TABLE auth_user ADD COLUMN forbidden text']) await denied(authPool, sql)
  for (const sql of ['SELECT * FROM auth_session', 'SELECT * FROM auth_account', 'CREATE TABLE forbidden(id int)', "UPDATE users SET user_role='admin'", 'UPDATE users SET auth_user_id=NULL', 'DELETE FROM users']) await denied(app, sql)
  await denied(operator, 'UPDATE public.users SET auth_user_id=NULL')
  await denied(operator, 'DELETE FROM keepfin_security.identity_link_audit')
  await denied(app, `SELECT keepfin_security.link_identity('${randomUUID()}','${randomUUID()}','test reason','SUP-1')`)
  const flags = await admin.query<{ rolname: string; rolsuper: boolean; rolcreatedb: boolean; rolcreaterole: boolean }>("SELECT rolname,rolsuper,rolcreatedb,rolcreaterole FROM pg_roles WHERE rolname IN ('keepfin_runtime','auth_runtime','keepfin_migrator','auth_migrator')")
  assert.equal(flags.rows.length, 4)
  for (const role of flags.rows) assert.equal(role.rolsuper || role.rolcreatedb || role.rolcreaterole, false)

  // Authentication still creates users, sessions and rate-limit rows with the restricted role.
  const origin = 'http://localhost:8080'
  const auth = createAuth(readConfig({ NODE_ENV: 'test', APP_ORIGIN: origin,
    DATABASE_URL: urlFor('auth_runtime','security-test-auth'), BETTER_AUTH_SECRET: randomUUID()+randomUUID(),
    AUTH_MAIL_MODE: 'file', AUTH_MAIL_DIRECTORY: mailDirectory }), authPool)
  const request = (path: string, method = 'GET', body?: object, cookie?: string) => auth.handler(new Request(new URL(path, origin), {
    method, headers: { Origin: origin, 'X-Real-IP': '198.51.100.50', ...(body ? { 'Content-Type':'application/json' } : {}), ...(cookie ? { Cookie:cookie } : {}) },
    ...(body ? { body:JSON.stringify(body) } : {}),
  }))
  const email = 'legacy-security@example.test'
  assert.equal((await request('/api/auth/sign-in/magic-link','POST',{email})).status, 200)
  const file = (await readdir(mailDirectory))[0]; assert.ok(file)
  const link = string(object(JSON.parse(await readFile(join(mailDirectory,file),'utf8'))).url)
  const verified = await request(link)
  assert.equal(verified.status, 302)
  const cookie = verified.headers.getSetCookie().map(value=>value.split(';')[0]).join('; ')
  const session = await readSession(await request('/api/auth/get-session','GET',undefined,cookie)); assert.ok(session)
  const applicationID = '10000000-0000-4000-8000-000000000001'
  const args = [applicationID, session.user.id, '--reason','Verified legacy account migration','--ownership-reference','SUPPORT-42','--confirm']
  const run = (connectionString: string) => promisify(execFile)(process.execPath, [fileURLToPath(new URL('../dist/link-user.js', import.meta.url)), ...args], {
    env: { ...process.env, IDENTITY_OPERATOR_DATABASE_URL:connectionString },
  })
  await assert.rejects(run(urlFor('auth_runtime','security-test-auth')))
  const call = (id = applicationID, identity = session.user.id, reason = 'Verified legacy account migration', reference = 'SUPPORT-42') => operator.query<{ audit_id:string }>('SELECT keepfin_security.link_identity($1::uuid,$2::uuid,$3,$4) AS audit_id',[id,identity,reason,reference])
  await assert.rejects(call(applicationID,session.user.id,'',''))
  await admin.query('UPDATE auth_user SET "emailVerified"=false WHERE id=$1',[session.user.id])
  await assert.rejects(call())
  await admin.query('UPDATE auth_user SET "emailVerified"=true WHERE id=$1',[session.user.id])
  for (const update of ["user_status=0", "deleted_at=now()", "email='wrong@example.test'", `auth_user_id='${randomUUID()}'`]) {
    await admin.query(`UPDATE users SET ${update} WHERE id=$1`,[applicationID])
    await assert.rejects(call())
    await admin.query('UPDATE users SET user_status=1,deleted_at=NULL,email=$1,auth_user_id=NULL WHERE id=$2',[email,applicationID])
  }
  assert.equal((await admin.query('SELECT * FROM keepfin_security.identity_link_audit')).rowCount,0)

  // Audit insertion failure must roll the mapping back too.
  await admin.query("ALTER TABLE keepfin_security.identity_link_audit ADD CONSTRAINT test_reject_audit CHECK (reason <> 'Verified legacy account migration')")
  await assert.rejects(call())
  assert.equal((await admin.query('SELECT auth_user_id FROM users WHERE id=$1',[applicationID])).rows[0].auth_user_id,null)
  await admin.query('ALTER TABLE keepfin_security.identity_link_audit DROP CONSTRAINT test_reject_audit')
  const output = await run(operatorURL)
  assert.match(output.stdout,/Audit ID:/)
  const audit = (await admin.query('SELECT * FROM keepfin_security.identity_link_audit')).rows[0]
  assert.equal(audit.operator_role,'security_test_operator')
  assert.equal(audit.application_user_id,applicationID)
  assert.equal(audit.auth_user_id,session.user.id)
  assert.equal(audit.previous_auth_user_id,null)
  assert.equal(audit.ownership_reference,'SUPPORT-42')
  assert.ok(audit.occurred_at)
  assert.equal((await app.query('SELECT auth_user_id FROM users WHERE id=$1',[applicationID])).rows[0].auth_user_id,session.user.id)
  const retainedAccount = (await app.query('SELECT id,user_id,initial_balance FROM accounts WHERE user_id=$1',[applicationID])).rows[0]
  assert.equal(retainedAccount.id,'20000000-0000-4000-8000-000000000001')
  assert.equal(retainedAccount.user_id,applicationID)
  assert.equal(Number(retainedAccount.initial_balance),775.50)
  for (const pool of [app,authPool,operator]) await denied(pool,'TRUNCATE keepfin_security.identity_link_audit')
  await admin.query('REVOKE identity_operator FROM security_test_operator')
  await assert.rejects(call())
})
