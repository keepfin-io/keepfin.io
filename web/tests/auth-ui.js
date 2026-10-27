const { JSDOM, VirtualConsole } = require('jsdom')
const assert = require('node:assert/strict')
const fs = require('node:fs')
const path = require('node:path')
const { test } = require('node:test')
const build = path.resolve(process.env.AUTH_UI_BUILD || path.join(__dirname, '../dist'))
const pause = () => new Promise(resolve => setTimeout(resolve, 25))

async function launch(t, { route = '/login', authenticated = false, outage = false, conflict = false, width = 320 } = {}) {
  const errors = [], calls = []
  const vc = new VirtualConsole()
  vc.on('jsdomError', error => errors.push(error.message))
  vc.on('error', (...args) => errors.push(args.join(' ')))
  const dom = new JSDOM('<html><body><div id="app"></div></body></html>', {
    url: `http://localhost:8080${route}`, runScripts: 'outside-only', pretendToBeVisual: true, virtualConsole: vc,
  })
  t.after(() => dom.window.close())
  const w = dom.window
  const user = { id: 'app-uuid', email: 'test@example.com', theme: 'dark', base_currency: 'USD', locale: 'en-US', created_at: '2026-01-01T00:00:00Z' }
  // A forged legacy profile must never unlock protected routes.
  w.sessionStorage.setItem('keepnode_user', JSON.stringify(user))
  w.matchMedia = query => ({ matches: query.includes('max-width') && width < 768, media: query, addEventListener() {}, removeEventListener() {}, addListener() {}, removeListener() {} })
  w.ResizeObserver = class { observe() {} unobserve() {} disconnect() {} }
  w.IntersectionObserver = class { observe() {} unobserve() {} disconnect() {} }
  w.scrollTo = () => {}
  w.HTMLElement.prototype.scrollIntoView = () => {}
  for (const key of ['Request', 'Response', 'Headers', 'AbortController', 'AbortSignal', 'TextEncoder', 'TextDecoder']) w[key] = global[key]
  const state = { authenticated, outage, conflict, failSignOut: false, rateLimited: false }
  w.fetch = async (input, options = {}) => {
    const endpoint = new URL(input.url || String(input), w.location.origin).pathname
    const body = options.body ? JSON.parse(options.body) : null
    calls.push({ endpoint, method: options.method || 'GET', body })
    let data, status = 200
    if (endpoint === '/api/auth/get-session') {
      if (state.outage) { status = 503; data = { message: 'Unavailable' } }
      else data = state.authenticated ? { user: { id: 'auth-uuid', email: user.email, emailVerified: true }, session: { id: 'session', userId: 'auth-uuid', expiresAt: '2099-01-01T00:00:00Z' } } : null
    } else if (endpoint === '/api/auth/config') data = { google: true, email: true, developmentMail: true }
    else if (endpoint === '/api/auth/sign-in/magic-link') {
      status = state.rateLimited ? 429 : 200; data = state.rateLimited ? { message: 'Rate limited' } : { status: true }
    } else if (endpoint === '/api/auth/sign-out') {
      status = state.failSignOut ? 503 : 200
      if (!state.failSignOut) state.authenticated = false
      data = state.failSignOut ? { message: 'Unavailable' } : { success: true }
    } else if (endpoint === '/api/v1/user/me') {
      status = state.conflict ? 409 : 200
      data = state.conflict ? { error: true } : { error: false, user: { ...user } }
    } else if (endpoint === '/api/v1/user/theme') { user.theme = body.theme; data = { error: false } }
    else if (endpoint === '/api/v1/accounts') data = { error: false, accounts: [] }
    else throw new Error(`Unexpected request: ${endpoint}`)
    return new global.Response(JSON.stringify(data), { status, headers: { 'Content-Type': 'application/json' } })
  }
  const html = fs.readFileSync(path.join(build, 'index.html'), 'utf8')
  for (const [, src] of html.matchAll(/<script[^>]+src="([^"]+)"/g)) w.eval(fs.readFileSync(path.join(build, src.replace(/^\//, '')), 'utf8'))
  const find = id => w.document.querySelector(`[data-testid="${id}"]`)
  async function until(check) {
    for (let i = 0; i < 80; i++) { if (check()) return; await pause() }
    assert.ok(check(), `UI did not settle: ${w.document.body.textContent}; errors: ${errors}`)
  }
  await until(() => find('login-page') || find('app-layout') || find('auth-callback-error'))
  return { w, find, state, calls, errors, until }
}

test('passwordless form, latest branding, no cached-identity access, no duplicate sends', async t => {
  const ui = await launch(t, { route: '/accounts' })
  const { w, find, until, calls } = ui
  await until(() => find('login-email'))
  assert.equal(w.location.pathname, '/login')
  assert.equal(w.sessionStorage.getItem('keepnode_user'), null)
  assert.equal(find('login-password'), null)
  assert.equal(find('login-register-link'), null)
  assert.ok(find('login-brand').querySelector('img[src="/brand/keepfin.svg"]'))
  assert.ok(find('login-google'))
  const email = find('login-email'); email.value = 'person@example.com'; email.dispatchEvent(new w.Event('input', { bubbles: true }))
  find('login-form').dispatchEvent(new w.Event('submit', { bubbles: true, cancelable: true }))
  await until(() => find('login-link-sent'))
  assert.match(find('login-link-sent').textContent, /No email was sent/)
  assert.equal(find('login-submit').disabled, true)
  find('login-form').dispatchEvent(new w.Event('submit', { bubbles: true, cancelable: true }))
  await pause()
  assert.equal(calls.filter(call => call.endpoint.endsWith('/sign-in/magic-link')).length, 1)
  assert.equal(calls.find(call => call.endpoint.endsWith('/sign-in/magic-link')).body.callbackURL, '/auth/callback')
  assert.deepEqual(ui.errors, [])
})

test('register redirects and invalid links show a generic error', async t => {
  const ui = await launch(t, { route: '/register' })
  await ui.until(() => ui.find('login-email'))
  assert.equal(ui.w.location.pathname, '/login')
  const errorUI = await launch(t, { route: '/login?error=expired-token' })
  await errorUI.until(() => errorUI.find('login-error'))
  assert.match(errorUI.find('login-error').textContent, /expired or already been used/)
})

test('outage cannot unlock dashboard; migration conflict can sign out', async t => {
  const offline = await launch(t, { route: '/accounts', outage: true })
  await offline.until(() => offline.find('login-error'))
  assert.equal(offline.w.location.pathname, '/login')
  offline.state.outage = false
  offline.find('login-retry').click()
  await offline.until(() => !offline.find('login-error'))
  const collision = await launch(t, { authenticated: true, conflict: true })
  await collision.until(() => collision.find('login-sign-out'))
  assert.match(collision.find('login-error').textContent, /existing account needs/)
  collision.state.failSignOut = true
  collision.find('login-sign-out').click()
  await collision.until(() => collision.find('login-error')?.textContent.includes('Sign-out failed'))
  assert.equal(collision.state.authenticated, true)
  collision.state.failSignOut = false
  collision.find('login-sign-out').click()
  await collision.until(() => !collision.find('login-sign-out'))
  assert.equal(collision.state.authenticated, false)
})

test('verified callback loads profile; dark theme survives navigation', async t => {
  const ui = await launch(t, { route: '/auth/callback', authenticated: true, width: 1024 })
  await ui.until(() => ui.find('app-layout') && ui.w.location.pathname === '/' && ui.w.document.documentElement.classList.contains('dark'))
  await pause()
  assert.ok(ui.find('header-brand-link').querySelector('img[src="/brand/keepfin-mark.svg"]'))
  const collapsedLogo = ui.find('sidebar-brand-link').querySelector('span.hidden')
  assert.ok(collapsedLogo, 'collapsed brand remains hidden in the expanded sidebar')
  assert.equal(collapsedLogo.classList.contains('inline-flex'), false)
  ui.find('sidebar-nav-profile').click()
  await ui.until(() => ui.find('profile-theme-light'))
  assert.equal(ui.w.document.documentElement.classList.contains('dark'), true)
  ui.find('profile-theme-light').click()
  await ui.until(() => !ui.w.document.documentElement.classList.contains('dark'))
  assert.equal(ui.w.sessionStorage.getItem('keepnode_user'), null)
  assert.deepEqual(ui.errors, [])
})

test('rate-limit errors are visible and permit retry', async t => {
  const ui = await launch(t)
  await ui.until(() => ui.find('login-email'))
  ui.state.rateLimited = true
  const email = ui.find('login-email'); email.value = 'person@example.com'; email.dispatchEvent(new ui.w.Event('input', { bubbles: true }))
  ui.find('login-form').dispatchEvent(new ui.w.Event('submit', { bubbles: true, cancelable: true }))
  await ui.until(() => ui.find('login-error'))
  assert.match(ui.find('login-error').textContent, /Too many attempts/)
  assert.equal(ui.find('login-submit').disabled, false)
})
