<template>
  <div data-testid="holding-edit-page" class="space-y-6" :aria-busy="submitting">
    <!-- Which holding is being edited, as it appears in the holdings list. -->
    <div v-if="holding" data-testid="holding-edit-asset" class="flex flex-wrap items-center gap-x-4 gap-y-3 rounded-xl border bg-muted/30 p-4">
      <div class="min-w-0 flex-1"><AssetIdentity :asset="asset" /></div>
      <dl class="flex gap-6 text-right text-xs">
        <div>
          <dt class="text-muted-foreground">{{ $t('accountDetail.quantity') }}</dt>
          <dd class="mt-0.5 text-sm font-semibold tabular-nums">{{ formatNumber(holding.quantity) }}</dd>
        </div>
        <div v-if="holding.price_available">
          <dt class="text-muted-foreground">{{ $t('accountDetail.currentPrice') }}</dt>
          <dd data-testid="holding-edit-current-price" class="mt-0.5 text-sm font-semibold tabular-nums">{{ formatCurrency(holding.current_price, holding.buy_currency) }}</dd>
        </div>
      </dl>
    </div>
    <Alert
      data-testid="holding-edit-error"
      v-if="formError"
      variant="error"
      :description="formError"
    />
    <form
      data-testid="holding-edit-form"
      id="edit-holding-form"
      class="space-y-8"
      @submit.prevent="submitUpdate"
    >
      <section class="space-y-4" aria-labelledby="edit-position-title">
        <header class="flex gap-3">
          <span class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary"><LuReceipt class="size-4" aria-hidden="true" /></span>
          <div>
            <h2 id="edit-position-title" class="text-sm font-semibold">{{ $t('holdings.editSections.position') }}</h2>
            <p class="text-xs text-muted-foreground">{{ $t('holdings.editSections.positionHint') }}</p>
          </div>
        </header>
        <!-- Multi-trade positions are the sum of their activities. -->
        <div v-if="holding && !holding.can_edit_position" data-testid="holding-edit-locked" class="flex flex-wrap items-center gap-3 rounded-lg border border-amber-500/30 bg-amber-500/10 p-3 text-sm">
          <LuInfo class="size-4 shrink-0 text-amber-600 dark:text-amber-400" aria-hidden="true" />
          <p class="min-w-0 flex-1">{{ $t('portfolio.classificationHint') }}</p>
          <Button as-child variant="outline" size="sm">
            <router-link data-testid="holding-edit-activities" :to="{ name: 'AccountDetail', params: { id: accountId }, query: { tab: 'activities' } }">
              {{ $t('holdings.goToActivities') }}<LuArrowRight class="size-3.5" aria-hidden="true" />
            </router-link>
          </Button>
        </div>
        <div class="grid gap-4 sm:grid-cols-3">
          <div class="space-y-2">
            <Label for="edit-quantity"><LuHash class="size-4 text-muted-foreground" aria-hidden="true" />{{ $t('accountDetail.quantity') }}</Label>
            <Input
              id="edit-quantity"
              data-testid="holding-edit-quantity"
              type="number"
              v-model.number="form.quantity"
              step="any"
              min="0.00000001"
              :placeholder="$t('holdings.quantityPlaceholder')"
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
              :placeholder="$t('holdings.pricePlaceholder')"
              :disabled="submitting || !holding?.can_edit_position"
              required
            />
          </div>
          <div class="space-y-2">
            <Label for="edit-buyCurrency"><LuBanknote class="size-4 text-muted-foreground" aria-hidden="true" />{{ $t('accountDetail.currency') }}</Label>
            <FormSelect
              id="edit-buyCurrency"
              data-testid="holding-edit-currency"
              v-model="form.buy_currency"
              :disabled="submitting || !holding?.can_edit_position || !!lockedCurrency"
              :options="lockedCurrency ? [lockedCurrency] : [
                ...new Set([form.buy_currency, 'USD', 'EUR', 'GBP', 'TRY']),
              ]"
            />
          </div>
        </div>
        <p v-if="currencyCorrected" data-testid="holding-edit-currency-corrected" class="flex gap-2 rounded-lg border border-amber-500/30 bg-amber-500/10 p-3 text-xs text-amber-700 dark:text-amber-300">
          <LuTriangleAlert class="size-4 shrink-0" aria-hidden="true" />
          {{ $t('holdings.currencyCorrected', { symbol: holding.symbol, from: holding.buy_currency, currency: lockedCurrency }) }}
        </p>
        <p v-else-if="lockedCurrency" data-testid="holding-edit-currency-locked" class="text-xs text-muted-foreground">
          {{ $t('holdings.currencyLocked', { symbol: holding.symbol, currency: lockedCurrency }) }}
        </p>
        <div class="flex flex-wrap justify-between gap-2 rounded-lg bg-muted p-3 text-sm">
          <span class="flex items-center gap-2"><LuCalculator class="size-4 text-muted-foreground" aria-hidden="true" />{{ $t('holdings.totalCost') }}</span>
          <strong data-testid="holding-edit-total" class="break-all tabular-nums">
            {{ formatCurrency(form.quantity * form.buy_price, form.buy_currency) }}
          </strong>
        </div>
      </section>

      <section class="space-y-4 border-t pt-6" aria-labelledby="edit-classification-title">
        <header class="flex gap-3">
          <span class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary"><LuLayers class="size-4" aria-hidden="true" /></span>
          <div>
            <h2 id="edit-classification-title" class="text-sm font-semibold">{{ $t('holdings.editSections.classification') }}</h2>
            <p class="text-xs text-muted-foreground">{{ $t('holdings.editSections.classificationHint') }}</p>
          </div>
        </header>
        <div class="grid gap-4 sm:grid-cols-3">
          <div v-for="field in classificationFields" :key="field.key" class="space-y-2">
            <Label :for="`holding-${field.key}`"><component :is="field.icon" class="size-4 text-muted-foreground" aria-hidden="true" />{{ $t(`portfolio.${field.key}`) }}</Label>
            <Input :id="`holding-${field.key}`" v-model="form[field.key]" maxlength="100" :placeholder="$t(`holdings.classificationPlaceholders.${field.key}`)" :disabled="submitting" />
          </div>
        </div>
      </section>

      <section class="space-y-4 border-t pt-6" aria-labelledby="edit-notes-title">
        <header class="flex gap-3">
          <span class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary"><LuNotebookPen class="size-4" aria-hidden="true" /></span>
          <div>
            <h2 id="edit-notes-title" class="text-sm font-semibold"><label for="edit-notes">{{ $t('holdings.notes') }}</label></h2>
            <p class="text-xs text-muted-foreground">{{ $t('holdings.editSections.notesHint') }}</p>
          </div>
        </header>
        <Textarea
          id="edit-notes"
          data-testid="holding-edit-notes"
          v-model="form.notes"
          :placeholder="$t('holdings.notesPlaceholder')"
          :disabled="submitting"
        />
      </section>
    </form>
    <!-- Leaving the page is the page's own "Back to account" control. -->
    <div class="flex border-t pt-5 sm:justify-end">
      <Button
        data-testid="holding-edit-submit"
        class="w-full sm:w-auto"
        type="submit"
        form="edit-holding-form"
        :disabled="!canSubmit"
        :loading="submitting"
        :loading-text="$t('common.saving')"
      >
        <LuSave class="size-4" aria-hidden="true" />{{ $t('holdings.saveChanges') }}
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
import AssetIdentity from './AssetIdentity.vue'
import { holdingAPI } from '@/services/holdingApi'
import { LuArrowRight, LuBanknote, LuCalculator, LuCircleDollarSign, LuFactory, LuGlobe, LuHash, LuInfo, LuLayers, LuNotebookPen, LuReceipt, LuSave, LuTriangleAlert } from '@kalimahapps/vue-icons/lu'
import { tradeCurrency } from '@/lib/portfolio-format'
import { getLanguage } from '@/i18n'

const CLASSIFICATION_FIELDS = [
  { key: 'asset_class', icon: LuLayers },
  { key: 'sector', icon: LuFactory },
  { key: 'country', icon: LuGlobe },
]

export default {
  components: {
    Alert,
    AssetIdentity,
    Button,
    FormSelect,
    Input,
    Label,
    Textarea,
    LuArrowRight, LuBanknote, LuCalculator, LuCircleDollarSign, LuHash, LuInfo, LuLayers, LuNotebookPen, LuReceipt, LuSave, LuTriangleAlert,
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
  emits: ['success', 'busy'],
  setup() {
    return { classificationFields: CLASSIFICATION_FIELDS }
  },
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
    asset() {
      const symbol = this.holding.symbol, at = symbol.lastIndexOf('.')
      return {
        Code: at < 0 ? symbol : symbol.slice(0, at),
        Exchange: at < 0 ? '' : symbol.slice(at + 1),
        Currency: tradeCurrency(this.holding.quote_currency) || this.holding.buy_currency,
        Type: this.holding.asset_class,
        Name: this.holding.name && this.holding.name !== symbol ? this.holding.name : '',
      }
    },
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

    formatNumber(value) {
      return new Intl.NumberFormat(getLanguage(), { maximumFractionDigits: 8 }).format(value || 0)
    },

    formatCurrency(amount, currency) {
      return new Intl.NumberFormat(getLanguage(), {
        style: 'currency',
        currency: currency || 'USD',
      }).format(amount || 0)
    },
  },
}
</script>
