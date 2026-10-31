import { setTimeout as delay } from 'node:timers/promises'

export const internalTestingUser = Object.freeze({
  name: 'internal-testing',
  email: 'internal-testing@keepfin.test',
  authID: '0c0f4e9c-0b0b-44aa-a280-99857d958542',
  applicationID: 'caf65a0e-c81b-4dd9-ad2f-4bf7613cd594',
  credentialID: '913cf96b-d3f3-4a3b-8e3b-44e0a131ea45',
})

// Both identities are created together. Email matches never authorize linking an
// existing account, and restarts never overwrite a password or financial data.
export async function seedInternalTesting(auth, pool, config, { waitForSchema = false } = {}) {
  if (!config.internalTesting) return false
  if (config.production) throw new Error('Internal testing seed is development/test-only')
  if (waitForSchema) {
    for (let attempt = 0; attempt < 60; attempt++) {
      const { rows: [schema] } = await pool.query("SELECT to_regclass('users') AS users")
      if (schema.users) break
      if (attempt === 59) throw new Error('Application migrations must finish before the internal testing seed')
      await delay(500)
    }
  }
  const client = await pool.connect()
  const u = internalTestingUser
  try {
    await client.query('BEGIN')
    await client.query("SELECT pg_advisory_xact_lock(hashtext('keepfin-internal-testing-seed'))")
    const { rows: identities } = await client.query('SELECT id, email, "emailVerified" FROM auth_user WHERE id = $1 OR lower(email) = $2 FOR UPDATE', [u.authID, u.email])
    const { rows: users } = await client.query('SELECT id, email, auth_user_id, user_role, is_dummy, user_status, deleted_at FROM users WHERE id = $1 OR auth_user_id = $2 OR lower(email) = $3 FOR UPDATE', [u.applicationID, u.authID, u.email])
    if (identities.some(row => row.id !== u.authID || row.email !== u.email || !row.emailVerified) ||
        users.some(row => row.id !== u.applicationID || row.email !== u.email || row.auth_user_id !== u.authID || row.user_role !== 'admin' || !row.is_dummy || row.user_status !== 1 || row.deleted_at)) {
      throw new Error('Internal testing seed conflicts with an existing identity; no accounts were changed')
    }
    if (!identities.length) {
      await client.query('INSERT INTO auth_user (id, name, email, "emailVerified", "createdAt", "updatedAt") VALUES ($1, $2, $3, true, NOW(), NOW())', [u.authID, u.name, u.email])
    }
    const { rows: credentials } = await client.query('SELECT id, "userId", "accountId", "providerId", password FROM auth_account WHERE id = $1 OR ("userId" = $2 AND "providerId" = \'credential\') FOR UPDATE', [u.credentialID, u.authID])
    if (credentials.some(row => row.id !== u.credentialID || row.userId !== u.authID || row.accountId !== u.authID || row.providerId !== 'credential' || !row.password)) {
      throw new Error('Internal testing seed conflicts with an existing credential')
    }
    if (!credentials.length) {
      const context = await auth.$context
      const passwordHash = await context.password.hash('12345')
      await client.query('INSERT INTO auth_account (id, "accountId", "providerId", "userId", password, "createdAt", "updatedAt") VALUES ($1, $2::text, \'credential\', $2::uuid, $3, NOW(), NOW())', [u.credentialID, u.authID, passwordHash])
    }
    if (!users.length) {
      await client.query(`INSERT INTO users (id, hash_id, is_dummy, email, password_hash, auth_user_id, user_status, user_role, theme, base_currency, locale, created_at, updated_at)
        VALUES ($1, 'internal-testing', true, $2, '', $3, 1, 'admin', 'dark', 'USD', 'en-US', NOW(), NOW())`, [u.applicationID, u.email, u.authID])
    }
    await client.query('COMMIT')
    return true
  } catch (error) {
    await client.query('ROLLBACK')
    throw error
  } finally { client.release() }
}
