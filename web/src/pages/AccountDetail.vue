<template>
  <FullPageSpinner
    data-testid="account-detail-loading"
    :visible="loading || deleting"
    :message="spinnerMessage"
  />
  <section data-testid="account-detail-page" class="min-w-0 space-y-6">
    <!-- Action pages (add or edit a holding, activities, transfers) have their
         own title and "Back to account" control, so the account header is hidden. -->
    <div v-if="!$route.meta.accountAction" data-testid="account-detail-header" class="flex flex-wrap items-center gap-3">
      <Button variant="outline" size="icon" as-child>
        <router-link
          data-testid="account-detail-back"
          to="/accounts"
          :aria-label="$t('accountDetail.backToAccounts')"
          :title="$t('accountDetail.backToAccounts')"
        >
          <LuArrowLeft class="size-4" />
        </router-link>
      </Button>
      <div class="flex min-w-0 items-center gap-3">
        <div data-testid="account-detail-logo" class="flex size-11 shrink-0 items-center justify-center rounded-xl border bg-card" aria-hidden="true">
          <Favicon v-if="logoDomain" :domain="logoDomain" :size="24" />
          <LuWallet v-else class="size-5 text-muted-foreground" />
        </div>
        <div class="min-w-0">
          <p class="text-xs font-medium uppercase tracking-wide text-muted-foreground">{{ $t('accounts.account') }}</p>
          <h1
            data-testid="account-detail-name"
            class="break-words text-2xl font-semibold leading-tight tracking-tight"
          >
            {{ account?.account_name || $t('accounts.account') }}
          </h1>
          <p class="mt-0.5 flex flex-wrap items-center gap-x-2 text-sm">
            <span><span class="text-muted-foreground">{{ $t('accountDetail.institution') }}:</span> <span data-testid="account-detail-platform" class="font-medium">{{ account?.platform || $t('accounts.general') }}</span></span>
            <template v-if="account?.currency">
              <span class="text-muted-foreground" aria-hidden="true">·</span>
              <span><span class="text-muted-foreground">{{ $t('accounts.currency') }}:</span> <span data-testid="account-detail-currency" class="font-medium">{{ account.currency }}</span></span>
            </template>
          </p>
          <div v-if="accountTags.length" data-testid="account-detail-tags" class="mt-2 flex flex-wrap gap-1.5">
            <TagBadge v-for="tag in accountTags" :key="tag.id" :tag="tag" />
          </div>
        </div>
      </div>
      <div v-if="account" class="ml-auto flex flex-wrap gap-2">
        <TagPicker v-model="accountTags" kind="account" :id="account.id" testid="account-tags" />
        <Button variant="outline" as-child>
          <router-link data-testid="account-analysis-link" :to="{ name: 'AccountAnalysis', params: { id: account.id } }"><LuChartNoAxesCombined class="size-4" />{{ $t('portfolio.accountAnalysis') }}</router-link>
        </Button>
      </div>
    </div>
    <Alert
      data-testid="account-detail-error"
      v-if="error"
      variant="error"
      :description="error"
    />
    <Alert
      data-testid="account-detail-notification"
      v-if="alert.show"
      v-bind="alert"
      dismissible
      @dismiss="alert.show = false"
    />
    <template v-if="!loading && !error && account">
      <router-view v-if="$route.meta.accountAction" v-slot="{ Component }">
        <component :is="Component" :key="$route.path" :account="account" :holdings="holdings" :activities="activities" :holdings-error="holdingsError" :activities-error="activitiesError" @retry="fetchAccountDetails" @saved="onActionSaved" />
      </router-view>
      <template v-else>
      <div class="grid min-w-0 grid-cols-2 gap-3 lg:grid-cols-4">
        <Card
          :data-testid="`account-summary-${stat.label}`"
          v-for="stat in summaryCards"
          :key="stat.label"
          class="min-w-0 gap-2 py-4"
        >
          <CardHeader class="px-4">
            <CardDescription class="flex min-w-0 items-center gap-2"><component :is="stat.icon" class="size-4 shrink-0 text-primary/80" aria-hidden="true" /><span class="min-w-0 break-words">{{ $t('accountDetail.' + stat.label) }}</span>
              <ZenToggle v-if="stat.label === 'totalValue'" data-testid="account-zen-toggle" class="-my-1.5 ml-auto" />
            </CardDescription>
          </CardHeader>
          <CardContent class="space-y-1 px-4">
            <p class="break-all text-lg font-semibold tabular-nums" :class="stat.tone">{{ stat.value }}</p>
            <p v-if="stat.detail" :data-testid="`account-summary-${stat.label}-detail`" class="text-xs font-medium tabular-nums" :class="stat.tone">{{ stat.detail }}</p>
            <p v-if="stat.caption" :data-testid="`account-summary-${stat.label}-caption`" class="flex items-center gap-1.5 break-words text-xs text-muted-foreground tabular-nums"><LuCoins class="size-3.5 shrink-0" aria-hidden="true" />{{ stat.caption }}</p>
          </CardContent>
        </Card>
      </div>
      <!-- Prices are daily closes; say which close they are and when the page last loaded. -->
      <div v-if="holdings.length" data-testid="account-prices-status" class="-mt-2 flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-muted-foreground">
        <span class="inline-flex items-center gap-1.5"><LuClock class="size-3.5 shrink-0" aria-hidden="true" /><span data-testid="account-prices-as-of">{{ latestPriceDate ? $t('accountDetail.pricesAsOf', { date: formatDay(latestPriceDate) }) : $t('accountDetail.pricesUnavailable') }}</span></span>
        <template v-if="lastRefreshed">
          <span aria-hidden="true">·</span>
          <span data-testid="account-last-refreshed">{{ $t('accountDetail.lastRefreshed', { time: formatTime(lastRefreshed) }) }}</span>
        </template>
        <Button data-testid="account-refresh" variant="ghost" size="sm" class="h-7 gap-1.5 px-2 text-xs" :disabled="refreshing" @click="refreshPrices"><LuRefreshCw class="size-3.5" :class="refreshing && 'animate-spin motion-reduce:animate-none'" aria-hidden="true" />{{ $t('portfolio.refresh') }}</Button>
      </div>
      <Alert v-if="holdingsSummary && !holdingsSummary.complete" variant="warning" :description="holdingsSummary.warnings.join('; ')" />
      <Tabs v-model="activeTab" class="min-w-0 space-y-4">
        <TabsList class="grid w-full grid-cols-2 sm:w-fit">
          <TabsTrigger data-testid="holdings-tab" value="holdings">
            <LuBriefcase class="size-4" aria-hidden="true" />
            {{ $t('accountDetail.holdings') }}
            <span class="text-xs tabular-nums text-muted-foreground">{{ holdings.length }}</span>
          </TabsTrigger>
          <TabsTrigger data-testid="activities-tab" value="activities">
            <LuHistory class="size-4" aria-hidden="true" />
            {{ $t('accountDetail.activities') }}
            <span class="text-xs tabular-nums text-muted-foreground">{{ activities.length }}</span>
          </TabsTrigger>
        </TabsList>
        <TabsContent
          data-testid="holdings-panel"
          value="holdings"
          class="min-w-0 space-y-4"
        >
          <HoldingsList :account="account" :has-holdings="holdings.length > 0" :refresh-key="refreshKey"
            @add="$router.push({ name: 'HoldingCreate', params: { id: account.id } })"
            @edit="openEditHolding" @delete="confirmDeleteHolding" />
        </TabsContent>
        <TabsContent
          data-testid="activities-panel"
          value="activities"
          class="min-w-0 space-y-4"
        >
          <TransactionsPanel :account="account" :activities="activities" @changed="fetchAccountDetails" />
        </TabsContent>
      </Tabs>
      </template>
    </template>
    <ConfirmModal
      data-testid="holding-delete-confirmation"
      :visible="showDeleteConfirm && !$route.meta.accountAction"
      variant="danger"
      :title="$t('accountDetail.deleteHoldingTitle')"
      :message="
        $t('accountDetail.deleteHoldingMessage', {
          symbol: holdingToDelete?.symbol,
        })
      "
      :confirm-text="$t('common.delete')"
      :cancel-text="$t('common.cancel')"
      :loading-text="$t('common.deleting')"
      :loading="deleting"
      @cancel="showDeleteConfirm = false"
      @confirm="deleteHolding"
    />
  </section>
</template>
<script>
import {
  FeedbackAlert as Alert,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  Favicon,
  ConfirmModal,
  FullPageSpinner,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui'
import TagBadge from '@/components/wealth/TagBadge.vue'
import TagPicker from '@/components/wealth/TagPicker.vue'
import { zen } from '@/composables/useZen'
import ZenToggle from '@/components/wealth/ZenToggle.vue'
import TransactionsPanel from '@/components/portfolio/TransactionsPanel.vue'
import HoldingsList from '@/components/holdings/HoldingsList.vue'
import { LuArrowLeft, LuBriefcase, LuChartNoAxesCombined, LuChartPie, LuClock, LuCoins, LuRefreshCw, LuHistory, LuLandmark, LuTrendingDown, LuTrendingUp, LuWallet } from '@kalimahapps/vue-icons/lu'
import { accountAPI } from '@/services/accountApi'
import { platformDomain } from '@/lib/platforms'
import { holdingAPI } from '@/services/holdingApi'
import { activityAPI } from '@/services/activityApi'
import { wealthAPI } from '@/services/wealthApi'

export default {
  name: 'AccountDetailPage',
  components: {
    TransactionsPanel,
    TagBadge,
    TagPicker,
    HoldingsList,
    Alert,
    LuArrowLeft,
    LuBriefcase,
    LuChartNoAxesCombined,
    LuClock,
    LuCoins,
    LuRefreshCw,
    ZenToggle,
    LuHistory,
    LuWallet,
    Favicon,
    Button,
    Card,
    CardContent,
    CardDescription,
      CardHeader,
      ConfirmModal,
      FullPageSpinner,
                Tabs,
    TabsContent,
    TabsList,
    TabsTrigger,
  },
  data() {
    return {
      accountTags: [],
      lastRefreshed: null,
      refreshing: false,
      refreshKey: 0,
      account: null,
      holdings: [],
      holdingsError: '',
      activitiesError: '',
      holdingsSummary: null,
      activities: [],
      loading: true,
      error: null,
      activeTab: this.$route.query.tab === 'activities' ? 'activities' : 'holdings',
      showDeleteConfirm: false,
      holdingToDelete: null,
      deleting: false,
      alert: {
        show: false,
        variant: 'success',
        title: '',
        description: '',
      },
    }
  },
  computed: {
    summaryCards() {
      const currency = this.account?.currency
      const summary = this.holdingsSummary
      const profit = summary?.complete ? summary.total_profit_loss : null
      const income = (summary?.dividend || 0) + (summary?.interest || 0)
      const tone = profit == null || zen.value || Math.abs(profit) < 0.005 ? '' : profit < 0 ? 'text-destructive' : 'text-emerald-600 dark:text-emerald-400'
      return [
        { label: 'totalValue', icon: LuLandmark, value: this.formatCurrency(this.totalEquity == null ? null : this.totalEquity + (this.account?.cash_balance || 0), currency) },
        { label: 'cashBalance', icon: LuWallet, value: this.formatCurrency(this.account?.cash_balance, currency) },
        { label: 'equity', icon: LuChartPie, value: this.formatCurrency(this.totalEquity, currency) },
        {
          label: 'profitLoss',
          icon: profit < 0 ? LuTrendingDown : LuTrendingUp,
          tone,
          value: profit == null ? '—' : this.formatCurrency(profit, currency, true),
          // Return on cost, with an explicit sign so gains and losses read without color.
          detail: profit == null || zen.value ? '' : this.$t('accountDetail.returnOnCost', { percent: this.formatPercent(summary.total_performance) }),
          // Income is rare for most accounts; show it only when there is some.
          caption: Math.abs(income) >= 0.005 ? this.$t('accountDetail.incomeCaption', { amount: this.formatCurrency(income, currency) }) : '',
        },
      ]
    },
    logoDomain() {
      return platformDomain(this.account?.platform)
    },
    zen() {
      return zen.value
    },
    // Newest close among priced holdings; older closes stay visible per row.
    latestPriceDate() {
      return this.holdings.filter(row => row.price_available && row.price_date).map(row => row.price_date).sort().at(-1) || ''
    },
    totalEquity() {
      return this.holdingsSummary?.complete ? this.holdingsSummary.total_value : null
    },
    spinnerMessage() {
      if (this.deleting) return this.$t('common.deleting')
      return this.$t('accountDetail.loadingDetails')
    },
  },
  watch: {
    '$route.query.tab'(tab) { this.activeTab = tab === 'activities' ? 'activities' : 'holdings' },
    '$route.path'() { this.showDeleteConfirm = false; this.holdingToDelete = null },
  },
  mounted() {
    this.fetchAccountDetails()
  },
  methods: {
    async fetchAccountDetails() {
      this.loading = true
      this.error = null
      try {
        const response = await accountAPI.getAccount(this.$route.params.id)
        this.account = response.data
        // Fetch holdings and activities after account is loaded
        await Promise.all([this.fetchHoldings(), this.fetchActivities(), this.fetchAccountTags()])
      } catch (err) {
        this.error = err.message || 'Failed to load account details'
      } finally {
        this.loading = false
      }
    },
    async fetchAccountTags() {
      try {
        this.accountTags = (await wealthAPI.get(`/tags/account/${this.$route.params.id}`)).data || []
      } catch {
        // Tags are optional context; the account stays usable without them.
        this.accountTags = []
      }
    },
    async fetchHoldings() {
      this.holdingsError = ''
      try {
        const response = await holdingAPI.getHoldings(this.$route.params.id)
        this.holdings = response.data.holdings || []
        this.holdingsSummary = response.data.summary || null
        this.lastRefreshed = new Date()
      } catch (err) {
        this.holdingsError = err.message || this.$t('holdingDetail.loadError')
        this.holdings = []
        this.holdingsSummary = { complete: false, warnings: [err.message] }
      }
    },
    async fetchActivities() {
      this.activitiesError = ''
      try {
        const response = await activityAPI.getActivities(this.$route.params.id)
        this.activities = response.data || []
      } catch (err) {
        this.activitiesError = err.message || this.$t('holdingDetail.activitiesError')
        this.activities = []
      }
    },
    async onActionSaved(action) {
      const activities = action.startsWith('activity') || action === 'transfer'
      await this.$router.replace({ name: 'AccountDetail', params: { id: this.account.id }, query: activities ? { tab: 'activities' } : {} })
      await this.fetchAccountDetails()
      const message = { 'holding-create': 'accountDetail.holdingAddedDesc', 'holding-edit': 'accountDetail.holdingUpdatedDesc', 'account-edit': 'accounts.accountUpdatedDesc' }[action] || 'accountDetail.changesSaved'
      this.showAlert('success', this.$t('common.success'), this.$t(message))
    },
    openEditHolding(holding) {
      this.$router.push({ name: 'HoldingEdit', params: { id: this.account.id, holdingId: holding.id } })
    },
    confirmDeleteHolding(holding) {
      this.holdingToDelete = holding
      this.showDeleteConfirm = true
    },
    async deleteHolding() {
      if (!this.holdingToDelete || this.deleting) return
      this.deleting = true
      try {
        await holdingAPI.deleteHolding(
          this.$route.params.id,
          this.holdingToDelete.id
        )
        this.showDeleteConfirm = false
        this.holdingToDelete = null
        await this.fetchAccountDetails()
        this.showAlert(
          'success',
          this.$t('common.success'),
          this.$t('accountDetail.holdingDeletedDesc')
        )
      } catch (err) {
        console.error('Failed to delete holding:', err)
        this.showAlert(
          'error',
          this.$t('common.error'),
          err.message || 'Failed to delete holding'
        )
      } finally {
        this.deleting = false
      }
    },
    showAlert(variant, title, description) {
      this.alert = {
        show: true,
        variant,
        title,
        description,
      }
      if (variant === 'success') {
        setTimeout(() => {
          this.alert.show = false
        }, 3000)
      }
    },
    formatCurrency(amount, currency, signed = false) {
      if (amount == null) return '—'
      if (zen.value) return '••••'
      return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: currency || 'USD',
        signDisplay: signed ? 'exceptZero' : 'auto',
      }).format(Math.abs(amount || 0) < 0.005 ? 0 : amount || 0)
    },
    formatDay(day) {
      return new Date(`${day}T00:00:00Z`).toLocaleDateString(this.$i18n.locale, { day: 'numeric', month: 'short', year: 'numeric', timeZone: 'UTC' })
    },
    formatTime(date) {
      return date.toLocaleTimeString(this.$i18n.locale, { hour: '2-digit', minute: '2-digit' })
    },
    async refreshPrices() {
      if (this.refreshing) return
      this.refreshing = true
      try {
        await this.fetchHoldings()
        this.refreshKey += 1
      } finally {
        this.refreshing = false
      }
    },
    formatPercent(value) {
      if (!Number.isFinite(value)) return '—'
      const shown = Math.abs(value) < 0.005 ? 0 : value
      return `${shown > 0 ? '+' : ''}${shown.toFixed(2)}%`
    },
    formatDate(dateString) {
      if (!dateString) return '-'
      const date = new Date(dateString)
      return date.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
      })
    },
  },
}
</script>
