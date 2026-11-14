<template>
  <div class="space-y-6" data-testid="all-activities-page">
    <SectionHeader :icon="LuListOrdered" :title="t('wealth.activities')" :description="t('wealth.activitiesHint')" />
    <PageFeedback :error="error" :busy="busy" />

    <form data-testid="activity-filters" class="space-y-3 rounded-2xl border bg-card p-4 shadow-xs" @submit.prevent="apply">
      <div class="grid items-end gap-3 sm:grid-cols-2 lg:grid-cols-[minmax(0,2fr)_repeat(3,minmax(0,1fr))]">
        <div class="space-y-1.5">
          <Label for="activity-q" class="text-xs text-muted-foreground"><LuSearch class="size-3.5" aria-hidden="true" />{{ t('wealth.search') }}</Label>
          <Input id="activity-q" v-model="filters.q" maxlength="100" :placeholder="t('wealth.activitySearchHint')" />
        </div>
        <div class="space-y-1.5"><Label class="text-xs text-muted-foreground"><LuWallet class="size-3.5" aria-hidden="true" />{{ t('nav.accounts') }}</Label><FormSelect v-model="filters.account_id" :aria-label="t('nav.accounts')" :options="[{ value: 'all', label: t('wealth.allAccounts') }, ...accounts.map(a => ({ value: a.id, label: a.account_name }))]" /></div>
        <div class="space-y-1.5"><Label class="text-xs text-muted-foreground"><LuArrowRightLeft class="size-3.5" aria-hidden="true" />{{ t('portfolio.type') }}</Label><FormSelect v-model="filters.type" :aria-label="t('portfolio.type')" :options="[{ value: 'all', label: t('wealth.allTypes') }, ...types.map(value => ({ value, label: t(`portfolio.types.${value}`) }))]" /></div>
        <div class="space-y-1.5"><Label class="text-xs text-muted-foreground"><LuTags class="size-3.5" aria-hidden="true" />{{ t('wealth.tags') }}</Label><FormSelect v-model="filters.tag_id" :aria-label="t('wealth.tags')" :options="[{ value: 'all', label: t('wealth.allTags') }, ...tags.map(a => ({ value: a.id, label: a.name }))]" /></div>
      </div>
      <div class="grid items-end gap-3 sm:grid-cols-3 lg:grid-cols-[repeat(3,minmax(0,12rem))_1fr]">
        <div class="space-y-1.5"><Label for="activity-from" class="text-xs text-muted-foreground"><LuCalendar class="size-3.5" aria-hidden="true" />{{ t('wealth.from') }}</Label><Input id="activity-from" v-model="filters.from" type="date" :max="today" /></div>
        <div class="space-y-1.5"><Label for="activity-to" class="text-xs text-muted-foreground"><LuCalendar class="size-3.5" aria-hidden="true" />{{ t('wealth.to') }}</Label><Input id="activity-to" v-model="filters.to" type="date" :min="filters.from" :max="today" /></div>
        <div class="space-y-1.5"><Label class="text-xs text-muted-foreground"><LuArrowDownUp class="size-3.5" aria-hidden="true" />{{ t('wealth.order') }}</Label><FormSelect v-model="filters.direction" :aria-label="t('wealth.order')" :options="[{value:'desc', label:t('wealth.newest')},{value:'asc',label:t('wealth.oldest')}]" /></div>
        <div class="flex flex-wrap justify-end gap-2 sm:col-span-3 lg:col-span-1">
          <Button v-if="activeFilters" type="button" variant="ghost" :disabled="busy" @click="reset"><LuX class="size-4" aria-hidden="true" />{{ t('holdings.resetFilters') }}</Button>
          <Button type="submit" :disabled="busy"><LuListFilter class="size-4" aria-hidden="true" />{{ t('wealth.applyFilters') }}</Button>
        </div>
      </div>
    </form>

    <div v-if="result" data-testid="activity-list" class="overflow-hidden rounded-2xl border bg-card shadow-xs">
      <div class="overflow-x-auto">
        <Table v-if="result.total">
          <TableHeader class="bg-muted/40"><TableRow class="hover:bg-transparent">
            <TableHead v-for="column in COLUMNS" :key="column.key" class="h-10 whitespace-nowrap text-xs font-medium uppercase tracking-wide" :class="column.right && 'text-right'"><span class="inline-flex items-center gap-1.5"><component :is="column.icon" class="size-3.5 text-muted-foreground/80" aria-hidden="true" />{{ t(`wealth.${column.key}`) }}</span></TableHead>
          </TableRow></TableHeader>
          <TableBody>
            <TableRow v-for="a in result.activities" :key="a.id" :data-testid="`activity-row-${a.id}`">
              <TableCell class="whitespace-nowrap tabular-nums text-muted-foreground">{{ formatDate(a.date) }}</TableCell>
              <TableCell><router-link :to="`/accounts/${a.account_id}`" class="font-medium hover:text-primary hover:underline">{{ a.account_name }}</router-link></TableCell>
              <TableCell><StatusBadge :tone="typeTone(a.type)">{{ t(`portfolio.types.${a.type}`) }}</StatusBadge></TableCell>
              <TableCell class="font-medium">{{ a.symbol || '—' }}</TableCell>
              <TableCell class="text-right tabular-nums">{{ a.type === 'split' ? `${a.split_numerator}:${a.split_denominator}` : number(a.quantity) }}</TableCell>
              <TableCell class="text-right font-medium tabular-nums">{{ a.type === 'split' ? '—' : money(a.quantity * a.unit_price, a.currency) }}</TableCell>
              <TableCell><div class="flex min-w-24 flex-wrap gap-1"><TagBadge v-for="tag in a.tags" :key="tag.id" :tag="tag" /><span v-if="!a.tags.length" class="text-muted-foreground">—</span></div></TableCell>
              <TableCell><div class="flex justify-end gap-2"><TagPicker v-model="a.tags" kind="activity" :id="a.id" size="sm" variant="outline" :testid="`activity-tags-${a.id}`" /><Button v-if="a.ledger && !a.transfer_id" as-child variant="outline" size="sm"><router-link :to="`/accounts/${a.account_id}/activities/${a.id}/edit`"><LuPencil class="size-3.5" aria-hidden="true" />{{ t('common.edit') }}</router-link></Button></div></TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
      <EmptyState v-if="!result.total" :icon="LuListOrdered" :title="t('common.noResults')" :description="activeFilters ? t('wealth.noActivitiesFiltered') : t('wealth.noActivities')" />
      <ListPager v-else :page="page" :pages="result.pages" :total="result.total" :size="Number(size)" :disabled="busy" @update:page="go">
        <label class="mr-2 hidden items-center gap-2 sm:flex">{{ t('wealth.perPage') }}<FormSelect v-model="size" class="h-8 w-20" :options="[10,25,50,100].map(v => ({value:String(v),label:String(v)}))" :disabled="busy" @update:model-value="apply" /></label>
      </ListPager>
    </div>
  </div>
</template>
<script setup>
import { computed, onMounted, ref } from 'vue'
import { LuArrowDownUp, LuArrowRightLeft, LuCalendar, LuCircleDollarSign, LuHash, LuListFilter, LuListOrdered, LuPencil, LuSearch, LuSettings2, LuTags, LuTrendingUp, LuWallet, LuX } from '@kalimahapps/vue-icons/lu'
import { Button, Input, Label, FormSelect, Table, TableHeader, TableHead, TableRow, TableBody, TableCell } from '@/components/ui'
import PageFeedback from '@/components/wealth/PageFeedback.vue'
import SectionHeader from '@/components/wealth/SectionHeader.vue'
import EmptyState from '@/components/wealth/EmptyState.vue'
import StatusBadge from '@/components/wealth/StatusBadge.vue'
import ListPager from '@/components/wealth/ListPager.vue'
import TagBadge from '@/components/wealth/TagBadge.vue'
import TagPicker from '@/components/wealth/TagPicker.vue'
import { wealthAPI } from '@/services/wealthApi'
import { accountAPI } from '@/services/accountApi'
import { useWealthPage } from '@/composables/useWealthPage'
const COLUMNS = [
  { key: 'date', icon: LuCalendar }, { key: 'account', icon: LuWallet }, { key: 'type', icon: LuArrowRightLeft }, { key: 'symbol', icon: LuTrendingUp },
  { key: 'quantity', icon: LuHash, right: true }, { key: 'amount', icon: LuCircleDollarSign, right: true }, { key: 'tags', icon: LuTags }, { key: 'actions', icon: LuSettings2, right: true },
]
// Cash coming in reads green, cash going out red; trades and moves stay neutral.
const TONES = { deposit: 'success', dividend: 'success', interest: 'success', withdrawal: 'danger', fee: 'danger', buy: 'info', sell: 'warning', opening: 'info' }
const blank = () => ({ q: '', account_id: 'all', tag_id: 'all', type: 'all', from: '', to: '', direction: 'desc' })
const { t, locale, busy, error, run, money, today } = useWealthPage()
const filters = ref(blank())
const page = ref(1), size = ref('25'), result = ref(null), accounts = ref([]), tags = ref([])
const types = ['buy','sell','deposit','withdrawal','dividend','interest','fee','split','opening','transfer_in','transfer_out']
const activeFilters = computed(() => Object.entries(blank()).filter(([key, value]) => key !== 'direction' && filters.value[key] !== value).length)
const typeTone = type => TONES[type] || 'neutral'
const number = value => new Intl.NumberFormat(locale.value, { maximumFractionDigits: 8 }).format(value)
const formatDate = value => new Intl.DateTimeFormat(locale.value, { dateStyle: 'medium' }).format(new Date(`${value.slice(0, 10)}T00:00:00`))
async function load() {
  const params = Object.fromEntries(Object.entries(filters.value).map(([k, v]) => [k, v === 'all' ? '' : v]))
  const { data, meta } = await wealthAPI.get('/activities', { ...params, page: page.value, per_page: size.value })
  result.value = { activities: data, total: meta.pagination.total, pages: meta.pagination.total_pages }
}
function apply() { page.value = 1; run(load) }
function reset() { filters.value = { ...blank(), direction: filters.value.direction }; apply() }
function go(value) { page.value = value; run(load) }
onMounted(() => run(async () => { const [a, t] = await Promise.all([accountAPI.getAccounts(), wealthAPI.get('/tags')]); accounts.value = a.data; tags.value = t.data; await load() }))
</script>
