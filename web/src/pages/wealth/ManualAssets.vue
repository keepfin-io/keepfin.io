<template>
  <div class="space-y-6" data-testid="manual-assets-page">
    <SectionHeader :icon="LuGem" :title="t('wealth.manualAssets')" :description="t('wealth.manualHint')">
      <Button data-testid="manual-asset-new" class="w-full sm:w-auto" :disabled="busy" @click="openCreate"><LuPlus class="size-4" aria-hidden="true" />{{ t('wealth.createAsset') }}</Button>
    </SectionHeader>
    <PageFeedback :error="error" :busy="busy" :notice="notice" />

    <!-- How a manual asset becomes part of a portfolio. -->
    <ol data-testid="manual-assets-steps" class="grid gap-3 sm:grid-cols-3">
      <li v-for="(step, index) in STEPS" :key="step.key" class="flex gap-3 rounded-xl border bg-muted/30 p-3">
        <span class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary"><component :is="step.icon" class="size-4" aria-hidden="true" /></span>
        <div class="min-w-0">
          <p class="text-sm font-medium"><span class="text-muted-foreground">{{ index + 1 }}.</span> {{ t(`wealth.manualSteps.${step.key}.title`) }}</p>
          <p class="mt-0.5 text-xs leading-relaxed text-muted-foreground">{{ t(`wealth.manualSteps.${step.key}.text`) }}</p>
          <router-link v-if="step.key === 'position'" to="/accounts" class="mt-1 inline-flex items-center gap-1 text-xs font-medium text-primary hover:underline">{{ t('wealth.addAssetToAccount') }}<LuArrowRight class="size-3" aria-hidden="true" /></router-link>
        </div>
      </li>
    </ol>

    <div data-testid="manual-assets-list" class="overflow-hidden rounded-2xl border bg-card shadow-xs">
      <template v-if="assets.length">
        <div class="hidden border-b bg-muted/40 px-5 py-2.5 text-xs font-medium uppercase tracking-wide text-muted-foreground lg:grid lg:grid-cols-[minmax(0,1fr)_11rem_auto] lg:gap-6" aria-hidden="true">
          <span>{{ t('wealth.asset') }}</span><span class="text-right">{{ t('wealth.latestValuation') }}</span><span class="w-[19rem] text-right">{{ t('common.actions') }}</span>
        </div>
        <ul role="list" class="divide-y">
          <li v-for="asset in assets" :key="asset.id" :data-testid="`manual-asset-${asset.id}`" class="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-x-4 gap-y-3 px-5 py-4 transition-colors hover:bg-muted/30 lg:grid-cols-[minmax(0,1fr)_11rem_auto] lg:gap-6">
            <div class="flex min-w-0 items-center gap-3">
              <span class="flex size-10 shrink-0 items-center justify-center rounded-xl border bg-background text-primary" aria-hidden="true"><LuGem class="size-5" /></span>
              <div class="min-w-0">
                <p class="truncate font-semibold">{{ asset.name }}</p>
                <p class="truncate text-sm text-muted-foreground">{{ [asset.asset_class, asset.sector, asset.country].filter(Boolean).join(' · ') }}<span class="ml-1 rounded bg-muted px-1.5 py-0.5 font-mono text-[11px]">{{ asset.currency }}</span></p>
              </div>
            </div>
            <div class="text-right">
              <template v-if="asset.valuations.length">
                <p class="font-semibold tabular-nums">{{ money(asset.valuations[0].price, asset.currency) }}</p>
                <p class="text-xs text-muted-foreground">{{ formatDate(asset.valuations[0].date) }}</p>
              </template>
              <StatusBadge v-else tone="warning">{{ t('wealth.noValuation') }}</StatusBadge>
            </div>
            <div class="col-span-2 flex flex-wrap justify-end gap-2 lg:col-span-1">
              <Button :data-testid="`manual-asset-${asset.id}-value`" variant="outline" size="sm" class="h-8" :disabled="busy" @click="openValuation(asset)"><LuCalendarClock class="size-3.5" aria-hidden="true" />{{ t('wealth.enterValuation') }}</Button>
              <RecordActions :data-testid="`manual-asset-${asset.id}-actions`" :disabled="busy" @edit="openEdit(asset)" @delete="deleting = asset" />
            </div>
          </li>
        </ul>
      </template>
      <EmptyState v-else-if="loaded" :icon="LuGem" :title="t('wealth.emptyManualTitle')" :description="t('holdings.manualEmpty')">
        <Button variant="outline" size="sm" @click="openCreate"><LuPlus class="size-4" aria-hidden="true" />{{ t('wealth.createAsset') }}</Button>
      </EmptyState>
    </div>

    <FormDialog data-testid="manual-asset-dialog" :open="formOpen" :busy="busy" :error="error" :icon="LuGem" :title="t(editing ? 'wealth.editAsset' : 'wealth.createAsset')" :description="t('wealth.assetFormHint')" @close="formOpen = false">
      <form id="manual-asset-form" class="grid gap-4 sm:grid-cols-2" @submit.prevent="save">
        <div class="space-y-2 sm:col-span-2"><Label for="manual-name"><LuTag class="size-4 text-muted-foreground" aria-hidden="true" />{{ t('wealth.name') }}</Label><Input id="manual-name" v-model="form.name" required maxlength="100" :placeholder="t('wealth.manualNamePlaceholder')" :disabled="busy" /></div>
        <div class="space-y-2">
          <Label for="manual-currency"><LuBanknote class="size-4 text-muted-foreground" aria-hidden="true" />{{ t('portfolio.currency') }}</Label>
          <Input id="manual-currency" v-model="form.currency" required maxlength="3" class="uppercase" placeholder="USD" :disabled="busy || !!editing" />
          <p v-if="editing" class="text-xs text-muted-foreground">{{ t('wealth.currencyFixed') }}</p>
        </div>
        <div class="space-y-2"><Label for="manual-class"><LuLayers class="size-4 text-muted-foreground" aria-hidden="true" />{{ t('wealth.assetClass') }}</Label><Input id="manual-class" v-model="form.asset_class" maxlength="40" :placeholder="t('wealth.manualClassPlaceholder')" :disabled="busy" /></div>
        <div class="space-y-2"><Label for="manual-sector"><LuFactory class="size-4 text-muted-foreground" aria-hidden="true" />{{ t('wealth.sector') }}</Label><Input id="manual-sector" v-model="form.sector" maxlength="100" :placeholder="t('holdings.classificationPlaceholders.sector')" :disabled="busy" /></div>
        <div class="space-y-2"><Label for="manual-country"><LuGlobe class="size-4 text-muted-foreground" aria-hidden="true" />{{ t('wealth.country') }}</Label><Input id="manual-country" v-model="form.country" maxlength="100" :placeholder="t('holdings.classificationPlaceholders.country')" :disabled="busy" /></div>
      </form>
      <template #footer>
        <Button variant="outline" :disabled="busy" @click="formOpen = false">{{ t('common.cancel') }}</Button>
        <Button data-testid="manual-asset-save" type="submit" form="manual-asset-form" :loading="busy" :loading-text="t('common.saving')">{{ t(editing ? 'common.save' : 'wealth.createAsset') }}</Button>
      </template>
    </FormDialog>

    <FormDialog data-testid="manual-valuation-dialog" :open="!!valuing" :busy="busy" :error="error" :icon="LuCalendarClock" :title="valuing ? t('wealth.valuationTitle', { name: valuing.name }) : ''" :description="t('wealth.valuationHint')" @close="valuing = null">
      <form v-if="valuing" id="manual-valuation-form" class="grid gap-4 sm:grid-cols-2" @submit.prevent="saveValuation">
        <div class="space-y-2"><Label for="valuation-date"><LuCalendar class="size-4 text-muted-foreground" aria-hidden="true" />{{ t('portfolio.date') }}</Label><Input id="valuation-date" v-model="entry.date" type="date" :max="today" required :disabled="busy" /></div>
        <div class="space-y-2"><Label for="valuation-price"><LuCircleDollarSign class="size-4 text-muted-foreground" aria-hidden="true" />{{ t('wealth.unitValuation') }} ({{ valuing.currency }})</Label><Input id="valuation-price" v-model.number="entry.price" type="number" min="0" step="0.00000001" required placeholder="0.00" :disabled="busy" /></div>
      </form>
      <div v-if="valuing?.valuations.length" class="rounded-xl border">
        <p class="flex items-center gap-2 border-b bg-muted/40 px-3 py-2 text-xs font-medium text-muted-foreground"><LuHistory class="size-3.5" aria-hidden="true" />{{ t('wealth.history') }}</p>
        <ul class="max-h-56 divide-y overflow-y-auto">
          <li v-for="value in valuing.valuations" :key="value.id" class="flex items-center gap-3 px-3 py-1.5 text-sm">
            <span class="flex-1 text-muted-foreground">{{ formatDate(value.date) }}</span>
            <span class="font-medium tabular-nums">{{ money(value.price, valuing.currency) }}</span>
            <Button variant="ghost" size="icon" class="size-7" :aria-label="t('common.edit')" :title="t('common.edit')" @click="entry = { date: value.date, price: value.price }"><LuPencil class="size-3.5" aria-hidden="true" /></Button>
          </li>
        </ul>
      </div>
      <template #footer>
        <Button variant="outline" :disabled="busy" @click="valuing = null">{{ t('common.cancel') }}</Button>
        <Button data-testid="manual-valuation-save" type="submit" form="manual-valuation-form" :loading="busy" :loading-text="t('common.saving')">{{ t('wealth.saveValuation') }}</Button>
      </template>
    </FormDialog>

    <ConfirmModal data-testid="manual-asset-delete-confirmation" :visible="!!deleting" variant="danger" :title="t('wealth.deleteAssetTitle')" :message="deleting ? `${t('wealth.deleteAssetMessage', { name: deleting.name })} ${t('wealth.deleteUnused')}` : ''" :confirm-text="t('common.delete')" :cancel-text="t('common.cancel')" :loading-text="t('common.deleting')" :loading="busy" @cancel="deleting = null" @confirm="remove" />
  </div>
</template>
<script setup>
import { onMounted, ref } from 'vue'
import { LuArrowRight, LuBanknote, LuCalendar, LuCalendarClock, LuCircleDollarSign, LuFactory, LuGem, LuGlobe, LuHistory, LuLayers, LuPencil, LuPlus, LuTag, LuWallet } from '@kalimahapps/vue-icons/lu'
import { Button, ConfirmModal, Input, Label, RecordActions } from '@/components/ui'
import PageFeedback from '@/components/wealth/PageFeedback.vue'
import SectionHeader from '@/components/wealth/SectionHeader.vue'
import EmptyState from '@/components/wealth/EmptyState.vue'
import StatusBadge from '@/components/wealth/StatusBadge.vue'
import FormDialog from '@/components/wealth/FormDialog.vue'
import { useWealthPage } from '@/composables/useWealthPage'
import { wealthAPI } from '@/services/wealthApi'
const STEPS = [{ key: 'create', icon: LuGem }, { key: 'value', icon: LuCalendarClock }, { key: 'position', icon: LuWallet }]
const { t, locale, busy, error, notice, run, money, today } = useWealthPage()
const assets = ref([]), loaded = ref(false), editing = ref(null), deleting = ref(null), formOpen = ref(false), valuing = ref(null), entry = ref({ date: today, price: '' })
const empty = () => ({ name: '', currency: 'USD', asset_class: 'Other', sector: '', country: '' })
const form = ref(empty())
const formatDate = value => new Intl.DateTimeFormat(locale.value, { dateStyle: 'medium' }).format(new Date(`${value.slice(0, 10)}T00:00:00`))
function openCreate() { error.value = ''; editing.value = null; form.value = empty(); formOpen.value = true }
function openEdit(a) { error.value = ''; editing.value = a.id; form.value = { name: a.name, currency: a.currency, asset_class: a.asset_class, sector: a.sector, country: a.country }; formOpen.value = true }
function openValuation(a) { error.value = ''; valuing.value = a; entry.value = { date: today, price: '' } }
async function load() { assets.value = (await wealthAPI.get('/assets/manual')).data; loaded.value = true }
function save() { run(async () => { await wealthAPI.save(`/assets/manual${editing.value ? `/${editing.value}` : ''}`, { ...form.value, currency: form.value.currency.toUpperCase() }, editing.value ? 'PUT' : 'POST'); formOpen.value = false; await load() }, true) }
function saveValuation() { run(async () => { await wealthAPI.save(`/assets/manual/${valuing.value.id}/valuations`, entry.value, 'PUT'); valuing.value = null; await load() }, true) }
function remove() { run(async () => { await wealthAPI.remove(`/assets/manual/${deleting.value.id}`); deleting.value = null; await load() }) }
onMounted(() => run(load))
</script>
