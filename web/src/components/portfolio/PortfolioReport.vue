<template>
  <section :data-testid="accountId ? 'account-analysis-report' : 'dashboard-page'" class="dashboard-page" :aria-busy="loading">
    <div class="dashboard-heading">
      <div class="min-w-0"><h1>{{ $t(accountId ? 'portfolio.accountAnalysis' : 'portfolio.title') }}</h1><p class="dashboard-subtitle">{{ accountId ? $t('portfolio.accountSubtitle', { name: accountName }) : $t('portfolio.subtitle') }}</p></div>
      <Button v-if="!accountId && !readonly" as-child variant="outline" class="portfolio-touch accounts-button"><router-link data-testid="dashboard-accounts-link" to="/accounts"><WalletCards class="size-4" />{{ $t('nav.accounts') }}<ArrowUpRight class="size-4" /></router-link></Button>
    </div>

    <Alert v-if="error" data-testid="portfolio-error" variant="error" :description="error" />
    <div class="portfolio-toolbar">
      <div class="toolbar-field"><Label for="portfolio-range">{{ $t('portfolio.range') }}</Label><FormSelect id="portfolio-range" data-testid="portfolio-range" class="portfolio-select" v-model="range" :options="['1d', '1m', '3m', 'ytd', '1y', '5y', 'max'].map(value => ({ value, label: $t(`portfolio.ranges.${value}`) }))" /></div>
      <div v-if="!readonly" class="toolbar-field benchmark-field"><Label for="portfolio-benchmark">{{ $t('portfolio.benchmark') }}</Label><FormSelect id="portfolio-benchmark" data-testid="portfolio-benchmark" class="portfolio-select" v-model="benchmark" :options="[{ value: 'none', label: $t('portfolio.none') }, { value: 'SPY.US', label: 'S&P 500 (SPY)' }, { value: 'VT.US', label: $t('portfolio.globalStocks') + ' (VT)' }, { value: 'QQQ.US', label: 'Nasdaq 100 (QQQ)' }]" /></div>
      <div v-if="!readonly" class="toolbar-field"><Label for="performance-method">{{ $t('wealth.method') }}</Label><FormSelect id="performance-method" v-model="method" :options="['dietz','twr','mwr','roai'].map(value => ({value,label:$t(`wealth.method_${value}`)}))" /></div>
      <div v-if="!accountId && !readonly" class="toolbar-field"><Label for="portfolio-tag">{{ $t('wealth.tags') }}</Label><FormSelect id="portfolio-tag" v-model="tag" :options="[{value:'all',label:$t('wealth.allTags')},...tags.map(t => ({value:t.id,label:t.name}))]" /></div>
      <Button data-testid="portfolio-refresh" :disabled="loading" variant="outline" class="portfolio-touch refresh-button" :aria-label="$t('portfolio.refresh')" @click="load"><RefreshCw :class="['size-4', loading && 'animate-spin motion-reduce:animate-none']" /><span>{{ $t('portfolio.refresh') }}</span></Button>
    </div>
    <div class="report-meta"><p v-if="report">{{ report.from }} <span aria-hidden="true">—</span> {{ report.to }} <span class="meta-dot" /> {{ report.currency }}</p><span>{{ $t('portfolio.latestClose') }}</span><span v-if="loading" role="status">{{ $t('common.loading') }}</span></div>

    <div v-if="loading && !report" data-testid="portfolio-loading" class="summary-loading" aria-hidden="true"><div v-for="index in 3" :key="index"><span /><span /></div></div>
    <template v-if="report">
      <details v-if="report.warnings.length" data-testid="portfolio-warnings" class="portfolio-warnings" open><summary><CircleAlert class="size-4" /><span>{{ $t('portfolio.dataNotes') }} <span class="warning-count">{{ report.warnings.length }}</span></span><ChevronDown class="ml-auto size-4" /></summary><div role="status"><p v-for="warning in report.warnings" :key="warning">{{ warning }}</p></div></details>

      <section class="portfolio-summary" aria-labelledby="portfolio-value-heading">
        <div class="summary-main">
          <div class="balance-overview">
            <div data-testid="portfolio-stat-value" class="balance-total"><h2 id="portfolio-value-heading">{{ $t('portfolio.value') }}</h2><p>{{ amount('value') }}</p></div>
            <dl class="balance-split">
              <div data-testid="portfolio-stat-cash"><dt>{{ $t('portfolio.cash') }}</dt><dd>{{ amount('cash') }}</dd></div>
              <div data-testid="portfolio-stat-equity"><dt>{{ $t('portfolio.investedValue') }}</dt><dd>{{ amount('equity') }}</dd></div>
            </dl>
          </div>
          <div v-for="key in ['profit_loss', 'daily_change']" :key="key" class="headline-metric" :data-testid="`portfolio-stat-${key}`">
            <h2>{{ $t(`portfolio.${key}`) }}</h2>
            <p :class="['metric-value', report.complete && tone(report.summary[key])]">{{ amount(key) }}</p>
          </div>
        </div>
        <details data-testid="portfolio-summary-details" class="summary-details">
          <summary data-testid="portfolio-summary-toggle"><span>{{ $t('portfolio.summaryDetails') }}</span><ChevronDown class="size-4" aria-hidden="true" /></summary>
          <dl class="detail-metrics">
            <div v-for="metric in detailMetrics" :key="metric.key" :data-testid="`portfolio-stat-${metric.key}`">
              <dt>{{ $t(`portfolio.${metric.key}`) }}</dt>
              <dd :class="['metric-value', metric.signed && report.complete && tone(report.summary[metric.key])]">{{ amount(metric.key) }}</dd>
              <dd v-if="['dividends', 'interest', 'fees'].includes(metric.key)" class="metric-caption">{{ $t('portfolio.selectedPeriod') }}</dd>
            </div>
          </dl>
        </details>
      </section>

      <details v-if="report.performance" class="summary-details"><summary class="cursor-pointer py-3 text-sm">{{ $t('wealth.performanceDetails') }}</summary><div class="grid gap-3 sm:grid-cols-3"><div v-for="key in ['profit','price_effect','currency_effect']" :key="key"><p class="text-xs text-muted-foreground">{{ $t(`wealth.${key}`) }}</p><p class="mt-1 font-semibold">{{ money(report.performance[key]) }}</p></div></div><p class="mt-3 text-xs text-muted-foreground">{{ $t('wealth.fxHint') }}</p></details>
      <Tabs v-model="tab" class="portfolio-tabs">
        <TabsList class="portfolio-tab-list" :aria-label="$t('portfolio.sections')">
          <TabsTrigger v-for="value in tabs" :key="value" :value="value" :data-testid="`portfolio-tab-${value}`" class="portfolio-tab">{{ $t(`portfolio.${value}`) }}</TabsTrigger>
        </TabsList>
        <TabsContent value="overview" class="overview-grid">
          <section class="performance-panel" aria-labelledby="performance-heading">
            <div class="performance-heading"><div><h2 id="performance-heading" class="portfolio-panel-title">{{ $t('portfolio.performance') }}</h2><p class="portfolio-panel-description">{{ $t('portfolio.performanceHint') }}</p></div><div class="chart-switch" role="group" :aria-label="$t('portfolio.chartMode')"><Button v-for="mode in ['value', 'return']" :key="mode" :data-testid="`portfolio-chart-${mode}`" variant="ghost" class="portfolio-touch" :aria-pressed="chartMode === mode" @click="chartMode = mode">{{ $t(`portfolio.${mode}`) }}</Button></div></div>
            <PortfolioChart :points="report.history" :mode="chartMode" />
            <details class="calculation-note"><summary>{{ $t('portfolio.howCalculated') }}</summary><p>{{ report.method || $t('portfolio.returnMethod') }}</p></details>
          </section>
          <section class="mix-panel" aria-labelledby="mix-heading">
            <h2 id="mix-heading" class="portfolio-panel-title">{{ $t('portfolio.assetMix') }}</h2>
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
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ArrowRight, ArrowUpRight, ChartPie, ChevronDown, CircleAlert, RefreshCw, WalletCards } from 'lucide-vue-next'
import { wealthAPI } from '@/services/wealthApi'
import { portfolioAPI } from '@/services/portfolioApi'
import PortfolioChart from '@/components/portfolio/PortfolioChart.vue'
import PortfolioDetails from '@/components/portfolio/PortfolioDetails.vue'
import { Button, FormSelect, Label, FeedbackAlert as Alert, Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui'
import { formatMoney, formatPercent } from '@/lib/portfolio-format'
const props = defineProps({ accountId: { type: String, default: '' }, accountName: { type: String, default: '' }, readonly: Boolean, reportLoader: { type: Function, default: null } })
const { locale, t } = useI18n()
const range = ref('1m'), benchmark = ref('none'), chartMode = ref('value'), tab = ref('overview')
const method = ref('dietz'), tag = ref('all'), tags = ref([])
onMounted(async () => { if (!props.readonly) { try { tags.value=(await wealthAPI.get('/tags')).items } catch { /* The portfolio remains available without tag filters. */ } } })
const report = ref(null), loading = ref(false), error = ref('')
const detailMetrics = [
  { key: 'net_contributions' }, { key: 'realized', signed: true },
  { key: 'unrealized', signed: true }, { key: 'cost' },
  { key: 'dividends' }, { key: 'interest' }, { key: 'fees' },
]
const tabs = ['overview', 'holdings', 'income', 'allocation']
const assetMix = computed(() => (report.value?.allocations.asset_class || []).filter(row => row.value !== 0).slice(0, 4))
let request = 0
async function load() {
  const current = ++request; loading.value = true; error.value = ''
  try {
    const result = await (props.reportLoader || portfolioAPI.report)({ account_id: props.accountId, range: range.value, benchmark: benchmark.value === 'none' ? '' : benchmark.value, method: method.value, tag_id: tag.value === 'all' ? '' : tag.value })
    if (current === request) report.value = result.portfolio
  } catch (e) { if (current === request) { error.value = e.message; report.value = null } }
  finally { if (current === request) loading.value = false }
}
watch([range, benchmark, () => props.accountId, method, tag], (values, previous) => {
  if (previous?.[2] !== values[2]) report.value = null
  load()
}, { immediate: true })
onBeforeUnmount(() => { request++ })
function money(v) { return formatMoney(v, report.value?.currency || 'USD', locale.value) }
function percent(v) { return formatPercent(v, locale.value) }
function amount(key) { return report.value.complete ? money(report.value.summary[key]) : '—' }
function tone(v) { return v < 0 ? 'portfolio-negative' : v > 0 ? 'portfolio-positive' : '' }
function allocationName(name) { return name === 'Cash' ? t('portfolio.cash') : name === 'Unclassified' ? t('portfolio.unclassified') : name }
</script>
<style scoped>
.dashboard-page { container: portfolio / inline-size; min-width: 0; display: flex; flex-direction: column; gap: 20px; }
.dashboard-heading { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 16px; }
.dashboard-heading h1 { font-size: clamp(26px, 3.4cqw, 32px); font-weight: 600; line-height: 1.2; letter-spacing: -.04em; }
.dashboard-subtitle { margin-top: 8px; max-width: 480px; color: var(--muted-foreground); font-size: 13px; line-height: 1.6; }
.accounts-button { padding-inline: 14px; font-size: 12px; }
.portfolio-toolbar { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) 44px; align-items: end; gap: 12px; }
.toolbar-field { display: grid; min-width: 0; gap: 7px; }
.toolbar-field label { color: var(--muted-foreground); font-size: 11px; }
.refresh-button { padding: 0; box-shadow: none; }
.refresh-button > span { display: none; }
.report-meta { display: flex; align-items: center; flex-wrap: wrap; gap: 6px 16px; margin-top: -8px; color: var(--muted-foreground); font-size: 10px; line-height: 1.6; }
.report-meta p { display: flex; align-items: center; flex-wrap: wrap; gap: 8px; }
.report-meta > span:last-child:not(:first-child) { margin-left: auto; }
.meta-dot { width: 3px; height: 3px; border-radius: 50%; background: currentColor; }
.portfolio-warnings { border: 1px solid #b584333d; border-radius: 12px; background: #b5843308; font-size: 12px; }
.portfolio-warnings summary { display: flex; align-items: center; gap: 10px; min-height: 44px; padding: 12px 16px; cursor: pointer; list-style: none; }
.portfolio-warnings summary::-webkit-details-marker { display: none; }
.portfolio-warnings summary > svg:first-child { color: #b58433; }
.warning-count { margin-left: 5px; opacity: .6; }
.portfolio-warnings[open] summary > svg:last-child { transform: rotate(180deg); }
.portfolio-warnings > div { padding: 0 16px 16px 42px; color: var(--muted-foreground); line-height: 1.7; overflow-wrap: anywhere; }
.portfolio-warnings p + p { margin-top: 6px; }
.portfolio-summary { min-width: 0; padding: 22px 20px 0; border: 1px solid var(--border); border-radius: 18px; background: var(--card); }
.summary-main { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 24px 20px; }
.balance-overview { grid-column: 1 / -1; min-width: 0; }
.balance-total h2, .headline-metric h2 { color: var(--muted-foreground); font-size: 12px; font-weight: 400; line-height: 1.6; }
.balance-total p { margin-top: 8px; font-size: clamp(30px, 4.5cqw, 40px); font-weight: 500; letter-spacing: -.04em; font-variant-numeric: tabular-nums; line-height: 1.2; overflow-wrap: anywhere; }
.balance-split { display: flex; flex-wrap: wrap; gap: 10px 20px; margin-top: 18px; }
.balance-split > div { display: flex; flex-wrap: wrap; gap: 5px 8px; min-width: 0; font-size: 11px; line-height: 1.6; }
.balance-split dt { color: var(--muted-foreground); }
.balance-split dd { font-variant-numeric: tabular-nums; overflow-wrap: anywhere; }
.headline-metric { min-width: 0; }
.metric-value { margin-top: 10px; font-size: clamp(20px, 2.8cqw, 26px); font-weight: 500; letter-spacing: -.025em; font-variant-numeric: tabular-nums; line-height: 1.35; overflow-wrap: anywhere; }
.summary-details { margin-top: 20px; border-top: 1px solid var(--border); }
.summary-details > summary { display: flex; align-items: center; justify-content: space-between; gap: 12px; min-height: 48px; color: var(--muted-foreground); font-size: 12px; cursor: pointer; list-style: none; transition: color 200ms ease; }
.summary-details > summary::-webkit-details-marker { display: none; }
.summary-details > summary:hover, .summary-details[open] > summary { color: var(--brand-accent); }
.summary-details > summary:focus-visible { outline: 2px solid var(--ring); outline-offset: 3px; border-radius: 6px; }
.summary-details[open] > summary > svg { transform: rotate(180deg); }
.detail-metrics { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 20px; padding: 4px 0 24px; }
.detail-metrics > div { min-width: 0; }
.detail-metrics dt { color: var(--muted-foreground); font-size: 11px; line-height: 1.6; }
.detail-metrics .metric-value { margin-top: 7px; font-size: 17px; }
.metric-caption { margin-top: 5px; color: var(--muted-foreground); font-size: 10px; }
.portfolio-tabs { display: grid; min-width: 0; gap: 20px; }
.portfolio-tab-list { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); width: 100%; height: auto; gap: 0 16px; justify-content: start; padding: 0; border-bottom: 1px solid var(--border); border-radius: 0; background: transparent; }
.portfolio-tab { min-width: 0; min-height: 48px; padding: 10px 4px; border: 0; border-bottom: 2px solid transparent; border-radius: 0; background: transparent; color: var(--muted-foreground); box-shadow: none; font-size: 12px; font-weight: 500; line-height: 1.5; white-space: normal; }
.portfolio-tab:hover { color: var(--foreground); }
.portfolio-tab[data-state=active] { border-color: var(--brand-accent); color: var(--brand-accent); background: transparent; box-shadow: none; }
.portfolio-tab:focus-visible { outline: 2px solid var(--ring); outline-offset: 2px; border-radius: 4px; }
.overview-grid { display: grid; min-width: 0; gap: 24px; padding: 20px; border: 1px solid var(--border); border-radius: 18px; background: var(--card); }
.performance-panel { min-width: 0; }
.performance-heading { display: flex; align-items: flex-start; justify-content: space-between; flex-wrap: wrap; gap: 14px; margin-bottom: 20px; }
.performance-heading > div:first-child { min-width: 0; flex: 1 1 200px; }
.overview-grid .portfolio-panel-title { font-size: 14px; font-weight: 500; }
.overview-grid .portfolio-panel-description { font-size: 11px; }
.chart-switch { display: flex; min-width: 0; gap: 2px; padding: 3px; border: 1px solid var(--border); border-radius: 10px; }
.chart-switch button { min-width: 0; padding-inline: 10px; border-radius: 7px; font-size: 11px; white-space: normal; }
.chart-switch button[aria-pressed=true] { color: var(--brand-accent); background: var(--brand-soft); }
.calculation-note { margin-top: 5px; color: var(--muted-foreground); font-size: 11px; line-height: 1.7; }
.calculation-note summary { padding-block: 12px; min-height: 44px; cursor: pointer; }
.calculation-note p { padding-bottom: 8px; }
.mix-panel { display: flex; flex-direction: column; min-width: 0; padding-top: 24px; border-top: 1px solid var(--border); }
.mix-rows { display: grid; align-content: start; gap: 20px; margin-top: 24px; }
.mix-row { min-width: 0; }
.mix-row > div:first-child { display: flex; justify-content: space-between; gap: 12px; font-size: 12px; }
.mix-row span { overflow-wrap: anywhere; }
.mix-row strong { flex-shrink: 0; font-weight: 500; font-variant-numeric: tabular-nums; }
.mix-track { height: 4px; margin: 9px 0 6px; background: var(--muted); border-radius: 4px; overflow: hidden; }
.mix-track > span { display: block; height: 100%; border-radius: inherit; background: var(--brand-accent); }
.mix-row p { color: var(--muted-foreground); font-size: 11px; overflow-wrap: anywhere; }
.mix-action { justify-content: space-between; margin-top: 20px; padding-inline: 0; color: var(--brand-accent); font-size: 12px; white-space: normal; }
.summary-loading { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 24px; padding: 24px 20px; border: 1px solid var(--border); border-radius: 18px; background: var(--card); }
.summary-loading > div { display: grid; gap: 16px; }
.summary-loading > div:first-child { grid-column: 1 / -1; }
.summary-loading span { display: block; height: 12px; border-radius: 5px; background: var(--muted); }
.summary-loading span + span { width: 70%; height: 24px; }
@container portfolio (min-width: 600px) {
  .portfolio-toolbar { grid-template-columns: 150px 220px auto; width: fit-content; max-width: 100%; }
  .refresh-button { padding-inline: 14px; }.refresh-button > span { display: inline; }
  .portfolio-tab-list { display: flex; flex-wrap: wrap; gap: 0 24px; }.portfolio-tab { flex: 0 1 auto; padding-inline: 0; }
}
@container portfolio (min-width: 740px) {
  .portfolio-summary { padding: 24px 28px 0; }
  .summary-main { grid-template-columns: minmax(0, 1.4fr) repeat(2, minmax(0, 1fr)); gap: 28px; }
  .balance-overview { grid-column: auto; }
  .headline-metric { padding-left: 24px; border-left: 1px solid var(--border); }
  .detail-metrics { grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 24px; }
  .summary-loading { grid-template-columns: 1.4fr 1fr 1fr; padding: 24px 28px; }.summary-loading > div:first-child { grid-column: auto; }
}
@container portfolio (min-width: 1000px) {
  .overview-grid { grid-template-columns: minmax(0, 2.4fr) minmax(0, 1fr); gap: 28px; padding: 24px 28px; }
  .mix-panel { padding-top: 0; padding-left: 28px; border-top: 0; border-left: 1px solid var(--border); }
}
@media (prefers-reduced-motion: reduce) { .summary-details > summary { transition: none; } }
</style>
