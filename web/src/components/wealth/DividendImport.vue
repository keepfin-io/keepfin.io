<template>
  <div class="space-y-4 rounded-xl border p-4" data-testid="dividend-import"><p class="text-sm text-muted-foreground">{{ t('wealth.dividendHint') }}</p><PageFeedback :error="error" :busy="busy" :notice="notice" />
    <form class="grid items-end gap-3 sm:grid-cols-4" @submit.prevent="fetchRows"><label class="grid gap-1 text-sm">{{ t('portfolio.symbol') }}<Input v-model="symbol" required maxlength="20" :disabled="busy" /></label><label class="grid gap-1 text-sm">{{ t('wealth.from') }}<Input v-model="from" type="date" :max="today" required :disabled="busy" /></label><label class="grid gap-1 text-sm">{{ t('wealth.to') }}<Input v-model="to" type="date" :min="from" :max="today" required :disabled="busy" /></label><Button type="submit" variant="outline" :disabled="busy">{{ t('wealth.fetchDividends') }}</Button></form>
    <div v-for="(row,index) in rows" :key="row.activity.reference" class="flex flex-wrap items-center gap-3 border-t pt-3"><label class="flex items-center gap-2 text-sm"><input v-model="selected" type="checkbox" :value="index" :disabled="busy || !eligible(row)" @change="preview = null">{{ row.ex_date }} · {{ row.shares }} × {{ row.per_share ?? '—' }}</label><label class="grid gap-1 text-xs">{{ t('wealth.paymentDate') }}<Input v-model="row.activity.date" type="date" :max="today" :disabled="busy" @update:model-value="preview = null" /></label><span class="text-sm">{{ money(row.activity.unit_price,row.activity.currency || 'USD') }}</span><p v-if="!eligible(row)" class="w-full text-xs text-muted-foreground">{{ row.reason }}</p></div>
    <div v-if="rows.length" class="flex gap-3"><Button variant="outline" :disabled="busy || !selected.length" @click="check">{{ t('portfolio.preview') }}</Button><Button v-if="preview?.valid" :disabled="busy || previewPayload !== payload" @click="commit">{{ t('portfolio.applyImport') }}</Button></div>
    <div v-if="preview" class="text-sm"><p v-if="preview.error" class="text-destructive">{{ preview.error }}</p><p v-for="row in preview.rows" :key="row.line">{{ row.activity.symbol }} · {{ row.activity.date }} · {{ row.error || row.status }}</p></div>
  </div>
</template>
<script setup>
import { computed, ref } from 'vue'
import { Button, Input } from '@/components/ui'
import PageFeedback from './PageFeedback.vue'
import { wealthAPI } from '@/services/wealthApi'
import { portfolioAPI } from '@/services/portfolioApi'
import { useWealthPage } from '@/composables/useWealthPage'
const props = defineProps({ accountId: {type:String,required:true} })
const emit = defineEmits(['changed'])
const { t, busy, error, notice, run, money, today } = useWealthPage()
const symbol=ref(''),from=ref(`${today.slice(0,4)}-01-01`),to=ref(today),rows=ref([]),selected=ref([]),preview=ref(null),previewPayload=ref('')
const eligible = row => row.per_share > 0 && /^[A-Z]{3}$/.test(row.activity.currency) && /^\d{4}-\d{2}-\d{2}$/.test(row.activity.date) && row.activity.date <= today
const payload=computed(()=>JSON.stringify({version:1,activities:selected.value.map(i=>rows.value[i]).filter(eligible).map(r=>r.activity)}))
function fetchRows(){run(async()=>{preview.value=null;selected.value=[];rows.value=(await wealthAPI.get(`/accounts/${props.accountId}/dividends/preview`,{symbol:symbol.value.toUpperCase(),from:from.value,to:to.value})).data})}
function check(){run(async()=>{previewPayload.value=payload.value;preview.value=(await portfolioAPI.importJSON(props.accountId,payload.value,false)).data})}
function commit(){if(previewPayload.value!==payload.value)return;run(async()=>{const r=await portfolioAPI.importJSON(props.accountId,previewPayload.value,true);preview.value=r.data;if(r.data.valid){rows.value=[];selected.value=[];emit('changed')}},true)}
</script>
