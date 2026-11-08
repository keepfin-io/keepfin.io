export interface BaseAuthConfig {
  production: boolean
  internalTesting: boolean
  baseURL: string
  secret: string
  databaseURL: string
  mailMode: 'disabled' | 'file'
  mailDirectory: string
  trustedProxyCIDRs: string[]
  sessionRateLimit: number
}

// Enabled OAuth always has both credentials; callers cannot construct a partial setup.
export type AuthConfig = BaseAuthConfig & (
  | { google: false }
  | { google: true; googleClientID: string; googleClientSecret: string }
)
