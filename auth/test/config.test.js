import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readConfig } from '../src/config.js'
const base = { BETTER_AUTH_SECRET: 'a'.repeat(64), DATABASE_URL: 'postgres://localhost/test' }
test('reject unsafe production configuration', () => {
  assert.throws(() => readConfig({ ...base, NODE_ENV: 'production' }), /HTTPS/)
  assert.throws(() => readConfig({ ...base, BETTER_AUTH_SECRET: 'short' }), /32/)
  assert.throws(() => readConfig({ ...base, APP_ORIGIN: 'https://example.com/path' }), /origin/)
  assert.throws(() => readConfig({ ...base, NODE_ENV: 'production', APP_ORIGIN: 'https://example.com', AUTH_MAIL_MODE: 'file' }), /development-only/)
  assert.throws(() => readConfig({ ...base, GOOGLE_CLIENT_ID: 'id' }), /Both/)
  assert.equal(readConfig(base).mailMode, 'disabled')
})
