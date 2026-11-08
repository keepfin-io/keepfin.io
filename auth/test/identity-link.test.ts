import { test } from 'node:test'
import assert from 'node:assert/strict'
import { randomUUID } from 'node:crypto'
import { parseIdentityLinkArguments } from '../src/identity-link.ts'

test('identity linking requires reviewed IDs, reason, ownership evidence and confirmation', () => {
  const ids = [randomUUID(), randomUUID()]
  const args = [...ids, '--reason', 'Legacy account migration', '--ownership-reference', 'SUPPORT-123', '--confirm']
  assert.equal(parseIdentityLinkArguments(args).ownershipReference, 'SUPPORT-123')
  for (const invalid of [ids, [...ids, '--confirm'], args.slice(0,-1), [...args, '--actor', 'admin'], [...args, '--confirm'], ['invalid', ...args.slice(1)]]) {
    assert.throws(() => parseIdentityLinkArguments(invalid))
  }
})
