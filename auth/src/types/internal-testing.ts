export interface InternalTestingIdentity {
  name: string
  email: string
  authID: string
  applicationID: string
  credentialID: string
}

export interface SeedOptions {
  waitForSchema?: boolean
}

