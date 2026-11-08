export interface IdentityLinkInput {
  userID: string
  authUserID: string
  reason: string
  ownershipReference: string
}

export interface IdentityLinkResult {
  audit_id: string
}
