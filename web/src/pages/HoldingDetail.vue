<template>
  <section data-testid="holding-detail-page" class="min-w-0 space-y-6">
    <Button as-child variant="ghost" class="px-0 text-muted-foreground"><router-link data-testid="holding-detail-back" :to="{ name: 'AccountDetail', params: { id: account.id } }"><ArrowLeft class="size-4" />{{ t('accountDetail.backToAccount') }}</router-link></Button>
    <div v-if="holdingsError" class="space-y-3">
      <Alert data-testid="holding-detail-error" variant="error" :description="t('holdingDetail.loadError')" />
      <Button data-testid="holding-detail-retry" variant="outline" @click="$emit('retry')">{{ t('common.retry') }}</Button>
    </div>
    <Alert v-else-if="!holding" data-testid="holding-detail-not-found" variant="warning" :description="t('holdingDetail.notFound')" />
    <template v-else>
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div class="min-w-0 space-y-3">
          <h1 class="text-2xl font-semibold tracking-tight">{{ t('holdingDetail.title') }}</h1>
          <div data-testid="holding-detail-identity"><AssetIdentity :asset="identity" detailed /></div>
        </div>
        <Button v-if="account.is_active" as-child variant="outline"><router-link data-testid="holding-detail-edit" :to="{ name: 'HoldingEdit', params: { id: account.id, holdingId: holding.id } }"><Pencil class="size-4" />{{ t('common.edit') }}</router-link></Button>
      </div>
      <Alert v-if="!holding.price_available" data-testid="holding-detail-price-unavailable" variant="warning" :description="t('holdingDetail.priceUnavailable')" />
      <section class="min-w-0 rounded-2xl border bg-card p-5 sm:p-6" :aria-label="t('holdingDetail.position')">
        <div class="flex flex-wrap items-center justify-between gap-2"><h2 class="font-semibold">{{ t('holdingDetail.position') }}</h2><span class="text-xs text-muted-foreground">{{ holding.buy_currency }}</span></div>
        <dl class="mt-5 grid grid-cols-2 gap-6 lg:grid-cols-4">
          <div v-for="metric in metrics" :key="metric.key" :data-testid="`holding-detail-${metric.key}`" class="min-w-0">
            <dt class="flex items-center gap-2 text-xs text-muted-foreground"><component :is="metric.icon" class="size-4 shrink-0 text-primary/80" aria-hidden="true" />{{ metric.label }}</dt>
            <dd class="mt-2 break-all text-xl font-semibold tabular-nums" :class="metric.tone">{{ metric.value }}</dd>
          </div>
        </dl>
        <dl class="mt-6 grid grid-cols-2 gap-4 border-t pt-4 text-xs sm:grid-cols-4">
          <div><dt class="text-muted-foreground">{{ t('accountDetail.performance') }}</dt><dd data-testid="holding-detail-performance" class="mt-1 font-medium tabular-nums">{{ holding.price_available ? formatPercent(holding.performance, locale) : '—' }}</dd></div>
          <div><dt class="text-muted-foreground">{{ t('holdingDetail.firstActivity') }}</dt><dd class="mt-1">{{ date(holding.first_activity) }}</dd></div>
          <div><dt class="text-muted-foreground">{{ t('holdingDetail.priceDate') }}</dt><dd class="mt-1">{{ holding.price_available ? date(holding.price_date) : '—' }}</dd></div>
          <div><dt class="text-muted-foreground">{{ t('portfolio.sector') }}</dt><dd class="mt-1 break-words">{{ holding.sector || t('portfolio.unclassified') }}</dd></div>
        </dl>
        <p class="mt-4 text-xs leading-relaxed text-muted-foreground">{{ t('holdingDetail.positionHint') }}</p>
      </section>
      <HoldingPriceHistory :key="holding.symbol" :symbol="holding.symbol" :currency="holding.buy_currency" />
      <HoldingActivityHistory :account="account" :activities="holdingActivities" :error="activitiesError" @retry="$emit('retry')" />
    </template>
  </section>
</template>
<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ArrowLeft, Pencil } from 'lucide-vue-next'
import { LuLayers, LuTrendingUp, LuWallet, LuChartNoAxesCombined } from '@kalimahapps/vue-icons/lu'
import { Button, FeedbackAlert as Alert } from '@/components/ui'
import AssetIdentity from '@/components/holdings/AssetIdentity.vue'
import HoldingPriceHistory from '@/components/holdings/HoldingPriceHistory.vue'
import HoldingActivityHistory from '@/components/holdings/HoldingActivityHistory.vue'
import { formatMoney, formatNumber, formatPercent } from '@/lib/portfolio-format'

defineOptions({ inheritAttrs: false })
const props = defineProps({ account: { type: Object, required: true }, holdings: { type: Array, default: () => [] }, activities: { type: Array, default: () => [] }, holdingsError: { type: String, default: '' }, activitiesError: { type: String, default: '' } })
defineEmits(['retry'])
const route = useRoute()
const { t, locale } = useI18n()
const holding = computed(() => props.holdings.find(row => row.id === route.params.holdingId))
const holdingActivities = computed(() => props.activities.filter(row => holding.value && (row.holding_id === holding.value.id || row.symbol === holding.value.symbol)))
const identity = computed(() => {
  const row = holding.value
  const dot = row.symbol.lastIndexOf('.')
  return { Code: dot < 0 ? row.symbol : row.symbol.slice(0, dot), Exchange: dot < 0 ? '' : row.symbol.slice(dot + 1), Name: row.name, Currency: row.buy_currency, Type: row.asset_class, Country: row.country }
})
const metrics = computed(() => {
  const row = holding.value
  if (!row) return []
  const money = value => row.price_available ? formatMoney(value, row.buy_currency, locale.value) : '—'
  return [
    { key: 'quantity', label: t('portfolio.quantity'), icon: LuLayers, value: formatNumber(row.quantity, locale.value, 8) },
    { key: 'price', label: t('accountDetail.currentPrice'), icon: LuChartNoAxesCombined, value: money(row.current_price) },
    { key: 'value', label: t('accountDetail.value'), icon: LuWallet, value: money(row.value) },
    { key: 'profit-loss', label: t('portfolio.unrealized'), icon: LuTrendingUp, value: money(row.profit_loss), tone: row.price_available && row.profit_loss !== 0 ? (row.profit_loss < 0 ? 'portfolio-negative' : 'portfolio-positive') : '' },
  ]
})
function date(value) { return value && !value.startsWith('0001-') ? value.slice(0, 10) : '—' }
</script>
