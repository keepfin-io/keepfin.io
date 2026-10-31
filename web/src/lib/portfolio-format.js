export function formatMoney(value, currency = 'USD', locale = 'en') {
  if (value == null || !Number.isFinite(value)) return '—'
  return new Intl.NumberFormat(locale, { style: 'currency', currency, maximumFractionDigits: 2 }).format(value)
}

export function formatNumber(value, locale = 'en', digits = 2) {
  if (value == null || !Number.isFinite(value)) return '—'
  return new Intl.NumberFormat(locale, { maximumFractionDigits: digits }).format(value)
}

export function formatPercent(value, locale = 'en') {
  return value == null || !Number.isFinite(value) ? '—' : `${formatNumber(value, locale)}%`
}
