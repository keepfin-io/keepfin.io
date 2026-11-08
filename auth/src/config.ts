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
