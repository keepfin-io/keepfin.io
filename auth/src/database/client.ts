import { Kysely, PostgresDialect } from 'kysely'
import type { Pool } from 'pg'
import type { OperatorDatabase } from '../types/database.ts'

// Reuse the caller's pool. Its owner closes it after all auth/query work finishes;
// callers must not independently destroy this wrapper around a shared pool.
export function createQueryDatabase(pool: Pool): Kysely<OperatorDatabase> {
  return new Kysely<OperatorDatabase>({ dialect: new PostgresDialect({ pool }) })
}
