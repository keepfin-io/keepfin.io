import { getMigrations } from 'better-auth/db/migration'
import { auth, pool } from './runtime.ts'

try {
  const { runMigrations } = await getMigrations(auth.options)
  await runMigrations()
  console.log('Better Auth migrations completed')
} finally {
  await pool.end()
}
