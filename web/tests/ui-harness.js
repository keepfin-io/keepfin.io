const { JSDOM, VirtualConsole } = require('jsdom')
const assert = require('node:assert/strict')
const fs = require('node:fs')
const path = require('node:path')
const build = path.resolve(process.env.AUTH_UI_BUILD || path.join(__dirname, '../dist'))
const pause = () => new Promise(resolve => setTimeout(resolve, 25))

async function launch(t, { route = '/login', authenticated = false, outage = false, conflict = false, width = 320, internalTesting = false, google = true, sessionGate, profileGate, onBoot, locale = 'en-US', portfolioResponse, portfolioGate, measuredChartWidth, accountsResponse = [], holdingsResponse = [], activitiesResponse = [] } = {}) {
  const errors = [], calls = []
  const vc = new VirtualConsole()
  vc.on('jsdomError', error => errors.push(error.message))
  vc.on('error', (...args) => errors.push(args.join(' ')))
  const html = fs.readFileSync(path.join(build, 'index.html'), 'utf8')
  const dom = new JSDOM(html, {
    url: `http://localhost:8080${route}`, runScripts: 'outside-only', pretendToBeVisual: true, virtualConsole: vc,
  })
  t.after(() => dom.window.close())
  const w = dom.window
  w.localStorage.setItem('locale', locale)
  const user = { id: 'app-uuid', email: 'test@example.com', theme: 'dark', base_currency: 'USD', locale: 'en-US', created_at: '2026-01-01T00:00:00Z' }
  // A forged legacy profile must never unlock protected routes.
  w.sessionStorage.setItem('keepnode_user', JSON.stringify(user))
  w.matchMedia = query => ({ matches: query.includes('max-width') && width < 768, media: query, addEventListener() {}, removeEventListener() {}, addListener() {}, removeListener() {} })
  w.ResizeObserver = class {
    constructor(callback) { this.callback = callback; this.active = true }
    observe(target) {
      if (measuredChartWidth == null) return
      queueMicrotask(() => {
        if (!this.active) return
        this.callback([{ target, contentRect: { width: measuredChartWidth, height: 250 }, contentBoxSize: [{ inlineSize: measuredChartWidth, blockSize: 250 }], borderBoxSize: [{ inlineSize: measuredChartWidth, blockSize: 250 }] }])
      })
    }
    unobserve() {} disconnect() { this.active = false }
  }
  w.IntersectionObserver = class { observe() {} unobserve() {} disconnect() {} }
  w.scrollTo = () => {}
  w.HTMLElement.prototype.scrollIntoView = () => {}
  for (const key of ['Request', 'Response', 'Headers', 'AbortController', 'AbortSignal', 'TextEncoder', 'TextDecoder']) w[key] = global[key]
  const state = { authenticated, outage, conflict, failSignOut: false, rateLimited: false, accountCreateFailure: false, accountCreateGate: null, accounts: accountsResponse.map(account => ({ ...account })), portfolioFailure: false, portfolio: portfolioResponse || { currency: 'USD', summary: {}, holdings: [], history: [], income: [], allocations: {}, warnings: [], complete: true, from: '2026-01-01', to: '2026-01-01' } }
  Object.assign(state, { holdings: holdingsResponse.map(item => ({ ...item })), activities: activitiesResponse.map(item => ({ ...item })), mutationGate: null, mutationFailure: false })
  w.fetch = async (input, options = {}) => {
    const requestURL = new URL(input.url || String(input), w.location.origin)
    const endpoint = requestURL.pathname
    const body = options.body ? JSON.parse(options.body) : null
    calls.push({ endpoint, method: options.method || 'GET', body, query: Object.fromEntries(requestURL.searchParams) })
    let data, status = 200
    if (endpoint === '/api/auth/get-session') {
      if (sessionGate) await sessionGate
      if (state.outage) { status = 503; data = { message: 'Unavailable' } }
      else data = state.authenticated ? { user: { id: 'auth-uuid', email: user.email, emailVerified: true }, session: { id: 'session', userId: 'auth-uuid', expiresAt: '2099-01-01T00:00:00Z' } } : null
    } else if (endpoint === '/api/auth/config') data = { google, apple: false, internalTesting, email: true, developmentMail: true }
    else if (endpoint === '/api/auth/sign-in/magic-link') {
      status = state.rateLimited ? 429 : 200; data = state.rateLimited ? { message: 'Rate limited' } : { status: true }
    } else if (endpoint === '/api/auth/sign-out') {
      status = state.failSignOut ? 503 : 200
      if (!state.failSignOut) state.authenticated = false
      data = state.failSignOut ? { message: 'Unavailable' } : { success: true }
    } else if (endpoint === '/api/v1/user/me') {
      if (profileGate) await profileGate
      status = state.conflict ? 409 : 200
      data = state.conflict ? { error: true } : { error: false, user: { ...user } }
    } else if (endpoint === '/api/v1/user/theme') { user.theme = body.theme; data = { error: false } }
    else if (endpoint === '/api/auth/sign-in/email') {
      const valid = internalTesting && body.email === 'internal-testing@keepfin.test' && body.password === '12345'
      status = valid ? 200 : 401
      if (valid) { state.authenticated = true; user.user_role = 'admin' }
      data = valid ? { user, redirect: false } : { message: 'Invalid credentials' }
    }
    else if (endpoint === '/api/v1/portfolio') {
      if (portfolioGate) await portfolioGate
      status = state.portfolioFailure ? 503 : 200
      data = state.portfolioFailure ? { error: true, message: 'Portfolio service unavailable' } : { error: false, portfolio: state.portfolio }
    }
    else if (endpoint === '/api/v1/accounts') data = { error: false, accounts: state.accounts }
    else if (endpoint === '/api/v1/accounts/create' && options.method === 'POST') {
      if (state.accountCreateGate) await state.accountCreateGate
      if (state.accountCreateFailure) { status = 400; data = { error: true, message: 'Account creation failed' } }
      else {
        const account = { ...body, id: `account-${state.accounts.length + 1}`, cash_balance: body.initial_balance, is_active: true }
        state.accounts.push(account)
        status = 201; data = { error: false, account }
      }
    }
    else if (/^\/api\/v1\/accounts\/[^/]+$/.test(endpoint) && options.method === 'PUT') {
      const account = state.accounts.find(account => endpoint === `/api/v1/accounts/${account.id}`)
      if (!account) throw new Error(`Unknown account: ${endpoint}`)
      Object.assign(account, body)
      data = { error: false, account }
    }
    else if (/^\/api\/v1\/accounts\/[^/]+$/.test(endpoint) && options.method === 'GET') {
      const account = state.accounts.find(account => endpoint === `/api/v1/accounts/${account.id}`)
      status = account ? 200 : 404; data = account ? { error: false, account } : { error: true, message: 'Account not found' }
    }
    else if (/^\/api\/v1\/accounts\/[^/]+\/(holdings|activities)(\/[^/]+)?$/.test(endpoint)) {
      const [, accountId, collection, recordId] = endpoint.match(/^\/api\/v1\/accounts\/([^/]+)\/(holdings|activities)(?:\/([^/]+))?$/)
      const records = state[collection]
      if (options.method === 'GET') data = { error: false, [collection]: records.filter(item => !item.account_id || item.account_id === accountId), summary: { complete: true, total_value: 0, warnings: [] } }
      else {
        if (state.mutationGate) await state.mutationGate
        if (state.mutationFailure) { status = 400; data = { error: true, message: 'Transaction rejected' } }
        else {
          if (options.method === 'POST') records.push({ ...body, id: `${collection}-new`, account_id: accountId, name: body.symbol, ledger: true, price_available: false })
          else if (options.method === 'PUT') Object.assign(records.find(item => item.id === recordId), body)
          else if (options.method === 'DELETE') state[collection] = records.filter(item => item.id !== recordId)
          else throw new Error(`Unexpected method: ${options.method}`)
          data = { error: false }
        }
      }
    }
    else if (endpoint === '/api/v1/market/search') data = { results: [{ Code: 'QQQM', Exchange: 'US', Name: 'Invesco NASDAQ 100 ETF', Currency: 'USD' }] }
    else if (endpoint === '/api/v1/market/live/QQQM.US') data = { data: { close: 284.34, change_p: -3.46 } }
    else if (endpoint === '/api/v1/transfers' && options.method === 'POST') {
      if (state.mutationGate) await state.mutationGate
      status = state.mutationFailure ? 400 : 200
      data = state.mutationFailure ? { error: true, message: 'Transaction rejected' } : { error: false }
    }
    else throw new Error(`Unexpected request: ${endpoint}`)
    return new global.Response(JSON.stringify(data), { status, headers: { 'Content-Type': 'application/json' } })
  }
  const protectedPaints = []
  const observer = new w.MutationObserver(records => {
    for (const record of records) for (const node of record.addedNodes) {
      if (node.nodeType === 1 && (node.matches('[data-testid="app-layout"], [data-testid="app-header"], [data-testid="dashboard-page"]') || node.querySelector('[data-testid="app-layout"], [data-testid="app-header"], [data-testid="dashboard-page"]'))) protectedPaints.push(node.tagName)
    }
  })
  observer.observe(w.document.getElementById('app'), { childList: true, subtree: true })
  t.after(() => observer.disconnect())
  for (const [, src] of html.matchAll(/<script[^>]+src="([^"]+)"/g)) w.eval(fs.readFileSync(path.join(build, src.replace(/^\//, '')), 'utf8'))
  const find = id => w.document.querySelector(`[data-testid="${id}"]`)
  async function until(check) {
    for (let i = 0; i < 80; i++) { if (check()) return; await pause() }
    assert.ok(check(), `UI did not settle: ${w.document.body.textContent}; errors: ${errors}`)
  }
  if (onBoot) await onBoot({ w, find, calls, until, protectedPaints })
  await until(() => find('login-page') || find('app-layout') || find('auth-callback-error'))
  return { w, find, state, calls, errors, until, protectedPaints }
}


module.exports = { launch, pause }
