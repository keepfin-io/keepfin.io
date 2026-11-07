<template>
  <section data-testid="holding-price-history" class="min-w-0 rounded-2xl border bg-card p-5 sm:p-6" :aria-busy="loading" aria-labelledby="holding-history-heading">
    <div class="mb-5 flex flex-wrap items-start justify-between gap-4">
      <div class="min-w-0">
        <h2 id="holding-history-heading" class="font-semibold">{{ t('holdingDetail.priceHistory') }}</h2>
        <p class="mt-1 text-xs leading-relaxed text-muted-foreground">{{ t('holdingDetail.priceHistoryHint', { currency }) }}</p>
      </div>
      <div class="flex items-center gap-2">
        <FormSelect data-testid="holding-history-range" class="w-32" v-model="range" :aria-label="t('portfolio.range')" :options="['1m', '3m', '1y', '5y'].map(value => ({ value, label: t(`portfolio.ranges.${value}`) }))" />
        <Button data-testid="holding-history-refresh" size="icon" variant="outline" :disabled="loading" :aria-label="t('portfolio.refresh')" @click="load"><RefreshCw class="size-4" :class="loading && 'animate-spin motion-reduce:animate-none'" /></Button>
      </div>
    </div>
    <div v-if="loading" data-testid="holding-history-loading" class="flex min-h-64 items-center justify-center text-sm text-muted-foreground" role="status">{{ t('common.loading') }}</div>
    <div v-else-if="errorCode" class="space-y-3">
      <Alert data-testid="holding-history-error" variant="warning" :description="errorMessage" />
      <Button data-testid="holding-history-retry" variant="outline" @click="load">{{ t('common.retry') }}</Button>
    </div>
    <PortfolioChart v-else :points="points" mode="price" :currency="currency" />
  </section>
</template>
<script setup>
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { RefreshCw } from 'lucide-vue-next'
import { Button, FormSelect, FeedbackAlert as Alert } from '@/components/ui'
import PortfolioChart from '@/components/portfolio/PortfolioChart.vue'
import { marketAPI } from '@/services/marketApi'

const props = defineProps({ symbol: { type: String, required: true }, currency: { type: String, required: true } })
const { t } = useI18n()
const range = ref('1m'), points = ref([]), loading = ref(false), errorCode = ref('')
const errorMessage = computed(() => t({ market_daily_limit_reached: 'holdingDetail.historyDailyLimit', market_rate_limited: 'holdingDetail.historyRateLimit' }[errorCode.value] || 'holdingDetail.historyError'))
let request = 0

async function load() {
  const current = ++request
  loading.value = true
  points.value = []
  errorCode.value = ''
  const to = new Date()
  const from = new Date(to)
  const day = from.getUTCDate()
  from.setUTCDate(1)
  from.setUTCMonth(from.getUTCMonth() - { '1m': 1, '3m': 3, '1y': 12, '5y': 60 }[range.value])
  const lastDay = new Date(Date.UTC(from.getUTCFullYear(), from.getUTCMonth() + 1, 0)).getUTCDate()
  from.setUTCDate(Math.min(day, lastDay))
  try {
    const result = await marketAPI.getEODData(props.symbol, { from: from.toISOString().slice(0, 10), to: to.toISOString().slice(0, 10), period: 'd', order: 'a' })
    if (current !== request) return
    // Keep missing prices as gaps. These are raw closing prices, not portfolio returns.
    points.value = (result.data || []).filter(row => row.date).map(row => ({ date: row.date, price: Number.isFinite(row.close) ? row.close : null })).sort((a, b) => a.date.localeCompare(b.date))
  } catch (error) {
    if (current === request) errorCode.value = error.code || 'unavailable'
  } finally {
    if (current === request) loading.value = false
  }
}

watch([() => props.symbol, range], load, { immediate: true })
onBeforeUnmount(() => { request++ })
</script>
