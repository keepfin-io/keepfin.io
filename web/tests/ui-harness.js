const { JSDOM, VirtualConsole } = require('jsdom')
const assert = require('node:assert/strict')
const fs = require('node:fs')
const path = require('node:path')
const build = path.resolve(process.env.AUTH_UI_BUILD || path.join(__dirname, '../dist'))
const pause = () => new Promise(resolve => setTimeout(resolve, 25))

async function launch(t, { route = '/login', authenticated = false, outage = false, conflict = false, width = 320, internalTesting = false, google = true, sessionGate, profileGate, onBoot, locale = 'en-US', portfolioResponse, portfolioGate, measuredChartWidth, accountsResponse = [], holdingsResponse = [], activitiesResponse = [], marketHistoryResponse = [], marketHistoryFailure = '', holdingsFailure = false, activitiesFailure = false, userId = 'app-uuid', storage = {}, wealthHandler = null, zen = false } = {}) {
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
  for (const [key, value] of Object.entries(storage)) w.localStorage.setItem(key, value)
  w.localStorage.setItem('locale', locale)
  const user = { id: userId, email: 'test@example.com', theme: 'dark', base_currency: 'USD', locale: 'en-US', created_at: '2026-01-01T00:00:00Z' }
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
  Object.assign(state, { marketResults: [{ Code: 'QQQM', Exchange: 'US', Name: 'Invesco NASDAQ 100 ETF', Currency: 'USD', Type: 'ETF' }], searchHandler: null, quoteHandler: null, marketHistory: marketHistoryResponse, marketHistoryFailure, historyHandler: null, portfolioHandler: null, holdingsFailure, activitiesFailure })
  state.preferences = { zen, max_asset_weight:25,max_account_weight:70,max_currency_weight:80,max_fee_ratio:1,emergency_months:6,emergency_fund:0,monthly_expenses:0,monthly_contribution:0,annual_expenses:0,annual_return:5,withdrawal_rate:4,horizon_years:30,disabled_rules:[] }
  state.wealthHandler = wealthHandler
  w.fetch = async (input, options = {}) => {
    const requestURL = new URL(input.url || String(input), w.location.origin)
    const endpoint = requestURL.pathname
    const body = options.body ? JSON.parse(options.body) : null
    calls.push({ endpoint, method: options.method || 'GET', body, query: Object.fromEntries(requestURL.searchParams) })
    if (state.wealthHandler) {
      const custom = await state.wealthHandler({ endpoint, method:options.method || 'GET', body, query:Object.fromEntries(requestURL.searchParams), state })
      if (custom) return new global.Response(JSON.stringify(custom.data || custom), { status:custom.status || 200, headers:{'Content-Type':'application/json'} })
    }
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
      data = state.portfolioFailure ? { error: true, message: 'Portfolio service unavailable' } : { error: false, portfolio: state.portfolioHandler ? await state.portfolioHandler(Object.fromEntries(requestURL.searchParams)) : state.portfolio }
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
    else if (/^\/api\/v1\/accounts\/[^/]+\/holdings\/search$/.test(endpoint)) {
      const accountId = endpoint.split('/')[4]
      const params = Object.fromEntries(requestURL.searchParams)
      if (state.holdingsSearchHandler) data = await state.holdingsSearchHandler(params)
      else {
        let rows = state.holdings.filter(item => !item.account_id || item.account_id === accountId)
        const field = (row, key) => key === 'currency' ? row.buy_currency : key === 'exchange' ? row.symbol.split('.').at(-1) : row[key]
        rows = rows.filter(row => (!params.q || `${row.symbol} ${row.name}`.toLowerCase().includes(params.q.toLowerCase())) && ['currency', 'exchange', 'asset_class', 'sector', 'country'].every(key => !params[key] || field(row, key) === params[key]))
        const facets = Object.fromEntries(['currency', 'exchange', 'asset_class', 'sector', 'country'].map(key => [key, [...new Set(rows.map(row => field(row, key)).filter(Boolean))]]))
        rows.sort((a, b) => {
          const left = a[params.sort] ?? '', right = b[params.sort] ?? ''
          const order = typeof left === 'number' ? left - right : String(left).localeCompare(String(right))
          return (params.direction === 'desc' ? -order : order) || a.symbol.localeCompare(b.symbol)
        })
        const page = Number(params.page || 1), perPage = Number(params.per_page || 10)
        data = { holdings: rows.slice((page - 1) * perPage, page * perPage), total: rows.length, page, per_page: perPage, pages: Math.max(1, Math.ceil(rows.length / perPage)), source: state.holdingsSearchSource || 'typesense', facets }
      }
      if (state.holdingsSearchFailure) { status = 503; data = { message: 'Holdings temporarily unavailable' } }
    }
    else if (/^\/api\/v1\/accounts\/[^/]+\/(holdings|activities)(\/[^/]+)?$/.test(endpoint)) {
      const [, accountId, collection, recordId] = endpoint.match(/^\/api\/v1\/accounts\/([^/]+)\/(holdings|activities)(?:\/([^/]+))?$/)
      const records = state[collection]
      if (options.method === 'GET') {
        status = state[`${collection}Failure`] ? 503 : 200
        data = status === 503 ? { error: true, message: `${collection} unavailable` } : { error: false, [collection]: records.filter(item => !item.account_id || item.account_id === accountId), summary: { complete: true, total_value: 0, warnings: [] } }
      }
      else {
        if (state.mutationGate) await state.mutationGate
        if (state.mutationFailure) { status = 400; data = { error: true, message: 'Transaction rejected' } }
        else {
          if (options.method === 'POST') records.push({ ...body, id: `${collection}-new`, account_id: accountId, name: body.name || body.symbol, ledger: true, price_available: false })
          else if (options.method === 'PUT') Object.assign(records.find(item => item.id === recordId), body)
          else if (options.method === 'DELETE') state[collection] = records.filter(item => item.id !== recordId)
          else throw new Error(`Unexpected method: ${options.method}`)
          data = { error: false }
        }
      }
    }
    else if (endpoint === '/api/v1/market/search') {
      if (state.marketSearchFailure) { status = 503; data = { error: true, code: state.marketSearchFailure, message: 'Market provider unavailable' } }
      else data = { results: state.searchHandler ? await state.searchHandler(requestURL.searchParams.get('q')) : state.marketResults }
    }
    else if (endpoint.startsWith('/api/v1/market/eod/')) {
      status = state.marketHistoryFailure ? 503 : 200
      data = state.marketHistoryFailure ? { error: true, code: state.marketHistoryFailure, message: 'Market provider unavailable' } : { error: false, data: state.historyHandler ? await state.historyHandler(decodeURIComponent(endpoint.split('/').at(-1)), Object.fromEntries(requestURL.searchParams)) : state.marketHistory }
    }
    else if (endpoint.startsWith('/api/v1/market/live/')) data = { data: state.quoteHandler ? await state.quoteHandler(decodeURIComponent(endpoint.split('/').at(-1))) : { close: 284.34, change_p: -3.46 } }
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
  await until(() => find('login-page') || find('app-layout') || find('auth-callback-error') || find('shared-portfolio-page'))
  return { w, find, state, calls, errors, until, protectedPaints }
}


module.exports = { launch, pause }
