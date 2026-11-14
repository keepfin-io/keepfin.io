const assert = require('node:assert/strict')
const { test } = require('node:test')
const { launch, pause } = require('./ui-harness')

function input(ui, id, value) {
  const field = ui.find(id)
  field.value = value
  field.dispatchEvent(new ui.w.Event('input', { bubbles: true }))
}
function submit(ui) {
  ui.find('account-form').dispatchEvent(new ui.w.Event('submit', { bubbles: true, cancelable: true }))
}
async function currency(ui, value) {
  ui.find('account-form-currency').dispatchEvent(new ui.w.KeyboardEvent('keydown', { key: 'Enter', bubbles: true, cancelable: true }))
  await ui.until(() => ui.find(`account-form-currency-option-${value}`))
  ui.find(`account-form-currency-option-${value}`).dispatchEvent(new ui.w.KeyboardEvent('keydown', { key: 'Enter', bubbles: true, cancelable: true }))
  await ui.until(() => ui.find('account-form-currency').textContent.includes(value))
}
const creates = ui => ui.calls.filter(call => call.endpoint === '/api/v1/accounts/create')

for (const width of [320, 1280]) {
  test(`account creation opens as a protected page from both entry points at simulated ${width}px`, async t => {
    const ui = await launch(t, { route: '/accounts', authenticated: true, width, locale: 'tr-TR' })
    await ui.until(() => ui.find('accounts-empty'))
    for (const [entry, exit] of [['accounts-add', 'account-form-cancel'], ['accounts-create-first', 'account-create-back']]) {
      ui.find(entry).click()
      await ui.until(() => ui.find('account-create-page'))
      assert.equal(ui.w.location.pathname, '/accounts/new')
      assert.equal(ui.find('account-form-dialog'), null)
      assert.equal(ui.find('account-form-name').value, '')
      assert.equal(ui.find('account-form-balance').value, '0')
      assert.match(ui.find('app-page-title').textContent, /Hesaplar/)
      assert.ok(ui.find('app-footer'))
      input(ui, 'account-form-name', 'Unsaved account')
      ui.find(exit).click()
      await ui.until(() => ui.find('accounts-empty'))
      assert.equal(ui.w.location.pathname, '/accounts')
      assert.equal(creates(ui).length, 0)
    }
    assert.deepEqual(ui.errors, [])
  })
}

test('direct create-page entry retains field validation and does not request an account named new', async t => {
  const ui = await launch(t, { route: '/accounts/new', authenticated: true })
  await ui.until(() => ui.find('account-create-page'))
  assert.equal(ui.calls.some(call => call.endpoint === '/api/v1/accounts/new'), false)
  submit(ui)
  await ui.until(() => ui.find('account-form-name-error') && ui.find('account-form-currency-error'))
  assert.equal(ui.w.document.activeElement, ui.find('account-form-name'))
  assert.equal(creates(ui).length, 0)
  input(ui, 'account-form-name', 'Savings')
  await currency(ui, 'USD')
  input(ui, 'account-form-balance', '-1')
  await pause()
  submit(ui)
  await ui.until(() => ui.find('account-form-balance-error'))
  assert.equal(creates(ui).length, 0)
  assert.equal(ui.find('account-form-opening-date').required, true)
  assert.deepEqual(ui.errors, [])
})

test('create-page failure preserves all fields; retry saves once and returns to the refreshed list', async t => {
  const ui = await launch(t, { route: '/accounts/new', authenticated: true, width: 1280 })
  await ui.until(() => ui.find('account-create-page'))
  input(ui, 'account-form-name', 'Investment account')
  await currency(ui, 'EUR')
  input(ui, 'account-form-balance', '125.5')
  input(ui, 'account-form-opening-date', '2026-01-15')
  input(ui, 'account-form-description', 'Long-term savings')
  ui.find('account-form-excluded').click()
  ui.find('account-form-platform').click()
  await ui.until(() => ui.find('account-form-platform-option-interactivebrokers.com'))
  ui.find('account-form-platform-option-interactivebrokers.com').click()
  await ui.until(() => ui.find('account-form-platform').textContent.includes('Interactive Brokers'))
  ui.state.accountCreateFailure = true
  submit(ui)
  await ui.until(() => ui.find('account-form-error'))
  assert.equal(ui.find('account-form-name').value, 'Investment account')
  assert.equal(ui.find('account-form-balance').value, '125.5')
  assert.equal(ui.find('account-form-description').value, 'Long-term savings')
  assert.equal(ui.w.location.pathname, '/accounts/new')
  assert.equal(ui.find('account-form-submit').disabled, false)

  let release
  ui.state.accountCreateGate = new Promise(resolve => { release = resolve })
  t.after(() => release())
  ui.state.accountCreateFailure = false
  submit(ui)
  submit(ui)
  await ui.until(() => ui.find('account-form-submit').disabled)
  assert.equal(ui.find('account-form-cancel').disabled, true)
  assert.equal(ui.find('account-create-back').disabled, true)
  assert.equal(creates(ui).length, 2, 'the failed attempt and one pending retry are the only create requests')
  release()
  await ui.until(() => ui.find('account-row-account-1') && ui.w.location.search === '')
  assert.equal(ui.w.location.pathname, '/accounts')
  assert.match(ui.find('accounts-notification').textContent, /Account created/)
  assert.deepEqual(creates(ui)[1].body, {
    account_name: 'Investment account', currency: 'EUR', initial_balance: 125.5,
    opening_date: '2026-01-15T00:00:00Z', platform: 'Interactive Brokers',
    description: 'Long-term savings', is_excluded: true,
  })
  assert.equal(ui.state.accounts.length, 1)
  assert.equal(ui.find('account-create-page'), null)
  assert.deepEqual(ui.errors, [])
})

test('editing an existing account retains its confirmation and update endpoint', async t => {
  const account = { id: 'existing', account_name: 'Savings', currency: 'USD', initial_balance: 10, cash_balance: 10, opening_date: '2026-01-15T00:00:00Z', platform: '', description: '', is_excluded: false, is_active: true }
  const ui = await launch(t, { route: '/accounts', authenticated: true, width: 1280, accountsResponse: [account] })
  await ui.until(() => ui.find('account-row-existing-actions-edit'))
  ui.find('account-row-existing-actions-edit').click()
  await ui.until(() => ui.find('account-edit-page'))
  assert.equal(ui.w.location.pathname, '/accounts/existing/edit')
  assert.equal(ui.w.document.querySelector('[role="dialog"]'), null)
  input(ui, 'account-form-name', 'Updated savings')
  await pause()
  submit(ui)
  await ui.until(() => ui.find('account-update-confirmation'))
  assert.equal(ui.calls.some(call => call.method === 'PUT'), false)
  ui.find('account-update-confirmation-cancel').click()
  await ui.until(() => !ui.find('account-update-confirmation'))
  assert.equal(ui.calls.some(call => call.method === 'PUT'), false)
  submit(ui)
  await ui.until(() => ui.find('account-update-confirmation-confirm'))
  ui.find('account-update-confirmation-confirm').click()
  await ui.until(() => !ui.find('account-edit-page') && ui.find('account-detail-name').textContent.includes('Updated savings'))
  assert.equal(ui.calls.filter(call => call.endpoint === '/api/v1/accounts/existing' && call.method === 'PUT').length, 1)
  assert.equal(creates(ui).length, 0)
  assert.deepEqual(ui.errors, [])
})

const sampleAccount = { id: 'source', account_name: 'Investments', currency: 'USD', initial_balance: 1000, cash_balance: 1000, opening_date: '2026-01-01T00:00:00Z', is_active: true, is_excluded: false }
const sampleHolding = { id: 'holding-1', symbol: 'QQQM.US', name: 'Invesco', quantity: 2, buy_price: 200, entry_price: 195, buy_currency: 'USD', can_edit_position: true, notes: 'Original', price_available: false }
const sampleActivity = { id: 'activity-1', ledger: true, type: 'buy', date: '2026-02-01', symbol: 'QQQM.US', quantity: 2, unit_price: 200, fee: 1, currency: 'USD', notes: 'Original' }
function setField(ui, id, value) { const element = ui.w.document.getElementById(id); element.value = value; element.dispatchEvent(new ui.w.Event('input', { bubbles: true })) }
function submitForm(ui, id) { ui.find(id).dispatchEvent(new ui.w.Event('submit', { bubbles: true, cancelable: true })) }
function noDialog(ui) { assert.equal(ui.w.document.querySelector('[role="dialog"], [role="alertdialog"]'), null) }

for (const width of [320, 1280]) test(`holding search and purchase stay on a page and submit once at simulated ${width}px`, async t => {
  const ui = await launch(t, { route: '/accounts/source', authenticated: true, width, accountsResponse: [sampleAccount] })
  await ui.until(() => ui.find('holdings-add'))
  ui.find('holdings-add').click()
  await ui.until(() => ui.find('holding-add-page'))
  assert.equal(ui.w.location.pathname, '/accounts/source/holdings/new'); noDialog(ui)
  input(ui, 'holding-add-search', 'QQ')
  await ui.until(() => ui.find('holding-add-result-QQQM.US'))
  ui.find('holding-add-result-QQQM.US').click()
  await ui.until(() => ui.find('holding-add-price')?.value === '284.34')
  assert.ok(ui.find('holding-add-currency-locked')); assert.equal(ui.find('holding-add-currency').disabled, true)
  input(ui, 'holding-add-quantity', '2'); input(ui, 'holding-add-fee', '1'); input(ui, 'holding-add-notes', 'New position')
  setField(ui, 'holding-date', '2026-02-02'); setField(ui, 'holding-rate', '1.1')
  await pause()
  let release
  ui.state.mutationGate = new Promise(resolve => { release = resolve }); t.after(() => release())
  submitForm(ui, 'holding-add-form'); submitForm(ui, 'holding-add-form')
  await ui.until(() => ui.find('holding-add-submit').disabled)
  const requests = ui.calls.filter(call => call.method === 'POST' && call.endpoint.endsWith('/holdings'))
  assert.equal(requests.length, 1)
  assert.deepEqual(requests[0].body, { symbol: 'QQQM.US', name: 'Invesco NASDAQ 100 ETF', asset_class: 'ETF', country: '', quantity: 2, buy_price: 284.34, fee: 1, buy_currency: 'USD', notes: 'New position', date: '2026-02-02', exchange_rate: 1.1 })
  release()
  await ui.until(() => ui.find('holding-row-holdings-new'))
  for (const container of ['holding-row-holdings-new', 'holding-card-holdings-new']) {
    const logo = ui.find(container).querySelector('[data-testid="asset-icon"] img')
    assert.equal(logo.src, 'https://eodhd.com/img/logos/US/QQQM.png')
    assert.equal(logo.width, 32)
  }
  assert.equal(ui.w.location.pathname, '/accounts/source'); assert.ok(ui.find('account-detail-notification')); noDialog(ui)
  assert.deepEqual(ui.errors, [])
})

for (const [path, page, cancel] of [['holdings/new', 'holding-add-page', 'account-action-back'], ['activities/new', 'activity-page', 'activity-cancel'], ['transfer', 'transfer-page', 'transfer-cancel'], ['edit', 'account-edit-page', 'account-form-cancel']]) {
  test(`direct account action ${path} loads and cancels without a mutation`, async t => {
    const ui = await launch(t, { route: `/accounts/source/${path}`, authenticated: true, accountsResponse: [sampleAccount] })
    await ui.until(() => ui.find(page)); noDialog(ui); assert.ok(ui.find('app-footer'))
    if (path === 'holdings/new') {
      assert.equal(ui.find('holding-add-cancel'), null, 'back to account is the only way back')
      assert.equal(ui.find('account-detail-header'), null, 'action pages do not repeat the account header')
    }
    ui.find(cancel).click()
    await ui.until(() => ui.w.location.pathname === '/accounts/source' && !ui.find(page))
    assert.equal(ui.calls.some(call => ['POST', 'PUT', 'PATCH', 'DELETE'].includes(call.method)), false)
    assert.deepEqual(ui.errors, [])
  })
}

test('holding editing records a single trade in the listing currency', async t => {
  const ui = await launch(t, { route: '/accounts/source/holdings/holding-1/edit', authenticated: true, accountsResponse: [sampleAccount], holdingsResponse: [{ ...sampleHolding, buy_currency: 'TRY', quote_currency: 'USD' }] })
  await ui.until(() => ui.find('holding-edit-page'))
  assert.ok(ui.find('holding-edit-currency-corrected')); assert.equal(ui.find('holding-edit-currency').disabled, true)
  assert.match(ui.find('holding-edit-asset').textContent, /QQQM/)
  assert.equal(ui.find('holding-edit-cancel'), null, 'back to account is the only way back')
  assert.equal(ui.find('account-detail-header'), null, 'action pages do not repeat the account header')
  assert.equal(ui.find('holding-edit-locked'), null)
  submitForm(ui, 'holding-edit-form')
  await ui.until(() => ui.calls.find(call => call.method === 'PUT'))
  assert.equal(ui.calls.find(call => call.method === 'PUT').body.buy_currency, 'USD')
  assert.deepEqual(ui.errors, [])
})

test('holding editing preserves restrictions and missing records cannot open a form', async t => {
  const ui = await launch(t, { route: '/accounts/source/holdings/holding-1/edit', authenticated: true, accountsResponse: [sampleAccount], holdingsResponse: [{ ...sampleHolding, can_edit_position: false }] })
  await ui.until(() => ui.find('holding-edit-page'))
  assert.equal(ui.find('holding-edit-quantity').disabled, true); assert.equal(ui.find('holding-edit-price').disabled, true)
  assert.ok(ui.find('holding-edit-locked')); assert.match(ui.find('holding-edit-activities').getAttribute('href'), /\/accounts\/source\?tab=activities$/)
  input(ui, 'holding-edit-notes', 'Updated notes'); setField(ui, 'holding-sector', 'Technology'); await pause()
  submitForm(ui, 'holding-edit-form')
  await ui.until(() => ui.find('holding-row-holding-1'))
  assert.deepEqual(ui.calls.find(call => call.method === 'PUT').body, { notes: 'Updated notes', asset_class: '', sector: 'Technology', country: '' })
  const missing = await launch(t, { route: '/accounts/source/holdings/missing/edit', authenticated: true, accountsResponse: [sampleAccount] })
  await missing.until(() => missing.find('account-action-error'))
  assert.equal(missing.find('holding-edit-form'), null); noDialog(ui); noDialog(missing); assert.deepEqual(ui.errors, [])
})

test('activity creation preserves values after an API error and returns to the activity tab', async t => {
  const ui = await launch(t, { route: '/accounts/source?tab=activities', authenticated: true, accountsResponse: [sampleAccount] })
  await ui.until(() => ui.find('activities-add')); ui.find('activities-add').click()
  await ui.until(() => ui.find('activity-page'))
  input(ui, 'activity-symbol', 'QQQM.US'); input(ui, 'activity-date', '2026-02-01'); input(ui, 'activity-quantity', '3'); input(ui, 'activity-price', '200'); await pause()
  ui.state.mutationFailure = true; submitForm(ui, 'activity-form')
  await ui.until(() => ui.find('activity-error'))
  assert.equal(ui.find('activity-symbol').value, 'QQQM.US'); assert.equal(ui.find('activity-quantity').value, '3')
  ui.state.mutationFailure = false; submitForm(ui, 'activity-form')
  await ui.until(() => ui.find('activity-row-activities-new'))
  assert.equal(ui.w.location.search, '?tab=activities'); noDialog(ui); assert.deepEqual(ui.errors, [])
})

test('activity editing and inline deletion keep the correct API and cancellation behavior', async t => {
  const ui = await launch(t, { route: '/accounts/source?tab=activities', authenticated: true, accountsResponse: [sampleAccount], activitiesResponse: [sampleActivity] })
  await ui.until(() => ui.find('activity-edit-activity-1')); ui.find('activity-edit-activity-1').click()
  await ui.until(() => ui.find('activity-page'))
  assert.equal(ui.w.location.pathname, '/accounts/source/activities/activity-1/edit'); assert.equal(ui.find('activity-quantity').value, '2')
  input(ui, 'activity-quantity', '4'); await pause(); submitForm(ui, 'activity-form')
  await ui.until(() => ui.find('activity-row-activity-1'))
  assert.equal(ui.calls.find(call => call.method === 'PUT').endpoint, '/api/v1/accounts/source/activities/activity-1')
  ui.find('activity-delete-activity-1').click(); await ui.until(() => ui.find('activity-delete-confirmation')); noDialog(ui)
  ui.find('activity-delete-confirmation-cancel').click(); await pause(); assert.equal(ui.calls.some(call => call.method === 'DELETE'), false)
  ui.find('activity-delete-activity-1').click(); await ui.until(() => ui.find('activity-delete-confirmation-confirm')); ui.find('activity-delete-confirmation-confirm').click()
  await ui.until(() => !ui.find('activity-row-activity-1'))
  assert.equal(ui.calls.filter(call => call.method === 'DELETE').length, 1); assert.deepEqual(ui.errors, [])
})

test('transfer page preserves cross-currency amounts and excludes inactive and source accounts', async t => {
  const accounts = [sampleAccount, { ...sampleAccount, id: 'target', currency: 'EUR' }, { ...sampleAccount, id: 'inactive', is_active: false }]
  const ui = await launch(t, { route: '/accounts/source?tab=activities', authenticated: true, accountsResponse: accounts })
  await ui.until(() => ui.find('activities-transfer')); ui.find('activities-transfer').click(); await ui.until(() => ui.find('transfer-page'))
  ui.find('transfer-target').dispatchEvent(new ui.w.KeyboardEvent('keydown', { key: 'Enter', bubbles: true, cancelable: true }))
  await ui.until(() => ui.find('transfer-target-option-target'))
  assert.equal(ui.find('transfer-target-option-source'), null); assert.equal(ui.find('transfer-target-option-inactive'), null)
  ui.find('transfer-target-option-target').dispatchEvent(new ui.w.KeyboardEvent('keydown', { key: 'Enter', bubbles: true, cancelable: true }))
  await ui.until(() => ui.w.document.getElementById('transfer-received'))
  setField(ui, 'transfer-date', '2026-02-02'); input(ui, 'transfer-amount', '100'); setField(ui, 'transfer-received', '90'); setField(ui, 'transfer-fee', '2'); await pause()
  submitForm(ui, 'transfer-form'); submitForm(ui, 'transfer-form')
  await ui.until(() => ui.find('activities-transfer'))
  const transfers = ui.calls.filter(call => call.endpoint === '/api/v1/transfers')
  assert.equal(transfers.length, 1)
  assert.deepEqual(transfers[0].body, { to_account_id: 'target', date: '2026-02-02', amount: 100, received_amount: 90, fee: 2, from_account_id: 'source' })
  noDialog(ui); assert.deepEqual(ui.errors, [])
})
