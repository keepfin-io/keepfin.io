<template>
  <div data-testid="transfer-page" class="space-y-6" :aria-busy="busy">
      <Alert v-if="error" data-testid="transfer-error" variant="error" :description="error" />
      <form id="transfer-form" data-testid="transfer-form" class="space-y-3" @submit.prevent="save">
        <div><Label for="transfer-target">{{ $t('portfolio.destination') }}</Label><FormSelect id="transfer-target" data-testid="transfer-target" v-model="form.to_account_id" :disabled="busy" :options="accounts.map(a => ({ value: a.id, label: `${a.account_name} (${a.currency})` }))" /></div>
        <div><Label for="transfer-date">{{ $t('portfolio.date') }}</Label><Input id="transfer-date" v-model="form.date" type="date" :max="today" required :disabled="busy" /></div>
        <div><Label for="transfer-amount">{{ $t('portfolio.amount') }} ({{ account.currency }})</Label><Input id="transfer-amount" data-testid="transfer-amount" v-model.number="form.amount" type="number" min="0.0001" step="0.0001" required :disabled="busy" /></div>
        <div v-if="target && target.currency !== account.currency"><Label for="transfer-received">{{ $t('portfolio.received') }} ({{ target.currency }})</Label><Input id="transfer-received" v-model.number="form.received_amount" type="number" min="0.0001" step="0.0001" required :disabled="busy" /></div>
        <div><Label for="transfer-fee">{{ $t('portfolio.fees') }} ({{ account.currency }})</Label><Input id="transfer-fee" v-model.number="form.fee" type="number" min="0" step="0.0001" :disabled="busy" /></div>
      </form>
      <div class="flex flex-col-reverse gap-3 border-t pt-5 sm:flex-row sm:justify-end"><Button data-testid="transfer-cancel" variant="outline" :disabled="busy" @click="$emit('close')">{{ $t('common.cancel') }}</Button><Button type="submit" form="transfer-form" :disabled="busy || !target" :loading="busy" :loading-text="$t('common.saving')" data-testid="transfer-save">{{ $t('common.save') }}</Button></div>
  </div>
</template>
<script setup>
import { ref, computed, onBeforeUnmount, watch } from 'vue'
import { accountAPI } from '@/services/accountApi'
import { portfolioAPI } from '@/services/portfolioApi'
import { Button, FormSelect, Input, Label, FeedbackAlert as Alert } from '@/components/ui'
const props = defineProps({ account: { type: Object, required: true } })
const emit = defineEmits(['close', 'saved', 'busy'])
let active = true
onBeforeUnmount(() => { active = false })
const accounts = ref([]), busy = ref(false), error = ref(''), form = ref({})
const today = new Date().toISOString().slice(0, 10)
const target = computed(() => accounts.value.find(a => a.id === form.value.to_account_id))
watch(() => props.account.id, async () => { error.value = ''; form.value = { to_account_id: '', date: today, amount: '', received_amount: '', fee: 0 }; try { const data = await accountAPI.getAccounts(); if (active) accounts.value = data.accounts.filter(a => a.id !== props.account.id && a.is_active) } catch (e) { error.value = e.message } }, { immediate: true })
watch(busy, value => emit('busy', value))
async function save() {
  if (busy.value || !target.value) return
  busy.value = true; error.value = ''
  try { await portfolioAPI.transfer({ ...form.value, from_account_id: props.account.id, received_amount: target.value.currency === props.account.currency ? form.value.amount : form.value.received_amount }); if (active) emit('saved') }
  catch (e) { error.value = e.message; busy.value = false }
}
</script>
