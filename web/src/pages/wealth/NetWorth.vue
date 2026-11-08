<template>
  <div class="space-y-5" data-testid="net-worth-page">
    <div class="flex flex-wrap justify-between gap-3"><h2 class="text-xl font-semibold">{{ t('wealth.netWorth') }}</h2><Button variant="outline" :disabled="busy" @click="run(load)">{{ t('portfolio.refresh') }}</Button></div>
    <PageFeedback :error="error" :busy="busy" :notice="notice" />
    <div v-if="report" class="grid gap-3 sm:grid-cols-3"><Card v-for="key in ['value', 'liabilities', 'net_worth']" :key="key"><CardContent class="pt-6"><p class="text-xs text-muted-foreground">{{ t(`wealth.${key}`) }}</p><p class="mt-2 text-2xl font-semibold">{{ money(report.complete ? report.summary[key] : null, report.currency) }}</p></CardContent></Card></div>
    <p class="text-sm text-muted-foreground">{{ t('wealth.netWorthHint') }}</p>
    <div class="flex items-center gap-3"><label class="text-sm" for="net-worth-range">{{ t('portfolio.range') }}</label><FormSelect id="net-worth-range" v-model="range" class="max-w-48" :options="['1m','1y','5y','max'].map(value => ({ value, label: t(`portfolio.ranges.${value}`) }))" :disabled="busy" @update:model-value="run(load)" /></div>
    <PortfolioChart v-if="report" :points="report.history.map(p => ({ ...p, value: p.net_worth }))" />
    <p v-for="warning in report?.warnings || []" :key="warning" class="text-sm text-muted-foreground">{{ warning }}</p>
    <h3 class="font-semibold">{{ t('wealth.liabilities') }}</h3>
    <Card><CardContent class="pt-6"><form class="grid items-end gap-3 sm:grid-cols-2 lg:grid-cols-5" @submit.prevent="create">
      <label class="grid gap-2 text-sm">{{ t('wealth.name') }}<Input v-model="form.name" required maxlength="100" :disabled="busy" /></label><label class="grid gap-2 text-sm">{{ t('portfolio.currency') }}<Input v-model="form.currency" required maxlength="3" :disabled="busy" /></label><label class="grid gap-2 text-sm">{{ t('portfolio.date') }}<Input v-model="form.date" type="date" :max="today" required :disabled="busy" /></label><label class="grid gap-2 text-sm">{{ t('wealth.outstanding') }}<Input v-model.number="form.amount" type="number" min="0" step="0.01" required :disabled="busy" /></label><Button type="submit" :disabled="busy">{{ t('wealth.addLiability') }}</Button>
    </form></CardContent></Card>
    <p class="text-xs text-muted-foreground">{{ t('wealth.liabilityHint') }}</p>
    <Card v-for="debt in debts" :key="debt.id"><CardContent class="space-y-3 pt-6"><h3 class="font-semibold">{{ debt.name }} · {{ debt.currency }}</h3><form class="flex flex-wrap items-end gap-3" @submit.prevent="balance(debt)"><label class="grid gap-1 text-sm">{{ t('portfolio.date') }}<Input v-model="entries[debt.id].date" type="date" :max="today" required :disabled="busy" /></label><label class="grid flex-1 gap-1 text-sm">{{ t('wealth.outstanding') }}<Input v-model.number="entries[debt.id].amount" type="number" min="0" step="0.01" required :disabled="busy" /></label><Button type="submit" variant="outline" :disabled="busy">{{ t('common.save') }}</Button></form><details><summary class="cursor-pointer text-sm">{{ t('wealth.history') }}</summary><div class="max-h-56 overflow-auto"><div v-for="b in debt.balances" :key="b.id" class="flex justify-between border-t py-2 text-sm"><span>{{ b.date }}</span><span>{{ money(b.amount, debt.currency) }}</span></div></div></details></CardContent></Card>
  </div>
</template>
<script setup>
import { onMounted, ref } from 'vue'
import { Button, Input, Card, CardContent, FormSelect } from '@/components/ui'
import PortfolioChart from '@/components/portfolio/PortfolioChart.vue'
import PageFeedback from '@/components/wealth/PageFeedback.vue'
import { useWealthPage } from '@/composables/useWealthPage'
import { wealthAPI } from '@/services/wealthApi'
import { portfolioAPI } from '@/services/portfolioApi'
const { t, busy, error, notice, run, money, today } = useWealthPage()
const debts = ref([]), report = ref(null), range = ref('1y'), entries = ref({})
const form = ref({ name: '', currency: 'USD', date: today, amount: '' })
async function load() { const [d, r] = await Promise.all([wealthAPI.get('/liabilities'), portfolioAPI.report({ range: range.value })]); debts.value = d.data; report.value = r.data; for (const item of debts.value) entries.value[item.id] ||= { date: today, amount: item.balances[0]?.amount ?? '' } }
function create() { run(async () => { await wealthAPI.save('/liabilities', { ...form.value, currency: form.value.currency.toUpperCase() }); form.value.name = ''; form.value.amount = ''; await load() }, true) }
function balance(item) { run(async () => { await wealthAPI.save(`/liabilities/${item.id}/balances`, entries.value[item.id], 'PUT'); await load() }, true) }
onMounted(() => run(load))
</script>
