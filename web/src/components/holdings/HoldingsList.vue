<template>
  <section class="min-w-0 space-y-4" :aria-busy="loading">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div><h2 class="font-semibold">{{ t('accountDetail.holdings') }}</h2><p class="mt-1 text-xs text-muted-foreground">{{ t('holdings.listHint') }}</p></div>
      <Button data-testid="holdings-add" :disabled="!account.is_active" @click="$emit('add')"><LuPlus class="size-4" />{{ t('accountDetail.addHolding') }}</Button>
    </div>
    <div class="space-y-3 rounded-2xl border bg-card p-3 sm:p-4">
      <div class="flex flex-col gap-3 sm:flex-row">
        <div class="relative min-w-0 flex-1">
          <LuSearch class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
          <Input data-testid="holdings-search" :model-value="params.q" :aria-label="t('holdings.searchOwned')" :placeholder="t('holdings.searchOwned')" class="h-10 pl-9 pr-10" @update:model-value="setParam('q', $event)" />
          <Button v-if="params.q" data-testid="holdings-clear-search" type="button" variant="ghost" size="icon" class="absolute right-1 top-1/2 size-8 -translate-y-1/2" :aria-label="t('holdings.clearSearch')" @click="setParam('q', '')"><LuX class="size-4" /></Button>
        </div>
        <div class="flex gap-2">
          <FormSelect data-testid="holdings-sort" class="h-10 min-w-40" :aria-label="t('holdings.sort')" :model-value="`${params.sort}:${params.direction}`" :options="sortOptions" @update:model-value="setSort" />
          <Button data-testid="holdings-filter-toggle" type="button" variant="outline" class="h-10" :aria-expanded="showFilters" @click="showFilters = !showFilters"><LuSlidersHorizontal class="size-4" />{{ t('holdings.filters') }}<span v-if="filterCount" class="text-primary">{{ filterCount }}</span></Button>
        </div>
      </div>
      <div v-if="showFilters" data-testid="holdings-filters" class="grid gap-3 border-t pt-3 sm:grid-cols-2 lg:grid-cols-5">
        <div v-for="key in filterKeys" :key="key" class="min-w-0 space-y-2">
          <Label :for="`holdings-filter-${key}`">{{ t(`holdings.filterLabels.${key}`) }}</Label>
          <FormSelect :id="`holdings-filter-${key}`" :data-testid="`holdings-filter-${key}`" :model-value="params[key] || 'all'" :options="filterOptions(key)" @update:model-value="setParam(key, $event === 'all' ? '' : $event)" />
        </div>
      </div>
      <Button v-if="filterCount || params.q" data-testid="holdings-reset" variant="link" size="sm" class="h-auto p-0 text-xs" @click="resetFilters">{{ t('holdings.resetFilters') }}</Button>
    </div>
    <p v-if="source === 'database'" data-testid="holdings-search-fallback" role="status" class="text-xs text-muted-foreground">{{ t('holdings.searchFallback') }}</p>
    <div v-if="error" class="space-y-3"><Alert data-testid="holdings-list-error" variant="error" :description="error" /><Button data-testid="holdings-list-retry" variant="outline" @click="scheduleFetch">{{ t('common.retry') }}</Button></div>
    <div v-else-if="loading" data-testid="holdings-list-loading" class="flex min-h-40 items-center justify-center gap-2 text-sm text-muted-foreground" role="status"><LuLoaderCircle class="size-4 animate-spin" />{{ t('common.loading') }}</div>
    <div v-else-if="!rows.length" :data-testid="hasHoldings || params.q || filterCount ? 'holdings-no-results' : 'holdings-empty'" class="rounded-2xl border border-dashed py-12 text-center">
      <p class="text-sm text-muted-foreground">{{ t(hasHoldings || params.q || filterCount ? 'holdings.noMatchingHoldings' : 'accountDetail.noHoldingsYet') }}</p>
      <Button v-if="params.q || filterCount" variant="link" @click="resetFilters">{{ t('holdings.resetFilters') }}</Button>
    </div>
    <template v-else>
      <div data-testid="holdings-cards" class="grid gap-3 md:hidden">
        <Card v-for="holding in rows" :key="holding.id" :data-testid="`holding-card-${holding.id}`" class="gap-4 rounded-2xl p-4">
          <div class="flex items-start gap-2"><router-link class="min-w-0 flex-1 rounded-md focus-visible:outline-2 focus-visible:outline-ring" :to="{ name: 'HoldingDetail', params: { id: account.id, holdingId: holding.id } }"><AssetIdentity :asset="identity(holding)" compact /></router-link><RecordActions :view-to="{ name: 'HoldingDetail', params: { id: account.id, holdingId: holding.id } }" :data-testid="`holding-card-${holding.id}-actions`" @edit="$emit('edit', holding)" @delete="$emit('delete', holding)" /></div>
          <div class="grid grid-cols-2 gap-4 text-sm">
            <div v-for="field in metricFields" :key="field"><p class="text-xs text-muted-foreground">{{ t(`accountDetail.${field}`) }}</p><p class="mt-1 break-all font-medium tabular-nums" :class="field === 'profitLoss' && profitClass(holding)">{{ metric(holding, field) }}</p></div>
          </div>
        </Card>
      </div>
      <div class="hidden min-w-0 overflow-hidden rounded-2xl border bg-card md:block">
        <Table data-testid="holdings-table">
          <TableHeader><TableRow class="hover:bg-transparent"><TableHead class="px-4">{{ t('accountDetail.name') }}</TableHead><TableHead v-for="field in metricFields" :key="field" class="text-right">{{ t(`accountDetail.${field}`) }}</TableHead><TableHead class="px-4 text-right">{{ t('accountDetail.actions') }}</TableHead></TableRow></TableHeader>
          <TableBody><TableRow v-for="holding in rows" :key="holding.id" :data-testid="`holding-row-${holding.id}`" class="hover:bg-muted/30">
            <TableCell class="w-[28%] min-w-48 max-w-80 px-4 py-4"><router-link class="min-w-0 flex-1 rounded-md focus-visible:outline-2 focus-visible:outline-ring" :to="{ name: 'HoldingDetail', params: { id: account.id, holdingId: holding.id } }"><AssetIdentity :asset="identity(holding)" compact /></router-link></TableCell>
            <TableCell v-for="field in metricFields" :key="field" class="text-right tabular-nums" :class="field === 'profitLoss' && profitClass(holding)">{{ metric(holding, field) }}</TableCell>
            <TableCell class="px-4"><div class="flex justify-end"><RecordActions :view-to="{ name: 'HoldingDetail', params: { id: account.id, holdingId: holding.id } }" :data-testid="`holding-row-${holding.id}-actions`" @edit="$emit('edit', holding)" @delete="$emit('delete', holding)" /></div></TableCell>
          </TableRow></TableBody>
        </Table>
      </div>
    </template>
    <div v-if="!error && (total > 0 || params.page > 1)" data-testid="holdings-pagination" class="flex flex-wrap items-center justify-between gap-4 text-xs text-muted-foreground">
      <p aria-live="polite">{{ t('holdings.showing', { from: total ? (params.page - 1) * params.per_page + 1 : 0, to: Math.min(params.page * params.per_page, total), total }) }}</p>
      <div class="flex flex-wrap items-center gap-3">
        <div class="flex items-center gap-2"><Label for="holdings-page-size" class="text-xs font-normal">{{ t('holdings.perPage') }}</Label><FormSelect id="holdings-page-size" data-testid="holdings-page-size" class="w-20" :model-value="String(params.per_page)" :options="['10', '25', '50', '100']" :disabled="loading" @update:model-value="setParam('per_page', Number($event))" /></div>
        <nav class="flex items-center gap-1" :aria-label="t('holdings.pagination')">
          <Button data-testid="holdings-first" variant="ghost" size="icon" class="size-8" :disabled="loading || params.page <= 1" :aria-label="t('holdings.firstPage')" @click="params.page = 1"><LuChevronsLeft class="size-4" /></Button>
          <Button data-testid="holdings-prev" variant="ghost" size="icon" class="size-8" :disabled="loading || params.page <= 1" :aria-label="t('holdings.previousPage')" @click="params.page--"><LuChevronLeft class="size-4" /></Button>
          <span data-testid="holdings-page-number" class="min-w-14 text-center tabular-nums">{{ params.page }} / {{ pages }}</span>
          <Button data-testid="holdings-next" variant="ghost" size="icon" class="size-8" :disabled="loading || params.page >= pages" :aria-label="t('holdings.nextPage')" @click="params.page++"><LuChevronRight class="size-4" /></Button>
          <Button data-testid="holdings-last" variant="ghost" size="icon" class="size-8" :disabled="loading || params.page >= pages" :aria-label="t('holdings.lastPage')" @click="params.page = pages"><LuChevronsRight class="size-4" /></Button>
        </nav>
      </div>
    </div>
  </section>
</template>
<script setup>
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { LuSearch, LuSlidersHorizontal, LuPlus, LuX, LuLoaderCircle, LuChevronLeft, LuChevronRight, LuChevronsLeft, LuChevronsRight } from '@kalimahapps/vue-icons/lu'
import { Button, Card, FeedbackAlert as Alert, FormSelect, Input, Label, RecordActions, Table, TableHeader, TableHead, TableRow, TableBody, TableCell } from '@/components/ui'
import AssetIdentity from './AssetIdentity.vue'
import { holdingAPI } from '@/services/holdingApi'

const props = defineProps({ account: { type: Object, required: true }, hasHoldings: Boolean })
defineEmits(['add', 'edit', 'delete'])
const { t, locale } = useI18n()
const filterKeys = ['currency', 'exchange', 'asset_class', 'sector', 'country']
const metricFields = ['quantity', 'currentPrice', 'value', 'profitLoss', 'performance']
const params = reactive({ q: '', page: 1, per_page: 10, sort: 'created_at', direction: 'desc', currency: '', exchange: '', asset_class: '', sector: '', country: '' })
const rows = ref([]), total = ref(0), pages = ref(1), facets = ref({}), source = ref(''), error = ref(''), loading = ref(true), showFilters = ref(false)
let timer, recoveryTimer, revision = 0, active = true
const filterCount = computed(() => filterKeys.filter(key => params[key]).length)
const sortOptions = computed(() => ['created_at:desc', 'symbol:asc', 'symbol:desc', 'name:asc', 'quantity:desc', 'quantity:asc'].map(value => ({ value, label: t(`holdings.sortOptions.${value.replace(':', '_')}`) })))
function filterOptions(key) { return [{ value: 'all', label: t('holdings.all') }, ...[...new Set([...(facets.value[key] || []), params[key]])].filter(Boolean).map(value => ({ value, label: value }))] }
function setParam(key, value) { params[key] = value; params.page = 1 }
function setSort(value) { [params.sort, params.direction] = value.split(':'); params.page = 1 }
function resetFilters() { params.q = ''; for (const key of filterKeys) params[key] = ''; params.page = 1 }
function identity(holding) {
  const at = holding.symbol.lastIndexOf('.')
  return { Code: at < 0 ? holding.symbol : holding.symbol.slice(0, at), Exchange: at < 0 ? '' : holding.symbol.slice(at + 1), Currency: holding.buy_currency, Type: holding.asset_class, Name: holding.name && holding.name !== holding.symbol ? holding.name : '' }
}
function number(value, options = {}) { return new Intl.NumberFormat(locale.value, options).format(value) }
function metric(holding, field) {
  if (field === 'quantity') return number(holding.quantity, { maximumFractionDigits: 8 })
  if (!holding.price_available) return '—'
  if (field === 'performance') return Number.isFinite(holding.performance) ? `${number(Math.abs(holding.performance) < 0.005 ? 0 : holding.performance, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}%` : '—'
  const value = holding[{ currentPrice: 'current_price', value: 'value', profitLoss: 'profit_loss' }[field]]
  if (!Number.isFinite(value)) return '—'
  return number(Math.abs(value) < 0.005 ? 0 : value, { style: 'currency', currency: holding.buy_currency || props.account.currency })
}
function profitClass(holding) { return !holding.price_available || Math.abs(holding.profit_loss) < 0.005 ? 'text-muted-foreground' : holding.profit_loss < 0 ? 'text-destructive' : 'text-primary' }
function scheduleFetch() {
  clearTimeout(timer); clearTimeout(recoveryTimer)
  const request = ++revision
  loading.value = true; error.value = ''
  timer = setTimeout(() => fetchPage(request), params.q ? 250 : 0)
}
async function fetchPage(request) {
  try {
    const result = await holdingAPI.searchHoldings(props.account.id, { ...params })
    if (!active || request !== revision) return
    pages.value = result.pages; total.value = result.total
    if (params.page > result.pages) { params.page = result.pages; return }
    rows.value = result.holdings; source.value = result.source
    facets.value = { ...facets.value, ...result.facets }
    if (result.source === 'database') recoveryTimer = setTimeout(() => fetchPage(++revision), 10000)
  } catch (e) {
    if (active && request === revision) { error.value = e.message; rows.value = []; total.value = 0 }
  } finally { if (active && request === revision) loading.value = false }
}
watch([() => props.account.id, params], scheduleFetch, { deep: true, immediate: true })
onBeforeUnmount(() => { active = false; revision++; clearTimeout(timer); clearTimeout(recoveryTimer) })
</script>
