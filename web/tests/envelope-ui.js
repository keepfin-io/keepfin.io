const { test } = require('node:test')
const assert = require('node:assert/strict')
const { launch } = require('./ui-harness')

const report = {
  currency: 'USD', summary: { value: 1990, cash: 1990, equity: 0, profit_loss: -10, daily_change: 0 },
  holdings: [], history: [], income: [], allocations: {}, warnings: [], complete: true,
  from: '2026-10-31', to: '2026-11-07',
}

test('public shared reports use the same envelope without changing Better Auth responses', async t => {
  const ui = await launch(t, {
    route: '/shared#public-token',
    wealthHandler: ({ endpoint, body }) => {
      if (endpoint !== '/api/public/portfolio') return null
      assert.equal(body.token, 'public-token')
      return { data: report }
    },
  })
  await ui.until(() => ui.find('portfolio-stat-value'))
  assert.match(ui.find('portfolio-stat-value').textContent, /1,990/)
  assert.equal(ui.find('dashboard-accounts-link'), null)
  assert.equal(ui.find('portfolio-options-toggle'), null)
  assert.equal(ui.find('portfolio-options'), null)
  assert.deepEqual(ui.errors, [])
})

test('public shared report failures display the envelope error message', async t => {
  const ui = await launch(t, {
    route: '/shared#expired-token',
    wealthHandler: ({ endpoint }) => endpoint === '/api/public/portfolio'
      ? { status: 404, data: { code: 'not_found', message: 'This share has expired' } } : null,
  })
  await ui.until(() => ui.find('portfolio-error'))
  assert.match(ui.find('portfolio-error').textContent, /This share has expired/)
  assert.equal(ui.find('portfolio-stat-value'), null)
  assert.deepEqual(ui.errors, [])
})

test('consolidated activity list reads rows from data and totals from meta.pagination', async t => {
  const ui = await launch(t, {
    route: '/wealth/activities', authenticated: true,
    wealthHandler: ({ endpoint, query }) => {
      if (endpoint !== '/api/v1/activities') return null
      const page = Number(query.page)
      return {
        data: [{ id: `activity-${page}`, account_id: 'account', account_name: 'Test account',
          date: '2026-11-07', type: 'buy', symbol: `ABC${page}.US`, quantity: 2,
          unit_price: 10, currency: 'USD', tags: [], ledger: true }],
        pagination: { page, page_size: 25, total: 26, total_pages: 2 },
      }
    },
  })
  await ui.until(() => ui.w.document.body.textContent.includes('ABC1.US'))
  const next = [...ui.w.document.querySelectorAll('button')].find(button => button.textContent.trim() === 'Next')
  assert.ok(next && !next.disabled)
  next.click()
  await ui.until(() => ui.w.document.body.textContent.includes('ABC2.US'))
  assert.ok(ui.calls.some(call => call.endpoint === '/api/v1/activities' && call.query.page === '2'))
  assert.deepEqual(ui.errors, [])
})

test('wealth lists and planning pages consume direct data payloads', async t => {
  const watchlist = await launch(t, {
    route: '/wealth/watchlist', authenticated: true,
    wealthHandler: ({ endpoint }) => endpoint === '/api/v1/watchlist'
      ? { data: [{ id: 'watch', symbol: 'QQQ.US', name: 'Watchlist fund', currency: 'USD', asset_class: 'ETF' }] } : null,
  })
  await watchlist.until(() => watchlist.w.document.body.textContent.includes('Watchlist fund'))
  const risk = await launch(t, {
    route: '/wealth/risk', authenticated: true,
    wealthHandler: ({ endpoint }) => endpoint === '/api/v1/risk'
      ? { data: { currency: 'USD', rules: [{ id: 'asset', enabled: true, value: 30, unit: 'percent', limit: 25, status: 'warning', subject: 'ABC.US' }] } } : null,
  })
  await risk.until(() => risk.w.document.body.textContent.includes('ABC.US'))
  assert.deepEqual(watchlist.errors, [])
  assert.deepEqual(risk.errors, [])
})
