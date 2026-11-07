# Wealth feature implementation

Scope: dividend-to-asset linkage, stock splits, watchlists, manual assets and dated
valuations, liabilities/net worth, tags, consolidated activities, configurable risk
checks, FIRE projections, portfolio sharing, provider jobs, performance methods and
FX attribution, JSON/dividend import, PWA and Zen mode.

Implementation follows the current PostgreSQL ledger and ownership rules. Existing
worktree changes are retained. Tests must distinguish actual database/API evidence
from simulated UI checks and external provider/browser availability.

- [ ] Ledger: dividends, splits and history-preserving replay
- [ ] Watchlists, manual assets, valuations and liabilities
- [ ] Tags and consolidated activities
- [ ] Risk settings and FIRE projections
- [ ] Scoped, expiring, revocable read-only sharing
- [ ] Provider selection, durable market-data jobs and administration
- [ ] Performance methods and FX attribution
- [ ] JSON and dividend preview/import
- [ ] Vue pages, navigation, PWA and Zen mode
- [ ] Database/API, accounting, UI, build and race verification
