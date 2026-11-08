import type { Kysely } from 'kysely'
import type { OperatorDatabase } from '../types/database.ts'
import type { InternalTestingIdentity } from '../types/internal-testing.ts'

export class InternalTestingQueries {
  private readonly db: Kysely<OperatorDatabase>

  constructor(db: Kysely<OperatorDatabase>) { this.db = db }

  transaction<T>(work: (queries: InternalTestingQueries) => Promise<T>): Promise<T> {
    return this.db.transaction().execute(transaction => work(new InternalTestingQueries(transaction)))
  }

  async applicationSchemaExists(): Promise<boolean> {
    const result = await this.db.selectNoFrom(eb => eb.fn<string | null>('to_regclass', [eb.val('users')]).as('table'))
      .executeTakeFirstOrThrow()
    return result.table !== null
  }

  async lockSeed(): Promise<void> {
    await this.db.selectNoFrom(eb => eb.fn<unknown>('pg_advisory_xact_lock', [
      eb.fn<number>('hashtext', [eb.val('keepfin-internal-testing-seed')]),
    ]).as('lock')).execute()
  }

  findAuthIdentities(user: InternalTestingIdentity) {
    return this.db.selectFrom('auth_user').select(['id', 'email', 'emailVerified'])
      .where(eb => eb.or([eb('id', '=', user.authID), eb(eb.fn<string>('lower', ['email']), '=', user.email)]))
      .forUpdate().execute()
  }

  findApplicationUsers(user: InternalTestingIdentity) {
    return this.db.selectFrom('users')
      .select(['id', 'email', 'auth_user_id', 'user_role', 'is_dummy', 'deleted_at'])
      .select(eb => eb('user_status', '=', '1').as('is_active'))
      .where(eb => eb.or([
        eb('id', '=', user.applicationID), eb('auth_user_id', '=', user.authID),
        eb(eb.fn<string>('lower', ['email']), '=', user.email),
      ])).forUpdate().execute()
  }

  findCredentials(user: InternalTestingIdentity) {
    return this.db.selectFrom('auth_account').select(['id', 'userId', 'accountId', 'providerId', 'password'])
      .where(eb => eb.or([
        eb('id', '=', user.credentialID), eb.and([eb('userId', '=', user.authID), eb('providerId', '=', 'credential')]),
      ])).forUpdate().execute()
  }

  async createAuthIdentity(user: InternalTestingIdentity): Promise<void> {
    await this.db.insertInto('auth_user').values(eb => ({
      id: user.authID, name: user.name, email: user.email, emailVerified: true,
      createdAt: eb.fn<Date>('now'), updatedAt: eb.fn<Date>('now'),
    })).execute()
  }

  async createCredential(user: InternalTestingIdentity, passwordHash: string): Promise<void> {
    await this.db.insertInto('auth_account').values(eb => ({
      id: user.credentialID, accountId: user.authID, providerId: 'credential', userId: user.authID,
      password: passwordHash, createdAt: eb.fn<Date>('now'), updatedAt: eb.fn<Date>('now'),
    })).execute()
  }

  async createApplicationUser(user: InternalTestingIdentity): Promise<void> {
    await this.db.insertInto('users').values(eb => ({
      id: user.applicationID, hash_id: 'internal-testing', is_dummy: true, email: user.email,
      password_hash: '', auth_user_id: user.authID, user_status: 1, user_role: 'admin',
      theme: 'dark', base_currency: 'USD', locale: 'en-US',
      created_at: eb.fn<Date>('now'), updated_at: eb.fn<Date>('now'),
    })).execute()
  }
}
