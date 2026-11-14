<template>
  <AccountForm v-if="action === 'account-edit'" :account="account" @close="close" @saved="$emit('saved', action)" />
  <section v-else data-testid="account-action-page" class="mx-auto w-full space-y-6" :class="action === 'holding-create' ? 'max-w-4xl' : 'max-w-3xl'">
    <div class="space-y-4">
      <Button data-testid="account-action-back" variant="ghost" class="-ml-3 min-h-11 rounded-xl text-muted-foreground" :disabled="busy" @click="close"><ArrowLeft class="size-4" />{{ $t('accountDetail.backToAccount') }}</Button>
      <div><h1 class="text-2xl font-semibold tracking-tight">{{ $t(title) }}</h1><p class="mt-2 text-sm text-muted-foreground">{{ $t(hint) }}</p></div>
    </div>
    <Alert v-if="unavailable" data-testid="account-action-error" variant="error" :description="$t('accountDetail.actionUnavailable')" />
    <Card v-else class="rounded-2xl" :class="action === 'holding-create' && 'py-0'"><CardContent class="p-5 sm:p-7">
      <AddHoldingForm v-if="action === 'holding-create'" :account-id="account.id" @busy="busy = $event" @success="$emit('saved', action)" />
      <EditHoldingForm v-else-if="action === 'holding-edit'" :account-id="account.id" :holding="holding" @busy="busy = $event" @success="$emit('saved', action)" />
      <ActivityForm v-else-if="action === 'activity-create' || action === 'activity-edit'" :account="account" :activity="activity" @busy="busy = $event" @close="close" @saved="$emit('saved', action)" />
      <TransferForm v-else-if="action === 'transfer'" :account="account" @busy="busy = $event" @close="close" @saved="$emit('saved', action)" />
    </CardContent></Card>
  </section>
</template>
<script setup>
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft } from 'lucide-vue-next'
import { Button, Card, CardContent, FeedbackAlert as Alert } from '@/components/ui'
import AddHoldingForm from '@/components/holdings/AddHoldingForm.vue'
import EditHoldingForm from '@/components/holdings/EditHoldingForm.vue'
import ActivityForm from '@/components/portfolio/ActivityForm.vue'
import TransferForm from '@/components/portfolio/TransferForm.vue'
import AccountForm from './AccountForm.vue'
const props = defineProps({ account: { type: Object, required: true }, holdings: { type: Array, required: true }, activities: { type: Array, required: true } })
defineEmits(['saved'])
const route = useRoute(), router = useRouter(), busy = ref(false)
const action = computed(() => route.meta.accountAction || '')
const holding = computed(() => props.holdings.find(item => item.id === route.params.holdingId))
const activity = computed(() => props.activities.find(item => item.id === route.params.activityId) || null)
const unavailable = computed(() => {
  if (action.value === 'holding-edit') return !holding.value
  if (action.value === 'activity-edit') return !activity.value?.ledger || !!activity.value?.transfer_id
  return !props.account.is_active
})
const title = computed(() => ({ 'holding-create': 'accountDetail.addHolding', 'holding-edit': 'holdings.editHolding', 'activity-create': 'portfolio.addActivity', 'activity-edit': 'portfolio.editActivity', transfer: 'portfolio.transfer' })[action.value])
const hint = computed(() => action.value === 'holding-create' ? 'holdings.addHint' : action.value === 'holding-edit' ? 'holdings.editHint' : action.value === 'transfer' ? 'portfolio.transferHint' : 'portfolio.activityHint')
function close() {
  if (busy.value) return
  router.replace({ name: 'AccountDetail', params: { id: props.account.id }, query: action.value.startsWith('activity') || action.value === 'transfer' ? { tab: 'activities' } : {} })
}
</script>
