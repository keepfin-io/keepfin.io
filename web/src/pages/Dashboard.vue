<template>
  <section data-testid="dashboard-page" class="dashboard-page" :aria-busy="loading">
    <div class="dashboard-heading">
      <div class="min-w-0"><p class="dashboard-eyebrow"><span />{{ $t('portfolio.eyebrow') }}</p><h1>{{ $t('portfolio.title') }}</h1><p class="dashboard-subtitle">{{ $t('portfolio.subtitle') }}</p></div>
      <Button as-child class="portfolio-touch portfolio-primary accounts-button"><router-link data-testid="dashboard-accounts-link" to="/accounts"><WalletCards class="size-4" />{{ $t('nav.accounts') }}<ArrowUpRight class="size-4" /></router-link></Button>
    </div>

    <Alert v-if="error" data-testid="portfolio-error" variant="error" :description="error" />
    <div class="portfolio-toolbar">
      <div class="toolbar-field"><Label for="portfolio-range">{{ $t('portfolio.range') }}</Label><FormSelect id="portfolio-range" data-testid="portfolio-range" class="portfolio-select" v-model="range" :options="['1d', '1m', '3m', 'ytd', '1y', '5y', 'max'].map(value => ({ value, label: $t(`portfolio.ranges.${value}`) }))" /></div>
      <div class="toolbar-field benchmark-field"><Label for="portfolio-benchmark">{{ $t('portfolio.benchmark') }}</Label><FormSelect id="portfolio-benchmark" data-testid="portfolio-benchmark" class="portfolio-select" v-model="benchmark" :options="[{ value: 'none', label: $t('portfolio.none') }, { value: 'SPY.US', label: 'S&P 500 (SPY)' }, { value: 'VT.US', label: $t('portfolio.globalStocks') + ' (VT)' }, { value: 'QQQ.US', label: 'Nasdaq 100 (QQQ)' }]" /></div>
      <Button data-testid="portfolio-refresh" :disabled="loading" variant="outline" class="portfolio-touch refresh-button" :aria-label="$t('portfolio.refresh')" @click="load"><RefreshCw :class="['size-4', loading && 'animate-spin motion-reduce:animate-none']" /><span>{{ $t('portfolio.refresh') }}</span></Button>
    </div>
    <div class="report-meta"><p v-if="report">{{ report.from }} <span aria-hidden="true">—</span> {{ report.to }} <span class="meta-dot" /> {{ report.currency }}</p><span>{{ $t('portfolio.latestClose') }}</span><span v-if="loading" role="status">{{ $t('common.loading') }}</span></div>

    <div v-if="loading && !report" data-testid="portfolio-loading" class="summary-loading" aria-hidden="true"><div v-for="index in 4" :key="index"><span /><span /></div></div>
    <template v-if="report">
      <details v-if="report.warnings.length" data-testid="portfolio-warnings" class="portfolio-warnings" open><summary><CircleAlert class="size-4" /><span>{{ $t('portfolio.dataNotes') }} <span class="warning-count">{{ report.warnings.length }}</span></span><ChevronDown class="ml-auto size-4" /></summary><div role="status"><p v-for="warning in report.warnings" :key="warning">{{ warning }}</p></div></details>

      <div class="summary-grid">
        <article class="balance-card">
          <div class="balance-top"><span class="balance-icon"><Layers3 class="size-5" /></span><span class="currency-badge">{{ report.currency }}</span></div>
          <div data-testid="portfolio-stat-value" class="balance-total"><h2>{{ $t('portfolio.value') }}</h2><p>{{ amount('value') }}</p></div>
          <div class="balance-split"><div data-testid="portfolio-stat-cash"><h3>{{ $t('portfolio.cash') }}</h3><p>{{ amount('cash') }}</p></div><div data-testid="portfolio-stat-equity"><h3>{{ $t('portfolio.investedValue') }}</h3><p>{{ amount('equity') }}</p></div></div>
        </article>
        <div class="metrics-grid">
          <article v-for="metric in metrics" :key="metric.key" class="metric-card" :data-testid="`portfolio-stat-${metric.key}`">
            <div class="metric-heading"><component :is="metric.icon" class="size-4" aria-hidden="true" /><h2>{{ $t(`portfolio.${metric.key}`) }}</h2></div>
            <p :class="['metric-value', metric.signed && report.complete && tone(report.summary[metric.key])]">{{ amount(metric.key) }}</p>
            <span v-if="metric.key === 'dividends'" class="metric-caption">{{ $t('portfolio.selectedPeriod') }}</span>
          </article>
        </div>
      </div>

      <Tabs v-model="tab" class="portfolio-tabs">
        <TabsList class="portfolio-tab-list" :aria-label="$t('portfolio.sections')">
          <TabsTrigger v-for="item in tabs" :key="item.value" :value="item.value" :data-testid="`portfolio-tab-${item.value}`" class="portfolio-tab"><component :is="item.icon" class="size-4 shrink-0" /><span>{{ $t(`portfolio.${item.value}`) }}</span></TabsTrigger>
        </TabsList>
        <TabsContent value="overview" class="overview-grid">
          <section class="portfolio-panel performance-panel" aria-labelledby="performance-heading">
            <div class="performance-heading"><div><h2 id="performance-heading" class="portfolio-panel-title">{{ $t('portfolio.performance') }}</h2><p class="portfolio-panel-description">{{ $t('portfolio.performanceHint') }}</p></div><div class="chart-switch" role="group" :aria-label="$t('portfolio.chartMode')"><Button v-for="mode in ['value', 'return']" :key="mode" :data-testid="`portfolio-chart-${mode}`" variant="ghost" class="portfolio-touch" :aria-pressed="chartMode === mode" @click="chartMode = mode">{{ $t(`portfolio.${mode}`) }}</Button></div></div>
            <PortfolioChart :points="report.history" :mode="chartMode" />
            <details class="calculation-note"><summary>{{ $t('portfolio.howCalculated') }}</summary><p>{{ $t('portfolio.returnMethod') }}</p></details>
          </section>
          <section class="portfolio-panel mix-panel" aria-labelledby="mix-heading">
            <div class="mix-heading"><span class="mix-icon"><ChartPie class="size-5" /></span><div><h2 id="mix-heading" class="portfolio-panel-title">{{ $t('portfolio.assetMix') }}</h2><p class="portfolio-panel-description">{{ $t('portfolio.assetMixHint') }}</p></div></div>
            <div v-if="!report.complete || !assetMix.length" class="portfolio-empty"><ChartPie /><p>{{ $t(report.complete ? 'portfolio.noAllocations' : 'portfolio.incomplete') }}</p></div>
            <div v-else class="mix-rows"><div v-for="(row, index) in assetMix" :key="row.name" class="mix-row"><div><span>{{ allocationName(row.name) }}</span><strong>{{ percent(row.weight) }}</strong></div><div class="mix-track" aria-hidden="true"><span :style="{ width: `${Math.max(0, Math.min(100, row.weight))}%`, opacity: 1 - index * 0.15 }" /></div><p>{{ money(row.value) }}</p></div></div>
            <Button data-testid="portfolio-view-allocation" variant="ghost" class="portfolio-touch mix-action" @click="tab = 'allocation'">{{ $t('portfolio.viewAllocation') }}<ArrowRight class="size-4" /></Button>
          </section>
        </TabsContent>
        <PortfolioDetails :report="report" />
      </Tabs>
    </template>
  </section>
</template>
<script setup>
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ArrowRight, ArrowUpRight, CalendarDays, ChartNoAxesCombined, ChartPie, ChevronDown, CircleAlert, CircleCheck, Coins, Layers3, LayoutDashboard, RefreshCw, TrendingUp, WalletCards } from 'lucide-vue-next'
import { portfolioAPI } from '@/services/portfolioApi'
import PortfolioChart from '@/components/portfolio/PortfolioChart.vue'
import PortfolioDetails from '@/components/portfolio/PortfolioDetails.vue'
import { Button, FormSelect, Label, FeedbackAlert as Alert, Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui'
import { formatMoney, formatPercent } from '@/lib/portfolio-format'
const { locale, t } = useI18n()
const range = ref('1m'), benchmark = ref('none'), chartMode = ref('value'), tab = ref('overview')
const report = ref(null), loading = ref(false), error = ref('')
const metrics = [
  { key: 'net_contributions', icon: WalletCards }, { key: 'profit_loss', icon: TrendingUp, signed: true },
  { key: 'realized', icon: CircleCheck, signed: true }, { key: 'unrealized', icon: ChartNoAxesCombined, signed: true },
  { key: 'dividends', icon: Coins }, { key: 'daily_change', icon: CalendarDays, signed: true },
]
const tabs = [{ value: 'overview', icon: LayoutDashboard }, { value: 'holdings', icon: Layers3 }, { value: 'income', icon: Coins }, { value: 'allocation', icon: ChartPie }]
const assetMix = computed(() => (report.value?.allocations.asset_class || []).filter(row => row.value !== 0).slice(0, 4))
let request = 0
async function load() {
  const current = ++request; loading.value = true; error.value = ''
  try {
    const result = await portfolioAPI.report({ range: range.value, benchmark: benchmark.value === 'none' ? '' : benchmark.value })
    if (current === request) report.value = result.portfolio
  } catch (e) { if (current === request) { error.value = e.message; report.value = null } }
  finally { if (current === request) loading.value = false }
}
watch([range, benchmark], load, { immediate: true })
function money(v) { return formatMoney(v, report.value?.currency || 'USD', locale.value) }
function percent(v) { return formatPercent(v, locale.value) }
function amount(key) { return report.value.complete ? money(report.value.summary[key]) : '—' }
function tone(v) { return v < 0 ? 'portfolio-negative' : v > 0 ? 'portfolio-positive' : '' }
function allocationName(name) { return name === 'Cash' ? t('portfolio.cash') : name === 'Unclassified' ? t('portfolio.unclassified') : name }
</script>
<style scoped>
.dashboard-page { container: portfolio / inline-size; min-width: 0; display: flex; flex-direction: column; gap: 20px; }
.dashboard-heading { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 20px; margin-bottom: 4px; }.dashboard-heading h1 { margin-top: 10px; font-size: clamp(26px, 3.8cqw, 34px); font-weight: 600; line-height: 1.15; letter-spacing: -.045em; }
.dashboard-eyebrow { display: flex; align-items: center; gap: 8px; color: var(--brand-accent); font-size: 10px; font-weight: 600; letter-spacing: .13em; text-transform: uppercase; }.dashboard-eyebrow > span { width: 5px; height: 5px; border-radius: 50%; background: currentColor; }.dashboard-subtitle { margin-top: 12px; max-width: 480px; font-size: 13px; line-height: 1.7; color: var(--muted-foreground); }
.accounts-button { font-size: 12px; padding-inline: 16px; }.portfolio-toolbar { display: grid; grid-template-columns: minmax(0, 1fr) 44px; gap: 14px 12px; padding: 16px; border: 1px solid var(--border); border-radius: 18px; background: var(--card); }
.toolbar-field { display: grid; min-width: 0; gap: 8px; }.toolbar-field label { font-size: 11px; color: var(--muted-foreground); }.benchmark-field { grid-column: 1 / -1; grid-row: 2; }.refresh-button { grid-column: 2; grid-row: 1; align-self: end; padding: 0; box-shadow: none; }.refresh-button > span { display: none; }
.report-meta { display: flex; flex-wrap: wrap; gap: 6px 14px; margin-top: -9px; font-size: 10px; color: var(--muted-foreground); }.report-meta p { display: flex; flex-wrap: wrap; gap: 8px; align-items: center; }.meta-dot { width: 3px; height: 3px; border-radius: 50%; background: currentColor; }.report-meta > span:last-child:not(:first-child) { margin-left: auto; }
.portfolio-warnings { border: 1px solid #b584333d; border-radius: 14px; background: #b5843308; font-size: 12px; }.portfolio-warnings summary { display: flex; align-items: center; gap: 10px; min-height: 48px; padding: 12px 16px; cursor: pointer; list-style: none; }.portfolio-warnings summary::-webkit-details-marker { display: none; }.portfolio-warnings summary > svg:first-child { color: #b58433; }.warning-count { margin-left: 5px; opacity: .6; }.portfolio-warnings[open] summary > svg:last-child { transform: rotate(180deg); }.portfolio-warnings > div { padding: 0 16px 16px 42px; color: var(--muted-foreground); line-height: 1.7; overflow-wrap: anywhere; }.portfolio-warnings p + p { margin-top: 6px; }
.summary-grid { display: grid; gap: 14px; }.balance-card { min-width: 0; display: flex; flex-direction: column; padding: 24px; border: 1px solid #2c9b7730; border-radius: 22px; background: var(--brand-panel); }.balance-top { display: flex; justify-content: space-between; align-items: center; }.balance-icon { display: grid; place-items: center; width: 40px; height: 40px; border: 1px solid #2d9e7f30; border-radius: 13px; color: var(--brand-accent); background: var(--brand-soft); }.currency-badge { padding: 5px 8px; border: 1px solid #2d9e7f30; border-radius: 7px; color: var(--brand-accent); font-size: 10px; font-weight: 500; letter-spacing: .07em; }
.balance-total { flex: 1; margin-top: 25px; margin-bottom: 28px; }.balance-total h2 { font-size: 12px; color: var(--muted-foreground); }.balance-total p { margin-top: 10px; font-size: clamp(26px, 4.5cqw, 38px); font-weight: 500; letter-spacing: -.045em; font-variant-numeric: tabular-nums; line-height: 1.2; overflow-wrap: anywhere; }.balance-split { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 20px; padding-top: 18px; border-top: 1px solid #2d9e7f25; }.balance-split h3 { font-size: 10px; color: var(--muted-foreground); }.balance-split p { margin-top: 7px; font-size: 14px; font-weight: 500; font-variant-numeric: tabular-nums; overflow-wrap: anywhere; }
.metrics-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; }.metric-card { min-width: 0; padding: clamp(14px, 2cqw, 20px); border: 1px solid var(--border); border-radius: 18px; background: var(--card); }.metric-heading { display: flex; align-items: flex-start; gap: 8px; color: var(--muted-foreground); }.metric-heading > svg { flex-shrink: 0; margin-top: 1px; }.metric-heading h2 { font-size: 11px; line-height: 1.6; }.metric-value { margin-top: 15px; font-size: clamp(17px, 2.6cqw, 23px); font-weight: 500; font-variant-numeric: tabular-nums; line-height: 1.3; letter-spacing: -.025em; overflow-wrap: anywhere; }.metric-caption { display: block; margin-top: 5px; color: var(--muted-foreground); font-size: 9px; }
.portfolio-tabs { min-width: 0; display: grid; gap: 20px; }.portfolio-tab-list { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); width: 100%; height: auto; justify-content: start; gap: 4px; padding: 5px; border: 1px solid var(--border); border-radius: 15px; background: color-mix(in srgb, var(--muted) 35%, transparent); }.portfolio-tab { min-width: 0; min-height: 44px; padding: 8px 12px; justify-content: flex-start; gap: 8px; white-space: normal; border-radius: 10px; font-size: 12px; line-height: 1.4; }.portfolio-tab[data-state=active] { color: var(--brand-accent); background: var(--card); box-shadow: 0 2px 5px #00000008; }
.overview-grid { display: grid; min-width: 0; gap: 16px; }.performance-heading { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; flex-wrap: wrap; margin-bottom: 20px; }.performance-heading > div:first-child { min-width: 0; flex: 1 1 200px; }.chart-switch { display: flex; min-width: 0; gap: 2px; padding: 3px; border-radius: 12px; background: var(--muted); }.chart-switch button { min-width: 0; padding-inline: 12px; font-size: 11px; white-space: normal; }.chart-switch button[aria-pressed=true] { color: var(--brand-accent); background: var(--card); box-shadow: 0 2px 4px #00000009; }
.calculation-note { margin-top: 5px; color: var(--muted-foreground); font-size: 11px; line-height: 1.7; }.calculation-note summary { padding-block: 12px; min-height: 44px; cursor: pointer; }.calculation-note p { padding-bottom: 8px; }.mix-panel { display: flex; flex-direction: column; }.mix-heading { display: flex; align-items: center; gap: 12px; }.mix-icon { display: grid; place-items: center; width: 40px; height: 40px; flex-shrink: 0; border-radius: 12px; color: var(--brand-accent); background: var(--brand-soft); }.mix-rows { display: grid; flex: 1; align-content: start; gap: 24px; margin-top: 28px; }.mix-row { min-width: 0; }.mix-row > div:first-child { display: flex; justify-content: space-between; gap: 12px; font-size: 12px; }.mix-row span { overflow-wrap: anywhere; }.mix-row strong { flex-shrink: 0; font-weight: 500; font-variant-numeric: tabular-nums; }.mix-track { height: 5px; margin: 10px 0 7px; background: var(--muted); border-radius: 6px; overflow: hidden; }.mix-track > span { display: block; height: 100%; border-radius: inherit; background: var(--brand-accent); }.mix-row p { color: var(--muted-foreground); font-size: 11px; overflow-wrap: anywhere; }.mix-action { justify-content: space-between; margin-top: 28px; color: var(--brand-accent); font-size: 12px; white-space: normal; }
.summary-loading { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; }.summary-loading > div { display: grid; gap: 20px; padding: 24px; border: 1px solid var(--border); border-radius: 18px; background: var(--card); }.summary-loading span { display: block; height: 12px; border-radius: 5px; background: var(--muted); }.summary-loading span + span { width: 70%; height: 24px; }
@container portfolio (min-width: 600px) { .portfolio-toolbar { grid-template-columns: minmax(120px, .7fr) minmax(0, 1.4fr) auto; gap: 20px; }.benchmark-field, .refresh-button { grid-column: auto; grid-row: auto; }.refresh-button { padding-inline: 16px; gap: 8px; }.refresh-button > span { display: inline; }.portfolio-tab-list { display: flex; width: fit-content; max-width: 100%; flex-wrap: wrap; }.portfolio-tab { justify-content: center; }.metrics-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); } }
@container portfolio (min-width: 900px) { .summary-grid { grid-template-columns: minmax(0, 1fr) minmax(0, 1.65fr); gap: 16px; }.overview-grid { grid-template-columns: minmax(0, 2fr) minmax(0, 1fr); } }
@container portfolio (max-width: 350px) { .metric-heading > svg { display: none; }.metric-value { font-size: 17px; }.balance-card { padding: 20px; }.portfolio-tab { font-size: 11px; padding-inline: 8px; } }
</style>
