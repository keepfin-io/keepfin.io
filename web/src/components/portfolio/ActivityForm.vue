<template>
  <div data-testid="activity-page" class="space-y-6" :aria-busy="busy">
      <Alert v-if="error" data-testid="activity-error" variant="error" :description="error" />
      <form id="activity-form" data-testid="activity-form" class="space-y-3" @submit.prevent="save">
        <div class="grid gap-3 sm:grid-cols-2">
          <div><Label for="activity-type">{{ $t('portfolio.type') }}</Label><FormSelect id="activity-type" data-testid="activity-type" v-model="form.type" :disabled="busy || activity?.type === 'opening'" :options="types.map(value => ({ value, label: $t(`portfolio.types.${value}`) }))" /></div>
          <div><Label for="activity-date">{{ $t('portfolio.date') }}</Label><Input id="activity-date" data-testid="activity-date" v-model="form.date" type="date" :max="today" :min="account.opening_date?.slice(0, 10)" required :disabled="busy" /></div>
          <div v-if="trade"><Label for="activity-symbol">{{ $t('portfolio.symbol') }}</Label><Input id="activity-symbol" data-testid="activity-symbol" v-model="form.symbol" placeholder="AAPL.US" maxlength="20" required :disabled="busy" /></div>
          <div><Label for="activity-currency">{{ $t('portfolio.currency') }}</Label><Input id="activity-currency" data-testid="activity-currency" v-model="form.currency" maxlength="3" required :disabled="busy" /></div>
          <div v-if="trade"><Label for="activity-quantity">{{ $t('portfolio.quantity') }}</Label><Input id="activity-quantity" data-testid="activity-quantity" v-model.number="form.quantity" type="number" min="0.00000001" step="0.00000001" required :disabled="busy" /></div>
          <div><Label for="activity-price">{{ $t(trade ? 'portfolio.unitPrice' : 'portfolio.amount') }}</Label><Input id="activity-price" data-testid="activity-price" v-model.number="form.unit_price" type="number" min="0.0001" step="0.0001" required :disabled="busy" /></div>
          <div><Label for="activity-fee">{{ $t('portfolio.fees') }}</Label><Input id="activity-fee" v-model.number="form.fee" type="number" min="0" step="0.0001" required :disabled="busy" /></div>
          <div v-if="form.currency.toUpperCase() !== account.currency"><Label for="activity-rate">1 {{ form.currency.toUpperCase() }} → {{ account.currency }}</Label><Input id="activity-rate" data-testid="activity-rate" v-model.number="form.exchange_rate" type="number" min="0.0000000001" step="any" :disabled="busy" :placeholder="$t('portfolio.automaticRate')" /></div>
        </div>
        <div><Label for="activity-notes">{{ $t('portfolio.notes') }}</Label><Textarea id="activity-notes" v-model="form.notes" maxlength="2000" :disabled="busy" /></div>
      </form>
      <div class="flex flex-col-reverse gap-3 border-t pt-5 sm:flex-row sm:justify-end"><Button data-testid="activity-cancel" variant="outline" :disabled="busy" @click="$emit('close')">{{ $t('common.cancel') }}</Button><Button type="submit" form="activity-form" data-testid="activity-save" :loading="busy" :loading-text="$t('common.saving')">{{ $t('common.save') }}</Button></div>
  </div>
</template>
<script setup>
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { activityAPI } from '@/services/activityApi'
import { Button, FormSelect, Input, Textarea, Label, FeedbackAlert as Alert } from '@/components/ui'
const props = defineProps({ account: { type: Object, required: true }, activity: { type: Object, default: null } })
const emit = defineEmits(['close', 'saved', 'busy'])
let active = true
onBeforeUnmount(() => { active = false })
const form = ref({ type: 'buy', currency: 'USD' }), busy = ref(false), error = ref('')
const today = new Date().toISOString().slice(0, 10)
const types = computed(() => props.activity?.type === 'opening' ? ['opening'] : ['buy', 'sell', 'deposit', 'withdrawal', 'dividend', 'interest', 'fee'])
const trade = computed(() => ['buy', 'sell', 'opening'].includes(form.value.type))
watch(() => props.activity, () => {
  error.value = ''
  const a = props.activity
  form.value = a ? { type: a.type, date: a.date.slice(0, 10), symbol: a.symbol, quantity: a.quantity, unit_price: ['buy', 'sell', 'opening'].includes(a.type) ? a.unit_price : a.unit_price * a.quantity, fee: a.fee, currency: a.currency, exchange_rate: a.exchange_rate || '', notes: a.notes || '' } : { type: 'buy', date: today, symbol: '', quantity: 1, unit_price: '', fee: 0, currency: props.account.currency, exchange_rate: '', notes: '' }
}, { immediate: true })
watch(busy, value => emit('busy', value))
async function save() {
  if (busy.value) return
  busy.value = true; error.value = ''
  try {
    const data = { ...form.value, quantity: trade.value ? Number(form.value.quantity) : 1, symbol: trade.value || props.activity ? form.value.symbol : '', unit_price: Number(form.value.unit_price), fee: Number(form.value.fee), exchange_rate: Number(form.value.exchange_rate) || 0 }
    if (props.activity) await activityAPI.updateActivity(props.account.id, props.activity.id, data)
    else await activityAPI.createActivity(props.account.id, data)
    if (active) emit('saved')
  } catch (e) { error.value = e.message; busy.value = false }
}
</script>
