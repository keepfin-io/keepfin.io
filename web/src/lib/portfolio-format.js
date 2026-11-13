import { zen } from '@/composables/useZen'
export function formatMoney(value, currency = 'USD', locale = 'en') {
  if (zen.value) return '••••'
  if (value == null || !Number.isFinite(value)) return '—'
  return new Intl.NumberFormat(locale, { style: 'currency', currency, maximumFractionDigits: 2 }).format(value)
}

export function formatNumber(value, locale = 'en', digits = 2) {
  if (zen.value) return '••••'
  if (value == null || !Number.isFinite(value)) return '—'
  return new Intl.NumberFormat(locale, { maximumFractionDigits: digits }).format(value)
}

export function formatPercent(value, locale = 'en') {
  return value == null || !Number.isFinite(value) ? '—' : `${formatNumber(value, locale)}%`
}

const MINOR_UNITS = { GBX: 'GBP', ZAC: 'ZAR', ILA: 'ILS' }

// Currency that trades of a listing must use. Providers quote some listings in
// minor units (LSE pence as GBX/GBp), which are traded in the major currency.
export function tradeCurrency(quote) {
  const code = String(quote || '').trim()
  const upper = code === 'GBp' ? 'GBX' : code === 'ZAc' ? 'ZAC' : code.toUpperCase()
  const major = MINOR_UNITS[upper] || upper
  return /^[A-Z]{3}$/.test(major) ? major : ''
}
