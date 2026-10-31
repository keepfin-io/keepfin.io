<template>
  <TabsContent value="holdings" class="min-w-0">
    <section class="portfolio-panel" aria-labelledby="holdings-heading">
      <div class="panel-heading"><div><h2 id="holdings-heading" class="portfolio-panel-title">{{ $t('portfolio.holdings') }}</h2><p class="portfolio-panel-description">{{ $t('portfolio.holdingsHint') }}</p></div><span class="count-badge">{{ report.holdings.length }}</span></div>
      <div v-if="!report.holdings.length" class="portfolio-empty"><Layers3 /><p>{{ $t('portfolio.noHoldings') }}</p></div>
      <template v-else>
        <div class="mobile-records" data-testid="portfolio-holdings-mobile">
          <article v-for="h in report.holdings" :key="`${h.symbol}-${h.currency}`" class="record-card">
            <div class="record-heading"><div class="min-w-0"><h3>{{ h.symbol }}</h3><p>{{ h.currency }} · {{ h.price_date || '—' }}</p></div><span class="record-value">{{ h.complete ? money(h.value) : '—' }}</span></div>
            <p class="record-accounts">{{ h.accounts.join(', ') }}</p>
            <dl class="record-metrics"><div><dt>{{ $t('portfolio.quantity') }}</dt><dd>{{ number(h.quantity, 8) }}</dd></div><div><dt>{{ $t('portfolio.weight') }}</dt><dd>{{ report.complete ? percent(h.weight) : '—' }}</dd></div><div><dt>{{ $t('portfolio.cost') }}</dt><dd>{{ h.complete ? money(h.cost) : '—' }}</dd></div><div><dt>{{ $t('portfolio.profit_loss') }}</dt><dd :class="h.complete && tone(h.profit_loss)">{{ h.complete ? money(h.profit_loss) : '—' }}</dd></div></dl>
          </article>
        </div>
        <div class="desktop-records" tabindex="0" role="region" :aria-label="$t('portfolio.holdings')">
          <Table data-testid="portfolio-holdings"><TableHeader><TableRow><TableHead v-for="key in ['symbol', 'accounts', 'quantity', 'value', 'cost', 'profit_loss', 'weight']" :key="key">{{ $t(`portfolio.${key}`) }}</TableHead></TableRow></TableHeader><TableBody><TableRow v-for="h in report.holdings" :key="`${h.symbol}-${h.currency}`"><TableCell class="font-medium">{{ h.symbol }}<p class="mt-1 text-xs text-muted-foreground">{{ h.currency }} · {{ h.price_date || '—' }}</p></TableCell><TableCell class="max-w-48 whitespace-normal break-words text-xs text-muted-foreground">{{ h.accounts.join(', ') }}</TableCell><TableCell>{{ number(h.quantity, 8) }}</TableCell><TableCell>{{ h.complete ? money(h.value) : '—' }}</TableCell><TableCell>{{ h.complete ? money(h.cost) : '—' }}</TableCell><TableCell :class="h.complete && tone(h.profit_loss)">{{ h.complete ? money(h.profit_loss) : '—' }}</TableCell><TableCell>{{ report.complete ? percent(h.weight) : '—' }}</TableCell></TableRow></TableBody></Table>
        </div>
      </template>
    </section>
  </TabsContent>

  <TabsContent value="income" class="min-w-0">
    <section class="portfolio-panel" aria-labelledby="income-heading">
      <div class="panel-heading"><div><h2 id="income-heading" class="portfolio-panel-title">{{ $t('portfolio.income') }}</h2><p class="portfolio-panel-description">{{ $t('portfolio.incomeHint') }}</p></div><div class="panel-filter"><Label for="portfolio-income-period" class="sr-only">{{ $t('portfolio.period') }}</Label><FormSelect id="portfolio-income-period" data-testid="portfolio-income-period" class="portfolio-select" v-model="incomeMode" :options="[{ value: 'month', label: $t('portfolio.monthly') }, { value: 'year', label: $t('portfolio.yearly') }]" /></div></div>
      <div v-if="!incomeRows.length" class="portfolio-empty"><Coins /><p>{{ $t('portfolio.noIncome') }}</p></div>
      <template v-else>
        <div class="mobile-records" data-testid="portfolio-income-mobile">
          <article v-for="row in incomeRows" :key="row.period" class="record-card"><div class="record-heading"><h3>{{ row.period }}</h3><span :class="['record-value', report.complete && tone(row.net)]">{{ report.complete ? money(row.net) : '—' }}</span></div><dl class="record-metrics"><div v-for="key in ['dividends', 'interest', 'fees', 'net']" :key="key"><dt>{{ $t(`portfolio.${key}`) }}</dt><dd>{{ report.complete ? money(row[key]) : '—' }}</dd></div></dl></article>
        </div>
        <div class="desktop-records" tabindex="0" role="region" :aria-label="$t('portfolio.income')">
          <Table data-testid="portfolio-income"><TableHeader><TableRow><TableHead v-for="key in ['period', 'dividends', 'interest', 'fees', 'net']" :key="key">{{ $t(`portfolio.${key}`) }}</TableHead></TableRow></TableHeader><TableBody><TableRow v-for="row in incomeRows" :key="row.period"><TableCell class="font-medium">{{ row.period }}</TableCell><TableCell v-for="key in ['dividends', 'interest', 'fees', 'net']" :key="key" :class="key === 'net' && report.complete && tone(row.net)">{{ report.complete ? money(row[key]) : '—' }}</TableCell></TableRow></TableBody></Table>
        </div>
      </template>
    </section>
  </TabsContent>

  <TabsContent value="allocation" class="min-w-0">
    <section class="portfolio-panel" aria-labelledby="allocation-heading">
      <div class="panel-heading"><div><h2 id="allocation-heading" class="portfolio-panel-title">{{ $t('portfolio.allocation') }}</h2><p class="portfolio-panel-description">{{ $t('portfolio.allocationHint') }}</p></div><div class="panel-filter"><Label for="portfolio-allocation-kind" class="sr-only">{{ $t('portfolio.allocation') }}</Label><FormSelect id="portfolio-allocation-kind" data-testid="portfolio-allocation-kind" class="portfolio-select" v-model="allocationKind" :options="['account', 'asset_class', 'currency', 'sector', 'country'].map(value => ({ value, label: $t(`portfolio.${value}`) }))" /></div></div>
      <div v-if="!report.complete" class="portfolio-empty"><ChartPie /><p>{{ $t('portfolio.incomplete') }}</p></div>
      <div v-else-if="!allocations.length" class="portfolio-empty"><ChartPie /><p>{{ $t('portfolio.noAllocations') }}</p></div>
      <div v-else class="allocation-grid" data-testid="portfolio-allocations"><article v-for="(row, index) in allocations" :key="row.name" class="allocation-row"><div class="allocation-label"><span><span class="allocation-dot" :style="{ opacity: 1 - (index % 4) * 0.18 }" />{{ row.name }}</span><strong>{{ percent(row.weight) }}</strong></div><p>{{ money(row.value) }}</p><div class="allocation-track" aria-hidden="true"><span :style="{ width: `${Math.max(0, Math.min(100, row.weight))}%`, opacity: 1 - (index % 4) * 0.18 }" /></div></article></div>
    </section>
  </TabsContent>
</template>
<script setup>
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ChartPie, Coins, Layers3 } from 'lucide-vue-next'
import { FormSelect, Label, TabsContent, Table, TableHeader, TableHead, TableRow, TableBody, TableCell } from '@/components/ui'
import { formatMoney, formatNumber, formatPercent } from '@/lib/portfolio-format'
const props = defineProps({ report: { type: Object, required: true } })
const { locale } = useI18n()
const incomeMode = ref('month'), allocationKind = ref('asset_class')
const allocations = computed(() => props.report.allocations[allocationKind.value] || [])
const incomeRows = computed(() => {
  if (incomeMode.value === 'month') return props.report.income
  const years = new Map()
  for (const row of props.report.income) {
    const year = row.period.slice(0, 4)
    const total = years.get(year) || { period: year, dividends: 0, interest: 0, fees: 0, net: 0 }
    for (const key of ['dividends', 'interest', 'fees', 'net']) total[key] += row[key]
    years.set(year, total)
  }
  return [...years.values()]
})
function money(v) { return formatMoney(v, props.report.currency, locale.value) }
function number(v, digits = 2) { return formatNumber(v, locale.value, digits) }
function percent(v) { return formatPercent(v, locale.value) }
function tone(v) { return v < 0 ? 'portfolio-negative' : 'portfolio-positive' }
</script>
<style scoped>
.panel-heading { display: flex; align-items: flex-start; justify-content: space-between; flex-wrap: wrap; gap: 16px; margin-bottom: 24px; }
.panel-heading > div:first-child { min-width: 0; flex: 1 1 220px; }.panel-filter { flex: 0 1 180px; min-width: 0; }
.count-badge { display: grid; place-items: center; min-width: 28px; height: 28px; border-radius: 9px; background: var(--brand-soft); color: var(--brand-accent); font-size: 12px; font-weight: 600; }
.mobile-records { display: grid; gap: 12px; }.desktop-records { display: none; min-width: 0; overflow-x: auto; border-radius: 10px; }
.record-card { min-width: 0; padding: 16px; border: 1px solid var(--border); border-radius: 16px; }
.record-heading { display: flex; align-items: flex-start; justify-content: space-between; flex-wrap: wrap; gap: 12px; }.record-heading h3 { font-size: 14px; font-weight: 600; overflow-wrap: anywhere; }
.record-heading p, .record-accounts { margin-top: 5px; font-size: 11px; color: var(--muted-foreground); overflow-wrap: anywhere; line-height: 1.6; }.record-accounts { margin-top: 12px; }
.record-value { max-width: 100%; font-size: 16px; font-weight: 600; font-variant-numeric: tabular-nums; overflow-wrap: anywhere; }
.record-metrics { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 16px 12px; margin-top: 16px; padding-top: 16px; border-top: 1px solid var(--border); }
.record-metrics dt { font-size: 10px; color: var(--muted-foreground); line-height: 1.5; }.record-metrics dd { margin-top: 5px; font-size: 12px; font-weight: 500; font-variant-numeric: tabular-nums; overflow-wrap: anywhere; }
.allocation-grid { display: grid; gap: 22px; }.allocation-row { min-width: 0; }.allocation-label { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; font-size: 13px; }
.allocation-label > span { display: flex; align-items: center; gap: 8px; min-width: 0; overflow-wrap: anywhere; }.allocation-label strong { flex-shrink: 0; font-weight: 500; font-variant-numeric: tabular-nums; }
.allocation-dot { width: 7px; height: 7px; flex-shrink: 0; border-radius: 50%; background: var(--brand-accent); }.allocation-row p { margin: 6px 0 12px; color: var(--muted-foreground); font-size: 12px; overflow-wrap: anywhere; }
.allocation-track { height: 6px; overflow: hidden; border-radius: 6px; background: var(--muted); }.allocation-track > span { display: block; height: 100%; border-radius: inherit; background: var(--brand-accent); }
@container portfolio (min-width: 640px) { .mobile-records { display: none; }.desktop-records { display: block; }.allocation-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 28px 40px; } }
</style>
