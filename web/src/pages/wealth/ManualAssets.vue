<template>
  <div class="space-y-5" data-testid="manual-assets-page">
    <h2 class="text-xl font-semibold">{{ t('wealth.manualAssets') }}</h2><p class="text-sm text-muted-foreground">{{ t('wealth.manualHint') }}</p>
    <PageFeedback :error="error" :busy="busy" :notice="notice" />
    <Card><CardContent class="pt-6"><form class="grid gap-4 sm:grid-cols-2" @submit.prevent="save">
      <label class="grid gap-2 text-sm">{{ t('wealth.name') }}<Input v-model="form.name" required maxlength="100" :disabled="busy" /></label>
      <label class="grid gap-2 text-sm">{{ t('portfolio.currency') }}<Input v-model="form.currency" required maxlength="3" :disabled="busy || !!editing" /></label>
      <label class="grid gap-2 text-sm">{{ t('wealth.assetClass') }}<Input v-model="form.asset_class" maxlength="40" :disabled="busy" /></label>
      <label class="grid gap-2 text-sm">{{ t('wealth.sector') }}<Input v-model="form.sector" maxlength="100" :disabled="busy" /></label>
      <label class="grid gap-2 text-sm">{{ t('wealth.country') }}<Input v-model="form.country" maxlength="100" :disabled="busy" /></label>
      <div class="flex items-end justify-end gap-2"><Button v-if="editing" type="button" variant="ghost" :disabled="busy" @click="reset">{{ t('common.cancel') }}</Button><Button type="submit" :disabled="busy">{{ t(editing ? 'common.save' : 'wealth.createAsset') }}</Button></div>
    </form></CardContent></Card>
    <Card v-for="asset in assets" :key="asset.id"><CardContent class="space-y-4 pt-6">
      <div class="flex flex-wrap items-center gap-3"><div class="flex-1"><h3 class="font-semibold">{{ asset.name }}</h3><p class="text-xs text-muted-foreground">{{ asset.currency }} · {{ asset.asset_class }} · {{ asset.symbol }}</p></div><Button variant="ghost" size="sm" :disabled="busy" @click="edit(asset)">{{ t('common.edit') }}</Button><Button variant="ghost" size="sm" :disabled="busy" @click="deleting = asset">{{ t('common.delete') }}</Button></div>
      <form class="flex flex-wrap items-end gap-3" @submit.prevent="valuation(asset)">
        <label class="grid gap-1 text-sm">{{ t('portfolio.date') }}<Input v-model="entries[asset.id].date" type="date" :max="today" required :disabled="busy" /></label>
        <label class="grid flex-1 gap-1 text-sm">{{ t('wealth.unitValuation') }}<Input v-model.number="entries[asset.id].price" type="number" min="0" step="0.00000001" required :disabled="busy" /></label><Button type="submit" variant="outline" :disabled="busy">{{ t('wealth.saveValuation') }}</Button>
      </form>
      <details v-if="asset.valuations.length"><summary class="cursor-pointer text-sm text-muted-foreground">{{ t('wealth.history') }}</summary><div class="mt-3 max-h-64 overflow-auto"><div v-for="value in asset.valuations" :key="value.id" class="flex justify-between gap-3 border-t py-2 text-sm"><span>{{ value.date }}</span><span>{{ money(value.price, asset.currency) }}</span><Button variant="ghost" size="sm" @click="entries[asset.id] = { date: value.date, price: value.price }">{{ t('common.edit') }}</Button></div></div></details>
      <Button as-child variant="link" class="px-0"><router-link to="/accounts">{{ t('wealth.addAssetToAccount') }}</router-link></Button>
    </CardContent></Card>
    <InlineConfirmation :visible="!!deleting" :title="t('common.delete')" :message="t('wealth.deleteUnused')" :confirm-text="t('common.delete')" :cancel-text="t('common.cancel')" :loading="busy" @cancel="deleting = null" @confirm="remove" />
  </div>
</template>
<script setup>
import { onMounted, ref } from 'vue'
import { Button, Input, Card, CardContent, InlineConfirmation } from '@/components/ui'
import PageFeedback from '@/components/wealth/PageFeedback.vue'
import { useWealthPage } from '@/composables/useWealthPage'
import { wealthAPI } from '@/services/wealthApi'
const { t, busy, error, notice, run, money, today } = useWealthPage()
const assets = ref([]), entries = ref({}), editing = ref(null), deleting = ref(null)
const empty = () => ({ name: '', currency: 'USD', asset_class: 'Other', sector: '', country: '' })
const form = ref(empty())
function reset() { editing.value = null; form.value = empty() }
function edit(a) { editing.value = a.id; form.value = { name: a.name, currency: a.currency, asset_class: a.asset_class, sector: a.sector, country: a.country } }
async function load() { assets.value = (await wealthAPI.get('/assets/manual')).items; for (const a of assets.value) entries.value[a.id] ||= { date: today, price: '' } }
function save() { run(async () => { await wealthAPI.save(`/assets/manual${editing.value ? `/${editing.value}` : ''}`, { ...form.value, currency: form.value.currency.toUpperCase() }, editing.value ? 'PUT' : 'POST'); reset(); await load() }, true) }
function valuation(a) { run(async () => { await wealthAPI.save(`/assets/manual/${a.id}/valuations`, entries.value[a.id], 'PUT'); await load() }, true) }
function remove() { run(async () => { await wealthAPI.remove(`/assets/manual/${deleting.value.id}`); deleting.value = null; await load() }) }
onMounted(() => run(load))
</script>
