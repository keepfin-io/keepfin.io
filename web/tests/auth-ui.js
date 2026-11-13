const assert = require('node:assert/strict')
const { test } = require('node:test')
const { launch, pause } = require('./ui-harness')

test('profile shows active devices, the current session and safe fallbacks without persisting tokens', async t => {
  const ui = await launch(t, { route: '/profile', authenticated: true, locale: 'tr-TR', authSessions: [
    { id: 'phone', token: 'private-phone-token', userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) Version/18.0 Mobile/15E148 Safari/604.1', ipAddress: '2001:db8:1234:5678:abcd:ef01:2345:6789', createdAt: '2026-11-08T08:00:00Z', expiresAt: '2026-11-15T08:00:00Z' },
    { id: 'unknown', token: 'private-unknown-token', userAgent: null, ipAddress: null, createdAt: null, expiresAt: 'invalid' },
    { id: 'session', token: 'private-current-token', userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Chrome/142.0.0.0 Safari/537.36', ipAddress: '203.0.113.10', createdAt: '2026-11-07T08:00:00Z', expiresAt: '2026-11-14T08:00:00Z' },
    { id: 'tablet', userAgent: 'Mozilla/5.0 (Linux; Android 14; Tablet) Chrome/142.0.0.0 Safari/537.36', createdAt: '2026-11-06T08:00:00Z' },
    { id: 'edge', userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/142.0.0.0 Safari/537.36 Edg/142.0.0.0', createdAt: '2026-11-05T08:00:00Z' },
    { id: 'linux', userAgent: 'Mozilla/5.0 (X11; Linux x86_64; rv:144.0) Gecko/20100101 Firefox/144.0', createdAt: '2026-11-04T08:00:00Z' },
    { id: 'chromebook', userAgent: 'Mozilla/5.0 (X11; CrOS x86_64 14541.0.0) AppleWebKit/537.36 Chrome/142.0.0.0 Safari/537.36', createdAt: '2026-11-03T08:00:00Z' },
  ] })
  await ui.until(() => ui.find('profile-current-session'))
  const rows = [...ui.w.document.querySelectorAll('[data-testid="profile-session"]')]
  assert.equal(rows.length, 7)
  assert.match(rows[0].textContent, /Chrome · macOS/)
  assert.match(rows[0].textContent, /Bu oturum/)
  assert.match(rows[0].textContent, /Girişteki IP adresi.*203\.0\.113\.10/)
  assert.match(rows[0].textContent, /Giriş zamanı.*2026/)
  assert.match(rows[1].textContent, /Safari · iOS.*Telefon/)
  assert.match(rows[2].textContent, /Chrome · Android.*Tablet/)
  assert.match(rows[3].textContent, /Edge · Windows.*Bilgisayar/)
  assert.match(rows[4].textContent, /Firefox · Linux.*Bilgisayar/)
  assert.match(rows[5].textContent, /Chrome · ChromeOS.*Bilgisayar/)
  assert.match(rows[6].textContent, /Bilinmeyen tarayıcı · Bilinmeyen işletim sistemi.*Bilinmeyen cihaz/)
  assert.equal(rows[6].querySelectorAll('dd')[0].textContent, '—')
  assert.equal(rows[6].querySelectorAll('dd')[1].textContent, '—')
  assert.equal(rows[6].querySelectorAll('dd')[2].textContent, '—')
  const icons = rows.map(row => row.querySelector('svg').innerHTML)
  assert.equal(icons[0], icons[1], 'macOS and iOS share the Apple icon')
  assert.equal(new Set(icons).size, 6, 'each operating system and the unknown fallback have distinct icons')
  assert.equal(ui.find('profile-sessions-refresh'), null)
  assert.equal(ui.w.document.querySelectorAll('[data-testid="profile-current-session"]').length, 1)
  assert.doesNotMatch(ui.w.document.body.innerHTML, /private-.*-token/)
  assert.doesNotMatch(JSON.stringify({ ...ui.w.localStorage, ...ui.w.sessionStorage }), /private-.*-token/)
  const call = ui.calls.find(call => call.endpoint === '/api/auth/list-sessions')
  assert.equal(call.method, 'GET')
  assert.deepEqual(call.query, {}, 'session scope is determined by the server, not a supplied user ID')
  ui.find('profile-theme-light').click()
  await ui.until(() => !ui.w.document.documentElement.classList.contains('dark'))
  assert.equal(ui.calls.filter(call => call.endpoint === '/api/auth/list-sessions').length, 1, 'preference changes must not reload sessions')
  assert.deepEqual(ui.errors, [])
})

test('profile session list has loading, retry, empty and recent-sign-in states', async t => {
  let release
  const sessionsGate = new Promise(resolve => { release = resolve })
  t.after(() => release())
  const ui = await launch(t, { route: '/profile', authenticated: true, sessionsGate, sessionsFailure: 'UNAVAILABLE' })
  await ui.until(() => ui.find('profile-sessions-loading'))
  assert.ok(ui.find('profile-email'), 'sessions must not block the rest of the profile')
  assert.equal(ui.find('profile-sessions-refresh'), null)
  assert.equal(ui.find('profile-sessions-empty'), null)
  release()
  await ui.until(() => ui.find('profile-sessions-error'))
  assert.match(ui.find('profile-sessions-error').textContent, /could not be loaded/)
  ui.state.sessionsFailure = ''
  ui.find('profile-sessions-retry').click()
  await ui.until(() => ui.find('profile-sessions-empty'))
  assert.deepEqual(ui.errors, [])
  const reauth = await launch(t, { route: '/profile', authenticated: true, sessionsFailure: 'SESSION_NOT_FRESH' })
  await reauth.until(() => reauth.find('profile-sessions-error'))
  assert.match(reauth.find('profile-sessions-error').textContent, /sign out and sign in again/)
  assert.equal(reauth.find('profile-sessions-retry'), null)
  assert.equal(reauth.calls.some(call => call.endpoint === '/api/auth/sign-out'), false, 'reauthentication must be an explicit user action')
  reauth.state.failSignOut = true
  reauth.find('profile-sessions-sign-out').click()
  await reauth.until(() => reauth.find('profile-sessions-error')?.textContent.includes('Sign-out failed'))
  assert.equal(reauth.w.location.pathname, '/profile')
  reauth.state.failSignOut = false
  reauth.find('profile-sessions-sign-out').click()
  await reauth.until(() => reauth.find('login-email'))
  assert.equal(reauth.find('profile-sessions'), null)
  assert.deepEqual(reauth.errors, [])
})

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
  assert.equal(ui.w.document.querySelectorAll('[data-testid="app-footer"]').length, 1)
  const collapsedLogo = ui.find('sidebar-brand-link').querySelector('span.hidden')
  assert.ok(collapsedLogo, 'collapsed brand remains hidden in the expanded sidebar')
  assert.equal(collapsedLogo.classList.contains('inline-flex'), false)
  assert.equal(ui.find('sidebar-nav-profile'), null)
  ui.find('sidebar-user-menu-trigger').dispatchEvent(new ui.w.KeyboardEvent('keydown', { key: 'Enter', bubbles: true }))
  await ui.until(() => ui.find('sidebar-profile-link'))
  ui.find('sidebar-profile-link').click()
  await ui.until(() => ui.find('profile-theme-light'))
  assert.ok(ui.find('footer-legal-trigger'))
  assert.equal(ui.w.document.querySelectorAll('[data-testid="app-footer"]').length, 1)
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


for (const route of ['/login', '/accounts', '/accounts/new']) {
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

for (const locale of ['en-US', 'tr-TR']) {
  test(`shared footer provides a dismissible legal notice without sign-in or requests: ${locale}`, async t => {
    const ui = await launch(t, { locale })
    await ui.until(() => ui.find('login-email'))
    assert.equal(ui.w.document.querySelectorAll('[data-testid="app-footer"]').length, 1)
    const source = ui.find('footer-source-link')
    assert.equal(source.href, 'https://github.com/keepfin-io/keepfin.io')
    assert.match(source.rel, /noopener/)
    const trigger = ui.find('footer-legal-trigger')
    const callsBeforeNotice = ui.calls.length
    trigger.focus()
    trigger.click()
    await ui.until(() => ui.find('footer-legal-dialog'))
    const dialog = ui.find('footer-legal-dialog')
    assert.equal(dialog.getAttribute('role'), 'dialog')
    assert.match(dialog.textContent, locale === 'tr-TR' ? /hukuken sınırlandırılamayan sorumlulukları/ : /liabilities that cannot lawfully be limited/)
    dialog.dispatchEvent(new ui.w.KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }))
    await ui.until(() => !ui.find('footer-legal-dialog') && ui.w.document.activeElement === trigger)
    trigger.click()
    await ui.until(() => ui.find('footer-legal-close'))
    ui.find('footer-legal-close').click()
    await ui.until(() => !ui.find('footer-legal-dialog'))
    assert.equal(ui.calls.length, callsBeforeNotice)
    assert.equal(ui.w.location.pathname, '/login')
    assert.equal(ui.find('login-submit').disabled, false)
    assert.deepEqual(ui.errors, [])
  })
}
