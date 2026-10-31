<template>
  <Dialog :open="open" @update:open="!$event && !busy && $emit('close')">
    <DialogContent :show-close-button="!busy" @interact-outside="busy && $event.preventDefault()" @escape-key-down="busy && $event.preventDefault()">
      <DialogHeader><DialogTitle>{{ $t('portfolio.transfer') }}</DialogTitle><DialogDescription>{{ $t('portfolio.transferHint') }}</DialogDescription></DialogHeader>
      <Alert v-if="error" variant="error" :description="error" />
      <form id="transfer-form" data-testid="transfer-form" class="space-y-3" @submit.prevent="save">
        <div><Label for="transfer-target">{{ $t('portfolio.destination') }}</Label><FormSelect id="transfer-target" v-model="form.to_account_id" :disabled="busy" :options="accounts.map(a => ({ value: a.id, label: `${a.account_name} (${a.currency})` }))" /></div>
        <div><Label for="transfer-date">{{ $t('portfolio.date') }}</Label><Input id="transfer-date" v-model="form.date" type="date" :max="today" required :disabled="busy" /></div>
        <div><Label for="transfer-amount">{{ $t('portfolio.amount') }} ({{ account.currency }})</Label><Input id="transfer-amount" data-testid="transfer-amount" v-model.number="form.amount" type="number" min="0.0001" step="0.0001" required :disabled="busy" /></div>
        <div v-if="target && target.currency !== account.currency"><Label for="transfer-received">{{ $t('portfolio.received') }} ({{ target.currency }})</Label><Input id="transfer-received" v-model.number="form.received_amount" type="number" min="0.0001" step="0.0001" required :disabled="busy" /></div>
        <div><Label for="transfer-fee">{{ $t('portfolio.fees') }} ({{ account.currency }})</Label><Input id="transfer-fee" v-model.number="form.fee" type="number" min="0" step="0.0001" :disabled="busy" /></div>
      </form>
      <DialogFooter><Button variant="outline" :disabled="busy" @click="$emit('close')">{{ $t('common.cancel') }}</Button><Button type="submit" form="transfer-form" :disabled="busy || !target" data-testid="transfer-save">{{ $t('common.save') }}</Button></DialogFooter>
    </DialogContent>
  </Dialog>
</template>
<script setup>
import { ref, computed, watch } from 'vue'
import { accountAPI } from '@/services/accountApi'
import { portfolioAPI } from '@/services/portfolioApi'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter, Button, FormSelect, Input, Label, FeedbackAlert as Alert } from '@/components/ui'
const props = defineProps({ open: Boolean, account: { type: Object, required: true } })
const emit = defineEmits(['close', 'saved'])
const accounts = ref([]), busy = ref(false), error = ref(''), form = ref({})
const today = new Date().toISOString().slice(0, 10)
const target = computed(() => accounts.value.find(a => a.id === form.value.to_account_id))
watch(() => props.open, async open => { if (!open) return; error.value = ''; form.value = { to_account_id: '', date: today, amount: '', received_amount: '', fee: 0 }; try { const data = await accountAPI.getAccounts(); accounts.value = data.accounts.filter(a => a.id !== props.account.id && a.is_active) } catch (e) { error.value = e.message } })
async function save() {
  if (busy.value || !target.value) return
  busy.value = true; error.value = ''
  try { await portfolioAPI.transfer({ ...form.value, from_account_id: props.account.id, received_amount: target.value.currency === props.account.currency ? form.value.amount : form.value.received_amount }); emit('saved'); emit('close') }
  catch (e) { error.value = e.message } finally { busy.value = false }
}
</script>
