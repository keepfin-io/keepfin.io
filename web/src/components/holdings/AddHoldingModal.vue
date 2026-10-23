<template>
  <Dialog :open="show" @update:open="!$event && close()">
    <DialogContent
      :show-close-button="!submitting"
      @interact-outside="submitting && $event.preventDefault()"
      @escape-key-down="submitting && $event.preventDefault()"
    >
      <DialogHeader>
        <DialogTitle>{{ $t('accountDetail.addHolding') }}</DialogTitle>
        <DialogDescription>{{ $t('holdings.formHint') }}</DialogDescription>
      </DialogHeader>
      <Alert v-if="formError" variant="error" :description="formError" />
      <div v-if="!selectedAsset" class="space-y-3">
        <Label for="assetSearch">{{ $t('holdings.searchAsset') }}</Label>
        <Input
          id="assetSearch"
          v-model="searchQuery"
          @update:model-value="onSearchInput"
          :placeholder="$t('holdings.searchPlaceholder')"
        />
        <p
          v-if="searchLoading"
          class="flex items-center gap-2 text-sm text-muted-foreground"
        >
          <LoaderCircle class="size-4 animate-spin" />
          {{ $t('common.loading') }}
        </p>
        <Command v-if="searchResults.length" class="rounded-lg border">
          <CommandList>
            <CommandGroup>
              <CommandItem
                v-for="asset in searchResults"
                :key="`${asset.Code}-${asset.Exchange}`"
                :value="`${asset.Code}.${asset.Exchange}`"
                @select="selectAsset(asset)"
              >
                <div class="min-w-0 flex-1">
                  <p class="font-medium">{{ asset.Code }}</p>
                  <p class="truncate text-xs text-muted-foreground">
                    {{ asset.Name }}
                  </p>
                </div>
                <Badge variant="secondary">{{ asset.Exchange }}</Badge>
              </CommandItem>
            </CommandGroup>
          </CommandList>
        </Command>
        <p
          v-else-if="searchQuery.length >= 2 && !searchLoading"
          class="text-sm text-muted-foreground"
        >
          {{ $t('common.noResults') }}
        </p>
      </div>
      <form
        id="add-holding-form"
        v-if="selectedAsset"
        class="space-y-4"
        @submit.prevent="submitHolding"
      >
        <div
          class="flex items-start justify-between gap-3 rounded-lg border p-3"
        >
          <div class="min-w-0">
            <p class="font-medium">
              {{ selectedAsset.Code }}
              <Badge variant="secondary">{{ selectedAsset.Exchange }}</Badge>
            </p>
            <p class="break-words text-sm text-muted-foreground">
              {{ selectedAsset.Name }}
            </p>
          </div>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            :disabled="submitting"
            @click="clearSelection"
          >
            {{ $t('common.edit') }}
          </Button>
        </div>
        <p v-if="livePrice" class="text-sm text-muted-foreground">
          {{ $t('accountDetail.currentPrice') }}:
          {{ formatCurrency(livePrice.close, selectedAsset.Currency) }}
          <span class="tabular-nums">
            ({{ livePrice.change_p?.toFixed(2) }}%)
          </span>
        </p>
        <p v-if="priceLoading" class="text-sm text-muted-foreground">
          {{ $t('common.loading') }}
        </p>
        <div class="grid gap-4 sm:grid-cols-2">
          <div class="space-y-2">
            <Label for="add-quantity">{{ $t('accountDetail.quantity') }}</Label>
            <Input
              id="add-quantity"
              type="number"
              v-model.number="form.quantity"
              step="any"
              min="0.00000001"
              :disabled="submitting"
              required
            />
          </div>
          <div class="space-y-2">
            <Label for="add-buy_price">{{ $t('accountDetail.price') }}</Label>
            <Input
              id="add-buy_price"
              type="number"
              v-model.number="form.buy_price"
              step="any"
              min="0"
              :disabled="submitting"
              required
            />
          </div>
          <div class="space-y-2">
            <Label for="add-fee">{{ $t('accountDetail.fee') }}</Label>
            <Input
              id="add-fee"
              type="number"
              v-model.number="form.fee"
              step="any"
              min="0"
              :disabled="submitting"
            />
          </div>
          <div class="space-y-2">
            <Label for="add-buyCurrency">
              {{ $t('accountDetail.currency') }}
            </Label>
            <FormSelect
              id="add-buyCurrency"
              v-model="form.buy_currency"
              :disabled="submitting"
              :options="[
                ...new Set([form.buy_currency, 'USD', 'EUR', 'GBP', 'TRY']),
              ]"
            />
          </div>
        </div>
        <div class="space-y-2">
          <Label for="add-notes">{{ $t('holdings.notes') }}</Label>
          <Textarea
            id="add-notes"
            v-model="form.notes"
            :disabled="submitting"
          />
        </div>
        <div
          class="flex flex-wrap justify-between gap-2 rounded-lg bg-muted p-3 text-sm"
        >
          <span>{{ $t('holdings.totalCost') }}</span>
          <strong class="break-all tabular-nums">
            {{
              formatCurrency(
                form.quantity * form.buy_price + (form.fee || 0),
                form.buy_currency
              )
            }}
          </strong>
        </div>
      </form>
      <DialogFooter>
        <Button
          type="button"
          variant="outline"
          :disabled="submitting"
          @click="close"
        >
          {{ $t('common.cancel') }}
        </Button>
        <Button
          type="submit"
          form="add-holding-form"
          :disabled="!canSubmit"
          :loading="submitting"
          :loading-text="$t('common.saving')"
        >
          {{ $t('accountDetail.addHolding') }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
<script>
import {
  FeedbackAlert as Alert,
  Badge,
  Button,
  Command,
  CommandGroup,
  CommandItem,
  CommandList,
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
import { LoaderCircle } from 'lucide-vue-next'
import { marketAPI } from '@/services/marketApi'
import { holdingAPI } from '@/services/holdingApi'

export default {
  components: {
    Alert,
    Badge,
    Button,
    Command,
    CommandGroup,
    CommandItem,
    CommandList,
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    FormSelect,
    Input,
    Label,
    LoaderCircle,
    Textarea,
  },
  name: 'AddHoldingModal',
  props: {
    show: {
      type: Boolean,
      default: false,
    },
    accountId: {
      type: String,
      required: true,
    },
  },
  emits: ['close', 'success'],
  data() {
    return {
      searchQuery: '',
      searchResults: [],
      searchLoading: false,
      searchTimeout: null,
      selectedAsset: null,
      livePrice: null,
      priceLoading: false,
      submitting: false,
      formError: null,
      form: {
        symbol: '',
        quantity: null,
        buy_price: null,
        fee: 0,
        buy_currency: 'USD',
        notes: '',
      },
    }
  },
  computed: {
    canSubmit() {
      return (
        this.selectedAsset &&
        this.form.quantity > 0 &&
        this.form.buy_price !== null &&
        this.form.buy_price !== '' &&
        this.form.buy_price >= 0
      )
    },
  },
  watch: {
    show(newVal) {
      if (!newVal) {
        this.resetForm()
      }
    },
  },
  methods: {
    onSearchInput() {
      clearTimeout(this.searchTimeout)

      if (this.searchQuery.length < 2) {
        this.searchResults = []
        return
      }

      this.searchTimeout = setTimeout(() => {
        this.performSearch()
      }, 300) // 300ms debounce
    },

    async performSearch() {
      const query = this.searchQuery
      this.searchLoading = true
      try {
        const response = await marketAPI.searchAssets(this.searchQuery, 10)
        if (this.show && this.searchQuery === query && !this.selectedAsset)
          this.searchResults = response.results || []
      } catch (err) {
        console.error('Search failed:', err)
        this.searchResults = []
      } finally {
        this.searchLoading = false
      }
    },

    async selectAsset(asset) {
      this.selectedAsset = asset
      this.searchResults = []
      this.form.symbol = `${asset.Code}.${asset.Exchange}`
      this.form.buy_currency = asset.Currency || 'USD'

      // Fetch live price
      this.priceLoading = true
      try {
        const response = await marketAPI.getLivePrice(
          `${asset.Code}.${asset.Exchange}`
        )
        if (!this.show || this.selectedAsset !== asset) return
        this.livePrice = response.data
        // Pre-fill buy price with current price
        if (this.livePrice?.close) {
          this.form.buy_price = this.livePrice.close
        }
      } catch (err) {
        console.error('Failed to fetch live price:', err)
      } finally {
        this.priceLoading = false
      }
    },

    clearSelection() {
      this.selectedAsset = null
      this.livePrice = null
      this.searchQuery = ''
      this.form.symbol = ''
      this.form.buy_price = null
    },

    async submitHolding() {
      if (!this.canSubmit || this.submitting) return

      this.submitting = true
      this.formError = null
      try {
        await holdingAPI.addHolding(this.accountId, {
          symbol: this.form.symbol,
          quantity: this.form.quantity,
          buy_price: this.form.buy_price,
          fee: this.form.fee || 0,
          buy_currency: this.form.buy_currency,
          notes: this.form.notes,
        })
        this.$emit('success')
        this.$emit('close')
      } catch (err) {
        console.error('Failed to add holding:', err)
        this.formError = err.message || 'Failed to add holding'
      } finally {
        this.submitting = false
      }
    },

    close() {
      if (this.submitting) return
      this.$emit('close')
    },

    resetForm() {
      clearTimeout(this.searchTimeout)
      this.formError = null
      this.searchQuery = ''
      this.searchResults = []
      this.selectedAsset = null
      this.livePrice = null
      this.form = {
        symbol: '',
        quantity: null,
        buy_price: null,
        fee: 0,
        buy_currency: 'USD',
        notes: '',
      }
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
