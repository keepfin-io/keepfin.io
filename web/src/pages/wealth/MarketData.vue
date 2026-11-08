<template>
  <div class="space-y-5" data-testid="market-data-page"><h2 class="text-xl font-semibold">{{ t('wealth.marketData') }}</h2><PageFeedback :error="error" :busy="busy" :notice="notice" />
    <Card v-if="configuration"><CardContent class="space-y-3 pt-6"><p>{{ t('wealth.currentProvider') }}: <strong>{{ configuration.provider }}</strong></p><p class="text-xs text-muted-foreground">{{ t('wealth.providerHint') }}</p><div v-if="admin" class="flex flex-wrap gap-3"><Button v-for="p in configuration.providers" :key="p.id" variant="outline" :disabled="busy || !p.configured || configuration.provider === p.id" @click="selectProvider(p.id)">{{ p.id }} · {{ t(p.configured ? 'wealth.configured' : 'wealth.notConfigured') }}</Button></div></CardContent></Card>
    <Card><CardContent class="space-y-3 pt-6"><h3 class="font-medium">{{ t('wealth.refreshHistory') }}</h3><p class="text-sm text-muted-foreground">{{ t('wealth.jobHint') }}</p><form class="grid items-end gap-3 sm:grid-cols-2 lg:grid-cols-4" @submit.prevent="queue"><label class="grid gap-2 text-sm">{{ t('portfolio.symbol') }}<Input v-model="form.symbol" placeholder="AAPL.US" maxlength="20" required :disabled="busy" /></label><label class="grid gap-2 text-sm">{{ t('wealth.from') }}<Input v-model="form.from" type="date" :max="today" required :disabled="busy" /></label><label class="grid gap-2 text-sm">{{ t('wealth.to') }}<Input v-model="form.to" type="date" :min="form.from" :max="today" required :disabled="busy" /></label><Button type="submit" :disabled="busy">{{ t('wealth.queueJob') }}</Button></form></CardContent></Card>
    <div class="flex items-center justify-between"><h3 class="font-medium">{{ t('wealth.jobs') }}</h3><Button variant="ghost" :disabled="busy" @click="run(load)">{{ t('portfolio.refresh') }}</Button></div>
    <div class="divide-y rounded-xl border"><div v-for="job in jobs" :key="job.id" class="flex flex-wrap items-center gap-3 p-4"><div class="flex-1"><p class="font-medium">{{ job.symbol }} <Badge variant="outline">{{ job.status }}</Badge></p><p class="mt-1 text-xs text-muted-foreground">{{ job.from }} — {{ job.to }} · {{ t('wealth.attempts') }}: {{ job.attempts }}</p><p v-if="job.error" class="mt-1 text-xs text-destructive">{{ job.error }}</p></div><Button v-if="job.status === 'failed'" variant="outline" size="sm" :disabled="busy" @click="retry(job)">{{ t('common.retry') }}</Button></div></div>
    <template v-if="admin"><h3 class="font-medium">{{ t('wealth.corrections') }}</h3><p class="text-xs text-muted-foreground">{{ t('wealth.correctionHint') }}</p><form class="grid items-end gap-3 sm:grid-cols-2" @submit.prevent="correct"><label class="grid gap-2 text-sm">{{ t('portfolio.symbol') }}<Input v-model="correction.symbol" maxlength="20" required :disabled="busy" /></label><label class="grid gap-2 text-sm">{{ t('portfolio.date') }}<Input v-model="correction.date" type="date" :max="today" required :disabled="busy" /></label><label class="grid gap-2 text-sm">{{ t('holdingDetail.closePrice') }}<Input v-model.number="correction.close" type="number" min="0.0000000001" step="any" required :disabled="busy" /></label><label class="grid gap-2 text-sm">{{ t('wealth.reason') }}<Input v-model="correction.reason" maxlength="500" required :disabled="busy" /></label><Button type="submit" :disabled="busy">{{ t('wealth.saveCorrection') }}</Button></form><div class="divide-y rounded-xl border"><div v-for="row in corrections" :key="`${row.symbol}-${row.date}`" class="flex flex-wrap items-center gap-3 p-3 text-sm"><div class="flex-1"><strong>{{ row.symbol }}</strong> · {{ row.date }} · {{ row.close }}<p class="text-xs text-muted-foreground">{{ row.reason }}</p></div><Button variant="ghost" size="sm" :disabled="busy" @click="deleteCorrection(row)">{{ t('wealth.useProviderPrice') }}</Button></div></div></template>
  </div>
</template>
<script setup>
import { computed, onMounted, ref } from 'vue'
import { Button, Input, Card, CardContent, Badge } from '@/components/ui'
import PageFeedback from '@/components/wealth/PageFeedback.vue'
import { useWealthPage } from '@/composables/useWealthPage'
import { useAuth } from '@/composables/useAuth'
import { wealthAPI } from '@/services/wealthApi'
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
