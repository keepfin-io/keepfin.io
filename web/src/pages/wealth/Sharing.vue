<template>
  <div class="space-y-6" data-testid="sharing-page">
    <SectionHeader :icon="LuShare2" :title="t('wealth.sharing')" :description="t('wealth.shareHint')" />
    <PageFeedback :error="error" :busy="busy" :notice="notice" />

    <ul class="grid gap-3 sm:grid-cols-3">
      <li v-for="point in POINTS" :key="point.key" class="flex gap-3 rounded-xl border bg-muted/30 p-3">
        <span class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary"><component :is="point.icon" class="size-4" aria-hidden="true" /></span>
        <div class="min-w-0"><p class="text-sm font-medium">{{ t(`wealth.sharePoints.${point.key}.title`) }}</p><p class="mt-0.5 text-xs leading-relaxed text-muted-foreground">{{ t(`wealth.sharePoints.${point.key}.text`) }}</p></div>
      </li>
    </ul>

    <!-- A new link's key is shown once, so it stays on top until dismissed. -->
    <div v-if="link" data-testid="share-link" class="space-y-3 rounded-2xl border border-primary/40 bg-primary/5 p-5">
      <p class="flex items-center gap-2 font-medium"><LuLink class="size-4 text-primary" aria-hidden="true" />{{ t('wealth.shareLink') }}</p>
      <p class="text-sm text-muted-foreground">{{ t('wealth.copyLinkHint') }}</p>
      <div class="flex flex-col gap-2 sm:flex-row">
        <label class="flex-1"><span class="sr-only">{{ t('wealth.shareLink') }}</span><Input :model-value="link" readonly class="bg-background font-mono text-xs" @focus="$event.target.select()" /></label>
        <Button @click="copy"><component :is="copied ? LuCheck : LuCopy" class="size-4" aria-hidden="true" />{{ t(copied ? 'wealth.copied' : 'wealth.copyLink') }}</Button>
      </div>
    </div>

    <form data-testid="share-form" class="overflow-hidden rounded-2xl border bg-card shadow-xs" @submit.prevent="create">
      <div class="border-b px-5 py-4"><h3 class="flex items-center gap-2 font-semibold"><LuPlus class="size-4 text-primary" aria-hidden="true" />{{ t('wealth.newShare') }}</h3></div>
      <div class="grid gap-4 p-5 sm:grid-cols-2">
        <div class="space-y-2"><Label for="share-name"><LuTag class="size-4 text-muted-foreground" aria-hidden="true" />{{ t('wealth.name') }}</Label><Input id="share-name" v-model="form.name" required maxlength="100" :placeholder="t('wealth.shareNamePlaceholder')" :disabled="busy" /></div>
        <div class="space-y-2"><Label for="share-expires"><LuCalendarClock class="size-4 text-muted-foreground" aria-hidden="true" />{{ t('wealth.expires') }}</Label><Input id="share-expires" v-model="form.expires" type="date" :min="tomorrow" :max="maxExpiry" required :disabled="busy" /></div>
        <div class="space-y-2 sm:col-span-2">
          <Label for="share-recipient"><LuMail class="size-4 text-muted-foreground" aria-hidden="true" />{{ t('wealth.recipient') }}</Label>
          <Input id="share-recipient" v-model="form.recipient_email" type="email" :placeholder="t('wealth.publicShare')" :disabled="busy" />
          <p class="flex items-center gap-1.5 text-xs text-muted-foreground"><component :is="form.recipient_email ? LuLock : LuGlobe" class="size-3.5" aria-hidden="true" />{{ t(form.recipient_email ? 'wealth.privateExplained' : 'wealth.publicExplained') }}</p>
        </div>
        <fieldset class="space-y-2 sm:col-span-2">
          <legend class="mb-2 flex items-center gap-2 text-sm font-medium"><LuWallet class="size-4 text-muted-foreground" aria-hidden="true" />{{ t('wealth.sharedAccounts') }}</legend>
          <div class="flex flex-wrap gap-2">
            <label v-for="account in accounts" :key="account.id" class="inline-flex cursor-pointer items-center gap-2 rounded-full border px-3 py-1.5 text-sm transition-colors has-[:checked]:border-primary/50 has-[:checked]:bg-primary/10 has-[:checked]:text-primary">
              <input v-model="form.account_ids" type="checkbox" class="size-3.5 accent-primary" :value="account.id" :disabled="busy">{{ account.account_name }}
            </label>
          </div>
          <p v-if="!accounts.length" class="text-sm text-muted-foreground">{{ t('accounts.noAccountsYet') }}</p>
        </fieldset>
      </div>
      <div class="flex flex-wrap items-center justify-between gap-3 border-t bg-muted/30 px-5 py-4">
        <p class="text-xs text-muted-foreground">{{ t('wealth.selectedAccounts', { count: form.account_ids.length }) }}</p>
        <Button type="submit" class="w-full sm:w-auto" :disabled="busy || !form.account_ids.length"><LuLink class="size-4" aria-hidden="true" />{{ t('wealth.createShare') }}</Button>
      </div>
    </form>

    <section class="space-y-3" :aria-label="t('wealth.shares')">
      <h3 class="font-semibold">{{ t('wealth.shares') }}</h3>
      <div data-testid="shares-list" class="overflow-hidden rounded-2xl border bg-card shadow-xs">
        <ul v-if="shares.length" role="list" class="divide-y">
          <li v-for="share in shares" :key="share.id" :data-testid="`share-${share.id}`" class="flex flex-wrap items-center gap-x-4 gap-y-2 px-5 py-4 transition-colors hover:bg-muted/30" :class="status(share) !== 'active' && 'text-muted-foreground'">
            <span class="flex size-10 shrink-0 items-center justify-center rounded-xl border bg-background" aria-hidden="true"><component :is="share.recipient_id ? LuLock : LuGlobe" class="size-5" :class="status(share) === 'active' ? 'text-primary' : 'text-muted-foreground'" /></span>
            <div class="min-w-0 flex-1">
              <p class="truncate font-semibold" :class="status(share) === 'active' && 'text-foreground'">{{ share.name }}</p>
              <p class="text-sm text-muted-foreground">{{ share.recipient_id ? t('wealth.private') : t('wealth.public') }} · {{ t('wealth.accountCount', { count: share.account_ids.length }) }} · {{ t('wealth.expiresOn', { date: formatDate(share.expires_at) }) }}</p>
            </div>
            <StatusBadge :tone="STATUS_TONES[status(share)]">{{ t(`wealth.${status(share)}`) }}</StatusBadge>
            <Button v-if="status(share) === 'active'" size="sm" variant="outline" class="text-destructive hover:bg-destructive/10 hover:text-destructive" :disabled="busy" @click="revoking = share"><LuBan class="size-3.5" aria-hidden="true" />{{ t('wealth.revoke') }}</Button>
          </li>
        </ul>
        <EmptyState v-else :icon="LuShare2" :title="t('wealth.emptySharesTitle')" :description="t('wealth.emptyShares')" />
      </div>
    </section>

    <ConfirmModal data-testid="share-revoke-confirmation" :visible="!!revoking" variant="danger" :title="t('wealth.revokeTitle')" :message="revoking ? t('wealth.revokeMessage', { name: revoking.name }) : ''" :confirm-text="t('wealth.revoke')" :cancel-text="t('common.cancel')" :loading="busy" @cancel="revoking = null" @confirm="revoke" />
  </div>
</template>
<script setup>
import { onMounted, ref } from 'vue'
import { LuBan, LuCalendarClock, LuCheck, LuCopy, LuEye, LuGlobe, LuLink, LuLock, LuMail, LuPlus, LuShare2, LuTag, LuWallet } from '@kalimahapps/vue-icons/lu'
import { Button, ConfirmModal, Input, Label } from '@/components/ui'
import PageFeedback from '@/components/wealth/PageFeedback.vue'
import SectionHeader from '@/components/wealth/SectionHeader.vue'
import EmptyState from '@/components/wealth/EmptyState.vue'
import StatusBadge from '@/components/wealth/StatusBadge.vue'
import { wealthAPI } from '@/services/wealthApi'
import { accountAPI } from '@/services/accountApi'
import { useWealthPage } from '@/composables/useWealthPage'
const POINTS = [{ key: 'readonly', icon: LuEye }, { key: 'expiry', icon: LuCalendarClock }, { key: 'access', icon: LuLock }]
const STATUS_TONES = { active: 'success', expired: 'neutral', revoked: 'danger' }
const { t, locale, busy, error, notice, run } = useWealthPage()
const accounts = ref([]), shares = ref([]), link = ref(''), copied = ref(false), revoking = ref(null)
const tomorrow = new Date(Date.now()+86400000).toISOString().slice(0,10), maxExpiry = new Date(Date.now()+364*86400000).toISOString().slice(0,10)
const form = ref({ name:'', expires: new Date(Date.now()+30*86400000).toISOString().slice(0,10), recipient_email:'', account_ids:[] })
const status = share => (share.revoked_at ? 'revoked' : new Date(share.expires_at) < new Date() ? 'expired' : 'active')
const formatDate = value => new Intl.DateTimeFormat(locale.value, { dateStyle: 'medium' }).format(new Date(value))
async function load() { shares.value = (await wealthAPI.get('/shares')).data }
function create() { run(async () => { const data = await wealthAPI.save('/shares',{ name:form.value.name, account_ids:form.value.account_ids, recipient_email:form.value.recipient_email, expires_at:`${form.value.expires}T23:59:59Z` }); link.value = `${window.location.origin}/shared#${data.data.token}`; copied.value = false; form.value.name = ''; await load() },true) }
function revoke() { run(async () => { await wealthAPI.remove(`/shares/${revoking.value.id}`); revoking.value = null; link.value=''; await load() },true) }
function copy() { run(async () => { await navigator.clipboard.writeText(link.value); copied.value = true }) }
onMounted(() => run(async () => { accounts.value=(await accountAPI.getAccounts()).data; await load() }))
</script>
