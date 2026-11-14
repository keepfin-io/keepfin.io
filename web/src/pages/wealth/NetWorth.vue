<template>
  <div class="space-y-6" data-testid="net-worth-page">
    <SectionHeader :icon="LuScale" :title="t('wealth.netWorth')" :description="t('wealth.netWorthHint')">
      <Button variant="outline" :disabled="busy" @click="run(load)"><LuRefreshCw class="size-4" :class="busy && 'animate-spin'" aria-hidden="true" />{{ t('portfolio.refresh') }}</Button>
    </SectionHeader>
    <PageFeedback :error="error" :busy="busy" :notice="notice" />

    <!-- Assets − liabilities = net worth, read left to right. -->
    <div v-if="report" data-testid="net-worth-summary" class="grid items-stretch gap-3 lg:grid-cols-[1fr_auto_1fr_auto_1fr]">
      <template v-for="(card, index) in SUMMARY" :key="card.key">
        <span v-if="index" class="hidden items-center justify-center text-muted-foreground lg:flex" aria-hidden="true"><component :is="index === 1 ? LuMinus : LuEqual" class="size-5" /></span>
        <Card class="gap-0 py-4" :class="card.key === 'net_worth' && 'border-primary/40 bg-primary/5'">
          <CardContent class="space-y-1 px-4">
            <p class="flex items-center gap-2 text-sm text-muted-foreground"><component :is="card.icon" class="size-4 text-primary/80" aria-hidden="true" />{{ t(`wealth.${card.key}`) }}</p>
            <p :data-testid="`net-worth-${card.key}`" class="text-2xl font-semibold tabular-nums" :class="card.key === 'liabilities' && report.summary?.liabilities > 0 && 'text-destructive'">{{ money(report.complete ? report.summary[card.key] : null, report.currency) }}</p>
            <p class="text-xs text-muted-foreground">{{ t(`wealth.summaryHints.${card.key}`) }}</p>
          </CardContent>
        </Card>
      </template>
    </div>

    <Card v-if="report" class="gap-0 py-0">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b px-5 py-3">
        <h3 class="flex items-center gap-2 text-sm font-semibold"><LuChartLine class="size-4 text-primary" aria-hidden="true" />{{ t('wealth.netWorthHistory') }}</h3>
        <div class="flex items-center gap-2"><Label for="net-worth-range" class="text-xs text-muted-foreground">{{ t('portfolio.range') }}</Label><FormSelect id="net-worth-range" v-model="range" class="w-36" :options="['1m','1y','5y','max'].map(value => ({ value, label: t(`portfolio.ranges.${value}`) }))" :disabled="busy" @update:model-value="run(load)" /></div>
      </div>
      <CardContent class="p-4"><PortfolioChart :points="report.history.map(p => ({ ...p, value: p.net_worth }))" /></CardContent>
    </Card>
    <div v-if="report?.warnings?.length" class="space-y-1 rounded-lg border border-amber-500/30 bg-amber-500/10 p-3 text-sm text-amber-800 dark:text-amber-300">
      <p v-for="warning in report.warnings" :key="warning" class="flex gap-2"><LuTriangleAlert class="mt-0.5 size-4 shrink-0" aria-hidden="true" />{{ warning }}</p>
    </div>

    <section class="space-y-4" :aria-label="t('wealth.liabilities')">
      <SectionHeader :icon="LuCreditCard" level="h3" :title="t('wealth.liabilities')" :description="t('wealth.liabilityHint')">
        <Button data-testid="liability-new" variant="outline" class="w-full sm:w-auto" :disabled="busy" @click="openCreate"><LuPlus class="size-4" aria-hidden="true" />{{ t('wealth.addLiability') }}</Button>
      </SectionHeader>
      <div data-testid="liabilities-list" class="overflow-hidden rounded-2xl border bg-card shadow-xs">
        <template v-if="debts.length">
          <div class="hidden border-b bg-muted/40 px-5 py-2.5 text-xs font-medium uppercase tracking-wide text-muted-foreground sm:grid sm:grid-cols-[minmax(0,1fr)_11rem_auto] sm:gap-6" aria-hidden="true">
            <span>{{ t('wealth.liability') }}</span><span class="text-right">{{ t('wealth.outstanding') }}</span><span class="w-40" />
          </div>
          <ul role="list" class="divide-y">
            <li v-for="debt in debts" :key="debt.id" :data-testid="`liability-${debt.id}`" class="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-x-4 gap-y-3 px-5 py-4 transition-colors hover:bg-muted/30 sm:grid-cols-[minmax(0,1fr)_11rem_auto] sm:gap-6">
              <div class="flex min-w-0 items-center gap-3">
                <span class="flex size-10 shrink-0 items-center justify-center rounded-xl border bg-background text-destructive/80" aria-hidden="true"><LuCreditCard class="size-5" /></span>
                <div class="min-w-0">
                  <p class="truncate font-semibold">{{ debt.name }}</p>
                  <p class="text-sm text-muted-foreground"><span class="rounded bg-muted px-1.5 py-0.5 font-mono text-[11px]">{{ debt.currency }}</span><template v-if="debt.balances[0]?.amount === 0"> · {{ t('wealth.paidOff') }}</template></p>
                </div>
              </div>
              <div class="text-right">
                <p class="font-semibold tabular-nums" :class="debt.balances[0]?.amount > 0 && 'text-destructive'">{{ debt.balances.length ? money(debt.balances[0].amount, debt.currency) : '—' }}</p>
                <p v-if="debt.balances[0]" class="text-xs text-muted-foreground">{{ formatDate(debt.balances[0].date) }}</p>
              </div>
              <div class="col-span-2 flex justify-end sm:col-span-1 sm:w-40">
                <Button :data-testid="`liability-${debt.id}-update`" variant="outline" size="sm" class="h-8" :disabled="busy" @click="openBalance(debt)"><LuCalendarClock class="size-3.5" aria-hidden="true" />{{ t('wealth.updateBalance') }}</Button>
              </div>
            </li>
          </ul>
        </template>
        <EmptyState v-else-if="report" :icon="LuCreditCard" :title="t('wealth.emptyLiabilitiesTitle')" :description="t('wealth.emptyLiabilities')" />
      </div>
    </section>

    <FormDialog data-testid="liability-dialog" :open="createOpen" :busy="busy" :error="error" :icon="LuCreditCard" :title="t('wealth.addLiability')" :description="t('wealth.liabilityFormHint')" @close="createOpen = false">
      <form id="liability-form" class="grid gap-4 sm:grid-cols-2" @submit.prevent="create">
        <div class="space-y-2 sm:col-span-2"><Label for="liability-name"><LuTag class="size-4 text-muted-foreground" aria-hidden="true" />{{ t('wealth.name') }}</Label><Input id="liability-name" v-model="form.name" required maxlength="100" :placeholder="t('wealth.liabilityNamePlaceholder')" :disabled="busy" /></div>
        <div class="space-y-2"><Label for="liability-currency"><LuBanknote class="size-4 text-muted-foreground" aria-hidden="true" />{{ t('portfolio.currency') }}</Label><Input id="liability-currency" v-model="form.currency" required maxlength="3" class="uppercase" placeholder="TRY" :disabled="busy" /></div>
        <div class="space-y-2"><Label for="liability-date"><LuCalendar class="size-4 text-muted-foreground" aria-hidden="true" />{{ t('portfolio.date') }}</Label><Input id="liability-date" v-model="form.date" type="date" :max="today" required :disabled="busy" /></div>
        <div class="space-y-2 sm:col-span-2"><Label for="liability-amount"><LuCircleDollarSign class="size-4 text-muted-foreground" aria-hidden="true" />{{ t('wealth.outstanding') }}</Label><Input id="liability-amount" v-model.number="form.amount" type="number" min="0" step="0.01" required placeholder="0.00" :disabled="busy" /></div>
      </form>
      <template #footer>
        <Button variant="outline" :disabled="busy" @click="createOpen = false">{{ t('common.cancel') }}</Button>
        <Button data-testid="liability-save" type="submit" form="liability-form" :loading="busy" :loading-text="t('common.saving')">{{ t('wealth.addLiability') }}</Button>
      </template>
    </FormDialog>

    <FormDialog data-testid="liability-balance-dialog" :open="!!updating" :busy="busy" :error="error" :icon="LuCalendarClock" :title="updating ? t('wealth.balanceTitle', { name: updating.name }) : ''" :description="t('wealth.balanceHint')" @close="updating = null">
      <form v-if="updating" id="liability-balance-form" class="grid gap-4 sm:grid-cols-2" @submit.prevent="saveBalance">
        <div class="space-y-2"><Label for="balance-date"><LuCalendar class="size-4 text-muted-foreground" aria-hidden="true" />{{ t('portfolio.date') }}</Label><Input id="balance-date" v-model="entry.date" type="date" :max="today" required :disabled="busy" /></div>
        <div class="space-y-2"><Label for="balance-amount"><LuCircleDollarSign class="size-4 text-muted-foreground" aria-hidden="true" />{{ t('wealth.outstanding') }} ({{ updating.currency }})</Label><Input id="balance-amount" v-model.number="entry.amount" type="number" min="0" step="0.01" required :disabled="busy" /></div>
      </form>
      <div v-if="updating?.balances.length" class="rounded-xl border">
        <p class="flex items-center gap-2 border-b bg-muted/40 px-3 py-2 text-xs font-medium text-muted-foreground"><LuHistory class="size-3.5" aria-hidden="true" />{{ t('wealth.history') }}</p>
        <ul class="max-h-56 divide-y overflow-y-auto">
          <li v-for="b in updating.balances" :key="b.id" class="flex justify-between gap-3 px-3 py-1.5 text-sm"><span class="text-muted-foreground">{{ formatDate(b.date) }}</span><span class="font-medium tabular-nums">{{ money(b.amount, updating.currency) }}</span></li>
        </ul>
      </div>
      <template #footer>
        <Button variant="outline" :disabled="busy" @click="updating = null">{{ t('common.cancel') }}</Button>
        <Button data-testid="liability-balance-save" type="submit" form="liability-balance-form" :loading="busy" :loading-text="t('common.saving')">{{ t('common.save') }}</Button>
      </template>
    </FormDialog>
  </div>
</template>
<script setup>
import { onMounted, ref } from 'vue'
import { LuBanknote, LuCalendar, LuCalendarClock, LuChartLine, LuCircleDollarSign, LuCreditCard, LuEqual, LuHistory, LuMinus, LuPlus, LuRefreshCw, LuScale, LuTag, LuTriangleAlert, LuWallet } from '@kalimahapps/vue-icons/lu'
import { Button, Card, CardContent, FormSelect, Input, Label } from '@/components/ui'
import PortfolioChart from '@/components/portfolio/PortfolioChart.vue'
import PageFeedback from '@/components/wealth/PageFeedback.vue'
import SectionHeader from '@/components/wealth/SectionHeader.vue'
import EmptyState from '@/components/wealth/EmptyState.vue'
import FormDialog from '@/components/wealth/FormDialog.vue'
import { useWealthPage } from '@/composables/useWealthPage'
import { wealthAPI } from '@/services/wealthApi'
import { portfolioAPI } from '@/services/portfolioApi'
const SUMMARY = [{ key: 'value', icon: LuWallet }, { key: 'liabilities', icon: LuCreditCard }, { key: 'net_worth', icon: LuScale }]
const { t, locale, busy, error, notice, run, money, today } = useWealthPage()
const debts = ref([]), report = ref(null), range = ref('1y'), createOpen = ref(false), updating = ref(null), entry = ref({ date: today, amount: '' })
const form = ref({ name: '', currency: 'USD', date: today, amount: '' })
const formatDate = value => new Intl.DateTimeFormat(locale.value, { dateStyle: 'medium' }).format(new Date(`${value.slice(0, 10)}T00:00:00`))
async function load() { const [d, r] = await Promise.all([wealthAPI.get('/liabilities'), portfolioAPI.report({ range: range.value })]); debts.value = d.data; report.value = r.data }
function openCreate() { error.value = ''; form.value = { name: '', currency: form.value.currency, date: today, amount: '' }; createOpen.value = true }
function openBalance(item) { error.value = ''; updating.value = item; entry.value = { date: today, amount: item.balances[0]?.amount ?? '' } }
function create() { run(async () => { await wealthAPI.save('/liabilities', { ...form.value, currency: form.value.currency.toUpperCase() }); createOpen.value = false; await load() }, true) }
function saveBalance() { run(async () => { await wealthAPI.save(`/liabilities/${updating.value.id}/balances`, entry.value, 'PUT'); updating.value = null; await load() }, true) }
onMounted(() => run(load))
</script>
