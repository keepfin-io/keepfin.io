<template>
  <div data-testid="holding-edit-page" class="space-y-6" :aria-busy="submitting">
      <Alert
        data-testid="holding-edit-error"
        v-if="formError"
        variant="error"
        :description="formError"
      />
      <form
        data-testid="holding-edit-form"
        id="edit-holding-form"
        class="space-y-4"
        @submit.prevent="submitUpdate"
      >
        <div class="rounded-lg border p-3">
          <p class="font-medium">{{ holding?.symbol }}</p>
          <p class="break-words text-sm text-muted-foreground">
            {{ holding?.name }}
          </p>
        </div>
        <div class="grid gap-4 sm:grid-cols-2">
          <div class="space-y-2">
            <Label for="edit-quantity">
              <LuHash class="size-4 text-muted-foreground" aria-hidden="true" />
              {{ $t('accountDetail.quantity') }}
            </Label>
            <Input
              id="edit-quantity"
              data-testid="holding-edit-quantity"
              type="number"
              v-model.number="form.quantity"
              step="any"
              min="0.00000001"
              :disabled="submitting || !holding?.can_edit_position"
              required
            />
          </div>
          <div class="space-y-2">
            <Label for="edit-buy_price"><LuCircleDollarSign class="size-4 text-muted-foreground" aria-hidden="true" />{{ $t('accountDetail.price') }}</Label>
            <Input
              id="edit-buy_price"
              data-testid="holding-edit-price"
              type="number"
              v-model.number="form.buy_price"
              step="any"
              min="0"
              :disabled="submitting || !holding?.can_edit_position"
              required
            />
          </div>
          <div class="space-y-2">
            <Label for="edit-buyCurrency">
              <LuBanknote class="size-4 text-muted-foreground" aria-hidden="true" />
              {{ $t('accountDetail.currency') }}
            </Label>
            <FormSelect
              id="edit-buyCurrency"
              data-testid="holding-edit-currency"
              v-model="form.buy_currency"
              :disabled="submitting || !holding?.can_edit_position || !!lockedCurrency"
              :options="lockedCurrency ? [lockedCurrency] : [
                ...new Set([form.buy_currency, 'USD', 'EUR', 'GBP', 'TRY']),
              ]"
            />
            <p v-if="currencyCorrected" data-testid="holding-edit-currency-corrected" class="text-xs text-amber-600 dark:text-amber-400">
              {{ $t('holdings.currencyCorrected', { symbol: holding.symbol, from: holding.buy_currency, currency: lockedCurrency }) }}
            </p>
            <p v-else-if="lockedCurrency" data-testid="holding-edit-currency-locked" class="text-xs text-muted-foreground">
              {{ $t('holdings.currencyLocked', { symbol: holding.symbol, currency: lockedCurrency }) }}
            </p>
          </div>
        </div>
        <p class="text-xs text-muted-foreground">{{ $t('portfolio.classificationHint') }}</p>
        <div class="grid gap-3 sm:grid-cols-3"><div v-for="field in ['asset_class', 'sector', 'country']" :key="field"><Label :for="`holding-${field}`"><component :is="{ asset_class: 'LuLayers', sector: 'LuFactory', country: 'LuGlobe' }[field]" class="size-4 text-muted-foreground" aria-hidden="true" />{{ $t(`portfolio.${field}`) }}</Label><Input :id="`holding-${field}`" v-model="form[field]" maxlength="100" :disabled="submitting" /></div></div>
        <div class="space-y-2">
          <Label for="edit-notes"><LuNotebookPen class="size-4 text-muted-foreground" aria-hidden="true" />{{ $t('holdings.notes') }}</Label>
          <Textarea
            id="edit-notes"
            data-testid="holding-edit-notes"
            v-model="form.notes"
            :disabled="submitting"
          />
        </div>
        <div
          class="flex flex-wrap justify-between gap-2 rounded-lg bg-muted p-3 text-sm"
        >
          <span class="flex items-center gap-2"><LuCalculator class="size-4 text-muted-foreground" aria-hidden="true" />{{ $t('holdings.totalCost') }}</span>
          <strong
            data-testid="holding-edit-total"
            class="break-all tabular-nums"
          >
            {{
              formatCurrency(form.quantity * form.buy_price, form.buy_currency)
            }}
          </strong>
        </div>
      </form>
      <div class="flex flex-col-reverse gap-3 border-t pt-5 sm:flex-row sm:justify-end">
        <Button
          type="button"
          variant="outline"
          :disabled="submitting"
          data-testid="holding-edit-cancel"
          @click="close"
        >
          {{ $t('common.cancel') }}
        </Button>
        <Button
          data-testid="holding-edit-submit"
          type="submit"
          form="edit-holding-form"
          :disabled="!canSubmit"
          :loading="submitting"
          :loading-text="$t('common.saving')"
        >
          {{ $t('holdings.editHolding') }}
        </Button>
      </div>
  </div>
</template>
<script>
import {
  FeedbackAlert as Alert,
  Button,
  FormSelect,
  Input,
  Label,
  Textarea,
} from '@/components/ui'
import { holdingAPI } from '@/services/holdingApi'
import { LuBanknote, LuCalculator, LuCircleDollarSign, LuFactory, LuGlobe, LuHash, LuLayers, LuNotebookPen } from '@kalimahapps/vue-icons/lu'
import { tradeCurrency } from '@/lib/portfolio-format'

export default {
  components: {
    Alert,
    Button,
    FormSelect,
    Input,
    Label,
    Textarea,
    LuBanknote, LuCalculator, LuCircleDollarSign, LuFactory, LuGlobe, LuHash, LuLayers, LuNotebookPen,
  },
  name: 'EditHoldingForm',
  props: {
    accountId: {
      type: String,
      required: true,
    },
    holding: {
      type: Object,
      default: null,
    },
  },
  emits: ['close', 'success', 'busy'],
  data() {
    return {
      active: true,
      submitting: false,
      formError: null,
      form: {
        quantity: null,
        buy_price: null,
        buy_currency: 'USD',
        notes: '',
      },
    }
  },
  computed: {
    // Only single-trade positions can change currency; others are fixed by their activities.
    lockedCurrency() {
      return this.holding?.can_edit_position ? tradeCurrency(this.holding.quote_currency) : ''
    },
    // Legacy positions may record a listing's price in another currency.
    currencyCorrected() {
      return !!this.lockedCurrency && this.lockedCurrency !== this.holding.buy_currency
    },
    canSubmit() {
      return (
        this.form.quantity > 0 &&
        this.form.buy_price !== null &&
        this.form.buy_price !== '' &&
        this.form.buy_price >= 0
      )
    },
  },
  watch: {
    holding: {
      immediate: true,
      handler(value) { if (value) this.populateForm() },
    },
    submitting(value) { this.$emit('busy', value) },
  },
  beforeUnmount() { this.active = false },
  methods: {
    populateForm() {
      this.formError = null
      this.form = {
        quantity: this.holding.quantity,
        buy_price: this.holding.can_edit_position ? this.holding.entry_price : this.holding.buy_price,
        buy_currency: this.lockedCurrency || this.holding.buy_currency || 'USD',
        notes: this.holding.notes || '',
        asset_class: this.holding.asset_class || '', sector: this.holding.sector || '', country: this.holding.country || '',
      }
    },

    async submitUpdate() {
      if (!this.canSubmit || this.submitting) return

      this.submitting = true
      this.formError = null
      try {
        await holdingAPI.updateHolding(this.accountId, this.holding.id, {
          quantity: this.holding.can_edit_position ? this.form.quantity : undefined,
          buy_price: this.holding.can_edit_position ? this.form.buy_price : undefined,
          buy_currency: this.holding.can_edit_position ? this.form.buy_currency : undefined,
          notes: this.form.notes,
          asset_class: this.form.asset_class, sector: this.form.sector, country: this.form.country,
        })
        if (this.active) this.$emit('success')
      } catch (err) {
        console.error('Failed to update holding:', err)
        this.formError = err.message || 'Failed to update holding'
        this.submitting = false
      }
    },

    close() {
      if (this.submitting) return
      this.$emit('close')
    },

    formatCurrency(amount, currency) {
      return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: currency || 'USD',
      }).format(amount || 0)
    },
  },
}
</script>
