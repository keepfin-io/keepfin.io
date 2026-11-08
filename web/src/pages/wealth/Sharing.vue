<template>
  <div class="space-y-5" data-testid="sharing-page"><h2 class="text-xl font-semibold">{{ t('wealth.sharing') }}</h2><p class="text-sm text-muted-foreground">{{ t('wealth.shareHint') }}</p><PageFeedback :error="error" :busy="busy" :notice="notice" />
    <Card><CardContent class="pt-6"><form class="grid gap-4 sm:grid-cols-2" @submit.prevent="create"><label class="grid gap-2 text-sm">{{ t('wealth.name') }}<Input v-model="form.name" required maxlength="100" :disabled="busy" /></label><label class="grid gap-2 text-sm">{{ t('wealth.expires') }}<Input v-model="form.expires" type="date" :min="tomorrow" :max="maxExpiry" required :disabled="busy" /></label><label class="grid gap-2 text-sm sm:col-span-2">{{ t('wealth.recipient') }}<Input v-model="form.recipient_email" type="email" :placeholder="t('wealth.publicShare')" :disabled="busy" /></label><fieldset class="space-y-2 sm:col-span-2"><legend class="mb-2 text-sm">{{ t('wealth.sharedAccounts') }}</legend><label v-for="account in accounts" :key="account.id" class="mr-4 inline-flex items-center gap-2 text-sm"><input v-model="form.account_ids" type="checkbox" :value="account.id" :disabled="busy">{{ account.account_name }}</label></fieldset><Button type="submit" :disabled="busy || !form.account_ids.length">{{ t('wealth.createShare') }}</Button></form></CardContent></Card>
    <Card v-if="link"><CardContent class="space-y-3 pt-6"><p class="text-sm">{{ t('wealth.copyLinkHint') }}</p><label class="grid gap-2 text-sm">{{ t('wealth.shareLink') }}<Input :model-value="link" readonly @focus="$event.target.select()" /></label><Button variant="outline" @click="copy">{{ t('wealth.copyLink') }}</Button></CardContent></Card>
    <div class="divide-y rounded-xl border"><div v-for="share in shares" :key="share.id" class="flex flex-wrap items-center gap-3 p-4"><div class="flex-1"><p class="font-medium">{{ share.name }}</p><p class="text-xs text-muted-foreground">{{ share.recipient_id ? t('wealth.private') : t('wealth.public') }} · {{ share.expires_at.slice(0,10) }} · {{ share.account_ids.length }} {{ t('nav.accounts') }}</p></div><span class="text-sm text-muted-foreground">{{ share.revoked_at ? t('wealth.revoked') : new Date(share.expires_at) < new Date() ? t('wealth.expired') : t('wealth.active') }}</span><Button v-if="!share.revoked_at" size="sm" variant="outline" :disabled="busy" @click="revoke(share)">{{ t('wealth.revoke') }}</Button></div></div>
  </div>
</template>
<script setup>
import { onMounted, ref } from 'vue'
import { Button, Input, Card, CardContent } from '@/components/ui'
import PageFeedback from '@/components/wealth/PageFeedback.vue'
import { wealthAPI } from '@/services/wealthApi'
import { accountAPI } from '@/services/accountApi'
import { useWealthPage } from '@/composables/useWealthPage'
const { t, busy, error, notice, run } = useWealthPage()
const accounts = ref([]), shares = ref([]), link = ref('')
const tomorrow = new Date(Date.now()+86400000).toISOString().slice(0,10), maxExpiry = new Date(Date.now()+364*86400000).toISOString().slice(0,10)
const form = ref({ name:'', expires: new Date(Date.now()+30*86400000).toISOString().slice(0,10), recipient_email:'', account_ids:[] })
async function load() { shares.value = (await wealthAPI.get('/shares')).data }
function create() { run(async () => { const data = await wealthAPI.save('/shares',{ name:form.value.name, account_ids:form.value.account_ids, recipient_email:form.value.recipient_email, expires_at:`${form.value.expires}T23:59:59Z` }); link.value = `${window.location.origin}/shared#${data.data.token}`; await load() },true) }
function revoke(item) { run(async () => { await wealthAPI.remove(`/shares/${item.id}`); link.value=''; await load() },true) }
function copy() { run(async () => { await navigator.clipboard.writeText(link.value) },true) }
onMounted(() => run(async () => { accounts.value=(await accountAPI.getAccounts()).data; await load() }))
</script>
