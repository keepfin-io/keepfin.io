const assert = require('node:assert/strict')
const { test } = require('node:test')
const { launch, pause } = require('./ui-harness')

const account = { id: 'investments', account_name: 'Investments', currency: 'TRY', cash_balance: 1000, is_active: true }
const holding = { id: 'qqqm', account_id: account.id, symbol: 'QQQM.US', name: 'Invesco NASDAQ 100 ETF', buy_currency: 'USD', asset_class: 'ETF', country: 'USA', sector: '', quantity: 2, current_price: 250, value: 500, profit_loss: 100, performance: 25, price_available: true, price_date: '2026-10-30', first_activity: '2026-01-01T00:00:00Z', can_edit_position: true, buy_price: 200, entry_price: 200 }
const activity = { id: 'purchase', account_id: account.id, holding_id: holding.id, symbol: holding.symbol, ledger: true, type: 'buy', date: '2026-01-01T00:00:00Z', quantity: 2, unit_price: 200, fee: 1, currency: 'USD', notes: 'Initial purchase' }
const history = [{ date: '2026-10-29', close: 240 }, { date: '2026-10-28', close: null }, { date: '2026-10-30', close: 250 }]
const report = {
  currency: 'TRY', complete: true, from: '2026-01-01', to: '2026-10-30', warnings: [],
  summary: { value: 12500, cash: 2500, equity: 10000, cost: 8500, net_contributions: 11000, profit_loss: 1500, realized: 100, unrealized: 1400, dividends: 70, interest: 30, fees: 5, daily_change: null },
  history: [{ date: '2026-10-29', value: 12000, return: 0, net_flow: 0, benchmark: 0 }, { date: '2026-10-30', value: 12500, return: 4.17, net_flow: 0, benchmark: 2 }],
  holdings: [], income: [{ period: '2026-10', dividends: 70, interest: 30, fees: 5, net: 95 }], allocations: { asset_class: [{ name: 'ETF', value: 10000, weight: 80 }] },
}
const defaults = { authenticated: true, accountsResponse: [account], holdingsResponse: [holding], activitiesResponse: [activity], marketHistoryResponse: history }
const detailRoute = `/accounts/${account.id}/holdings/${holding.id}`
const analysisRoute = `/accounts/${account.id}/analysis`
const portfolioCalls = ui => ui.calls.filter(call => call.endpoint === '/api/v1/portfolio')
const historyCalls = ui => ui.calls.filter(call => call.endpoint.includes('/market/eod/'))

async function select(ui, id, value) {
  ui.find(id).dispatchEvent(new ui.w.KeyboardEvent('keydown', { key: 'Enter', bubbles: true, cancelable: true }))
  await ui.until(() => ui.find(`${id}-option-${value}`))
  ui.find(`${id}-option-${value}`).dispatchEvent(new ui.w.KeyboardEvent('keydown', { key: 'Enter', bubbles: true, cancelable: true }))
}
function tab(ui, id) {
  ui.find(id).dispatchEvent(new ui.w.MouseEvent('mousedown', { bubbles: true, button: 0 }))
  ui.find(id).click()
}

for (const width of [320, 1280]) {
  test(`account analysis and asset details are linked pages at simulated ${width}px`, async t => {
    const ui = await launch(t, { ...defaults, route: `/accounts/${account.id}`, width, portfolioResponse: report, measuredChartWidth: width === 320 ? 220 : 800 })
    await ui.until(() => ui.find('account-analysis-link') && ui.find(`holding-row-${holding.id}`))
    const cards = [...ui.w.document.querySelectorAll('[data-testid^="account-summary-"]')].map(card => card.dataset.testid)
    assert.deepEqual(cards, ['account-summary-totalValue', 'account-summary-cashBalance', 'account-summary-equity', 'account-summary-profitLoss'])
    assert.match(ui.find('account-summary-totalValue').textContent, /TRY\s*1,000.00/)
    await ui.until(() => ui.find('holdings-pagination'))
    assert.equal(ui.find('holdings-page-number').textContent, '1 / 1')
    assert.equal(ui.find('holdings-next').disabled, true)
    assert.equal(ui.find('entity-tags'), null)
    ui.find('account-tags-toggle').click()
    await ui.until(() => ui.find('entity-tags'))
    ui.find('account-analysis-link').click()
    await ui.until(() => ui.find('account-analysis-report') && ui.find('portfolio-stat-value'))
    assert.equal(ui.w.location.pathname, analysisRoute)
    assert.equal(ui.find('dashboard-page'), null)
    assert.equal(ui.find('dashboard-accounts-link'), null)
    assert.match(ui.find('account-analysis-report').textContent, /Investments/)
    assert.match(ui.find('portfolio-stat-value').textContent, /TRY\s*12,500.00/)
    ui.find('portfolio-summary-toggle').click()
    for (const [key, amount] of [['cost', '8,500.00'], ['interest', '30.00'], ['fees', '5.00']]) assert.ok(ui.find(`portfolio-stat-${key}`).textContent.includes(amount))
    tab(ui, 'portfolio-tab-income')
    await ui.until(() => ui.find('portfolio-income'))
    await select(ui, 'portfolio-range', '1y')
    await ui.until(() => portfolioCalls(ui).some(call => call.query.range === '1y'))
    ui.find('portfolio-options-toggle').click()
    await ui.until(() => ui.find('portfolio-options-toggle').getAttribute('aria-expanded') === 'true')
    assert.equal(ui.find('portfolio-tag'), null)
    await select(ui, 'portfolio-benchmark', 'QQQ.US')
    await ui.until(() => portfolioCalls(ui).some(call => call.query.benchmark === 'QQQ.US'))
    assert.ok(portfolioCalls(ui).every(call => call.query.account_id === account.id), 'every report request stays scoped to this account')
    assert.equal(ui.find('portfolio-tab-income').getAttribute('data-state'), 'active')
    ui.find('account-analysis-back').click()
    const viewId = `holding-${width < 768 ? 'card' : 'row'}-${holding.id}-link`
    await ui.until(() => ui.find(viewId))
    ui.find(viewId).click()
    await ui.until(() => ui.find('holding-detail-value') && !ui.find('holding-history-loading'))
    assert.equal(ui.w.location.pathname, detailRoute)
    assert.equal(ui.w.document.querySelector('[role="dialog"], [role="alertdialog"]'), null)
    assert.ok(ui.find('app-footer'))
    assert.match(ui.find('holding-detail-identity').textContent, /QQQM.*USD.*Invesco/)
    assert.match(ui.find('holding-detail-value').textContent, /\$500.00/)
    assert.match(ui.find('holding-detail-profit-loss').textContent, /\$100.00/)
    assert.equal(ui.find('holding-detail-performance').textContent, '25%')
    const chart = ui.find('holding-price-history').querySelector('svg[role="img"]')
    assert.match(chart.getAttribute('aria-label'), /closing price history/)
    assert.equal(chart.getAttribute('viewBox'), `0 0 ${width === 320 ? 220 : 800} 250`)
    assert.equal(ui.find('holding-price-history').querySelectorAll('table thead th').length, 2, 'price history never displays portfolio return or net flows')
    assert.equal(ui.find('holding-price-history').querySelector('tbody tr td').textContent, '2026-10-28', 'history is date sorted')
    assert.match(ui.find('holding-price-history').textContent, /—/)
    assert.ok(ui.find('holding-activity-purchase'))
    ui.find('holding-detail-back').click()
    await ui.until(() => ui.find(viewId))
    assert.deepEqual(ui.errors, [])
  })
}

test('direct account analysis preserves scope after failure and retry, and never invents unavailable totals', async t => {
  const ui = await launch(t, { ...defaults, route: analysisRoute, portfolioResponse: { ...report, complete: false, warnings: ['Missing price'] }, locale: 'tr-TR' })
  await ui.until(() => ui.find('portfolio-stat-value'))
  assert.match(ui.find('account-analysis-page').textContent, /Hesap analizi/)
  for (const key of ['value', 'profit_loss', 'cost', 'interest', 'fees']) assert.match(ui.find(`portfolio-stat-${key}`).textContent, /—/)
  ui.state.portfolioFailure = true
  ui.find('portfolio-refresh').click()
  await ui.until(() => ui.find('portfolio-error'))
  assert.equal(ui.find('portfolio-stat-value'), null)
  ui.state.portfolioFailure = false
  ui.state.portfolio = report
  ui.find('portfolio-refresh').click()
  await ui.until(() => ui.find('portfolio-stat-value'))
  assert.ok(portfolioCalls(ui).every(call => call.query.account_id === account.id))
  assert.deepEqual(ui.errors, [])
})

test('price provider failure leaves position and transactions usable and can be retried', async t => {
  const ui = await launch(t, { ...defaults, route: detailRoute, locale: 'tr-TR', marketHistoryFailure: 'market_daily_limit_reached', holdingsResponse: [{ ...holding, price_available: false }] })
  await ui.until(() => ui.find('holding-history-error'))
  assert.match(ui.find('holding-history-error').textContent, /günlük kotası/)
  assert.match(ui.find('holding-history-error').textContent, /fiyat geçmişi/)
  assert.ok(ui.find('holding-detail-price-unavailable'))
  for (const key of ['value', 'price', 'profit-loss', 'performance']) assert.match(ui.find(`holding-detail-${key}`).textContent, /—/)
  assert.match(ui.find('holding-detail-quantity').textContent, /2/)
  assert.ok(ui.find('holding-activity-purchase'))
  ui.state.marketHistoryFailure = ''
  ui.find('holding-history-retry').click()
  await ui.until(() => ui.find('holding-price-history').querySelector('svg[role="img"]'))
  assert.equal(ui.find('holding-history-error'), null)
  assert.equal(ui.find('holding-detail-performance').textContent, '—', 'a chart response must not fabricate a position valuation')
  assert.deepEqual(ui.errors, [])
})

test('holding history is scoped, sorted, paginated and separates archived records', async t => {
  const activities = Array.from({ length: 12 }, (_, index) => ({ ...activity, id: `purchase-${index}`, date: `2026-10-${String(index + 1).padStart(2, '0')}` }))
  activities.push({ ...activity, id: 'dividend', type: 'dividend', date: '2026-10-31', quantity: 1, unit_price: 10, ledger: false })
  activities.push({ ...activity, id: 'other-symbol', holding_id: 'other', symbol: 'AAPL.US' }, { ...activity, id: 'other-account', account_id: 'private' })
  const ui = await launch(t, { ...defaults, route: detailRoute, activitiesResponse: activities })
  await ui.until(() => ui.find('holding-activity-dividend'))
  assert.equal(ui.find('holding-activity-other-symbol'), null)
  assert.equal(ui.find('holding-activity-other-account'), null)
  assert.equal(ui.find('holding-activity-history').querySelectorAll('li').length, 10)
  assert.match(ui.find('holding-activity-dividend').textContent, /Earlier history/)
  assert.equal(ui.find('holding-activity-edit-dividend'), null)
  ui.find('holding-activities-next').click()
  await ui.until(() => ui.find('holding-activity-history').querySelectorAll('li').length === 3)
  await select(ui, 'holding-activity-type', 'dividend')
  await ui.until(() => ui.find('holding-activity-history').querySelectorAll('li').length === 1)
  assert.equal(ui.find('holding-activities-prev').disabled, true)
  ui.find('holding-all-activities').click()
  await ui.until(() => ui.find('activities-panel'))
  assert.equal(ui.w.location.search, '?tab=activities')
  assert.deepEqual(ui.errors, [])
})

test('missing holding and read failures are distinct, retryable states', async t => {
  const ui = await launch(t, { ...defaults, route: detailRoute, holdingsFailure: true, activitiesFailure: true })
  await ui.until(() => ui.find('holding-detail-error'))
  assert.equal(ui.find('holding-detail-not-found'), null)
  assert.equal(historyCalls(ui).length, 0)
  ui.state.holdingsFailure = false
  ui.find('holding-detail-retry').click()
  await ui.until(() => ui.find('holding-activities-error'))
  assert.equal(ui.find('holding-activities-empty'), null)
  ui.state.activitiesFailure = false
  ui.find('holding-activities-retry').click()
  await ui.until(() => ui.find('holding-activity-purchase'))
  const missing = await launch(t, { ...defaults, route: '/accounts/investments/holdings/missing' })
  await missing.until(() => missing.find('holding-detail-not-found'))
  assert.equal(historyCalls(missing).length, 0)
  assert.deepEqual(ui.errors, [])
  assert.deepEqual(missing.errors, [])
})

test('late price-history responses cannot replace the selected range', async t => {
  const ui = await launch(t, { ...defaults, route: detailRoute })
  await ui.until(() => ui.find('holding-price-history').querySelector('svg[role="img"]'))
  let release
  const pending = new Promise(resolve => { release = resolve })
  t.after(() => release())
  let requests = 0
  ui.state.historyHandler = async () => ++requests === 1 ? pending : [{ date: '2026-10-30', close: 321 }]
  ui.find('holding-history-refresh').click()
  await ui.until(() => ui.find('holding-history-loading'))
  await select(ui, 'holding-history-range', '1y')
  await ui.until(() => !ui.find('holding-history-loading'))
  assert.match(ui.find('holding-price-history').textContent, /321/)
  release([{ date: '2026-10-30', close: 999 }])
  await pause()
  assert.doesNotMatch(ui.find('holding-price-history').textContent, /999/)
  assert.equal(historyCalls(ui).at(-1).query.period, 'd')
  assert.equal(historyCalls(ui).at(-1).query.order, 'a')
  assert.deepEqual(ui.errors, [])
})

test('new detail routes require a session and do not expose account data to guests', async t => {
  for (const route of [analysisRoute, detailRoute]) {
    const ui = await launch(t, { ...defaults, route, authenticated: false })
    await ui.until(() => ui.find('login-page'))
    assert.equal(ui.calls.some(call => call.endpoint.startsWith('/api/v1/accounts') || call.endpoint === '/api/v1/portfolio' || call.endpoint.includes('/market/eod/')), false)
    assert.deepEqual(ui.errors, [])
  }
})

test('late account reports cannot overwrite a newer selected period', async t => {
  const ui = await launch(t, { ...defaults, route: analysisRoute, portfolioResponse: report })
  await ui.until(() => ui.find('portfolio-stat-value'))
  let release
  const pending = new Promise(resolve => { release = resolve })
  t.after(() => release())
  let requests = 0
  ui.state.portfolioHandler = async () => ++requests === 1 ? pending : { ...report, summary: { ...report.summary, value: 54321 } }
  ui.find('portfolio-refresh').click()
  await ui.until(() => requests === 1)
  await select(ui, 'portfolio-range', '3m')
  await ui.until(() => ui.find('portfolio-stat-value')?.textContent.includes('54,321.00'))
  release({ ...report, summary: { ...report.summary, value: 99999 } })
  await pause()
  assert.match(ui.find('portfolio-stat-value').textContent, /54,321.00/)
  assert.ok(portfolioCalls(ui).every(call => call.query.account_id === account.id))
  assert.deepEqual(ui.errors, [])
})

test('inactive accounts remain readable and a missing account cannot fall back to the global portfolio', async t => {
  const ui = await launch(t, { ...defaults, route: detailRoute, accountsResponse: [{ ...account, is_active: false }] })
  await ui.until(() => ui.find('holding-detail-value'))
  assert.equal(ui.find('holding-detail-edit'), null)
  assert.equal(ui.find('holding-activity-edit-purchase'), null)
  assert.ok(ui.find('holding-activity-purchase'))
  const missing = await launch(t, { ...defaults, route: '/accounts/missing/analysis' })
  await missing.until(() => missing.find('account-detail-error'))
  assert.equal(portfolioCalls(missing).length, 0)
  assert.equal(missing.find('account-analysis-page'), null)
  assert.deepEqual(ui.errors, [])
  assert.deepEqual(missing.errors, [])
})

test('account header names the account, its institution logo and currency', async t => {
  const ui = await launch(t, { ...defaults, route: `/accounts/${account.id}`, accountsResponse: [{ ...account, platform: 'Midas' }] })
  await ui.until(() => ui.find('account-detail-platform'))
  assert.equal(ui.find('account-detail-name').textContent.trim(), 'Investments')
  assert.equal(ui.find('account-detail-platform').textContent.trim(), 'Midas')
  assert.equal(ui.find('account-detail-currency').textContent.trim(), 'TRY')
  assert.match(ui.find('account-detail-logo').querySelector('img').src, /domain=getmidas\.com/)
  const general = await launch(t, { ...defaults, route: `/accounts/${account.id}` })
  await general.until(() => general.find('account-detail-platform'))
  assert.equal(general.find('account-detail-logo').querySelector('img'), null, 'accounts without an institution show a generic icon')
  assert.deepEqual(ui.errors, []); assert.deepEqual(general.errors, [])
})
