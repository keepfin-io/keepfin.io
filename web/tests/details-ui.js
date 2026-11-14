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
    const cards = [...ui.w.document.querySelectorAll('[data-testid^="account-summary-"]')].map(card => card.dataset.testid).filter(id => !/-(detail|caption)$/.test(id))
    assert.deepEqual(cards, ['account-summary-totalValue', 'account-summary-cashBalance', 'account-summary-equity', 'account-summary-profitLoss'])
    assert.match(ui.find('account-summary-totalValue').textContent, /TRY\s*1,000.00/)
    await ui.until(() => ui.find('holdings-pagination'))
    assert.equal(ui.find('holdings-page-number').textContent, '1 / 1')
    assert.equal(ui.find('holdings-next').disabled, true)
    assert.ok(ui.find('account-tags-toggle'))
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

test('account tags are created with a color, saved at once and shown on both account pages', async t => {
  const retirement = { id: 'tag-1', name: 'Retirement', color: 'green' }
  const ui = await launch(t, { ...defaults, route: `/accounts/${account.id}`, tagsResponse: [retirement] })
  await ui.until(() => ui.find('account-tags-toggle') && ui.find(`holding-row-${holding.id}`))
  assert.equal(ui.find('account-detail-tags'), null)
  ui.find('account-tags-toggle').click()
  await ui.until(() => ui.find('account-tags-option-tag-1'))
  ui.find('account-tags-option-tag-1').click()
  await ui.until(() => ui.find('account-detail-tags')?.textContent.includes('Retirement'))
  const search = ui.find('account-tags-search')
  search.value = 'Short term'
  search.dispatchEvent(new ui.w.Event('input', { bubbles: true }))
  await ui.until(() => ui.find('account-tags-create'))
  ui.find('account-tags-create').click()
  await ui.until(() => ui.find('account-tags-manager-name')?.value === 'Short term')
  assert.equal(ui.w.location.pathname, `/accounts/${account.id}`, 'tags are created without leaving the account')
  ui.find('account-tags-manager-color-violet').click()
  ui.find('account-tags-manager-add').click()
  await ui.until(() => ui.find('account-detail-tags')?.textContent.includes('Short term') && !ui.find('account-tags-manager-dialog'))
  assert.deepEqual(ui.calls.find(call => call.method === 'POST' && call.endpoint === '/api/v1/tags').body, { name: 'Short term', color: 'violet' })
  assert.deepEqual(ui.calls.filter(call => call.method === 'PUT' && call.endpoint === `/api/v1/tags/account/${account.id}`).at(-1).body, { tag_ids: ['tag-1', 'tag-2'] })
  assert.match(ui.find('tag-badge-tag-2').className, /violet/)
  assert.match(ui.find('tag-badge-tag-1').className, /green/)
  // Managing tags opens a dialog on the same page; edits update the header at once.
  ui.find('account-tags-toggle').click()
  await ui.until(() => ui.find('account-tags-manage'))
  ui.find('account-tags-manage').click()
  await ui.until(() => ui.find('account-tags-manager-tag-1'))
  assert.equal(ui.w.location.pathname, `/accounts/${account.id}`)
  ui.find('account-tags-manager-tag-1-color-red').click()
  await pause()
  ;[...ui.find('account-tags-manager-tag-1').querySelectorAll('button')].find(button => button.textContent.trim() === 'Save').click()
  await ui.until(() => /red/.test(ui.find('tag-badge-tag-1')?.className || ''))
  assert.deepEqual(ui.calls.find(call => call.method === 'PUT' && call.endpoint === '/api/v1/tags/tag-1').body, { name: 'Retirement', color: 'red' })

  const extra = [{ id: 'tag-2', name: 'Savings', color: 'blue' }, { id: 'tag-3', name: 'Travel', color: 'pink' }, { id: 'tag-4', name: 'Zeta', color: 'gray' }]
  const list = await launch(t, { ...defaults, route: '/accounts', width: 1280, tagsResponse: [retirement, ...extra], tagAssignments: { account: { [account.id]: ['tag-1', 'tag-2', 'tag-3', 'tag-4'] } } })
  await list.until(() => list.find(`account-row-${account.id}-tags`))
  assert.equal(list.find(`account-row-${account.id}-tags`).querySelectorAll('[data-testid^="tag-badge-"]').length, 4, 'rows show every tag')
  assert.equal(list.find('accounts-active-count').textContent.trim(), '1 active')
  assert.match(list.find('accounts-cash-totals').textContent, /TRY\s*1,000\.00/)
  assert.match(list.find(`account-row-${account.id}`).textContent, /General · .*TRY/)
  const rowLink = list.find(`account-row-${account.id}-name`)
  assert.equal(rowLink.getAttribute('href'), `/accounts/${account.id}`)
  assert.match(rowLink.className, /after:inset-0/, 'the account link covers the whole row')
  const details = list.find(`account-row-${account.id}-actions-view`)
  assert.equal(details.textContent.trim(), 'Details', 'the detail link is labelled instead of an eye icon')
  assert.equal(list.find(`account-row-${account.id}-actions-edit`).textContent.trim(), 'Edit')
  assert.equal(list.find(`account-row-${account.id}-actions-delete`).textContent.trim(), 'Delete')
  details.click()
  await list.until(() => list.w.location.pathname === `/accounts/${account.id}`)
  assert.deepEqual(ui.errors, []); assert.deepEqual(list.errors, [])
})

test('settings create tags with a color and recolor existing ones', async t => {
  const ui = await launch(t, { authenticated: true, route: '/settings', tagsResponse: [{ id: 'tag-1', name: 'Retirement', color: 'gray' }] })
  await ui.until(() => ui.find('settings-tag-tag-1'))
  const name = ui.find('settings-tag-name')
  name.value = 'Crypto'
  name.dispatchEvent(new ui.w.Event('input', { bubbles: true }))
  await pause()
  ui.find('settings-tag-color-amber').click()
  ui.find('settings-tag-add').click()
  await ui.until(() => ui.calls.some(call => call.method === 'POST' && call.endpoint === '/api/v1/tags'))
  assert.deepEqual(ui.calls.find(call => call.method === 'POST' && call.endpoint === '/api/v1/tags').body, { name: 'Crypto', color: 'amber' })
  await ui.until(() => ui.find('settings-tag-tag-2'))
  ui.find('settings-tag-tag-1-color-red').click()
  await pause()
  ;[...ui.find('settings-tag-tag-1').querySelectorAll('button')].find(button => button.textContent.trim() === 'Save').click()
  await ui.until(() => ui.calls.some(call => call.method === 'PUT' && call.endpoint === '/api/v1/tags/tag-1'))
  assert.deepEqual(ui.calls.find(call => call.method === 'PUT' && call.endpoint === '/api/v1/tags/tag-1').body, { name: 'Retirement', color: 'red' })
  assert.deepEqual(ui.errors, [])
})

test('hiding values sits beside the account values and theme switches animate from the toggle', async t => {
  const ui = await launch(t, { ...defaults, route: `/accounts/${account.id}`, width: 1280 })
  await ui.until(() => ui.find('account-summary-cashBalance'))
  assert.equal(ui.find('app-header').querySelector('[data-testid="zen-toggle"]'), null, 'the header no longer hides values')
  assert.ok(ui.find('account-summary-totalValue').querySelector('[data-testid="account-zen-toggle"]'))
  assert.match(ui.find('account-summary-cashBalance').textContent, /1,000\.00/)
  ui.find('account-zen-toggle').click()
  await ui.until(() => ui.find('account-summary-cashBalance').textContent.includes('••••'))
  assert.equal(ui.state.preferences.zen, true)
  assert.equal(ui.find('account-zen-toggle').getAttribute('aria-pressed'), 'true')
  ui.find('account-zen-toggle').click()
  await ui.until(() => ui.find('account-summary-cashBalance').textContent.includes('1,000.00'))

  const animations = []
  ui.w.document.startViewTransition = update => { update(); return { ready: Promise.resolve() } }
  ui.w.document.documentElement.animate = (keyframes, options) => animations.push({ keyframes, options })
  ui.find('header-theme-toggle').click()
  await ui.until(() => !ui.w.document.documentElement.classList.contains('dark') && animations.length)
  assert.match(animations[0].keyframes.clipPath[0], /^circle\(0px at /)
  assert.equal(animations[0].options.pseudoElement, '::view-transition-new(root)')
  assert.deepEqual(ui.errors, [])
})

test('gains show a sign, arrow and color, and the page says how fresh prices are', async t => {
  const losing = { ...holding, id: 'loser', symbol: 'LOSS.US', name: 'Losing fund', profit_loss: -40, performance: -8 }
  const ui = await launch(t, { ...defaults, route: `/accounts/${account.id}`, width: 1280, holdingsResponse: [holding, losing] })
  await ui.until(() => ui.find(`holding-row-${holding.id}`) && ui.find('holding-row-loser'))
  const cells = id => [...ui.find(`holding-row-${id}`).querySelectorAll('td')].map(cell => cell.textContent.trim())
  assert.ok(cells(holding.id).includes('+$100.00') && cells(holding.id).includes('+25.00%'))
  assert.ok(cells('loser').includes('-$40.00') && cells('loser').includes('-8.00%'))
  const gain = [...ui.find(`holding-row-${holding.id}`).querySelectorAll('td')].find(cell => cell.textContent.trim() === '+$100.00')
  assert.match(gain.className, /emerald/)
  assert.ok(gain.querySelector('svg'), 'gains carry an arrow, not only a color')
  assert.match([...ui.find('holding-row-loser').querySelectorAll('td')].find(cell => cell.textContent.trim() === '-$40.00').className, /destructive/)

  assert.equal(ui.find('account-prices-as-of').textContent.trim(), 'Prices from the Oct 30, 2026 close · updated daily')
  assert.ok(ui.find('account-last-refreshed'))
  const loads = () => ui.calls.filter(call => call.endpoint === `/api/v1/accounts/${account.id}/holdings`).length
  const pages = () => ui.calls.filter(call => call.endpoint.endsWith('/holdings/search')).length
  const [beforeLoads, beforePages] = [loads(), pages()]
  ui.find('account-refresh').click()
  await ui.until(() => loads() > beforeLoads && pages() > beforePages && !ui.find('account-refresh').disabled)

  ui.find('account-zen-toggle').click()
  await ui.until(() => cells(holding.id).includes('••••'))
  assert.ok(cells(holding.id).includes('+25.00%'), 'percentages stay visible while amounts are hidden')
  assert.deepEqual(ui.errors, [])
})

test('adding a holding starts with a short guide and keeps manual assets as a fallback', async t => {
  const ui = await launch(t, { ...defaults, route: `/accounts/${account.id}/holdings/new`, width: 1280 })
  await ui.until(() => ui.find('holding-add-guide'))
  assert.equal(ui.find('holding-add-guide').children.length, 3)
  assert.match(ui.find('holding-add-guide').textContent, /Find the asset.*Currency is automatic.*Daily prices/)
  assert.ok(ui.find('holding-add-filter-crypto').querySelector('svg'), 'asset type filters carry icons')
  const manual = ui.find('holding-add-manual')
  assert.equal(manual.open, false)
  assert.ok(manual.compareDocumentPosition(ui.find('holding-add-search')) & ui.w.Node.DOCUMENT_POSITION_PRECEDING, 'manual assets come after the search')
  manual.open = true
  manual.dispatchEvent(new ui.w.Event('toggle'))
  await ui.until(() => manual.textContent.includes('You have no manual assets yet.'))
  assert.deepEqual(ui.errors, [])
})

test('deleting a holding or an account asks in a modal dialog', async t => {
  const ui = await launch(t, { ...defaults, route: `/accounts/${account.id}`, width: 1280 })
  const remove = `holding-row-${holding.id}-actions-delete`
  await ui.until(() => ui.find(remove))
  ui.find(remove).click()
  await ui.until(() => ui.find('holding-delete-confirmation'))
  const dialog = ui.find('holding-delete-confirmation')
  assert.equal(dialog.getAttribute('role'), 'alertdialog')
  assert.equal(ui.find('account-detail-page').contains(dialog), false, 'the dialog overlays the page instead of sitting in it')
  assert.match(dialog.textContent, /QQQM\.US/)
  ui.find('holding-delete-confirmation-cancel').click()
  await ui.until(() => !ui.find('holding-delete-confirmation'))
  assert.equal(ui.calls.some(call => call.method === 'DELETE'), false)
  ui.find(remove).click()
  await ui.until(() => ui.find('holding-delete-confirmation-confirm'))
  ui.find('holding-delete-confirmation-confirm').click()
  await ui.until(() => ui.calls.some(call => call.method === 'DELETE' && call.endpoint.endsWith(`/holdings/${holding.id}`)) && !ui.find('holding-delete-confirmation'))

  const list = await launch(t, {
    ...defaults, route: '/accounts', width: 1280,
    wealthHandler: ({ endpoint, method, state }) => {
      if (method !== 'DELETE' || endpoint !== `/api/v1/accounts/${account.id}`) return null
      state.accounts = []
      return { data: null }
    },
  })
  await list.until(() => list.find(`account-row-${account.id}-actions-delete`))
  list.find(`account-row-${account.id}-actions-delete`).click()
  await list.until(() => list.find('account-delete-confirmation'))
  assert.equal(list.find('account-delete-confirmation').getAttribute('role'), 'alertdialog')
  assert.match(list.find('account-delete-confirmation').textContent, /Investments/)
  list.find('account-delete-confirmation-confirm').click()
  await list.until(() => list.calls.some(call => call.method === 'DELETE') && !list.find('account-delete-confirmation') && !list.find(`account-row-${account.id}`))
  assert.deepEqual(ui.errors, []); assert.deepEqual(list.errors, [])
})

test('wealth pages mark the current tab and edit manual assets in dialogs', async t => {
  const assets = [{ id: 'flat', name: 'City flat', currency: 'USD', asset_class: 'Real estate', sector: '', country: '', symbol: 'MANUAL.FLAT', valuations: [] }]
  const ui = await launch(t, {
    authenticated: true, route: '/wealth/assets', width: 1280,
    wealthHandler: ({ endpoint, method, body }) => {
      if (endpoint === '/api/v1/assets/manual' && method === 'GET') return { data: assets }
      if (endpoint === '/api/v1/assets/manual' && method === 'POST') { assets.push({ id: 'gold', ...body, symbol: 'MANUAL.GOLD', valuations: [] }); return { data: assets.at(-1) } }
      if (endpoint === '/api/v1/assets/manual/flat/valuations') { assets[0].valuations = [{ id: 'v1', ...body }]; return { data: null } }
      if (endpoint === '/api/v1/assets/manual/flat' && method === 'DELETE') { assets.splice(0, 1); return { data: null } }
      return null
    },
  })
  await ui.until(() => ui.find('manual-asset-flat'))
  assert.equal(ui.find('wealth-nav-assets').getAttribute('aria-current'), 'page')
  assert.equal(ui.find('wealth-nav-watchlist').getAttribute('aria-current'), null)
  assert.ok(ui.find('manual-assets-steps'), 'the page explains how a manual asset reaches a portfolio')
  assert.match(ui.find('manual-asset-flat').textContent, /No value yet/)

  ui.find('manual-asset-new').click()
  await ui.until(() => ui.find('manual-asset-dialog'))
  assert.equal(ui.find('manual-assets-page').contains(ui.find('manual-asset-dialog')), false, 'the form opens in a dialog')
  const name = ui.w.document.getElementById('manual-name')
  name.value = 'Gold bars'; name.dispatchEvent(new ui.w.Event('input', { bubbles: true }))
  await pause()
  ui.find('manual-asset-save').click()
  await ui.until(() => ui.find('manual-asset-gold') && !ui.find('manual-asset-dialog'))
  assert.equal(ui.calls.find(call => call.method === 'POST' && call.endpoint === '/api/v1/assets/manual').body.name, 'Gold bars')

  ui.find('manual-asset-flat-value').click()
  await ui.until(() => ui.find('manual-valuation-dialog'))
  const price = ui.w.document.getElementById('valuation-price')
  price.value = '250000'; price.dispatchEvent(new ui.w.Event('input', { bubbles: true }))
  await pause()
  ui.find('manual-valuation-save').click()
  await ui.until(() => !ui.find('manual-valuation-dialog') && /250,000/.test(ui.find('manual-asset-flat').textContent))

  ui.find('manual-asset-flat-actions-delete').click()
  await ui.until(() => ui.find('manual-asset-delete-confirmation'))
  assert.equal(ui.find('manual-asset-delete-confirmation').getAttribute('role'), 'alertdialog')
  assert.match(ui.find('manual-asset-delete-confirmation').textContent, /City flat/)
  ui.find('manual-asset-delete-confirmation-confirm').click()
  await ui.until(() => !ui.find('manual-asset-flat') && !ui.find('manual-asset-delete-confirmation'))
  assert.deepEqual(ui.errors, [])
})

test('revoking a share link asks in a modal dialog', async t => {
  const shares = [{ id: 'advisor', name: 'Advisor', recipient_id: null, account_ids: ['investments'], expires_at: '2099-01-01T23:59:59Z', revoked_at: null }]
  const ui = await launch(t, {
    authenticated: true, route: '/wealth/sharing', width: 1280, accountsResponse: [account],
    wealthHandler: ({ endpoint, method }) => {
      if (endpoint === '/api/v1/shares' && method === 'GET') return { data: shares }
      if (endpoint === '/api/v1/shares/advisor' && method === 'DELETE') { shares[0].revoked_at = '2026-11-14T00:00:00Z'; return { data: null } }
      return null
    },
  })
  await ui.until(() => ui.find('share-advisor'))
  assert.match(ui.find('share-advisor').textContent, /Active/)
  const revoke = [...ui.find('share-advisor').querySelectorAll('button')].find(button => /Revoke/.test(button.textContent))
  revoke.click()
  await ui.until(() => ui.find('share-revoke-confirmation'))
  ui.find('share-revoke-confirmation-cancel').click()
  await ui.until(() => !ui.find('share-revoke-confirmation'))
  assert.equal(ui.calls.some(call => call.method === 'DELETE'), false)
  revoke.click()
  await ui.until(() => ui.find('share-revoke-confirmation-confirm'))
  ui.find('share-revoke-confirmation-confirm').click()
  await ui.until(() => /Revoked/.test(ui.find('share-advisor').textContent) && !ui.find('share-revoke-confirmation'))
  assert.deepEqual(ui.errors, [])
})
