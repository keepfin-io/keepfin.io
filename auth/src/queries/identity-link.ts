import type { Kysely } from 'kysely'
import type { OperatorDatabase } from '../types/database.ts'
import type { IdentityLinkInput, IdentityLinkResult } from '../types/link-user.ts'

// Authorization, row locks, validation and audit stay inside the protected DB
// function. The operator never receives direct UPDATE access to users.
export function linkReviewedIdentity(db: Kysely<OperatorDatabase>, input: IdentityLinkInput): Promise<IdentityLinkResult> {
  return db.selectNoFrom(eb => eb.fn<string>('keepfin_security.link_identity', [
    eb.cast(eb.val(input.userID), 'uuid'), eb.cast(eb.val(input.authUserID), 'uuid'),
    eb.val(input.reason), eb.val(input.ownershipReference),
  ]).as('audit_id')).executeTakeFirstOrThrow()
}
