<template>
  <Dialog :open="show" @update:open="!$event && close()">
    <DialogContent
      data-testid="holding-edit-dialog"
      :show-close-button="!submitting"
      @interact-outside="submitting && $event.preventDefault()"
      @escape-key-down="submitting && $event.preventDefault()"
    >
      <DialogHeader>
        <DialogTitle>{{ $t('holdings.editHolding') }}</DialogTitle>
        <DialogDescription>{{ $t('holdings.formHint') }}</DialogDescription>
      </DialogHeader>
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
              {{ $t('accountDetail.quantity') }}
            </Label>
            <Input
              id="edit-quantity"
              data-testid="holding-edit-quantity"
              type="number"
              v-model.number="form.quantity"
              step="any"
              min="0.00000001"
              :disabled="submitting"
              required
            />
          </div>
          <div class="space-y-2">
            <Label for="edit-buy_price">{{ $t('accountDetail.price') }}</Label>
            <Input
              id="edit-buy_price"
              data-testid="holding-edit-price"
              type="number"
              v-model.number="form.buy_price"
              step="any"
              min="0"
              :disabled="submitting"
              required
            />
          </div>
          <div class="space-y-2">
            <Label for="edit-buyCurrency">
              {{ $t('accountDetail.currency') }}
            </Label>
            <FormSelect
              id="edit-buyCurrency"
              data-testid="holding-edit-currency"
              v-model="form.buy_currency"
              :disabled="submitting"
              :options="[
                ...new Set([form.buy_currency, 'USD', 'EUR', 'GBP', 'TRY']),
              ]"
            />
          </div>
        </div>
        <div class="space-y-2">
          <Label for="edit-notes">{{ $t('holdings.notes') }}</Label>
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
          <span>{{ $t('holdings.totalCost') }}</span>
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
      <DialogFooter>
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
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
<script>
import {
  FeedbackAlert as Alert,
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  FormSelect,
  Input,
  Label,
  Textarea,
} from '@/components/ui'
import { holdingAPI } from '@/services/holdingApi'

export default {
  components: {
    Alert,
    Button,
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    FormSelect,
    Input,
    Label,
    Textarea,
  },
  name: 'EditHoldingModal',
  props: {
    show: {
      type: Boolean,
      default: false,
    },
    accountId: {
      type: String,
      required: true,
    },
    holding: {
      type: Object,
      default: null,
    },
  },
  emits: ['close', 'success'],
  data() {
    return {
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
    show(newVal) {
      if (newVal && this.holding) {
        this.populateForm()
      }
    },
    holding(newVal) {
      if (newVal && this.show) {
        this.populateForm()
      }
    },
  },
  methods: {
    populateForm() {
      this.formError = null
      this.form = {
        quantity: this.holding.quantity,
        buy_price: this.holding.buy_price,
        buy_currency: this.holding.buy_currency || 'USD',
        notes: this.holding.notes || '',
      }
    },

    async submitUpdate() {
      if (!this.canSubmit || this.submitting) return

      this.submitting = true
      this.formError = null
      try {
        await holdingAPI.updateHolding(this.accountId, this.holding.id, {
          quantity: this.form.quantity,
          buy_price: this.form.buy_price,
          buy_currency: this.form.buy_currency,
          notes: this.form.notes,
        })
        this.$emit('success')
        this.$emit('close')
      } catch (err) {
        console.error('Failed to update holding:', err)
        this.formError = err.message || 'Failed to update holding'
      } finally {
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
