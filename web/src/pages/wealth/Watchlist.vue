<template>
  <div class="space-y-6" data-testid="watchlist-page">
    <SectionHeader :icon="LuEye" :title="t('wealth.watchlist')" :description="t('wealth.watchlistHint')" />
    <PageFeedback :error="error" :busy="busy" :notice="notice" />

    <!-- Step 1: find an asset to follow. -->
    <div class="rounded-2xl border bg-card p-4 shadow-xs sm:p-5">
      <form class="flex flex-col gap-3 sm:flex-row" @submit.prevent="search">
        <label class="relative flex-1">
          <span class="sr-only">{{ t('holdings.searchAsset') }}</span>
          <LuSearch class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
          <Input v-model="query" data-testid="watchlist-search" class="h-11 pl-10" minlength="2" maxlength="100" required :placeholder="t('wealth.watchSearchPlaceholder')" :disabled="busy" />
        </label>
        <Button type="submit" class="h-11" :disabled="busy"><LuSearch class="size-4" aria-hidden="true" />{{ t('wealth.search') }}</Button>
      </form>
      <div v-if="results.length" data-testid="watchlist-results" class="mt-4 overflow-hidden rounded-xl border">
        <div class="flex items-center justify-between border-b bg-muted/40 px-4 py-2 text-xs text-muted-foreground">
          <span>{{ t('holdings.resultCount', { count: results.length }) }}</span>
          <Button variant="ghost" size="sm" class="h-7" @click="results = []"><LuX class="size-3.5" aria-hidden="true" />{{ t('holdings.clearSearch') }}</Button>
        </div>
        <ul class="max-h-80 divide-y overflow-y-auto">
          <li v-for="asset in results" :key="`${asset.Code}.${asset.Exchange}`" class="flex items-center gap-3 px-4 py-3">
            <div class="min-w-0 flex-1"><AssetIdentity :asset="asset" compact /></div>
            <span v-if="watched.has(`${asset.Code}.${asset.Exchange}`)" class="inline-flex items-center gap-1 text-xs font-medium text-emerald-700 dark:text-emerald-400"><LuCheck class="size-3.5" aria-hidden="true" />{{ t('wealth.watching') }}</span>
            <Button v-else size="sm" variant="outline" :disabled="busy" @click="add(asset)"><LuPlus class="size-3.5" aria-hidden="true" />{{ t('wealth.addWatch') }}</Button>
          </li>
        </ul>
      </div>
    </div>

    <!-- Step 2: the followed assets and their latest prices. -->
    <div data-testid="watchlist-list" class="overflow-hidden rounded-2xl border bg-card shadow-xs">
      <div class="flex flex-wrap items-center gap-3 border-b p-3 sm:px-4">
        <label class="relative min-w-48 flex-1">
          <span class="sr-only">{{ t('wealth.filter') }}</span>
          <LuListFilter class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
          <Input v-model="filter" class="h-9 pl-9" :placeholder="t('wealth.filter')" @input="page = 1" />
        </label>
        <Button variant="outline" size="sm" :disabled="busy || !visible.length" @click="refreshQuotes"><LuRefreshCw class="size-4" :class="busy && 'animate-spin'" aria-hidden="true" />{{ t('wealth.refreshQuotes') }}</Button>
      </div>
      <template v-if="visible.length">
        <div class="hidden border-b bg-muted/40 px-4 py-2.5 text-xs font-medium uppercase tracking-wide text-muted-foreground sm:grid sm:grid-cols-[minmax(0,1fr)_9rem_7rem_6rem] sm:gap-4" aria-hidden="true">
          <span>{{ t('wealth.asset') }}</span><span class="text-right">{{ t('wealth.lastPrice') }}</span><span class="text-right">{{ t('wealth.dailyChange') }}</span><span />
        </div>
        <ul role="list" class="divide-y">
          <li v-for="item in visible" :key="item.id" :data-testid="`watch-row-${item.id}`" class="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-x-4 gap-y-2 px-4 py-3 transition-colors hover:bg-muted/30 sm:grid-cols-[minmax(0,1fr)_9rem_7rem_6rem]">
            <AssetIdentity :asset="asAsset(item)" compact />
            <div class="text-right">
              <p class="font-semibold tabular-nums">{{ quotes[item.symbol] ? money(quotes[item.symbol].close, item.currency) : '—' }}</p>
              <p class="text-xs text-muted-foreground">{{ quotes[item.symbol]?.timestamp ? quoteDate(quotes[item.symbol].timestamp) : t('wealth.notRefreshed') }}</p>
            </div>
            <p class="flex items-center gap-1 tabular-nums sm:justify-end" :class="changeClass(item)">
              <component :is="changeIcon(item)" v-if="changeIcon(item)" class="size-3.5" aria-hidden="true" />{{ change(item) }}
            </p>
            <div class="flex justify-end">
              <Button size="sm" variant="ghost" class="text-destructive hover:bg-destructive/10 hover:text-destructive" :disabled="busy" @click="remove(item)"><LuTrash2 class="size-3.5" aria-hidden="true" />{{ t('wealth.unwatch') }}</Button>
            </div>
          </li>
        </ul>
        <ListPager v-model:page="page" :pages="pages" :total="filtered.length" :size="PAGE_SIZE" />
      </template>
      <EmptyState v-else :icon="LuEye" :title="t(items.length ? 'common.noResults' : 'wealth.emptyWatchlistTitle')" :description="items.length ? '' : t('wealth.emptyWatchlist')" />
    </div>
    <p class="flex items-start gap-2 text-xs text-muted-foreground"><LuInfo class="mt-0.5 size-3.5 shrink-0" aria-hidden="true" />{{ t('wealth.quoteHint') }}</p>
  </div>
</template>
<script setup>
import { computed, onMounted, ref } from 'vue'
import { LuCheck, LuEye, LuInfo, LuListFilter, LuPlus, LuRefreshCw, LuSearch, LuTrash2, LuTrendingDown, LuTrendingUp, LuX } from '@kalimahapps/vue-icons/lu'
import { Button, Input } from '@/components/ui'
import AssetIdentity from '@/components/holdings/AssetIdentity.vue'
import PageFeedback from '@/components/wealth/PageFeedback.vue'
import SectionHeader from '@/components/wealth/SectionHeader.vue'
import EmptyState from '@/components/wealth/EmptyState.vue'
import ListPager from '@/components/wealth/ListPager.vue'
import { wealthAPI } from '@/services/wealthApi'
import { marketAPI } from '@/services/marketApi'
import { useWealthPage } from '@/composables/useWealthPage'
const PAGE_SIZE = 20
const { t, locale, busy, error, notice, run, money } = useWealthPage()
const items = ref([]), results = ref([]), quotes = ref({}), query = ref(''), filter = ref(''), page = ref(1)
const watched = computed(() => new Set(items.value.map(item => item.symbol)))
const filtered = computed(() => items.value.filter(a => `${a.symbol} ${a.name}`.toLowerCase().includes(filter.value.toLowerCase())))
const pages = computed(() => Math.max(1, Math.ceil(filtered.value.length / PAGE_SIZE)))
const visible = computed(() => filtered.value.slice((page.value - 1) * PAGE_SIZE, page.value * PAGE_SIZE))
const asAsset = item => ({ Code: item.symbol.split('.')[0], Exchange: item.symbol.split('.').slice(1).join('.'), Name: item.name, Currency: item.currency, Type: item.asset_class })
const quoteDate = seconds => new Intl.DateTimeFormat(locale.value, { dateStyle: 'medium' }).format(new Date(seconds * 1000))
const changeOf = item => quotes.value[item.symbol]?.change_p
function change(item) {
  const value = changeOf(item)
  return Number.isFinite(value) ? new Intl.NumberFormat(locale.value, { style: 'percent', maximumFractionDigits: 2, signDisplay: 'exceptZero' }).format(value / 100) : '—'
}
const changeIcon = item => (changeOf(item) > 0 ? LuTrendingUp : changeOf(item) < 0 ? LuTrendingDown : null)
const changeClass = item => (changeOf(item) > 0 ? 'text-emerald-600 dark:text-emerald-400' : changeOf(item) < 0 ? 'text-destructive' : 'text-muted-foreground')
async function load() { items.value = (await wealthAPI.get('/watchlist')).data; page.value = Math.min(page.value, pages.value) }
function search() { run(async () => { results.value = []; results.value = (await marketAPI.searchAssets(query.value.trim(), 20)).data }) }
function add(a) { run(async () => { await wealthAPI.save('/watchlist', { symbol: `${a.Code}.${a.Exchange}`, name: a.Name, currency: a.Currency, asset_class: a.Type }); await load() }, true) }
function remove(item) { run(async () => { await wealthAPI.remove(`/watchlist/${item.id}`); await load() }) }
function refreshQuotes() { run(async () => { const data = await marketAPI.getMultipleLivePrices(visible.value.map(a => a.symbol)); quotes.value = { ...quotes.value, ...Object.fromEntries(data.data.map(a => [a.code, a])) } }) }
onMounted(() => run(load))
</script>
