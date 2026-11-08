<template>
  <section :data-testid="account ? 'account-edit-page' : 'account-create-page'" class="mx-auto w-full max-w-3xl space-y-6" :aria-busy="saving">
    <div class="space-y-4">
      <Button data-testid="account-create-back" variant="ghost" class="-ml-3 min-h-11 rounded-xl text-muted-foreground" :disabled="saving" @click="cancel">
        <ArrowLeft class="size-4" aria-hidden="true" />{{ $t(account ? 'accountDetail.backToAccount' : 'accountDetail.backToAccounts') }}
      </Button>
      <div>
        <h1 class="text-2xl font-semibold tracking-tight">{{ $t(account ? 'accounts.editAccount' : 'accounts.createAccount') }}</h1>
        <p class="mt-2 text-sm text-muted-foreground">{{ $t(account ? 'accounts.subtitle' : 'accounts.createHint') }}</p>
      </div>
    </div>

    <Card class="rounded-2xl">
      <CardContent class="p-5 sm:p-7">
        <form ref="formElement" data-testid="account-form" class="space-y-6" @submit.prevent="submit">
          <fieldset :disabled="saving || confirming" class="min-w-0 space-y-6">
            <div class="space-y-2">
              <Label for="accountName">{{ $t('accounts.accountName') }} *</Label>
              <Input id="accountName" data-testid="account-form-name" v-model="form.account_name" maxlength="100" :placeholder="$t('accounts.accountNamePlaceholder')" :aria-invalid="!!fieldErrors.account_name" :aria-describedby="fieldErrors.account_name ? 'account-name-error' : undefined" />
              <p v-if="fieldErrors.account_name" id="account-name-error" data-testid="account-form-name-error" class="text-sm text-destructive">{{ fieldErrors.account_name }}</p>
            </div>

            <div class="grid gap-6 sm:grid-cols-2">
              <div class="space-y-2">
                <Label for="currency">{{ $t('accounts.currency') }} *</Label>
                <FormSelect id="currency" data-testid="account-form-currency" v-model="form.currency" :options="CURRENCY_CODES" :placeholder="$t('accounts.selectCurrency')" :invalid="!!fieldErrors.currency" :disabled="saving || confirming" :aria-describedby="fieldErrors.currency ? 'account-currency-error' : undefined" />
                <p v-if="fieldErrors.currency" id="account-currency-error" data-testid="account-form-currency-error" class="text-sm text-destructive">{{ fieldErrors.currency }}</p>
              </div>
              <div class="space-y-2">
                <Label for="initialBalance">{{ $t('accounts.cashBalance') }}</Label>
                <Input id="initialBalance" data-testid="account-form-balance" v-model.number="form.initial_balance" type="number" step="0.01" placeholder="0.00" :aria-invalid="!!fieldErrors.initial_balance" :aria-describedby="fieldErrors.initial_balance ? 'account-balance-error' : undefined" />
                <p v-if="fieldErrors.initial_balance" id="account-balance-error" data-testid="account-form-balance-error" class="text-sm text-destructive">{{ fieldErrors.initial_balance }}</p>
              </div>
              <div class="space-y-2">
                <Label for="platform">{{ $t('accounts.platform') }}</Label>
                <PlatformSelect id="platform" data-testid="account-form-platform" v-model="form.platform" :disabled="saving || confirming" :aria-invalid="!!fieldErrors.platform" :aria-describedby="fieldErrors.platform ? 'account-platform-error' : undefined" />
                <p v-if="fieldErrors.platform" id="account-platform-error" data-testid="account-form-platform-error" class="text-sm text-destructive">{{ fieldErrors.platform }}</p>
              </div>
              <div class="space-y-2">
                <Label for="account-opening-date">{{ $t('portfolio.openingDate') }}</Label>
                <Input id="account-opening-date" data-testid="account-form-opening-date" v-model="form.opening_date" type="date" :max="today" required aria-describedby="account-opening-hint" />
                <p id="account-opening-hint" class="text-xs leading-relaxed text-muted-foreground">{{ $t('portfolio.openingHint') }}</p>
              </div>
            </div>

            <div class="space-y-2">
              <Label for="description">{{ $t('accounts.description') }}</Label>
              <Textarea id="description" data-testid="account-form-description" v-model="form.description" maxlength="255" :placeholder="$t('accounts.descriptionPlaceholder')" :aria-invalid="!!fieldErrors.description" :aria-describedby="fieldErrors.description ? 'account-description-error' : undefined" />
              <p v-if="fieldErrors.description" id="account-description-error" data-testid="account-form-description-error" class="text-sm text-destructive">{{ fieldErrors.description }}</p>
            </div>
            <div class="flex items-start gap-3 rounded-xl border p-4">
              <Checkbox id="excluded" data-testid="account-form-excluded" v-model="form.is_excluded" :disabled="saving || confirming" aria-describedby="account-excluded-hint" />
              <div class="space-y-1.5">
                <Label for="excluded" class="leading-snug">{{ $t('accounts.excludeFromTotals') }}</Label>
                <p id="account-excluded-hint" class="text-xs leading-relaxed text-muted-foreground">{{ $t('accounts.excludeHint') }}</p>
              </div>
            </div>
          </fieldset>

          <Alert v-if="formError" data-testid="account-form-error" variant="error" :description="formError" />
          <InlineConfirmation data-testid="account-update-confirmation" :visible="confirming" :title="$t('accounts.updateConfirmTitle')" :message="`${$t('accounts.updateConfirmMessage')} ${form.account_name}?`" :confirm-text="$t('accounts.updateAccount')" :cancel-text="$t('common.cancel')" :loading="saving" :loading-text="$t('common.saving')" @cancel="confirming = false" @confirm="save" />
          <div v-if="!confirming" class="flex flex-col-reverse gap-3 border-t pt-5 sm:flex-row sm:justify-end">
            <Button data-testid="account-form-cancel" type="button" variant="outline" class="min-h-11 rounded-xl" :disabled="saving" @click="cancel">{{ $t('common.cancel') }}</Button>
            <Button data-testid="account-form-submit" type="submit" class="min-h-11 rounded-xl" :loading="saving" :disabled="saving" :loading-text="$t('common.saving')">{{ $t(account ? 'accounts.updateAccount' : 'accounts.createAccount') }}</Button>
          </div>
        </form>
      </CardContent>
    </Card>
  </section>
</template>

<script setup>
import { nextTick, onBeforeUnmount, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ArrowLeft } from 'lucide-vue-next'
import { Button, Card, CardContent, Checkbox, FeedbackAlert as Alert, FormSelect, InlineConfirmation, Input, Label, PlatformSelect, Textarea } from '@/components/ui'
import { accountAPI } from '@/services/accountApi'
import { CURRENCY_CODES, createAccountSchema, updateAccountSchema } from '@/schemas/account'

const props = defineProps({ account: { type: Object, default: null } })
const emit = defineEmits(['close', 'saved'])
const router = useRouter()
const { t } = useI18n()
const today = new Date().toISOString().slice(0, 10)
const form = ref(props.account ? { ...props.account, opening_date: props.account.opening_date?.slice(0, 10) || today, platform: props.account.platform || '', description: props.account.description || '' } : { account_name: '', currency: '', initial_balance: 0, opening_date: today, platform: '', description: '', is_excluded: false })
const formElement = ref(null)
const fieldErrors = ref({})
const formError = ref(null)
const saving = ref(false)
const confirming = ref(false)
let payload
let mounted = true
onBeforeUnmount(() => { mounted = false })

function cancel() {
  if (saving.value) return
  if (props.account) emit('close')
  else router.replace({ name: 'Accounts' })
}

async function submit() {
  if (saving.value || confirming.value) return
  formError.value = null
  fieldErrors.value = {}
  const result = (props.account ? updateAccountSchema : createAccountSchema).safeParse(form.value)
  if (!result.success) {
    fieldErrors.value = Object.fromEntries(Object.entries(result.error.flatten().fieldErrors).map(([key, messages]) => [key, messages[0]]))
    await nextTick()
    formElement.value?.querySelector('[aria-invalid="true"]')?.focus()
    return
  }
  payload = { ...result.data, opening_date: `${form.value.opening_date}T00:00:00Z` }
  if (props.account) confirming.value = true
  else await save()
}

async function save() {
  if (saving.value || !payload) return
  saving.value = true
  try {
    if (props.account) {
      await accountAPI.updateAccount(props.account.id, payload)
      if (mounted) emit('saved')
    } else {
      const response = await accountAPI.createAccount(payload)
      if (mounted) await router.replace({ name: 'Accounts', query: { created: response.data.id } })
    }
  } catch (error) {
    formError.value = error.message || t('accounts.saveError')
    saving.value = false
    confirming.value = false
  }
}
</script>
