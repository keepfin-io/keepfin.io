export function readConfig(env = process.env) {
  const production = env.NODE_ENV === 'production'
  const internalTesting = env.AUTH_INTERNAL_TESTING === 'true'
  if (internalTesting && !['development', 'test'].includes(env.NODE_ENV)) {
    throw new Error('Internal testing login is development/test-only')
  }
  const baseURL = new URL(env.APP_ORIGIN || 'http://localhost:8080')
  if (baseURL.origin !== (env.APP_ORIGIN || 'http://localhost:8080') ||
      !['http:', 'https:'].includes(baseURL.protocol) || baseURL.username || baseURL.password) {
    throw new Error('APP_ORIGIN must be an origin without a trailing slash')
  }
  if (production && baseURL.protocol !== 'https:') throw new Error('Production requires HTTPS')
  if (!env.BETTER_AUTH_SECRET || env.BETTER_AUTH_SECRET.length < 32) {
    throw new Error('BETTER_AUTH_SECRET must contain at least 32 random characters')
  }
  if (!env.DATABASE_URL) throw new Error('DATABASE_URL is required')
  const mailMode = env.AUTH_MAIL_MODE || 'disabled'
  if (!['disabled', 'file'].includes(mailMode)) throw new Error('Unsupported AUTH_MAIL_MODE')
  if (production && mailMode === 'file') throw new Error('File mail delivery is development-only')
  const google = Boolean(env.GOOGLE_CLIENT_ID && env.GOOGLE_CLIENT_SECRET)
  if (Boolean(env.GOOGLE_CLIENT_ID) !== Boolean(env.GOOGLE_CLIENT_SECRET)) {
    throw new Error('Both Google credentials are required')
  }
  return {
    production, internalTesting, baseURL: baseURL.origin, secret: env.BETTER_AUTH_SECRET,
    databaseURL: env.DATABASE_URL, google,
    googleClientID: env.GOOGLE_CLIENT_ID, googleClientSecret: env.GOOGLE_CLIENT_SECRET,
    mailMode, mailDirectory: env.AUTH_MAIL_DIRECTORY || './.mail',
  }
}
