const assert = require('node:assert/strict')
const { test } = require('node:test')
const { launch, pause } = require('./ui-harness')

const account = { id: 'source', account_name: 'Investments', currency: 'USD', is_active: true }
const apple = { Code: 'AAPL', Exchange: 'US', Name: 'Apple Inc.', Currency: 'USD', Type: 'Common Stock' }
const fund = { Code: 'QQQM', Exchange: 'US', Name: 'Invesco NASDAQ 100 ETF', Currency: 'USD', Type: 'ETF' }
const key = 'keepfin:recent-assets:v1:app-uuid'
const options = { route: '/accounts/source/holdings/new', authenticated: true, accountsResponse: [account] }
const input = (ui, id, value) => { const field = ui.find(id); field.value = value; field.dispatchEvent(new ui.w.Event('input', { bubbles: true })) }
const recents = ui => [...ui.w.document.querySelectorAll('[data-testid^="holding-add-recent-"]')].map(item => item.dataset.testid)
const quotes = ui => ui.calls.filter(call => call.endpoint.startsWith('/api/v1/market/live/'))

test('provider daily quota is explained in Turkish without signing the user out', async t => {
  const ui = await launch(t, { ...options, locale: 'tr-TR' })
  await ui.until(() => ui.find('holding-add-search'))
  ui.state.marketSearchFailure = 'market_daily_limit_reached'
  input(ui, 'holding-add-search', 'QQ')
  await ui.until(() => ui.find('holding-add-search-error'))
  assert.match(ui.find('holding-add-search-error').textContent, /günlük kotası doldu/)
  assert.equal(ui.w.location.pathname, options.route)
  assert.equal(ui.find('holding-add-result-QQQM.US'), null)
  ui.state.marketSearchFailure = null
  ui.find('holding-add-search-retry').click()
  await ui.until(() => ui.find('holding-add-result-QQQM.US'))
  assert.equal(ui.find('holding-add-search-error'), null)
  assert.deepEqual(ui.errors, [])
})

for (const width of [320, 1280]) test(`asset results display logos and a visible fallback at simulated ${width}px`, async t => {
  const ui = await launch(t, { ...options, width, locale: 'tr-TR' })
  await ui.until(() => ui.find('holding-add-search'))
  assert.match(ui.find('holding-add-recents').textContent, /Son seçilenler/)
  assert.equal(ui.find('holding-add-submit'), null)
  ui.state.marketResults = [apple, fund]
  input(ui, 'holding-add-search', 'AA')
  await ui.until(() => ui.find('holding-add-result-AAPL.US'))
  const row = ui.find('holding-add-result-AAPL.US')
  assert.equal(row.querySelector('img').src, 'https://eodhd.com/img/logos/US/AAPL.png')
  assert.match(row.textContent, /US · USD/)
  row.querySelector('img').dispatchEvent(new ui.w.Event('error'))
  await ui.until(() => row.querySelector('img')?.src === 'https://eodhd.com/img/logos/US/aapl.png')
  const logo = row.querySelector('img')
  logo.dispatchEvent(new ui.w.Event('load'))
  await ui.until(() => logo.classList.contains('opacity-100'))
  logo.dispatchEvent(new ui.w.Event('error'))
  await ui.until(() => !row.querySelector('img'))
  assert.equal(row.querySelector('[data-testid="asset-icon"]').textContent, 'AA')
  const fundRow = ui.find('holding-add-result-QQQM.US')
  assert.equal(fundRow.querySelector('img').src, 'https://eodhd.com/img/logos/US/QQQM.png')
  fundRow.querySelector('img').dispatchEvent(new ui.w.Event('load'))
  await ui.until(() => fundRow.querySelector('img').classList.contains('opacity-100'))
  fundRow.querySelector('img').dispatchEvent(new ui.w.Event('error'))
  await ui.until(() => fundRow.querySelector('img')?.src === 'https://eodhd.com/img/logos/US/qqqm.png')
  fundRow.querySelector('img').dispatchEvent(new ui.w.Event('error'))
  await ui.until(() => !fundRow.querySelector('img'))
  assert.ok(fundRow.querySelector('[data-testid="asset-icon"] svg'))
  assert.equal(ui.find('holding-add-recents'), null)
  ui.find('header-theme-toggle').click()
  await ui.until(() => !ui.w.document.documentElement.classList.contains('dark'))
  fundRow.click()
  await ui.until(() => ui.find('holding-add-form'))
  assert.ok(ui.find('holding-add-form').querySelector('[data-testid="asset-icon"]'))
  assert.deepEqual(ui.errors, [])
})

test('recent selections persist, deduplicate by exchange, reorder and fetch a new price when selected', async t => {
  const otherExchange = { ...apple, Exchange: 'LSE', Currency: 'GBP' }
  const ui = await launch(t, { ...options, storage: { [key]: JSON.stringify([apple, otherExchange, fund]) } })
  await ui.until(() => ui.find('holding-add-recent-QQQM.US'))
  ui.find('holding-add-recent-QQQM.US').click()
  await ui.until(() => ui.find('holding-add-price')?.value === '284.34')
  assert.equal(ui.calls.some(call => call.endpoint === '/api/v1/market/search'), false)
  ui.find('holding-add-change-asset').click()
  await ui.until(() => ui.find('holding-add-recent-QQQM.US'))
  assert.deepEqual(recents(ui), ['holding-add-recent-QQQM.US', 'holding-add-recent-AAPL.US', 'holding-add-recent-AAPL.LSE'])
  ui.state.quoteHandler = async () => ({ close: 300, change_p: 1 })
  ui.find('holding-add-recent-QQQM.US').click()
  await ui.until(() => ui.find('holding-add-price')?.value === '300')
  assert.equal(quotes(ui).length, 2)
  assert.equal(JSON.parse(ui.w.localStorage.getItem(key)).length, 3)
  assert.equal(ui.calls.some(call => call.method === 'POST'), false)
  const storage = { [key]: ui.w.localStorage.getItem(key) }
  const reloaded = await launch(t, { ...options, storage })
  await reloaded.until(() => reloaded.find('holding-add-recent-QQQM.US'))
  assert.equal(recents(reloaded)[0], 'holding-add-recent-QQQM.US')
  const differentUser = await launch(t, { ...options, storage, userId: 'other-user' })
  await differentUser.until(() => differentUser.find('holding-add-recents-empty'))
  assert.equal(recents(differentUser).length, 0)
  reloaded.find('holding-add-clear-recents').click()
  await reloaded.until(() => reloaded.find('holding-add-recents-empty'))
  assert.equal(reloaded.w.localStorage.getItem(key), null)
  assert.deepEqual(ui.errors, [])
  assert.deepEqual(reloaded.errors, [])
  assert.deepEqual(differentUser.errors, [])
})

test('recent history rejects invalid entries, caps its size and survives malformed or unavailable storage', async t => {
  const seed = [null, { Code: '../private', Exchange: 'US', Name: 'Invalid' }, apple, apple, ...Array.from({ length: 8 }, (_, i) => ({ ...fund, Code: `FUND${i}` }))]
  const ui = await launch(t, { ...options, storage: { [key]: JSON.stringify(seed) } })
  await ui.until(() => ui.find('holding-add-recent-AAPL.US'))
  assert.equal(recents(ui).length, 6)
  const malformed = await launch(t, { ...options, storage: { [key]: '{invalid' } })
  await malformed.until(() => malformed.find('holding-add-recents-empty'))
  const blocked = await launch(t, { ...options, onBoot: ({ w }) => {
    const get = w.Storage.prototype.getItem, set = w.Storage.prototype.setItem
    w.Storage.prototype.getItem = function (name) { if (name.startsWith('keepfin:recent-assets:')) throw new w.DOMException('Blocked', 'SecurityError'); return get.call(this, name) }
    w.Storage.prototype.setItem = function (name, value) { if (name.startsWith('keepfin:recent-assets:')) throw new w.DOMException('Blocked', 'SecurityError'); return set.call(this, name, value) }
  } })
  await blocked.until(() => blocked.find('holding-add-search'))
  input(blocked, 'holding-add-search', 'QQ')
  await blocked.until(() => blocked.find('holding-add-result-QQQM.US'))
  blocked.find('holding-add-result-QQQM.US').click()
  await blocked.until(() => blocked.find('holding-add-form'))
  blocked.find('holding-add-change-asset').click()
  await blocked.until(() => blocked.find('holding-add-recent-QQQM.US'))
  assert.deepEqual(blocked.errors, [])
  assert.deepEqual(malformed.errors, [])
})

test('late search responses cannot replace newer results or interrupt a recent selection', async t => {
  const ui = await launch(t, { ...options, storage: { [key]: JSON.stringify([fund]) } })
  await ui.until(() => ui.find('holding-add-search'))
  let release
  const pending = new Promise(resolve => { release = resolve }); t.after(() => release([apple]))
  ui.state.searchHandler = query => query === 'AA' ? pending : [fund]
  input(ui, 'holding-add-search', 'AA')
  await ui.until(() => ui.calls.some(call => call.query.q === 'AA'))
  input(ui, 'holding-add-search', 'QQ')
  await ui.until(() => ui.find('holding-add-result-QQQM.US'))
  input(ui, 'holding-add-search', '')
  await ui.until(() => ui.find('holding-add-recent-QQQM.US'))
  ui.find('holding-add-recent-QQQM.US').click()
  release([apple]); await pause()
  assert.equal(ui.find('holding-add-result-AAPL.US'), null)
  assert.match(ui.find('holding-add-form').textContent, /QQQM/)
  assert.deepEqual(ui.errors, [])
})

test('recent selection tolerates quote failure and delayed quotes preserve manually entered prices', async t => {
  const ui = await launch(t, { ...options, storage: { [key]: JSON.stringify([fund, apple]) } })
  await ui.until(() => ui.find('holding-add-recent-QQQM.US'))
  ui.state.quoteHandler = async () => { throw new Error('Quote unavailable') }
  ui.find('holding-add-recent-QQQM.US').click()
  await ui.until(() => ui.find('holding-add-price-error'))
  input(ui, 'holding-add-quantity', '2'); input(ui, 'holding-add-price', '250'); await pause()
  assert.equal(ui.find('holding-add-submit').disabled, false)
  ui.find('holding-add-change-asset').click()
  await ui.until(() => ui.find('holding-add-recent-AAPL.US'))
  let release
  const pending = new Promise(resolve => { release = resolve }); t.after(() => release({ close: 999 }))
  ui.state.quoteHandler = () => pending
  ui.find('holding-add-recent-AAPL.US').click()
  await ui.until(() => ui.find('holding-add-price'))
  assert.equal(ui.find('holding-add-price').value, '')
  input(ui, 'holding-add-price', '200'); await pause(); release({ close: 999 })
  await ui.until(() => !ui.find('holding-add-price-loading'))
  assert.equal(ui.find('holding-add-price').value, '200')
  assert.equal(ui.find('holding-add-price-error'), null)
  assert.deepEqual(ui.errors, [])
})

test('asset type filters and keyboard navigation distinguish exchanges and prioritize exact symbols', async t => {
  const ui = await launch(t, options)
  await ui.until(() => ui.find('holding-add-search'))
  const qqqUS = { ...fund, Code: 'QQQ', Country: 'USA' }
  const qqqMX = { ...qqqUS, Exchange: 'MX', Currency: 'MXN', Country: 'Mexico' }
  const crypto = { Code: 'QQQX-USD', Exchange: 'CC', Name: 'Tokenized ETF', Currency: 'USD', Type: 'Crypto' }
  ui.state.marketResults = [fund, apple, crypto, qqqUS, qqqMX]
  input(ui, 'holding-add-search', 'QQQ')
  await ui.until(() => ui.find('holding-add-result-QQQ.US'))
  const rows = () => [...ui.find('holding-add-search-results').querySelectorAll('[role="option"]')].map(row => row.dataset.testid)
  assert.deepEqual(rows().slice(0, 2), ['holding-add-result-QQQ.US', 'holding-add-result-QQQ.MX'])
  assert.match(ui.find('holding-add-result-QQQ.MX').textContent, /MX · MXN/)
  assert.match(ui.find('holding-add-result-QQQ.MX').textContent, /Mexico/)
  ui.find('holding-add-filter-stock').click()
  await ui.until(() => rows().length === 1)
  assert.deepEqual(rows(), ['holding-add-result-AAPL.US'])
  ui.find('holding-add-filter-crypto').click()
  await ui.until(() => ui.find('holding-add-result-QQQX-USD.CC'))
  assert.equal(rows().length, 1)
  ui.find('holding-add-filter-forex').click()
  await ui.until(() => ui.find('holding-add-search-empty'))
  assert.match(ui.find('holding-add-search-empty').textContent, /No assets of this type/)
  ui.find('holding-add-search-empty').querySelector('button').click()
  await ui.until(() => ui.find('holding-add-result-QQQ.US'))
  assert.equal(ui.calls.filter(call => call.endpoint === '/api/v1/market/search').length, 1)
  const search = ui.find('holding-add-search')
  const press = key => search.dispatchEvent(new ui.w.KeyboardEvent('keydown', { key, bubbles: true }))
  press('ArrowDown'); await pause()
  assert.equal(search.getAttribute('aria-activedescendant'), 'asset-result-0')
  press('ArrowDown'); await pause()
  assert.equal(ui.find('holding-add-result-QQQ.MX').getAttribute('aria-selected'), 'true')
  press('Enter')
  await ui.until(() => ui.find('holding-add-form'))
  assert.equal(quotes(ui).at(-1).endpoint, '/api/v1/market/live/QQQ.MX')
  assert.equal(JSON.parse(ui.w.localStorage.getItem(key))[0].Exchange, 'MX')
  assert.deepEqual(ui.errors, [])
})

test('failed searches can be retried and Escape clears results to show recent assets', async t => {
  const ui = await launch(t, { ...options, storage: { [key]: JSON.stringify([fund]) } })
  await ui.until(() => ui.find('holding-add-search'))
  ui.state.searchHandler = () => { throw new Error('Market unavailable') }
  input(ui, 'holding-add-search', 'QQ')
  await ui.until(() => ui.find('holding-add-search-error'))
  assert.equal(ui.find('holding-add-search-empty'), null)
  ui.state.searchHandler = () => [fund]
  ui.find('holding-add-search-retry').click()
  await ui.until(() => ui.find('holding-add-result-QQQM.US'))
  assert.equal(ui.find('holding-add-search-error'), null)
  ui.find('holding-add-search').dispatchEvent(new ui.w.KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
  await ui.until(() => ui.find('holding-add-recent-QQQM.US'))
  assert.equal(ui.find('holding-add-search-results'), null)
  assert.equal(ui.find('holding-add-search').value, '')
  assert.deepEqual(ui.errors, [])
})
