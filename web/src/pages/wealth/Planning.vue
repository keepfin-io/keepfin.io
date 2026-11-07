<template>
  <div class="space-y-5" :data-testid="`${mode}-page`">
    <h2 class="text-xl font-semibold">{{ t(`wealth.${mode}`) }}</h2><p class="text-sm text-muted-foreground">{{ t(`wealth.${mode}Hint`) }}</p><PageFeedback :error="error" :busy="busy" :notice="notice" />
    <Card v-if="preferences"><CardContent class="pt-6"><form class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3" @submit.prevent="save">
      <label v-for="key in fields" :key="key" class="grid gap-2 text-sm">{{ t(`wealth.${key}`) }}<Input v-model.number="preferences[key]" type="number" :min="key === 'annual_return' ? -99 : 0" :step="key === 'horizon_years' ? 1 : 'any'" required :disabled="busy" /></label>
      <div v-if="mode === 'risk'" class="space-y-2 sm:col-span-2 lg:col-span-3"><p class="text-sm">{{ t('wealth.disabledRules') }}</p><div class="flex flex-wrap gap-4"><label v-for="id in ['asset','account','currency','fees','emergency','liquidity']" :key="id" class="flex gap-2 text-sm"><input v-model="preferences.disabled_rules" type="checkbox" :value="id" :disabled="busy">{{ t(`wealth.rule_${id}`) }}</label></div></div>
      <div class="flex items-end"><Button type="submit" :disabled="busy">{{ t('wealth.saveAndCalculate') }}</Button></div>
    </form></CardContent></Card>
    <template v-if="report && mode === 'risk'"><div class="grid gap-3 md:grid-cols-2"><Card v-for="rule in report.rules" :key="rule.id" :class="!rule.enabled && 'opacity-60'"><CardContent class="space-y-2 pt-6"><div class="flex items-center justify-between gap-3"><h3 class="font-medium">{{ t(`wealth.rule_${rule.id}`) }}</h3><Badge variant="outline">{{ t(`wealth.${rule.enabled ? rule.status : 'disabled'}`) }}</Badge></div><p class="text-2xl font-semibold">{{ rule.value == null ? '—' : rule.unit === 'money' ? money(rule.value,report.currency) : `${rule.value.toFixed(2)}${rule.unit === 'percent' ? '%' : ''}` }}</p><p class="text-xs text-muted-foreground">{{ rule.subject }} · {{ t('wealth.threshold') }}: {{ rule.limit }}</p></CardContent></Card></div><p class="text-xs text-muted-foreground">{{ t('wealth.feeRuleHint') }}</p></template>
    <template v-if="report && mode === 'fire'"><div class="grid gap-3 sm:grid-cols-3"><Card v-for="key in ['investable','target','years_to_target']" :key="key"><CardContent class="pt-6"><p class="text-sm text-muted-foreground">{{ t(`wealth.${key}`) }}</p><p class="mt-2 text-2xl font-semibold">{{ key === 'years_to_target' ? report[key] == null ? '—' : report[key].toFixed(1) : money(report[key],report.currency) }}</p></CardContent></Card></div><PortfolioChart :points="report.projection.map(p => ({ date: `${t('wealth.year')} ${p.year}`, value: p.value, net_flow: p.contributions }))" /><p class="text-xs text-muted-foreground">{{ t('wealth.projectionHint') }}</p></template>
    <p v-for="warning in report?.warnings || []" :key="warning" class="text-sm text-muted-foreground">{{ warning }}</p>
  </div>
</template>
<script setup>
import { computed, ref, watch } from 'vue'
import { Button, Input, Card, CardContent, Badge } from '@/components/ui'
import PageFeedback from '@/components/wealth/PageFeedback.vue'
import PortfolioChart from '@/components/portfolio/PortfolioChart.vue'
import { useWealthPage } from '@/composables/useWealthPage'
import { wealthAPI } from '@/services/wealthApi'
const props = defineProps({ mode: { type: String, required: true } })
const { t, busy, error, notice, run, money } = useWealthPage()
const preferences = ref(null), report = ref(null)
const fields = computed(() => props.mode === 'risk' ? ['max_asset_weight','max_account_weight','max_currency_weight','max_fee_ratio','emergency_fund','monthly_expenses','emergency_months'] : ['annual_expenses','monthly_contribution','annual_return','withdrawal_rate','horizon_years','emergency_fund'])
async function calculate() { report.value = (await wealthAPI.get(`/${props.mode}`)).report }
function save() { run(async () => { preferences.value = (await wealthAPI.save('/preferences', preferences.value, 'PUT')).preferences; await calculate() }, true) }
watch(() => props.mode, () => run(async () => { report.value = null; preferences.value = (await wealthAPI.get('/preferences')).preferences; await calculate() }), { immediate: true })
</script>
