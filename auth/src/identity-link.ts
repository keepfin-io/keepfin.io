import type { IdentityLinkInput } from './types/link-user.ts'

export function parseIdentityLinkArguments(args: string[]): IdentityLinkInput {
  const [userID, authUserID, ...flags] = args
  const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i
  const values = new Map<string, string>()
  let confirmed = false
  for (let i = 0; i < flags.length; i++) {
    const flag = flags[i]
    if (flag === '--confirm' && !confirmed) { confirmed = true; continue }
    if ((flag !== '--reason' && flag !== '--ownership-reference') || values.has(flag)) throw new Error('Invalid or duplicate identity-link option')
    const value = flags[++i]
    if (!value || value.startsWith('--')) throw new Error(`Missing value for ${flag}`)
    values.set(flag, value.trim())
  }
  const reason = values.get('--reason') ?? ''
  const ownershipReference = values.get('--ownership-reference') ?? ''
  if (!userID || !authUserID || !uuid.test(userID) || !uuid.test(authUserID) || !confirmed ||
      reason.length < 10 || reason.length > 1000 || ownershipReference.length < 3 || ownershipReference.length > 200) {
    throw new Error('Usage: link-user APPLICATION_UUID AUTH_UUID --reason "Migration reason" --ownership-reference "Reviewed support ticket" --confirm')
  }
  return { userID, authUserID, reason, ownershipReference }
}
