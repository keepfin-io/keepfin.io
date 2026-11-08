// One-shot development provisioning; the HTTP server has no users-table access.
import { auth, config, pool } from './runtime.ts'
import { seedInternalTesting } from './internal-testing.ts'

try {
  await seedInternalTesting(auth, pool, config)
} finally {
  await pool.end()
}
