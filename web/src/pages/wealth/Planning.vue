<template>
  <div class="space-y-6" :data-testid="`${mode}-page`">
    <SectionHeader :icon="mode === 'risk' ? LuShieldCheck : LuFlame" :title="t(`wealth.${mode}`)" :description="t(`wealth.${mode}Hint`)" />
    <PageFeedback :error="error" :busy="busy" :notice="notice" />

    <template v-if="report && mode === 'risk'">
      <div data-testid="risk-summary" class="grid gap-3 sm:grid-cols-3">
        <div v-for="item in riskSummary" :key="item.status" class="flex items-center gap-3 rounded-xl border bg-card p-4 shadow-xs">
          <span class="flex size-10 items-center justify-center rounded-xl" :class="item.tile" aria-hidden="true"><component :is="item.icon" class="size-5" /></span>
          <div><p class="text-2xl font-semibold tabular-nums">{{ item.count }}</p><p class="text-xs text-muted-foreground">{{ t(`wealth.${item.status}`) }}</p></div>
        </div>
      </div>
      <div class="grid gap-3 md:grid-cols-2">
        <article v-for="rule in report.rules" :key="rule.id" :data-testid="`risk-rule-${rule.id}`" class="space-y-3 rounded-2xl border bg-card p-5 shadow-xs" :class="!rule.enabled && 'opacity-60'">
          <div class="flex items-start gap-3">
            <span class="flex size-9 shrink-0 items-center justify-center rounded-lg bg-muted text-muted-foreground" aria-hidden="true"><component :is="RULE_ICONS[rule.id]" class="size-4" /></span>
            <div class="min-w-0 flex-1">
              <h3 class="font-medium">{{ t(`wealth.rule_${rule.id}`) }}</h3>
              <p class="text-xs text-muted-foreground">{{ t(`wealth.ruleHints.${rule.id}`) }}</p>
            </div>
            <StatusBadge :tone="rule.enabled ? STATUS_TONES[rule.status] : 'neutral'">{{ t(`wealth.${rule.enabled ? rule.status : 'disabled'}`) }}</StatusBadge>
          </div>
          <div class="flex items-end justify-between gap-3">
            <p class="text-2xl font-semibold tabular-nums">{{ formatRule(rule, rule.value) }}</p>
            <p class="text-right text-xs text-muted-foreground">{{ t(rule.unit === 'percent' ? 'wealth.maximum' : 'wealth.minimum') }}: <span class="font-medium text-foreground">{{ formatRule(rule, rule.limit) }}</span></p>
          </div>
          <div v-if="rule.unit !== 'money' && rule.value != null && rule.limit > 0" class="h-1.5 overflow-hidden rounded-full bg-muted" aria-hidden="true">
            <div class="h-full rounded-full transition-all" :class="BAR[rule.enabled ? rule.status : 'unknown']" :style="{ width: `${Math.min(100, (rule.value / rule.limit) * 100)}%` }" />
          </div>
          <p v-if="rule.subject" class="truncate text-xs text-muted-foreground">{{ t('wealth.highest') }}: <span class="font-medium text-foreground">{{ rule.subject }}</span></p>
        </article>
      </div>
      <p class="flex items-start gap-2 text-xs text-muted-foreground"><LuInfo class="mt-0.5 size-3.5 shrink-0" aria-hidden="true" />{{ t('wealth.feeRuleHint') }}</p>
    </template>

    <template v-if="report && mode === 'fire'">
      <div data-testid="fire-summary" class="grid gap-3 sm:grid-cols-3">
        <Card v-for="card in FIRE_CARDS" :key="card.key" class="gap-0 py-4" :class="card.key === 'years_to_target' && 'border-primary/40 bg-primary/5'">
          <CardContent class="space-y-1 px-4">
            <p class="flex items-center gap-2 text-sm text-muted-foreground"><component :is="card.icon" class="size-4 text-primary/80" aria-hidden="true" />{{ t(`wealth.${card.key}`) }}</p>
            <p class="text-2xl font-semibold tabular-nums">{{ card.key === 'years_to_target' ? report[card.key] == null ? '—' : t('wealth.yearsValue', { years: report[card.key].toFixed(1) }) : money(report[card.key], report.currency) }}</p>
            <p class="text-xs text-muted-foreground">{{ t(`wealth.fireHints.${card.key}`) }}</p>
          </CardContent>
        </Card>
      </div>
      <Card class="gap-0 py-0">
        <h3 class="flex items-center gap-2 border-b px-5 py-3 text-sm font-semibold"><LuChartLine class="size-4 text-primary" aria-hidden="true" />{{ t('wealth.projection') }}</h3>
        <CardContent class="p-4"><PortfolioChart :points="report.projection.map(p => ({ date: `${t('wealth.year')} ${p.year}`, value: p.value, net_flow: p.contributions }))" /></CardContent>
      </Card>
      <p class="flex items-start gap-2 text-xs text-muted-foreground"><LuInfo class="mt-0.5 size-3.5 shrink-0" aria-hidden="true" />{{ t('wealth.projectionHint') }}</p>
    </template>
    <div v-if="report?.warnings?.length" class="space-y-1 rounded-lg border border-amber-500/30 bg-amber-500/10 p-3 text-sm text-amber-800 dark:text-amber-300">
      <p v-for="warning in report.warnings" :key="warning" class="flex gap-2"><LuTriangleAlert class="mt-0.5 size-4 shrink-0" aria-hidden="true" />{{ warning }}</p>
    </div>

    <!-- The assumptions behind the results above. -->
    <form v-if="preferences" :data-testid="`${mode}-settings`" class="overflow-hidden rounded-2xl border bg-card shadow-xs" @submit.prevent="save">
      <div class="border-b px-5 py-4">
        <h3 class="flex items-center gap-2 font-semibold"><LuSlidersHorizontal class="size-4 text-primary" aria-hidden="true" />{{ t(mode === 'risk' ? 'wealth.riskSettings' : 'wealth.fireSettings') }}</h3>
        <p class="mt-1 text-xs text-muted-foreground">{{ t('wealth.settingsHint') }}</p>
      </div>
      <div class="space-y-6 p-5">
        <fieldset v-for="group in groups" :key="group.key" class="space-y-3">
          <legend class="mb-3 text-xs font-medium uppercase tracking-wide text-muted-foreground">{{ t(`wealth.settingGroups.${group.key}`) }}</legend>
          <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <div v-for="key in group.fields" :key="key" class="space-y-2">
              <Label :for="`pref-${key}`">{{ t(`wealth.${key}`) }}</Label>
              <Input :id="`pref-${key}`" v-model.number="preferences[key]" type="number" :min="key === 'annual_return' ? -99 : 0" :step="key === 'horizon_years' ? 1 : 'any'" required :disabled="busy" />
            </div>
          </div>
        </fieldset>
        <fieldset v-if="mode === 'risk'" class="space-y-3">
          <legend class="mb-3 text-xs font-medium uppercase tracking-wide text-muted-foreground">{{ t('wealth.disabledRules') }}</legend>
          <div class="flex flex-wrap gap-2">
            <label v-for="id in RULES" :key="id" class="inline-flex cursor-pointer items-center gap-2 rounded-full border px-3 py-1.5 text-sm transition-colors has-[:checked]:border-muted-foreground/40 has-[:checked]:bg-muted has-[:checked]:text-muted-foreground has-[:checked]:line-through">
              <input v-model="preferences.disabled_rules" type="checkbox" class="size-3.5 accent-primary" :value="id" :disabled="busy">{{ t(`wealth.rule_${id}`) }}
            </label>
          </div>
        </fieldset>
      </div>
      <div class="flex justify-end border-t bg-muted/30 px-5 py-4">
        <Button type="submit" class="w-full sm:w-auto" :loading="busy" :loading-text="t('common.saving')"><LuCalculator class="size-4" aria-hidden="true" />{{ t('wealth.saveAndCalculate') }}</Button>
      </div>
    </form>
  </div>
</template>
<script setup>
import { computed, ref, watch } from 'vue'
import { LuBanknote, LuCalculator, LuChartLine, LuChartPie, LuCircleCheck, LuCircleHelp, LuCoins, LuFlame, LuHourglass, LuInfo, LuPiggyBank, LuReceipt, LuShieldCheck, LuSlidersHorizontal, LuTarget, LuTriangleAlert, LuUmbrella, LuWallet } from '@kalimahapps/vue-icons/lu'
import { Button, Input, Label, Card, CardContent } from '@/components/ui'
import PageFeedback from '@/components/wealth/PageFeedback.vue'
import SectionHeader from '@/components/wealth/SectionHeader.vue'
import StatusBadge from '@/components/wealth/StatusBadge.vue'
import PortfolioChart from '@/components/portfolio/PortfolioChart.vue'
import { useWealthPage } from '@/composables/useWealthPage'
import { wealthAPI } from '@/services/wealthApi'
const RULES = ['asset', 'account', 'currency', 'fees', 'emergency', 'liquidity']
const RULE_ICONS = { asset: LuChartPie, account: LuWallet, currency: LuBanknote, fees: LuReceipt, emergency: LuUmbrella, liquidity: LuCoins }
const STATUS_TONES = { ok: 'success', warning: 'warning', unknown: 'neutral' }
const BAR = { ok: 'bg-emerald-500', warning: 'bg-amber-500', unknown: 'bg-muted-foreground/40' }
const FIRE_CARDS = [{ key: 'investable', icon: LuPiggyBank }, { key: 'target', icon: LuTarget }, { key: 'years_to_target', icon: LuHourglass }]
const GROUPS = {
  risk: [{ key: 'concentration', fields: ['max_asset_weight', 'max_account_weight', 'max_currency_weight'] }, { key: 'costs', fields: ['max_fee_ratio'] }, { key: 'safety', fields: ['emergency_fund', 'monthly_expenses', 'emergency_months'] }],
  fire: [{ key: 'spending', fields: ['annual_expenses', 'withdrawal_rate'] }, { key: 'growth', fields: ['monthly_contribution', 'annual_return', 'horizon_years'] }, { key: 'safety', fields: ['emergency_fund'] }],
}
const props = defineProps({ mode: { type: String, required: true } })
const { t, locale, busy, error, notice, run, money } = useWealthPage()
const preferences = ref(null), report = ref(null)
const groups = computed(() => GROUPS[props.mode])
const riskSummary = computed(() => {
  const rules = (report.value?.rules || []).filter(rule => rule.enabled)
  return [
    { status: 'ok', icon: LuCircleCheck, tile: 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400' },
    { status: 'warning', icon: LuTriangleAlert, tile: 'bg-amber-500/10 text-amber-600 dark:text-amber-400' },
    { status: 'unknown', icon: LuCircleHelp, tile: 'bg-muted text-muted-foreground' },
  ].map(item => ({ ...item, count: rules.filter(rule => rule.status === item.status).length }))
})
function formatRule(rule, value) {
  if (value == null) return '—'
  if (rule.unit === 'money') return money(value, report.value.currency)
  const number = new Intl.NumberFormat(locale.value, { maximumFractionDigits: 2 }).format(value)
  return rule.unit === 'percent' ? `${number}%` : t('wealth.monthsValue', { months: number })
}
async function calculate() { report.value = (await wealthAPI.get(`/${props.mode}`)).data }
function save() { run(async () => { preferences.value = (await wealthAPI.save('/preferences', preferences.value, 'PUT')).data; await calculate() }, true) }
watch(() => props.mode, () => run(async () => { report.value = null; preferences.value = (await wealthAPI.get('/preferences')).data; await calculate() }), { immediate: true })
</script>
