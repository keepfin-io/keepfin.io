const { test } = require('node:test')
const assert = require('node:assert/strict')
const { launch } = require('./ui-harness')

function report() {
  return {
    currency: 'USD', complete: true, from: '2026-01-01', to: '2026-02-01', warnings: [],
    summary: { value: 125000.25, equity: 100000, cash: 25000.25, net_contributions: 120000, profit_loss: 5000.25, realized: 1200, unrealized: -500, dividends: 120, daily_change: null },
    holdings: [{ symbol: 'ABC.US', currency: 'USD', quantity: 123.12345678, value: 100000, cost: 100500, profit_loss: -500, weight: 79.99, complete: true, accounts: ['Retirement account with a very long name', 'Broker account'], price_date: '2026-02-01' }],
    history: [{ date: '2026-01-01', value: 120000, net_flow: 120000, return: 0, benchmark: 0 }, { date: '2026-01-02', value: null, net_flow: 0, return: null, benchmark: null }, { date: '2026-02-01', value: 125000.25, net_flow: 0, return: 4.17, benchmark: 2 }],
    income: [{ period: '2026-01', dividends: 70, interest: 10, fees: 5, net: 75 }, { period: '2026-02', dividends: 50, interest: 5, fees: 2, net: 53 }],
    allocations: { asset_class: [{ name: 'Equities', value: 100000, weight: 80 }, { name: 'Cash', value: 25000.25, weight: 20 }], currency: [{ name: 'USD', value: 125000.25, weight: 100 }] },
  }
}
function activate(ui, id) {
  const element = ui.find(id)
  assert.ok(element, `missing ${id}`)
  element.dispatchEvent(new ui.w.MouseEvent('mousedown', { bubbles: true, button: 0 }))
  element.click()
}
async function select(ui, id, value) {
  const trigger = ui.find(id)
  trigger.dispatchEvent(new ui.w.KeyboardEvent('keydown', { key: 'Enter', bubbles: true, cancelable: true }))
  await ui.until(() => ui.find(`${id}-option-${value}`))
  ui.find(`${id}-option-${value}`).dispatchEvent(new ui.w.KeyboardEvent('keydown', { key: 'Enter', bubbles: true, cancelable: true }))
}

for (const width of [320, 390, 768, 1024, 1440]) {
  test(`dashboard controls, data views and sidebar behavior at simulated ${width}px`, async t => {
    const ui = await launch(t, { route: '/', authenticated: true, width, portfolioResponse: report(), measuredChartWidth: Math.max(220, width - 340) })
    await ui.until(() => ui.find('portfolio-stat-value'))
    assert.match(ui.find('portfolio-stat-value').textContent, /125,000.25/)
    assert.match(ui.find('portfolio-stat-unrealized').textContent, /-\$500.00/)
    assert.match(ui.find('portfolio-stat-daily_change').textContent, /—/)
    assert.ok(ui.w.document.querySelector('.metric-value.portfolio-negative'))
    const chart = ui.w.document.querySelector('.chart-svg')
    assert.equal(chart.getAttribute('viewBox'), `0 0 ${Math.max(220, width - 340)} 250`)
    activate(ui, 'portfolio-chart-return')
    await ui.until(() => ui.find('portfolio-chart-return').getAttribute('aria-pressed') === 'true')
    assert.equal(chart.querySelectorAll('circle').length, 2, 'isolated observations remain visible across missing data')
    activate(ui, 'portfolio-tab-holdings')
    await ui.until(() => ui.find('portfolio-holdings-mobile'))
    assert.match(ui.find('portfolio-holdings-mobile').textContent, /123.12345678/)
    assert.match(ui.find('portfolio-holdings').textContent, /Retirement account with a very long name/)
    activate(ui, 'portfolio-tab-income')
    await ui.until(() => ui.find('portfolio-income-mobile'))
    assert.match(ui.find('portfolio-income-mobile').textContent, /2026-02/)
    activate(ui, 'portfolio-tab-allocation')
    await ui.until(() => ui.find('portfolio-allocations'))
    assert.match(ui.find('portfolio-allocations').textContent, /80%/)
    // Closing and navigating the mobile drawer must not leave an overlay behind.
    if (width < 768) {
      ui.find('sidebar-toggle').click()
      await ui.until(() => ui.find('sidebar-nav-accounts'))
      assert.ok(ui.w.document.querySelector('[data-mobile="true"]'))
      ui.find('sidebar-nav-accounts').click()
      await ui.until(() => ui.w.location.pathname === '/accounts' && !ui.w.document.querySelector('[data-mobile="true"]'))
    } else {
      assert.ok(ui.find('sidebar-nav-accounts'))
      ui.find('sidebar-toggle').click()
      await ui.until(() => ui.w.document.querySelector('[data-slot="sidebar"][data-state="collapsed"]'))
    }
    assert.deepEqual(ui.errors, [])
  })
}

test('period and benchmark requests preserve the active tab; annual income aggregates correctly', async t => {
  const ui = await launch(t, { route: '/', authenticated: true, width: 1024, portfolioResponse: report() })
  await ui.until(() => ui.find('portfolio-stat-value'))
  activate(ui, 'portfolio-tab-income')
  await ui.until(() => ui.find('portfolio-income'))
  await select(ui, 'portfolio-income-period', 'year')
  await ui.until(() => ui.find('portfolio-income-mobile').textContent.includes('$128.00'))
  await select(ui, 'portfolio-range', '1y')
  await ui.until(() => ui.calls.some(call => call.endpoint === '/api/v1/portfolio' && call.query.range === '1y'))
  await ui.until(() => ui.find('portfolio-income'))
  assert.equal(ui.find('portfolio-tab-income').getAttribute('data-state'), 'active')
  await select(ui, 'portfolio-benchmark', 'SPY.US')
  await ui.until(() => ui.calls.some(call => call.query.benchmark === 'SPY.US'))
  assert.deepEqual(ui.errors, [])
})

test('incomplete valuation stays unknown in totals, holdings and allocation', async t => {
  const data = report(); data.complete = false; data.warnings = ['Missing EUR/USD exchange rate']; data.holdings[0].complete = false
  const ui = await launch(t, { route: '/', authenticated: true, portfolioResponse: data })
  await ui.until(() => ui.find('portfolio-stat-value'))
  assert.match(ui.find('portfolio-stat-value').textContent, /—/)
  assert.doesNotMatch(ui.find('portfolio-stat-value').textContent, /125,000/)
  assert.match(ui.find('portfolio-warnings').textContent, /Missing EUR/)
  activate(ui, 'portfolio-tab-holdings')
  await ui.until(() => ui.find('portfolio-holdings-mobile'))
  assert.doesNotMatch(ui.find('portfolio-holdings-mobile').textContent, /100,000|100,500|500.00/)
  activate(ui, 'portfolio-tab-allocation')
  await ui.until(() => ui.find('portfolio-tab-allocation').getAttribute('data-state') === 'active')
  assert.equal(ui.find('portfolio-allocations'), null)
  assert.deepEqual(ui.errors, [])
})

test('initial loading and API failure allow retry without displaying stale values', async t => {
  let release
  const portfolioGate = new Promise(resolve => { release = resolve })
  t.after(() => release())
  const ui = await launch(t, { route: '/', authenticated: true, portfolioResponse: report(), portfolioGate })
  await ui.until(() => ui.find('portfolio-loading'))
  assert.equal(ui.find('portfolio-stat-value'), null)
  assert.equal(ui.find('portfolio-refresh').disabled, true)
  release()
  await ui.until(() => ui.find('portfolio-stat-value'))
  ui.state.portfolioFailure = true
  ui.find('portfolio-refresh').click()
  await ui.until(() => ui.find('portfolio-error'))
  assert.equal(ui.find('portfolio-stat-value'), null)
  ui.state.portfolioFailure = false
  ui.find('portfolio-refresh').click()
  await ui.until(() => ui.find('portfolio-stat-value'))
  assert.equal(ui.find('portfolio-error'), null)
  assert.deepEqual(ui.errors, [])
})

test('empty Turkish portfolio and long currency amounts render without NaN or missing translations', async t => {
  const data = report(); data.holdings = []; data.income = []; data.allocations = {}; data.history = []; data.summary.value = 987654321012.34
  const ui = await launch(t, { route: '/', authenticated: true, locale: 'tr-TR', width: 320, portfolioResponse: data })
  await ui.until(() => ui.find('portfolio-stat-value'))
  assert.match(ui.find('portfolio-stat-value').textContent, /987\.654\.321\.012,34/)
  activate(ui, 'portfolio-view-allocation')
  await ui.until(() => ui.find('portfolio-tab-allocation').getAttribute('data-state') === 'active')
  assert.match(ui.find('dashboard-page').textContent, /varlık veya nakit ekle/)
  assert.doesNotMatch(ui.find('dashboard-page').textContent, /NaN|undefined|portfolio\./)
  assert.deepEqual(ui.errors, [])
})
