import { test } from 'node:test'
import assert from 'node:assert/strict'
import { createClientIPResolver } from '../src/proxy.ts'

test('only trusted peers can supply an auth rate-limit IP', () => {
  const resolve = createClientIPResolver(['192.0.2.10/32', '2001:db8:1::/64'])
  assert.equal(resolve('192.0.2.10', '198.51.100.1'), '198.51.100.1')
  assert.equal(resolve('192.0.2.11', '198.51.100.1'), '192.0.2.11')
  assert.equal(resolve('192.0.2.10', ['198.51.100.1', '198.51.100.2']), '192.0.2.10')
  assert.equal(resolve('::ffff:192.0.2.10', '::ffff:198.51.100.1'), '198.51.100.1')
  assert.equal(resolve('2001:db8:1::1', '2001:db8:2::4'), '2001:db8:2::4')
  assert.throws(() => createClientIPResolver(['garbage']))
})
