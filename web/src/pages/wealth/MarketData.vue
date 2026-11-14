<template>
  <div class="space-y-6" data-testid="market-data-page">
    <SectionHeader :icon="LuDatabase" :title="t('wealth.marketData')" :description="t('wealth.marketDataHint')" />
    <PageFeedback :error="error" :busy="busy" :notice="notice" />

    <div v-if="configuration" data-testid="market-provider" class="overflow-hidden rounded-2xl border bg-card shadow-xs">
      <div class="flex flex-wrap items-center gap-3 border-b px-5 py-4">
        <LuServer class="size-4 text-primary" aria-hidden="true" />
        <h3 class="flex-1 font-semibold">{{ t('wealth.currentProvider') }}</h3>
        <span class="rounded-full bg-primary/10 px-3 py-1 text-sm font-semibold text-primary">{{ configuration.provider }}</span>
      </div>
      <ul role="list" class="divide-y">
        <li v-for="p in configuration.providers" :key="p.id" class="flex flex-wrap items-center gap-3 px-5 py-3">
          <p class="flex-1 font-medium">{{ p.id }}</p>
          <StatusBadge :tone="p.configured ? 'success' : 'neutral'">{{ t(p.configured ? 'wealth.configured' : 'wealth.notConfigured') }}</StatusBadge>
          <StatusBadge v-if="configuration.provider === p.id" tone="info">{{ t('wealth.inUse') }}</StatusBadge>
          <Button v-else-if="admin" size="sm" variant="outline" :disabled="busy || !p.configured" @click="selectProvider(p.id)">{{ t('wealth.useProvider') }}</Button>
        </li>
      </ul>
      <p class="flex items-start gap-2 border-t bg-muted/30 px-5 py-3 text-xs text-muted-foreground"><LuInfo class="mt-0.5 size-3.5 shrink-0" aria-hidden="true" />{{ t('wealth.providerHint') }}</p>
    </div>

    <form data-testid="market-job-form" class="overflow-hidden rounded-2xl border bg-card shadow-xs" @submit.prevent="queue">
      <div class="border-b px-5 py-4">
        <h3 class="flex items-center gap-2 font-semibold"><LuHistory class="size-4 text-primary" aria-hidden="true" />{{ t('wealth.refreshHistory') }}</h3>
        <p class="mt-1 text-xs text-muted-foreground">{{ t('wealth.jobHint') }}</p>
      </div>
      <div class="grid items-end gap-4 p-5 sm:grid-cols-2 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_minmax(0,1fr)_auto]">
        <div class="space-y-2"><Label for="job-symbol"><LuTrendingUp class="size-4 text-muted-foreground" aria-hidden="true" />{{ t('portfolio.symbol') }}</Label><Input id="job-symbol" v-model="form.symbol" placeholder="AAPL.US" maxlength="20" required :disabled="busy" /></div>
        <div class="space-y-2"><Label for="job-from"><LuCalendar class="size-4 text-muted-foreground" aria-hidden="true" />{{ t('wealth.from') }}</Label><Input id="job-from" v-model="form.from" type="date" :max="today" required :disabled="busy" /></div>
        <div class="space-y-2"><Label for="job-to"><LuCalendar class="size-4 text-muted-foreground" aria-hidden="true" />{{ t('wealth.to') }}</Label><Input id="job-to" v-model="form.to" type="date" :min="form.from" :max="today" required :disabled="busy" /></div>
        <Button type="submit" :disabled="busy"><LuRefreshCw class="size-4" aria-hidden="true" />{{ t('wealth.queueJob') }}</Button>
      </div>
    </form>

    <section class="space-y-3" :aria-label="t('wealth.jobs')">
      <div class="flex items-center justify-between gap-3">
        <h3 class="font-semibold">{{ t('wealth.jobs') }}</h3>
        <Button variant="ghost" size="sm" :disabled="busy" @click="run(load)"><LuRefreshCw class="size-4" :class="busy && 'animate-spin'" aria-hidden="true" />{{ t('portfolio.refresh') }}</Button>
      </div>
      <div data-testid="market-jobs" class="overflow-hidden rounded-2xl border bg-card shadow-xs">
        <ul v-if="jobs.length" role="list" class="divide-y">
          <li v-for="job in jobs" :key="job.id" :data-testid="`market-job-${job.id}`" class="flex flex-wrap items-center gap-x-4 gap-y-2 px-5 py-3">
            <div class="min-w-0 flex-1">
              <p class="font-medium">{{ job.symbol }}</p>
              <p class="text-xs text-muted-foreground">{{ job.from }} → {{ job.to }} · {{ t('wealth.attempts') }}: {{ job.attempts }}</p>
              <p v-if="job.error" class="mt-1 text-xs text-destructive">{{ job.error }}</p>
            </div>
            <StatusBadge :tone="JOB_TONES[job.status] || 'neutral'" :pulse="job.status === 'running'">{{ t(`wealth.jobStatus.${job.status}`, job.status) }}</StatusBadge>
            <Button v-if="job.status === 'failed'" variant="outline" size="sm" :disabled="busy" @click="retry(job)"><LuRotateCcw class="size-3.5" aria-hidden="true" />{{ t('common.retry') }}</Button>
          </li>
        </ul>
        <EmptyState v-else-if="configuration" :icon="LuHistory" :title="t('wealth.emptyJobsTitle')" :description="t('wealth.emptyJobs')" />
      </div>
    </section>

    <section v-if="admin" data-testid="market-corrections" class="overflow-hidden rounded-2xl border bg-card shadow-xs" :aria-label="t('wealth.corrections')">
      <div class="border-b px-5 py-4">
        <h3 class="flex items-center gap-2 font-semibold"><LuWrench class="size-4 text-primary" aria-hidden="true" />{{ t('wealth.corrections') }}</h3>
        <p class="mt-1 flex items-start gap-1.5 text-xs text-amber-700 dark:text-amber-400"><LuTriangleAlert class="mt-0.5 size-3.5 shrink-0" aria-hidden="true" />{{ t('wealth.correctionHint') }}</p>
      </div>
      <form class="grid items-end gap-4 p-5 sm:grid-cols-2" @submit.prevent="correct">
        <div class="space-y-2"><Label for="correction-symbol">{{ t('portfolio.symbol') }}</Label><Input id="correction-symbol" v-model="correction.symbol" placeholder="AAPL.US" maxlength="20" required :disabled="busy" /></div>
        <div class="space-y-2"><Label for="correction-date">{{ t('portfolio.date') }}</Label><Input id="correction-date" v-model="correction.date" type="date" :max="today" required :disabled="busy" /></div>
        <div class="space-y-2"><Label for="correction-close">{{ t('holdingDetail.closePrice') }}</Label><Input id="correction-close" v-model.number="correction.close" type="number" min="0.0000000001" step="any" required :disabled="busy" /></div>
        <div class="space-y-2"><Label for="correction-reason">{{ t('wealth.reason') }}</Label><Input id="correction-reason" v-model="correction.reason" maxlength="500" required :disabled="busy" /></div>
        <div class="flex justify-end sm:col-span-2"><Button type="submit" :disabled="busy">{{ t('wealth.saveCorrection') }}</Button></div>
      </form>
      <ul v-if="corrections.length" role="list" class="divide-y border-t">
        <li v-for="row in corrections" :key="`${row.symbol}-${row.date}`" class="flex flex-wrap items-center gap-3 px-5 py-3 text-sm">
          <div class="min-w-0 flex-1"><p><strong>{{ row.symbol }}</strong> · {{ row.date }} · <span class="tabular-nums">{{ row.close }}</span></p><p class="text-xs text-muted-foreground">{{ row.reason }}</p></div>
          <Button variant="outline" size="sm" :disabled="busy" @click="deleteCorrection(row)"><LuRotateCcw class="size-3.5" aria-hidden="true" />{{ t('wealth.useProviderPrice') }}</Button>
        </li>
      </ul>
    </section>
  </div>
</template>
<script setup>
import { computed, onMounted, ref } from 'vue'
import { LuCalendar, LuDatabase, LuHistory, LuInfo, LuRefreshCw, LuRotateCcw, LuServer, LuTrendingUp, LuTriangleAlert, LuWrench } from '@kalimahapps/vue-icons/lu'
import { Button, Input, Label } from '@/components/ui'
import PageFeedback from '@/components/wealth/PageFeedback.vue'
import SectionHeader from '@/components/wealth/SectionHeader.vue'
import EmptyState from '@/components/wealth/EmptyState.vue'
import StatusBadge from '@/components/wealth/StatusBadge.vue'
import { useWealthPage } from '@/composables/useWealthPage'
import { useAuth } from '@/composables/useAuth'
import { wealthAPI } from '@/services/wealthApi'
const JOB_TONES = { queued: 'neutral', running: 'info', complete: 'success', failed: 'danger' }
const { t, busy, error, notice, run, today } = useWealthPage()
const { user } = useAuth(), admin = computed(() => user.value?.user_role === 'admin')
const configuration = ref(null), jobs = ref([]), corrections = ref([])
const form = ref({ symbol:'', from:new Date(Date.now()-30*86400000).toISOString().slice(0,10), to:today })
const correction = ref({ symbol:'', date:today, close:'', reason:'' })
async function load() { const [c,j] = await Promise.all([wealthAPI.get('/market/providers'),wealthAPI.get('/market/jobs')]); configuration.value=c.data;jobs.value=j.data;if(admin.value) corrections.value=(await wealthAPI.get('/market/corrections')).data }
function selectProvider(provider) { run(async () => {await wealthAPI.save('/market/providers',{provider},'PUT');await load()},true) }
function queue() {run(async () => {await wealthAPI.save('/market/jobs',form.value);await load()},true)}
function retry(job) {run(async () => {await wealthAPI.save(`/market/jobs/${job.id}/retry`,{});await load()},true)}
function correct() {run(async () => {await wealthAPI.save('/market/corrections',correction.value,'PUT');await load()},true)}
function deleteCorrection(row) {run(async () => {await wealthAPI.remove(`/market/corrections?${new URLSearchParams({symbol:row.symbol,date:row.date})}`);await load()},true)}
onMounted(() => run(load))
</script>
