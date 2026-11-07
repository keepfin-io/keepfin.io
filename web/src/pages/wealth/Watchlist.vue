<template>
  <div class="space-y-5" data-testid="watchlist-page">
    <h2 class="text-xl font-semibold">{{ t('wealth.watchlist') }}</h2>
    <PageFeedback :error="error" :busy="busy" :notice="notice" />
    <form class="flex flex-wrap items-end gap-3" @submit.prevent="search">
      <label class="grid min-w-56 flex-1 gap-2 text-sm">{{ t('holdings.searchAsset') }}<Input v-model="query" minlength="2" maxlength="100" required :disabled="busy" /></label>
      <Button type="submit" :disabled="busy">{{ t('wealth.search') }}</Button>
    </form>
    <div v-if="results.length" class="divide-y rounded-xl border">
      <div v-for="asset in results" :key="`${asset.Code}.${asset.Exchange}`" class="flex items-center gap-3 p-3"><AssetIdentity :asset="asset" /><Button class="ml-auto" size="sm" variant="outline" :disabled="busy" @click="add(asset)">{{ t('wealth.addWatch') }}</Button></div>
    </div>
    <div class="flex flex-wrap items-center gap-3"><label class="flex-1"><span class="sr-only">{{ t('wealth.filter') }}</span><Input v-model="filter" :placeholder="t('wealth.filter')" @input="page = 1" /></label><Button variant="outline" :disabled="busy || !visible.length" @click="refreshQuotes">{{ t('wealth.refreshQuotes') }}</Button></div>
    <p class="text-xs text-muted-foreground">{{ t('wealth.quoteHint') }}</p>
    <div v-if="visible.length" class="divide-y rounded-xl border">
      <div v-for="item in visible" :key="item.id" class="flex flex-wrap items-center gap-3 p-4">
        <AssetIdentity :asset="asAsset(item)" />
        <div class="ml-auto text-right text-sm"><p>{{ money(quotes[item.symbol]?.close, item.currency) }}</p><p v-if="quotes[item.symbol]?.timestamp" class="text-xs text-muted-foreground">{{ new Date(quotes[item.symbol].timestamp * 1000).toISOString().slice(0, 10) }}</p></div>
        <Button size="sm" variant="ghost" :disabled="busy" @click="remove(item)">{{ t('common.delete') }}</Button>
      </div>
    </div>
    <p v-else class="py-8 text-sm text-muted-foreground">{{ t('wealth.emptyWatchlist') }}</p>
    <div class="flex items-center justify-end gap-3 text-sm"><Button variant="outline" size="sm" :disabled="page <= 1" @click="page--">{{ t('wealth.previous') }}</Button><span>{{ page }} / {{ pages }}</span><Button variant="outline" size="sm" :disabled="page >= pages" @click="page++">{{ t('wealth.next') }}</Button></div>
  </div>
</template>
<script setup>
import { computed, onMounted, ref } from 'vue'
import { Button, Input } from '@/components/ui'
import AssetIdentity from '@/components/holdings/AssetIdentity.vue'
import PageFeedback from '@/components/wealth/PageFeedback.vue'
import { wealthAPI } from '@/services/wealthApi'
import { marketAPI } from '@/services/marketApi'
import { useWealthPage } from '@/composables/useWealthPage'
const { t, busy, error, notice, run, money } = useWealthPage()
const items = ref([]), results = ref([]), quotes = ref({}), query = ref(''), filter = ref(''), page = ref(1)
const filtered = computed(() => items.value.filter(a => `${a.symbol} ${a.name}`.toLowerCase().includes(filter.value.toLowerCase())))
const pages = computed(() => Math.max(1, Math.ceil(filtered.value.length / 20)))
const visible = computed(() => filtered.value.slice((page.value - 1) * 20, page.value * 20))
const asAsset = item => ({ Code: item.symbol.split('.')[0], Exchange: item.symbol.split('.').slice(1).join('.'), Name: item.name, Currency: item.currency, Type: item.asset_class })
async function load() { items.value = (await wealthAPI.get('/watchlist')).items; page.value = Math.min(page.value, pages.value) }
function search() { run(async () => { results.value = []; results.value = (await marketAPI.searchAssets(query.value.trim(), 20)).results }) }
function add(a) { run(async () => { await wealthAPI.save('/watchlist', { symbol: `${a.Code}.${a.Exchange}`, name: a.Name, currency: a.Currency, asset_class: a.Type }); await load() }, true) }
function remove(item) { run(async () => { await wealthAPI.remove(`/watchlist/${item.id}`); await load() }) }
function refreshQuotes() { run(async () => { quotes.value = {}; const data = await marketAPI.getMultipleLivePrices(visible.value.map(a => a.symbol)); quotes.value = Object.fromEntries(data.data.map(a => [a.code, a])) }) }
onMounted(() => run(load))
</script>
