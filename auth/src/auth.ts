import { betterAuth } from 'better-auth'
import type { Pool } from 'pg'
import { createBetterAuthOptions } from './config.ts'
import type { AuthConfig } from './types/config.ts'

export function createAuth(config: AuthConfig, pool: Pool) {
  return betterAuth({
    ...createBetterAuthOptions(config),
    database: pool,
  })
}
