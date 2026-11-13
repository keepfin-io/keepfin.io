// Supported account platforms, grouped as the picker shows them. Names must
// match the backend allowlist in app/models/platform/platform.go.
export const PLATFORM_GROUPS = [
  {
    label: 'Banks',
    options: [
      { name: 'Bank of America', domain: 'bankofamerica.com' },
      { name: 'Chase', domain: 'chase.com' },
      { name: 'Wells Fargo', domain: 'wellsfargo.com' },
      { name: 'Citibank', domain: 'citibank.com' },
      { name: 'HSBC', domain: 'hsbc.com' },
      { name: 'Barclays', domain: 'barclays.co.uk' },
      { name: 'Deutsche Bank', domain: 'db.com' },
      { name: 'ING', domain: 'ing.com' },
      { name: 'N26', domain: 'n26.com' },
      { name: 'Revolut', domain: 'revolut.com' },
    ],
  },
  {
    label: 'Brokerages',
    options: [
      { name: 'Interactive Brokers', domain: 'interactivebrokers.com' },
      { name: 'Fidelity', domain: 'fidelity.com' },
      { name: 'Charles Schwab', domain: 'schwab.com' },
      { name: 'TD Ameritrade', domain: 'tdameritrade.com' },
      { name: 'E*TRADE', domain: 'etrade.com' },
      { name: 'Robinhood', domain: 'robinhood.com' },
      { name: 'Webull', domain: 'webull.com' },
      { name: 'Trading 212', domain: 'trading212.com' },
      { name: 'eToro', domain: 'etoro.com' },
      { name: 'Degiro', domain: 'degiro.com' },
    ],
  },
  {
    label: 'Crypto',
    options: [
      { name: 'Coinbase', domain: 'coinbase.com' },
      { name: 'Binance', domain: 'binance.com' },
      { name: 'Kraken', domain: 'kraken.com' },
      { name: 'Crypto.com', domain: 'crypto.com' },
      { name: 'Gemini', domain: 'gemini.com' },
      { name: 'Bybit', domain: 'bybit.com' },
      { name: 'OKX', domain: 'okx.com' },
    ],
  },
  {
    label: 'Payment Services',
    options: [
      { name: 'PayPal', domain: 'paypal.com' },
      { name: 'Wise', domain: 'wise.com' },
      { name: 'Stripe', domain: 'stripe.com' },
      { name: 'Venmo', domain: 'venmo.com' },
      { name: 'Cash App', domain: 'cash.app' },
    ],
  },
  {
    label: 'Turkish Banks',
    options: [
      { name: 'Garanti BBVA', domain: 'garantibbva.com.tr' },
      { name: 'İş Bankası', domain: 'isbank.com.tr' },
      { name: 'Yapı Kredi', domain: 'yapikredi.com.tr' },
      { name: 'Akbank', domain: 'akbank.com' },
      { name: 'QNB Finansbank', domain: 'qnbfinansbank.com' },
      { name: 'Ziraat Bankası', domain: 'ziraatbank.com.tr' },
      { name: 'Midas', domain: 'getmidas.com' },
    ],
  },
]

const DOMAINS = Object.fromEntries(
  PLATFORM_GROUPS.flatMap(group => group.options.map(option => [option.name, option.domain]))
)

// Website domain used to show a platform's logo, or '' when it has none.
export function platformDomain(name) {
  return DOMAINS[name] || ''
}
