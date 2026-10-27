// Operator-only migration tool; not exposed by the auth HTTP server.
import pg from 'pg'
const [userID, authUserID, confirmation] = process.argv.slice(2)
if (!userID || !authUserID || confirmation !== '--confirm') {
  throw new Error('Usage: node src/link-user.js APPLICATION_UUID AUTH_UUID --confirm (verify ownership first)')
}
const pool = new pg.Pool({ connectionString: process.env.DATABASE_URL })
const client = await pool.connect()
try {
  await client.query('BEGIN')
  const { rows: [identity] } = await client.query('SELECT id, email, "emailVerified" FROM auth_user WHERE id = $1 FOR UPDATE', [authUserID])
  const { rows: [user] } = await client.query('SELECT id, email, auth_user_id, deleted_at, user_status FROM users WHERE id = $1 FOR UPDATE', [userID])
  if (!identity?.emailVerified || !user || user.deleted_at || user.user_status !== 1 ||
      user.email.toLowerCase() !== identity.email.toLowerCase() || (user.auth_user_id && user.auth_user_id !== identity.id)) {
    throw new Error('Identity mapping rejected: check verification, email, account status and existing mapping')
  }
  await client.query('UPDATE users SET auth_user_id = $1, updated_at = NOW() WHERE id = $2', [identity.id, user.id])
  await client.query('COMMIT')
  console.log('Identity linked; application UUID and financial records retained')
} catch (error) {
  await client.query('ROLLBACK')
  throw error
} finally {
  client.release()
  await pool.end()
}
