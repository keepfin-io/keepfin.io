import { setTimeout as delay } from 'node:timers/promises'
import type { Pool } from 'pg'
import type { AuthInstance } from './types/auth.ts'
import type { AuthConfig } from './types/config.ts'
import type { InternalTestingIdentity, SeedOptions } from './types/internal-testing.ts'
import { createQueryDatabase } from './database/client.ts'
import { InternalTestingQueries } from './queries/internal-testing.ts'

export const internalTestingUser = Object.freeze({
  name: 'internal-testing',
  email: 'internal-testing@keepfin.test',
  authID: '0c0f4e9c-0b0b-44aa-a280-99857d958542',
  applicationID: 'caf65a0e-c81b-4dd9-ad2f-4bf7613cd594',
  credentialID: '913cf96b-d3f3-4a3b-8e3b-44e0a131ea45',
} satisfies InternalTestingIdentity)

// Both identities are created together. Email matches never authorize linking an
// existing account, and restarts never overwrite a password or financial data.
export async function seedInternalTesting(
  auth: Pick<AuthInstance, '$context'>,
  pool: Pool,
  config: Pick<AuthConfig, 'internalTesting' | 'production'>,
  { waitForSchema = false }: SeedOptions = {},
): Promise<boolean> {
  if (!config.internalTesting) return false
  if (config.production) throw new Error('Internal testing seed is development/test-only')
  const queries = new InternalTestingQueries(createQueryDatabase(pool))
  if (waitForSchema) {
    for (let attempt = 0; attempt < 60; attempt++) {
      if (await queries.applicationSchemaExists()) break
      if (attempt === 59) throw new Error('Application migrations must finish before the internal testing seed')
      await delay(500)
    }
  }
  const u = internalTestingUser
  return queries.transaction(async transaction => {
    await transaction.lockSeed()
    const identities = await transaction.findAuthIdentities(u)
    const users = await transaction.findApplicationUsers(u)
    if (identities.some(row => row.id !== u.authID || row.email !== u.email || !row.emailVerified) ||
        users.some(row => row.id !== u.applicationID || row.email !== u.email || row.auth_user_id !== u.authID || row.user_role !== 'admin' || !row.is_dummy || !row.is_active || row.deleted_at)) {
      throw new Error('Internal testing seed conflicts with an existing identity; no accounts were changed')
    }
    if (!identities.length) {
      await transaction.createAuthIdentity(u)
    }
    const credentials = await transaction.findCredentials(u)
    if (credentials.some(row => row.id !== u.credentialID || row.userId !== u.authID || row.accountId !== u.authID || row.providerId !== 'credential' || !row.password)) {
      throw new Error('Internal testing seed conflicts with an existing credential')
    }
    if (!credentials.length) {
      const context = await auth.$context
      const passwordHash = await context.password.hash('12345')
      await transaction.createCredential(u, passwordHash)
    }
    if (!users.length) {
      await transaction.createApplicationUser(u)
    }
    return true
  })
}
