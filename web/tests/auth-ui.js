const assert = require('node:assert/strict')
const { test } = require('node:test')
const { launch, pause } = require('./ui-harness')

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


test('Google and Apple remain visible when not connected and do not send OAuth requests', async t => {
  const ui = await launch(t, { google: false })
  await ui.until(() => ui.find('login-email'))
  assert.equal(ui.find('login-google').disabled, true)
  assert.equal(ui.find('login-apple').disabled, true)
  assert.match(ui.find('login-google').textContent, /Coming soon/)
  assert.match(ui.find('login-apple').textContent, /Coming soon/)
  ui.find('login-google').click(); ui.find('login-apple').click()
  assert.equal(ui.calls.some(call => call.endpoint.includes('/sign-in/social')), false)
  assert.equal(ui.find('login-internal'), null)
  assert.deepEqual(ui.errors, [])
})

test('internal testing login reports invalid credentials then opens the verified admin session', async t => {
  const ui = await launch(t, { internalTesting: true, google: false })
  await ui.until(() => ui.find('login-internal'))
  ui.find('login-internal').open = true
  assert.equal(ui.find('login-username').value, 'internal-testing')
  assert.equal(ui.find('login-password').type, 'password')
  async function submit(value, duplicate = false) {
    const input = ui.find('login-password'); input.value = value
    input.dispatchEvent(new ui.w.Event('input', { bubbles: true }))
    await pause()
    const form = ui.find('login-internal-form')
    if (duplicate) form.dispatchEvent(new ui.w.Event('submit', { bubbles: true, cancelable: true }))
    form.dispatchEvent(new ui.w.Event('submit', { bubbles: true, cancelable: true }))
  }
  await submit('wrong')
  await ui.until(() => ui.find('login-error'))
  assert.match(ui.find('login-error').textContent, /Incorrect username or password/)
  await submit('12345', true)
  await ui.until(() => ui.find('dashboard-page'))
  assert.equal(ui.w.location.pathname, '/')
  const calls = ui.calls.filter(call => call.endpoint === '/api/auth/sign-in/email')
  assert.equal(calls.length, 2)
  assert.equal(calls[1].body.email, 'internal-testing@keepfin.test')
  assert.equal(ui.w.sessionStorage.getItem('keepnode_user'), null)
  assert.deepEqual(ui.errors, [])
})


for (const route of ['/login', '/accounts']) {
  test(`no protected layout is painted while the initial guest session resolves: ${route}`, async t => {
    let release
    const sessionGate = new Promise(resolve => { release = resolve })
    t.after(() => release())
    const ui = await launch(t, { route, sessionGate, onBoot: async ({ find, until, calls, protectedPaints }) => {
      await until(() => calls.some(call => call.endpoint === '/api/auth/get-session'))
      await pause()
      assert.equal(find('app-layout'), null)
      assert.ok(find('app-startup'))
      assert.equal(find('dashboard-page'), null)
      assert.deepEqual(protectedPaints, [], 'protected layout must never be mounted before session verification')
      assert.equal(calls.some(call => call.endpoint === '/api/v1/portfolio'), false)
      release()
    } })
    await ui.until(() => ui.find('login-email'))
    assert.equal(ui.w.location.pathname, '/login')
    assert.deepEqual(ui.protectedPaints, [])
    assert.deepEqual(ui.errors, [])
  })
}

test('authenticated entry waits for the application profile before rendering the dashboard', async t => {
  let release
  const profileGate = new Promise(resolve => { release = resolve })
  t.after(() => release())
  const ui = await launch(t, { authenticated: true, profileGate, onBoot: async ({ find, until, calls, protectedPaints }) => {
    await until(() => calls.some(call => call.endpoint === '/api/v1/user/me'))
    await pause()
    assert.equal(find('app-layout'), null)
    assert.deepEqual(protectedPaints, [])
    release()
  } })
  await ui.until(() => ui.find('dashboard-page'))
  assert.equal(ui.w.location.pathname, '/')
  assert.deepEqual(ui.errors, [])
})


test('Turkish login preserves theme choice and password visibility without sending requests', async t => {
  const ui = await launch(t, { locale: 'tr-TR', internalTesting: true, google: false })
  await ui.until(() => ui.find('login-email'))
  assert.equal(ui.find('login-email').placeholder, 'ornek@eposta.com')
  assert.match(ui.find('login-apple').textContent, /Yakında/)
  assert.equal(ui.w.document.documentElement.classList.contains('dark'), true)
  ui.find('login-theme').click()
  await ui.until(() => !ui.w.document.documentElement.classList.contains('dark'))
  assert.equal(ui.w.localStorage.getItem('theme'), 'light')
  ui.find('login-internal').open = true
  ui.find('login-password-toggle').click()
  await ui.until(() => ui.find('login-password').type === 'text')
  assert.equal(ui.find('login-password-toggle').getAttribute('aria-pressed'), 'true')
  ui.find('login-password-toggle').click()
  await ui.until(() => ui.find('login-password').type === 'password')
  assert.equal(ui.calls.some(call => call.endpoint === '/api/v1/user/theme' || call.endpoint === '/api/auth/sign-in/email'), false)
  assert.deepEqual(ui.errors, [])
})
