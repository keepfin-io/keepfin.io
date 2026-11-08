// Operator-only tool. Use a personal DB login with identity_operator membership.
import pg from 'pg'
import { parseIdentityLinkArguments } from './identity-link.ts'
import { createQueryDatabase } from './database/client.ts'
import { linkReviewedIdentity } from './queries/identity-link.ts'

const input = parseIdentityLinkArguments(process.argv.slice(2))
const connectionString = process.env.IDENTITY_OPERATOR_DATABASE_URL
if (!connectionString) throw new Error('IDENTITY_OPERATOR_DATABASE_URL is required; runtime credentials cannot link identities')
const pool = new pg.Pool({ connectionString, max: 1, connectionTimeoutMillis: 5000,
  statement_timeout: 15000, lock_timeout: 5000, application_name: 'keepfin-identity-operator' })
try {
  const result = await linkReviewedIdentity(createQueryDatabase(pool), input)
  console.log(`Identity linked; application UUID retained. Audit ID: ${result.audit_id}`)
} catch (error) {
  // SQL/connection error details may contain credentials or personal data.
  const code = error instanceof Error && 'code' in error ? String(error.code) : 'unknown'
  console.error(`Identity mapping failed (${code}); inspect the audit record before retrying`)
  process.exitCode = 1
} finally {
  await pool.end()
}
