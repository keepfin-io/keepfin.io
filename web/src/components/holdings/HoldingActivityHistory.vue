<template>
  <section data-testid="holding-activity-history" class="min-w-0 space-y-4 rounded-2xl border bg-card p-5 sm:p-6" aria-labelledby="holding-activities-heading">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div><h2 id="holding-activities-heading" class="font-semibold">{{ t('holdingDetail.activityHistory') }}</h2><p class="mt-1 text-xs text-muted-foreground">{{ t('holdingDetail.activityHint') }}</p></div>
      <Button as-child variant="outline" size="sm"><router-link data-testid="holding-all-activities" :to="{ name: 'AccountDetail', params: { id: account.id }, query: { tab: 'activities' } }">{{ t('holdingDetail.allActivities') }}<ArrowUpRight class="size-4" /></router-link></Button>
    </div>
    <template v-if="error">
      <Alert data-testid="holding-activities-error" variant="warning" :description="t('holdingDetail.activitiesError')" />
      <Button data-testid="holding-activities-retry" variant="outline" @click="$emit('retry')">{{ t('common.retry') }}</Button>
    </template>
    <template v-else>
      <p v-if="!activities.length" data-testid="holding-activities-empty" class="py-8 text-center text-sm text-muted-foreground">{{ t('accountDetail.noActivitiesRecorded') }}</p>
      <template v-else>
        <FormSelect data-testid="holding-activity-type" class="w-full sm:w-48" v-model="type" :aria-label="t('portfolio.type')" :options="typeOptions" />
        <ul class="divide-y">
          <li v-for="activity in visibleRows" :key="activity.id" :data-testid="`holding-activity-${activity.id}`" class="space-y-3 py-4">
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div class="flex flex-wrap items-center gap-2 text-sm"><span class="font-medium">{{ t(`portfolio.types.${activity.type}`) }}</span><Badge v-if="!activity.ledger" variant="outline">{{ t('portfolio.archived') }}</Badge><time class="text-xs text-muted-foreground" :datetime="activity.date">{{ activity.date?.slice(0, 10) || '—' }}</time></div>
              <Button v-if="account.is_active && activity.ledger && !activity.transfer_id" as-child size="sm" variant="ghost"><router-link :data-testid="`holding-activity-edit-${activity.id}`" :to="{ name: 'ActivityEdit', params: { id: account.id, activityId: activity.id } }"><Pencil class="size-3.5" />{{ t('common.edit') }}</router-link></Button>
            </div>
            <dl class="grid grid-cols-2 gap-3 text-sm sm:grid-cols-3">
              <div><dt class="text-xs text-muted-foreground">{{ t('portfolio.quantity') }}</dt><dd class="mt-1 break-all tabular-nums">{{ activity.type === 'split' ? `${activity.split_numerator}:${activity.split_denominator}` : formatNumber(activity.quantity, locale, 8) }}</dd></div>
              <div><dt class="text-xs text-muted-foreground">{{ t('portfolio.unitPrice') }}</dt><dd class="mt-1 break-all tabular-nums">{{ activity.type === 'split' ? '—' : formatMoney(activity.unit_price, activity.currency, locale) }}</dd></div>
              <div><dt class="text-xs text-muted-foreground">{{ t('portfolio.fees') }}</dt><dd class="mt-1 break-all tabular-nums">{{ formatMoney(activity.fee, activity.currency, locale) }}</dd></div>
            </dl>
            <p v-if="activity.notes" class="whitespace-pre-wrap break-words text-xs leading-relaxed text-muted-foreground">{{ activity.notes }}</p>
          </li>
        </ul>
        <nav v-if="filteredRows.length" class="flex flex-wrap items-center justify-between gap-3 border-t pt-4 text-xs text-muted-foreground" :aria-label="t('holdings.pagination')">
          <p data-testid="holding-activities-count" aria-live="polite">{{ t('holdings.showing', { from: (page - 1) * pageSize + 1, to: Math.min(page * pageSize, filteredRows.length), total: filteredRows.length }) }}</p>
          <div class="flex items-center gap-2"><Button data-testid="holding-activities-prev" variant="ghost" size="icon" :disabled="page === 1" :aria-label="t('holdings.previousPage')" @click="page--"><ChevronLeft class="size-4" /></Button><span>{{ page }} / {{ pages }}</span><Button data-testid="holding-activities-next" variant="ghost" size="icon" :disabled="page >= pages" :aria-label="t('holdings.nextPage')" @click="page++"><ChevronRight class="size-4" /></Button></div>
        </nav>
      </template>
    </template>
  </section>
</template>
<script setup>
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ArrowUpRight, ChevronLeft, ChevronRight, Pencil } from 'lucide-vue-next'
import { Badge, Button, FormSelect, FeedbackAlert as Alert } from '@/components/ui'
import { formatMoney, formatNumber } from '@/lib/portfolio-format'

const props = defineProps({ account: { type: Object, required: true }, activities: { type: Array, default: () => [] }, error: { type: String, default: '' } })
defineEmits(['retry'])
const { t, locale } = useI18n()
const type = ref('all'), page = ref(1), pageSize = 10
const typeOptions = computed(() => [{ value: 'all', label: t('holdingDetail.allTypes') }, ...[...new Set(props.activities.map(row => row.type))].sort().map(value => ({ value, label: t(`portfolio.types.${value}`) }))])
const filteredRows = computed(() => props.activities.filter(row => type.value === 'all' || row.type === type.value).slice().sort((a, b) => String(b.date).localeCompare(String(a.date)) || a.id.localeCompare(b.id)))
const pages = computed(() => Math.max(1, Math.ceil(filteredRows.value.length / pageSize)))
const visibleRows = computed(() => filteredRows.value.slice((page.value - 1) * pageSize, page.value * pageSize))
watch([type, () => props.activities], () => { page.value = 1 })
</script>
