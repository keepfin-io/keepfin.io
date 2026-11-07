const assert = require('node:assert/strict')
const { test } = require('node:test')
const { launch, pause } = require('./ui-harness')

const account = { id: 'source', account_name: 'Investments', currency: 'USD', cash_balance: 1000, is_active: true }
const holdings = Array.from({ length: 23 }, (_, i) => ({ id: `h${i}`, account_id: account.id, symbol: `ASSET${String(i).padStart(2, '0')}.US`, name: `Fund ${i}`, quantity: i + 1, buy_currency: 'USD', asset_class: i < 10 ? 'ETF' : 'Stock', price_available: true, current_price: 10, value: 10 * (i + 1), profit_loss: -0.00001, performance: -0.00001 }))
function input(ui, value) { const el = ui.find('holdings-search'); el.value = value; el.dispatchEvent(new ui.w.Event('input', { bubbles: true })) }
async function select(ui, id, value) {
  await ui.until(() => ui.find(id))
  ui.find(id).dispatchEvent(new ui.w.KeyboardEvent('keydown', { key: 'Enter', bubbles: true, cancelable: true }))
  await ui.until(() => ui.find(`${id}-option-${value}`))
  ui.find(`${id}-option-${value}`).dispatchEvent(new ui.w.KeyboardEvent('keydown', { key: 'Enter', bubbles: true, cancelable: true }))
}
const listRequests = ui => ui.calls.filter(call => call.endpoint.endsWith('/holdings/search'))

for (const width of [320, 1280]) test(`holdings pagination, filters and summaries at simulated ${width}px`, async t => {
  const ui = await launch(t, { route: '/accounts/source', authenticated: true, width, accountsResponse: [account], holdingsResponse: holdings })
  await ui.until(() => ui.find('holding-row-h0'))
  assert.equal(ui.w.document.querySelectorAll('tr[data-testid^="holding-row-"]').length, 10)
  assert.equal(ui.find('holding-row-h10'), null)
  assert.equal(ui.find('holdings-page-number').textContent, '1 / 3')
  assert.doesNotMatch(ui.find('holding-row-h0').textContent, /-\$0|NaN|-0\.00/)
  const summary = ui.find('account-summary-equity').textContent
  ui.find('holdings-next').click()
  await ui.until(() => ui.find('holding-row-h10'))
  assert.equal(listRequests(ui).at(-1).query.page, '2')
  assert.equal(ui.find('holding-row-h0'), null)
  assert.equal(ui.find('account-summary-equity').textContent, summary)
  ui.find('holdings-filter-toggle').click()
  await select(ui, 'holdings-filter-asset_class', 'ETF')
  await ui.until(() => ui.find('holding-row-h0'))
  assert.equal(listRequests(ui).at(-1).query.page, '1')
  assert.equal(ui.find('holdings-next').disabled, true)
  ui.find('holdings-reset').click()
  await ui.until(() => ui.find('holdings-page-number').textContent === '1 / 3' && !ui.find('holdings-next').disabled)
  await select(ui, 'holdings-page-size', '25')
  await ui.until(() => ui.find('holding-row-h22'))
  assert.equal(ui.find('holdings-page-number').textContent, '1 / 1')
  assert.ok(ui.find('holding-card-h22'))
  input(ui, 'not-found')
  await ui.until(() => ui.find('holdings-no-results'))
  ui.find('holdings-reset').click()
  await ui.until(() => ui.find('holding-row-h0'))
  assert.deepEqual(ui.errors, [])
})

test('holdings ignore stale searches, show fallback and recover from request failure', async t => {
  const ui = await launch(t, { route: '/accounts/source', authenticated: true, accountsResponse: [account], holdingsResponse: holdings })
  await ui.until(() => ui.find('holding-row-h0'))
  let release
  const delayed = new Promise(resolve => { release = resolve }); t.after(() => release())
  const result = rows => ({ holdings: rows, total: rows.length, page: 1, pages: 1, per_page: 10, source: 'database', facets: {} })
  ui.state.holdingsSearchHandler = async params => {
    if (params.q === 'older') { await delayed; return result([holdings[0]]) }
    return result([holdings[22]])
  }
  input(ui, 'older'); await ui.until(() => listRequests(ui).at(-1).query.q === 'older')
  input(ui, 'newer'); await ui.until(() => ui.find('holding-row-h22'))
  release(); await pause(); await pause()
  assert.equal(ui.find('holding-row-h0'), null)
  assert.ok(ui.find('holdings-search-fallback'))
  ui.state.holdingsSearchFailure = true
  input(ui, 'failure'); await ui.until(() => ui.find('holdings-list-error'))
  assert.equal(ui.find('holding-row-h22'), null)
  ui.state.holdingsSearchFailure = false
  ui.find('holdings-list-retry').click()
  await ui.until(() => ui.find('holding-row-h22'))
  assert.deepEqual(ui.errors, [])
})
