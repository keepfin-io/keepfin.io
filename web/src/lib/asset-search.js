export function assetCategory(asset) {
  const type = (asset.Type || '').toLowerCase()
  if (type.includes('etf') || type.includes('fund')) return 'fund'
  if (type.includes('crypto') || asset.Exchange === 'CC') return 'crypto'
  if (type.includes('currency') || asset.Exchange === 'FOREX') return 'forex'
  if (type.includes('stock') || type.includes('equity')) return 'stock'
  return 'other'
}
