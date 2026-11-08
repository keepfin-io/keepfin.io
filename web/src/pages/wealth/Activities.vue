<template>
  <div class="space-y-5" data-testid="all-activities-page">
    <h2 class="text-xl font-semibold">{{ t('wealth.activities') }}</h2><PageFeedback :error="error" :busy="busy" />
    <form class="grid items-end gap-3 sm:grid-cols-2 lg:grid-cols-4" @submit.prevent="apply">
      <label class="grid gap-1 text-sm">{{ t('wealth.search') }}<Input v-model="filters.q" maxlength="100" :placeholder="t('wealth.activitySearchHint')" /></label>
      <label class="grid gap-1 text-sm">{{ t('nav.accounts') }}<FormSelect v-model="filters.account_id" :options="[{ value: 'all', label: t('wealth.allAccounts') }, ...accounts.map(a => ({ value: a.id, label: a.account_name }))]" /></label>
      <label class="grid gap-1 text-sm">{{ t('portfolio.type') }}<FormSelect v-model="filters.type" :options="[{ value: 'all', label: t('wealth.allTypes') }, ...types.map(value => ({ value, label: t(`portfolio.types.${value}`) }))]" /></label>
      <label class="grid gap-1 text-sm">{{ t('wealth.tags') }}<FormSelect v-model="filters.tag_id" :options="[{ value: 'all', label: t('wealth.allTags') }, ...tags.map(a => ({ value: a.id, label: a.name }))]" /></label>
      <label class="grid gap-1 text-sm">{{ t('wealth.from') }}<Input v-model="filters.from" type="date" :max="today" /></label><label class="grid gap-1 text-sm">{{ t('wealth.to') }}<Input v-model="filters.to" type="date" :min="filters.from" :max="today" /></label>
      <label class="grid gap-1 text-sm">{{ t('wealth.order') }}<FormSelect v-model="filters.direction" :options="[{value:'desc', label:t('wealth.newest')},{value:'asc',label:t('wealth.oldest')}]" /></label><Button type="submit" :disabled="busy">{{ t('wealth.applyFilters') }}</Button>
    </form>
    <div v-if="result" class="space-y-3"><p class="text-xs text-muted-foreground">{{ t('wealth.totalRecords', { count: result.total }) }}</p><Card class="overflow-hidden"><CardContent class="overflow-auto p-0"><Table><TableHeader><TableRow><TableHead v-for="key in ['date','account','type','symbol','quantity','amount','tags','actions']" :key="key">{{ t(`wealth.${key}`) }}</TableHead></TableRow></TableHeader><TableBody><template v-for="a in result.activities" :key="a.id"><TableRow><TableCell>{{ a.date.slice(0,10) }}</TableCell><TableCell><router-link :to="`/accounts/${a.account_id}`" class="text-primary">{{ a.account_name }}</router-link></TableCell><TableCell>{{ t(`portfolio.types.${a.type}`) }}</TableCell><TableCell>{{ a.symbol }}</TableCell><TableCell>{{ a.type === 'split' ? `${a.split_numerator}:${a.split_denominator}` : a.quantity }}</TableCell><TableCell>{{ a.type === 'split' ? '—' : money(a.quantity * a.unit_price, a.currency) }}</TableCell><TableCell>{{ a.tags.map(t => t.name).join(', ') || '—' }}</TableCell><TableCell><div class="flex gap-2"><Button variant="ghost" size="sm" @click="editingTags = editingTags === a.id ? '' : a.id">{{ t('wealth.tags') }}</Button><Button v-if="a.ledger && !a.transfer_id" as-child variant="outline" size="sm"><router-link :to="`/accounts/${a.account_id}/activities/${a.id}/edit`">{{ t('common.edit') }}</router-link></Button></div></TableCell></TableRow><TableRow v-if="editingTags === a.id"><TableCell colspan="8"><EntityTags kind="activity" :id="a.id" @saved="run(load)" /></TableCell></TableRow></template></TableBody></Table></CardContent></Card><p v-if="!result.total" class="py-6 text-sm text-muted-foreground">{{ t('common.noResults') }}</p>
      <div class="flex flex-wrap items-center justify-end gap-3"><label class="flex items-center gap-2 text-sm">{{ t('wealth.perPage') }}<FormSelect v-model="size" class="w-24" :options="[10,25,50,100].map(v => ({value:String(v),label:String(v)}))" :disabled="busy" @update:model-value="apply" /></label><Button variant="outline" size="sm" :disabled="busy || page <= 1" @click="go(page-1)">{{ t('wealth.previous') }}</Button><span class="text-sm">{{ page }} / {{ result.pages }}</span><Button variant="outline" size="sm" :disabled="busy || page >= result.pages" @click="go(page+1)">{{ t('wealth.next') }}</Button></div>
    </div>
  </div>
</template>
<script setup>
import { onMounted, ref } from 'vue'
import { Button, Input, FormSelect, Card, CardContent, Table, TableHeader, TableHead, TableRow, TableBody, TableCell } from '@/components/ui'
import PageFeedback from '@/components/wealth/PageFeedback.vue'
import EntityTags from '@/components/wealth/EntityTags.vue'
import { wealthAPI } from '@/services/wealthApi'
import { accountAPI } from '@/services/accountApi'
import { useWealthPage } from '@/composables/useWealthPage'
const { t, busy, error, run, money, today } = useWealthPage()
const filters = ref({ q: '', account_id: 'all', tag_id: 'all', type: 'all', from: '', to: '', direction: 'desc' })
const page = ref(1), size = ref('25'), result = ref(null), accounts = ref([]), tags = ref([]), editingTags = ref('')
const types = ['buy','sell','deposit','withdrawal','dividend','interest','fee','split','opening','transfer_in','transfer_out']
async function load() {
  const params = Object.fromEntries(Object.entries(filters.value).map(([k, v]) => [k, v === 'all' ? '' : v]))
  const { data, meta } = await wealthAPI.get('/activities', { ...params, page: page.value, per_page: size.value })
  result.value = { activities: data, total: meta.pagination.total, pages: meta.pagination.total_pages }
}
function apply() { page.value = 1; run(load) }
function go(value) { page.value = value; run(load) }
onMounted(() => run(async () => { const [a, t] = await Promise.all([accountAPI.getAccounts(), wealthAPI.get('/tags')]); accounts.value = a.data; tags.value = t.data; await load() }))
</script>
