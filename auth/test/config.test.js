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

test('internal testing requires explicit development or test configuration', () => {
  assert.equal(readConfig(base).internalTesting, false)
  assert.equal(readConfig({ ...base, NODE_ENV: 'development', AUTH_INTERNAL_TESTING: 'true' }).internalTesting, true)
  assert.equal(readConfig({ ...base, NODE_ENV: 'test', AUTH_INTERNAL_TESTING: 'true' }).internalTesting, true)
  assert.throws(() => readConfig({ ...base, AUTH_INTERNAL_TESTING: 'true' }), /development\/test-only/)
  assert.throws(() => readConfig({ ...base, NODE_ENV: 'production', APP_ORIGIN: 'https://example.com', AUTH_INTERNAL_TESTING: 'true' }), /development\/test-only/)
})
