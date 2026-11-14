<template>
  <div data-testid="holding-add-page" class="space-y-6" :aria-busy="submitting">
      <ol class="flex flex-wrap items-center gap-2 border-b pb-5 text-sm" :aria-label="$t('holdings.addSteps')">
        <li class="flex items-center gap-2 rounded-full py-1 pl-1 pr-3" :class="!selectedAsset ? 'bg-primary/10 font-medium text-primary' : 'text-muted-foreground'" :aria-current="!selectedAsset ? 'step' : undefined">
          <span class="flex size-6 items-center justify-center rounded-full" :class="!selectedAsset ? 'bg-primary text-primary-foreground' : 'bg-primary/15 text-primary'"><LuCheck v-if="selectedAsset" class="size-3.5" aria-hidden="true" /><LuSearch v-else class="size-3.5" aria-hidden="true" /></span>
          <span><span class="sr-only">1. </span>{{ $t('holdings.chooseAsset') }}</span>
        </li>
        <li class="h-px w-6 bg-border" aria-hidden="true"></li>
        <li class="flex items-center gap-2 rounded-full py-1 pl-1 pr-3" :class="selectedAsset ? 'bg-primary/10 font-medium text-primary' : 'text-muted-foreground'" :aria-current="selectedAsset ? 'step' : undefined">
          <span class="flex size-6 items-center justify-center rounded-full" :class="selectedAsset ? 'bg-primary text-primary-foreground' : 'bg-muted'"><LuReceipt class="size-3.5" aria-hidden="true" /></span>
          <span><span class="sr-only">2. </span>{{ $t('holdings.purchaseDetails') }}</span>
        </li>
      </ol>
      <Alert
        data-testid="holding-add-error"
        v-if="formError"
        variant="error"
        :description="formError"
      />
      <div v-if="!selectedAsset" class="space-y-6">
        <!-- What to expect before adding: where prices and currencies come from. -->
        <section data-testid="holding-add-guide" class="grid gap-3 sm:grid-cols-3" :aria-label="$t('holdings.guideTitle')">
          <div v-for="item in guide" :key="item.key" class="flex gap-3 rounded-xl border bg-muted/30 p-3">
            <span class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary"><component :is="item.icon" class="size-4" aria-hidden="true" /></span>
            <div class="min-w-0">
              <p class="text-sm font-medium">{{ $t(`holdings.guide.${item.key}.title`) }}</p>
              <p class="mt-0.5 text-xs leading-relaxed text-muted-foreground">{{ $t(`holdings.guide.${item.key}.text`) }}</p>
            </div>
          </div>
        </section>
        <div class="space-y-4">
        <div class="space-y-2">
        <Label for="assetSearch">{{ $t('holdings.searchAsset') }}</Label>
        <div class="relative">
        <Search class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
        <Input
          id="assetSearch"
          data-testid="holding-add-search"
          v-model="searchQuery"
          @update:model-value="onSearchInput"
          :placeholder="$t('holdings.searchPlaceholder')"
          class="h-12 pl-10 pr-10"
          autocomplete="off"
          :aria-busy="searchLoading"
          aria-describedby="asset-search-hint"
          role="combobox"
          aria-autocomplete="list"
          aria-controls="asset-search-results"
          :aria-expanded="!!filteredResults.length"
          :aria-activedescendant="highlightedIndex >= 0 && filteredResults[highlightedIndex] ? `asset-result-${highlightedIndex}` : undefined"
          @keydown.down.prevent="moveHighlight(1)"
          @keydown.up.prevent="moveHighlight(-1)"
          @keydown.enter.prevent="selectHighlighted"
          @keydown.esc.prevent="clearSearch"
        />
        <Button v-if="searchQuery" data-testid="holding-add-clear-search" type="button" variant="ghost" size="icon" class="absolute right-1 top-1/2 size-9 -translate-y-1/2" :aria-label="$t('holdings.clearSearch')" @click="clearSearch"><X class="size-4" /></Button>
        </div>
        <p id="asset-search-hint" class="text-xs text-muted-foreground">{{ $t('holdings.searchHint') }}</p>
        </div>
        <div class="flex flex-wrap gap-2" role="group" :aria-label="$t('holdings.filterType')">
          <Button v-for="filter in ['all', 'stock', 'fund', 'crypto', 'forex']" :key="filter" :data-testid="`holding-add-filter-${filter}`" type="button" :variant="assetFilter === filter ? 'secondary' : 'ghost'" size="sm" class="rounded-full" :aria-pressed="assetFilter === filter" @click="assetFilter = filter"><component :is="filterIcons[filter]" class="size-3.5" aria-hidden="true" />{{ $t(`holdings.assetTypes.${filter}`) }}</Button>
        </div>
        <p
          data-testid="holding-add-search-loading"
          v-if="searchLoading"
          class="flex items-center gap-2 text-sm text-muted-foreground"
        >
          <LoaderCircle class="size-4 animate-spin" />
          {{ $t('common.loading') }}
        </p>
        <Alert v-if="searchError" data-testid="holding-add-search-error" variant="error" :description="searchError" />
        <div v-if="filteredResults.length" class="space-y-2">
          <p class="text-xs text-muted-foreground" role="status">{{ $t('holdings.resultCount', { count: filteredResults.length }) }}</p>
          <div id="asset-search-results" data-testid="holding-add-search-results" role="listbox" :aria-label="$t('holdings.searchResults')" class="divide-y divide-border/60 overflow-hidden rounded-xl border border-border/60">
              <button
                :data-testid="`holding-add-result-${asset.Code}.${asset.Exchange}`"
                v-for="(asset, index) in filteredResults"
                :key="`${asset.Code}-${asset.Exchange}`"
                :id="`asset-result-${index}`"
                type="button"
                role="option"
                :aria-selected="highlightedIndex === index"
                class="group flex min-h-20 w-full items-center gap-3 px-3 py-3 text-left transition-colors hover:bg-accent/60 focus-visible:bg-accent/60 focus-visible:outline-none sm:px-4"
                :class="highlightedIndex === index && 'bg-accent/60'"
                @click="selectAsset(asset)"
                @focus="highlightedIndex = index"
              >
                <AssetIdentity :asset="asset" detailed />
                <ArrowRight class="size-4 shrink-0 text-muted-foreground group-hover:text-primary" aria-hidden="true" />
              </button>
          </div>
        </div>
        <div
          data-testid="holding-add-search-empty"
          v-else-if="searchQuery.trim().length >= 2 && !searchLoading && !searchError"
          class="rounded-xl bg-muted/40 p-5 text-sm text-muted-foreground"
        >
          <p>{{ $t(searchResults.length ? 'holdings.noFilteredResults' : 'common.noResults') }}</p>
          <Button v-if="searchResults.length" type="button" variant="link" class="px-0" @click="assetFilter = 'all'">{{ $t('holdings.showAllResults') }}</Button>
        </div>
        <Button v-if="searchError" data-testid="holding-add-search-retry" type="button" variant="outline" @click="onSearchInput">{{ $t('common.retry') }}</Button>
        <section v-if="!searchQuery.trim()" data-testid="holding-add-recents" class="space-y-3 pt-2" :aria-label="$t('holdings.recentAssets')">
          <div class="flex items-center justify-between gap-3">
            <h2 class="flex items-center gap-2 text-sm font-medium"><History class="size-4 text-muted-foreground" aria-hidden="true" />{{ $t('holdings.recentAssets') }}</h2>
            <Button v-if="recentAssets.length" data-testid="holding-add-clear-recents" type="button" variant="ghost" size="sm" class="text-xs text-muted-foreground" @click="clearRecentAssets">{{ $t('holdings.clearRecents') }}</Button>
          </div>
          <div v-if="recentAssets.length" class="grid min-w-0 gap-2 sm:grid-cols-2">
            <Button v-for="asset in recentAssets" :key="`${asset.Code}.${asset.Exchange}`" :data-testid="`holding-add-recent-${asset.Code}.${asset.Exchange}`" type="button" variant="ghost" class="h-auto min-w-0 justify-start rounded-xl border border-border/60 p-3" @click="selectAsset(asset)">
              <AssetIdentity :asset="asset" />
            </Button>
          </div>
          <p v-else data-testid="holding-add-recents-empty" class="flex items-center gap-3 rounded-xl border border-dashed p-4 text-sm text-muted-foreground"><History class="size-4 shrink-0" aria-hidden="true" />{{ $t('holdings.recentEmpty') }}</p>
        </section>
        </div>
        <details data-testid="holding-add-manual" class="group rounded-xl border" @toggle="loadManualAssets($event.target.open)">
          <summary class="flex cursor-pointer list-none items-center gap-3 rounded-xl p-4 hover:bg-muted/30 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring [&::-webkit-details-marker]:hidden">
            <span class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-muted"><LuPackagePlus class="size-4" aria-hidden="true" /></span>
            <span class="min-w-0 flex-1">
              <span class="block text-sm font-medium">{{ $t('holdings.manualTitle') }}</span>
              <span class="block text-xs text-muted-foreground">{{ $t('holdings.manualHint') }}</span>
            </span>
            <LuChevronDown class="size-4 shrink-0 text-muted-foreground transition-transform group-open:rotate-180" aria-hidden="true" />
          </summary>
          <div class="space-y-3 border-t p-4">
            <p v-if="manualError" role="alert" class="text-sm text-destructive">{{ manualError }}</p>
            <div v-if="manualAssets.length" class="grid gap-2 sm:grid-cols-2">
              <Button v-for="a in manualAssets" :key="a.id" type="button" variant="outline" class="h-auto min-w-0 justify-between gap-3 rounded-xl p-3" @click="selectAsset({Code:a.symbol,Exchange:'',Name:a.name,Currency:a.currency,Type:a.asset_class,Country:a.country})">
                <span class="truncate">{{ a.name }}</span><span class="shrink-0 text-xs text-muted-foreground">{{ a.currency }}</span>
              </Button>
            </div>
            <p v-else-if="manualLoaded && !manualError" class="text-sm text-muted-foreground">{{ $t('holdings.manualEmpty') }}</p>
            <Button as-child variant="link" class="h-auto px-0">
              <router-link to="/wealth/assets"><LuPlus class="size-4" aria-hidden="true" />{{ $t('wealth.createAsset') }}</router-link>
            </Button>
          </div>
        </details>
      </div>
      <form
        data-testid="holding-add-form"
        id="add-holding-form"
        v-if="selectedAsset"
        class="space-y-4"
        @submit.prevent="submitHolding"
      >
        <div
          class="flex items-start justify-between gap-3 rounded-lg border p-3"
        >
          <AssetIdentity :asset="selectedAsset" />
          <Button
            type="button"
            variant="ghost"
            size="sm"
            :disabled="submitting"
            data-testid="holding-add-change-asset"
            @click="clearSelection"
          >
            {{ $t('common.edit') }}
          </Button>
        </div>
        <p
          data-testid="holding-add-live-price"
          v-if="livePrice"
          class="text-sm text-muted-foreground"
        >
          {{ $t('accountDetail.currentPrice') }}:
          {{ formatCurrency(livePrice.close, selectedAsset.Currency) }}
          <span class="tabular-nums">
            ({{ livePrice.change_p?.toFixed(2) }}%)
          </span>
        </p>
        <p
          data-testid="holding-add-price-loading"
          v-if="priceLoading"
          class="text-sm text-muted-foreground"
        >
          {{ $t('common.loading') }}
        </p>
        <p v-if="priceError" data-testid="holding-add-price-error" role="status" class="text-sm text-muted-foreground">{{ priceError }}</p>
        <div class="grid gap-4 sm:grid-cols-2">
          <div class="space-y-2">
            <Label for="holding-date"><LuCalendar class="size-4 text-muted-foreground" aria-hidden="true" />{{ $t('portfolio.date') }}</Label>
            <Input id="holding-date" v-model="form.date" type="date" :max="new Date().toISOString().slice(0, 10)" required />
          </div>
          <div class="space-y-2">
            <Label for="holding-rate"><LuArrowRightLeft class="size-4 text-muted-foreground" aria-hidden="true" />{{ $t('holdings.exchangeRate') }}</Label>
            <Input id="holding-rate" v-model.number="form.exchange_rate" type="number" min="0.0000000001" step="any" :placeholder="$t('holdings.exchangeRatePlaceholder')" />
          </div>
        </div>
        <div class="grid gap-4 sm:grid-cols-2">
          <div class="space-y-2">
            <Label for="add-quantity"><LuHash class="size-4 text-muted-foreground" aria-hidden="true" />{{ $t('accountDetail.quantity') }}</Label>
            <Input
              id="add-quantity"
              data-testid="holding-add-quantity"
              type="number"
              v-model.number="form.quantity"
              step="any"
              min="0.00000001"
              :placeholder="$t('holdings.quantityPlaceholder')"
              :disabled="submitting"
              required
            />
          </div>
          <div class="space-y-2">
            <Label for="add-buy_price"><LuCircleDollarSign class="size-4 text-muted-foreground" aria-hidden="true" />{{ $t('accountDetail.price') }}</Label>
            <Input
              id="add-buy_price"
              data-testid="holding-add-price"
              type="number"
              v-model.number="form.buy_price"
              step="any"
              min="0"
              :placeholder="$t('holdings.pricePlaceholder')"
              :disabled="submitting"
              required
            />
          </div>
          <div class="space-y-2">
            <Label for="add-fee"><LuReceipt class="size-4 text-muted-foreground" aria-hidden="true" />{{ $t('accountDetail.fee') }}</Label>
            <Input
              id="add-fee"
              data-testid="holding-add-fee"
              type="number"
              v-model.number="form.fee"
              step="any"
              min="0"
              :placeholder="$t('holdings.feePlaceholder')"
              :disabled="submitting"
            />
          </div>
          <div class="space-y-2">
            <Label for="add-buyCurrency">
              <LuBanknote class="size-4 text-muted-foreground" aria-hidden="true" />
              {{ $t('accountDetail.currency') }}
            </Label>
            <FormSelect
              id="add-buyCurrency"
              data-testid="holding-add-currency"
              v-model="form.buy_currency"
              :disabled="submitting || !!lockedCurrency"
              :options="lockedCurrency ? [lockedCurrency] : [
                ...new Set([form.buy_currency, 'USD', 'EUR', 'GBP', 'TRY']),
              ]"
            />
            <p v-if="lockedCurrency" data-testid="holding-add-currency-locked" class="text-xs text-muted-foreground">
              {{ $t('holdings.currencyLocked', { symbol: selectedAsset.Code, currency: lockedCurrency }) }}
            </p>
          </div>
        </div>
        <div class="space-y-2">
          <Label for="add-notes"><LuNotebookPen class="size-4 text-muted-foreground" aria-hidden="true" />{{ $t('holdings.notes') }}</Label>
          <Textarea
            id="add-notes"
            data-testid="holding-add-notes"
            v-model="form.notes"
            :placeholder="$t('holdings.notesPlaceholder')"
            :disabled="submitting"
          />
        </div>
        <div
          class="flex flex-wrap justify-between gap-2 rounded-lg bg-muted p-3 text-sm"
        >
          <span class="flex items-center gap-2"><LuCalculator class="size-4 text-muted-foreground" aria-hidden="true" />{{ $t('holdings.totalCost') }}</span>
          <strong
            data-testid="holding-add-total"
            class="break-all tabular-nums"
          >
            {{
              formatCurrency(
                form.quantity * form.buy_price + (form.fee || 0),
                form.buy_currency
              )
            }}
          </strong>
        </div>
      </form>
      <!-- Leaving the page is the page's own "Back to account" control. -->
      <div v-if="selectedAsset" class="flex border-t pt-5 sm:justify-end">
        <Button
          data-testid="holding-add-submit"
          class="w-full sm:w-auto"
          type="submit"
          form="add-holding-form"
          :disabled="!canSubmit"
          :loading="submitting"
          :loading-text="$t('common.saving')"
        >
          {{ $t('accountDetail.addHolding') }}
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
import { ArrowRight, History, LoaderCircle, Search, X } from 'lucide-vue-next'
import { LuArrowRightLeft, LuBanknote, LuBitcoin, LuBuilding2, LuCalculator, LuCalendar, LuCalendarClock, LuChartPie, LuCheck, LuChevronDown, LuCircleDollarSign, LuHash, LuLayoutGrid, LuNotebookPen, LuPackagePlus, LuPlus, LuReceipt, LuSearch } from '@kalimahapps/vue-icons/lu'
import AssetIdentity from './AssetIdentity.vue'
import { useAuth } from '@/composables/useAuth'
import { useRecentAssets } from '@/composables/useRecentAssets'
import { assetCategory } from '@/lib/asset-search'
import { tradeCurrency } from '@/lib/portfolio-format'

const GUIDE = [
  { key: 'find', icon: LuSearch },
  { key: 'currency', icon: LuArrowRightLeft },
  { key: 'prices', icon: LuCalendarClock },
]
const FILTER_ICONS = { all: LuLayoutGrid, stock: LuBuilding2, fund: LuChartPie, crypto: LuBitcoin, forex: LuArrowRightLeft }
import { wealthAPI } from '@/services/wealthApi'
import { marketAPI } from '@/services/marketApi'
import { holdingAPI } from '@/services/holdingApi'

export default {
  components: {
    Alert,
    AssetIdentity,
    Button,
    FormSelect,
    Input,
    Label,
    LoaderCircle,
    History,
    Search,
    ArrowRight,
    X,
    Textarea,
    LuArrowRightLeft, LuBanknote, LuCalculator, LuCalendar, LuCircleDollarSign, LuHash, LuNotebookPen, LuReceipt,
    LuCheck, LuChevronDown, LuPackagePlus, LuPlus, LuSearch,
  },
  name: 'AddHoldingForm',
  props: {
    accountId: {
      type: String,
      required: true,
    },
  },
  emits: ['success', 'busy'],
  setup() {
    const { user } = useAuth()
    // Icon components stay out of reactive state.
    return { ...useRecentAssets(user), guide: GUIDE, filterIcons: FILTER_ICONS }
  },
  data() {
    return {
      active: true,
      manualAssets: [], manualError: '', manualLoaded: false,
      searchQuery: '',
      searchResults: [],
      searchLoading: false,
      searchTimeout: null,
      searchRevision: 0,
      searchError: null,
      assetFilter: 'all',
      highlightedIndex: -1,
      selectedAsset: null,
      livePrice: null,
      priceLoading: false,
      priceRevision: 0,
      priceError: null,
      submitting: false,
      formError: null,
      form: {
        symbol: '',
        quantity: null,
        buy_price: null,
        fee: 0,
        buy_currency: 'USD',
        notes: '',
        date: new Date().toISOString().slice(0, 10), exchange_rate: '',
      },
    }
  },
  computed: {
    // Trade prices are valued against market quotes, so they use the listing currency.
    lockedCurrency() {
      return tradeCurrency(this.selectedAsset?.Currency)
    },
    filteredResults() {
      const query = this.searchQuery.trim().toUpperCase()
      return this.searchResults.filter(asset => this.assetFilter === 'all' || assetCategory(asset) === this.assetFilter)
        .slice().sort((a, b) => Number(b.Code.toUpperCase() === query) - Number(a.Code.toUpperCase() === query))
    },
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
    submitting(value) { this.$emit('busy', value) },
    filteredResults() { this.highlightedIndex = -1 },
  },
  beforeUnmount() {
    this.active = false
    clearTimeout(this.searchTimeout)
  },
  methods: {
    async loadManualAssets(open) { if (!open) return; this.manualError = ''; try { const r = await wealthAPI.get('/assets/manual');if (this.active) this.manualAssets=r.data } catch(e) { if (this.active) this.manualError=e.message } finally { if (this.active) this.manualLoaded = true } },
    moveHighlight(direction) {
      if (!this.filteredResults.length) return
      this.highlightedIndex = this.highlightedIndex < 0 ? (direction > 0 ? 0 : this.filteredResults.length - 1) : (this.highlightedIndex + direction + this.filteredResults.length) % this.filteredResults.length
      this.$nextTick(() => this.$el.querySelector(`#asset-result-${this.highlightedIndex}`)?.scrollIntoView({ block: 'nearest' }))
    },
    selectHighlighted() {
      const asset = this.filteredResults[this.highlightedIndex >= 0 ? this.highlightedIndex : 0]
      if (asset) this.selectAsset(asset)
    },
    clearSearch() {
      this.searchQuery = ''
      this.onSearchInput()
      this.$nextTick(() => this.$el.querySelector('#assetSearch')?.focus())
    },
    onSearchInput() {
      clearTimeout(this.searchTimeout)
      const revision = ++this.searchRevision
      const query = this.searchQuery.trim()
      this.searchResults = []
      this.searchError = null
      this.searchLoading = query.length >= 2
      if (!this.searchLoading) {
        return
      }
      this.searchTimeout = setTimeout(() => {
        this.performSearch(query, revision)
      }, 300) // 300ms debounce
    },

    async performSearch(query, revision) {
      try {
        const response = await marketAPI.searchAssets(query, 20)
        if (this.active && revision === this.searchRevision && !this.selectedAsset)
          this.searchResults = response.data || []
      } catch (err) {
        if (this.active && revision === this.searchRevision) {
          const key = `holdings.marketErrors.${err.code}`
          this.searchError = err.code && this.$te(key) ? this.$t(key) : err.message || this.$t('holdings.searchFailed')
          this.searchResults = []
        }
      } finally {
        if (this.active && revision === this.searchRevision) this.searchLoading = false
      }
    },

    async selectAsset(asset) {
      if (this.submitting) return
      clearTimeout(this.searchTimeout)
      this.searchRevision += 1
      const revision = ++this.priceRevision
      this.selectedAsset = asset
      if (asset.Exchange) this.rememberAsset(asset)
      this.searchResults = []
      this.searchLoading = false
      this.searchError = null
      this.formError = null
      this.livePrice = null
      this.priceError = null
      this.form.symbol = asset.Exchange ? `${asset.Code}.${asset.Exchange}` : asset.Code
      this.form.buy_currency = tradeCurrency(asset.Currency) || 'USD'
      this.form.buy_price = null

      // Fetch live price
      this.priceLoading = true
      try {
        const response = await marketAPI.getLivePrice(
          this.form.symbol
        )
        if (!this.active || revision !== this.priceRevision) return
        this.livePrice = response.data
        // A delayed quote must not replace a price the user has already entered.
        if (Number.isFinite(this.livePrice?.close) && this.livePrice.close >= 0 && (this.form.buy_price === null || this.form.buy_price === '')) {
          this.form.buy_price = this.livePrice.close
        }
      } catch {
        if (this.active && revision === this.priceRevision) this.priceError = this.$t('holdings.priceUnavailable')
      } finally {
        if (this.active && revision === this.priceRevision) this.priceLoading = false
      }
    },

    clearSelection() {
      this.priceRevision += 1
      this.searchRevision += 1
      this.selectedAsset = null
      this.livePrice = null
      this.priceLoading = false
      this.priceError = null
      this.formError = null
      this.searchQuery = ''
      this.searchResults = []
      this.form.symbol = ''
      this.form.buy_price = null
      this.$nextTick(() => this.$el.querySelector('#assetSearch')?.focus())
    },

    async submitHolding() {
      if (!this.canSubmit || this.submitting) return

      this.submitting = true
      this.formError = null
      try {
        await holdingAPI.addHolding(this.accountId, {
          symbol: this.form.symbol,
          name: this.selectedAsset.Name,
          asset_class: this.selectedAsset.Type || '',
          country: this.selectedAsset.Country || '',
          quantity: this.form.quantity,
          buy_price: this.form.buy_price,
          fee: this.form.fee || 0,
          buy_currency: this.form.buy_currency,
          notes: this.form.notes,
          date: this.form.date, exchange_rate: Number(this.form.exchange_rate) || 0,
        })
        if (this.active) this.$emit('success')
      } catch (err) {
        console.error('Failed to add holding:', err)
        this.formError = err.message || 'Failed to add holding'
        this.submitting = false
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
