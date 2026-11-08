import { randomUUID } from 'node:crypto'
import { mkdir, writeFile } from 'node:fs/promises'
import { resolve } from 'node:path'
import type { BetterAuthOptions } from 'better-auth'
import { APIError, createAuthMiddleware } from 'better-auth/api'
import { magicLink } from 'better-auth/plugins'
import { internalTestingUser } from './internal-testing.ts'
import type { AuthConfig, BaseAuthConfig } from "./types/config.ts";
import { createClientIPResolver } from "./proxy.ts";

export function readConfig(env: NodeJS.ProcessEnv = process.env): AuthConfig {
  const production = env.NODE_ENV === "production";
  const internalTesting = env.AUTH_INTERNAL_TESTING === "true";
  if (
    internalTesting &&
    env.NODE_ENV !== "development" &&
    env.NODE_ENV !== "test"
  ) {
    throw new Error("Internal testing login is development/test-only");
  }

  const baseURL = new URL(env.APP_ORIGIN || "http://localhost:8080");
  if (
    baseURL.origin !== (env.APP_ORIGIN || "http://localhost:8080") ||
    !["http:", "https:"].includes(baseURL.protocol) ||
    baseURL.username ||
    baseURL.password
  ) {
    throw new Error("APP_ORIGIN must be an origin without a trailing slash");
  }

  if (production && baseURL.protocol !== "https:")
    throw new Error("Production requires HTTPS");
  if (!env.BETTER_AUTH_SECRET || env.BETTER_AUTH_SECRET.length < 32) {
    throw new Error(
      "BETTER_AUTH_SECRET must contain at least 32 random characters",
    );
  }

  if (!env.DATABASE_URL) throw new Error("DATABASE_URL is required");
  const mailMode = env.AUTH_MAIL_MODE || "disabled";

  if (mailMode !== "disabled" && mailMode !== "file")
    throw new Error("Unsupported AUTH_MAIL_MODE");

  if (production && mailMode === "file")
    throw new Error("File mail delivery is development-only");
  const googleClientID = env.GOOGLE_CLIENT_ID;
  const trustedProxyCIDRs = (env.AUTH_TRUSTED_PROXY_CIDRS ?? "")
    .split(",")
    .map((value) => value.trim())
    .filter(Boolean);
  createClientIPResolver(trustedProxyCIDRs);
  const sessionRateLimit = Number(env.AUTH_SESSION_RATE_LIMIT ?? 900);

  if (
    !Number.isInteger(sessionRateLimit) ||
    sessionRateLimit < 1 ||
    sessionRateLimit > 100000
  )
    throw new Error("AUTH_SESSION_RATE_LIMIT must be between 1 and 100000");
  const googleClientSecret = env.GOOGLE_CLIENT_SECRET;

  if (Boolean(env.GOOGLE_CLIENT_ID) !== Boolean(env.GOOGLE_CLIENT_SECRET)) {
    throw new Error("Both Google credentials are required");
  }
  const config: BaseAuthConfig = {
    production,
    internalTesting,
    baseURL: baseURL.origin,
    secret: env.BETTER_AUTH_SECRET,
    trustedProxyCIDRs,
    sessionRateLimit,
    databaseURL: env.DATABASE_URL,
    mailMode,
    mailDirectory: env.AUTH_MAIL_DIRECTORY || "./.mail",
  };
  return googleClientID && googleClientSecret
    ? { ...config, google: true, googleClientID, googleClientSecret }
    : { ...config, google: false };
}

export function createBetterAuthOptions(config: AuthConfig) {
  // Check the options without widening plugin types used by AuthInstance.
  return {
    appName: 'Keepfin',
    baseURL: config.baseURL,
    basePath: '/api/auth',
    secret: config.secret,
    logger: {
      // Better Auth error metadata can include OAuth state. Keep it out of logs.
      log(level) { console.error(`Better Auth event: ${level}`) },
    },
    trustedOrigins: [config.baseURL],
    emailAndPassword: {
      enabled: config.internalTesting === true,
      disableSignUp: true,
      requireEmailVerification: true,
      minPasswordLength: config.internalTesting ? 5 : 8,
    },
    hooks: {
      before: createAuthMiddleware(async (ctx) => {
        // The development credential flow is reserved for the explicitly seeded
        // identity; it does not re-enable passwords for ordinary accounts.
        if (ctx.path !== '/sign-in/email') return
        const body: unknown = ctx.body
        const email = typeof body === 'object' && body !== null && 'email' in body ? body.email : undefined
        if (!config.internalTesting || typeof email !== 'string' || email.toLowerCase() !== internalTestingUser.email) {
          throw new APIError('UNAUTHORIZED', { message: 'Invalid email or password' })
        }
      }),
    },
    user: {
      modelName: 'auth_user',
      changeEmail: { enabled: false },
    },
    session: {
      modelName: 'auth_session',
      expiresIn: 60 * 60 * 24 * 7,
      updateAge: 60 * 60 * 24,
      cookieCache: { enabled: false },
    },
    account: {
      modelName: 'auth_account',
      encryptOAuthTokens: true,
      // Only link matching verified identities; never trust an unverified local account.
      accountLinking: {
        enabled: true,
        requireLocalEmailVerified: true,
        allowDifferentEmails: false,
      },
    },
    verification: { modelName: 'auth_verification' },
    rateLimit: {
      enabled: true,
      storage: 'database',
      modelName: 'auth_rate_limit',
      window: 60,
      max: 100,
      customRules: {
        '/get-session': { window: 60, max: config.sessionRateLimit },
        '/sign-in/social': { window: 60, max: 10 },
        '/sign-in/email': { window: 60, max: 5 },
      },
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
        await writeFile(
          resolve(config.mailDirectory, `${randomUUID()}.json`),
          JSON.stringify({ to: email, url, createdAt: new Date().toISOString() }),
          { mode: 0o600, flag: 'wx' },
        )
      },
    })] : [],
  } satisfies BetterAuthOptions
}
