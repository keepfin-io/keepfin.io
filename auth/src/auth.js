import { betterAuth } from 'better-auth'
import { magicLink } from 'better-auth/plugins'
import { mkdir, writeFile } from 'node:fs/promises'
import { randomUUID } from 'node:crypto'
import { resolve } from 'node:path'

export function createAuth(config, pool) {
  return betterAuth({
    appName: 'Keepfin',
    baseURL: config.baseURL,
    basePath: '/api/auth',
    secret: config.secret,
    logger: {
      // Better Auth error metadata can include OAuth state. Keep it out of logs.
      log(level) { console.error(`Better Auth event: ${level}`) },
    },
    database: pool,
    trustedOrigins: [config.baseURL],
    emailAndPassword: { enabled: false },
    user: { modelName: 'auth_user', changeEmail: { enabled: false } },
    session: {
      modelName: 'auth_session',
      expiresIn: 60 * 60 * 24 * 7,
      updateAge: 60 * 60 * 24,
      cookieCache: { enabled: false },
    },
    account: {
      modelName: 'auth_account',
      encryptOAuthTokens: true,
      // Passwordless accounts can share a verified email; do not trust an
      // unverified local account or permit linking different email addresses.
      accountLinking: { enabled: true, requireLocalEmailVerified: true, allowDifferentEmails: false },
    },
    verification: { modelName: 'auth_verification' },
    rateLimit: {
      enabled: true, storage: 'database', modelName: 'auth_rate_limit',
      window: 60, max: 100,
      customRules: { '/get-session': false, '/sign-in/social': { window: 60, max: 10 } },
    },
    advanced: {
      useSecureCookies: config.baseURL.startsWith('https:'),
      cookiePrefix: 'keepfin',
      ipAddress: { ipAddressHeaders: ['x-real-ip'] },
      database: { generateId: 'uuid' },
    },
    socialProviders: config.google ? {
      google: { clientId: config.googleClientID, clientSecret: config.googleClientSecret },
    } : {},
    plugins: config.mailMode === 'file' ? [magicLink({
      expiresIn: 300,
      storeToken: 'hashed',
      rateLimit: { window: 60, max: 5 },
      async sendMagicLink({ email, url }) {
        // No public inbox endpoint and no tokens in logs. Only local developers
        // with filesystem access can read these test messages.
        await mkdir(config.mailDirectory, { recursive: true, mode: 0o700 })
        await writeFile(resolve(config.mailDirectory, `${randomUUID()}.json`),
          JSON.stringify({ to: email, url, createdAt: new Date().toISOString() }),
          { mode: 0o600, flag: 'wx' })
      },
    })] : [],
  })
}
