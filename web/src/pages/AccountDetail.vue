<template>
  <FullPageSpinner
    data-testid="account-detail-loading"
    :visible="loading || deleting"
    :message="spinnerMessage"
  />
  <section data-testid="account-detail-page" class="min-w-0 space-y-6">
    <div class="flex flex-wrap items-center gap-3">
      <Button variant="outline" size="icon" as-child>
        <router-link
          data-testid="account-detail-back"
          to="/accounts"
          :aria-label="$t('accountDetail.backToAccounts')"
        >
          <ArrowLeft class="size-4" />
        </router-link>
      </Button>
      <div class="min-w-0">
        <h1
          data-testid="account-detail-name"
          class="break-words text-2xl font-semibold tracking-tight"
        >
          {{ account?.account_name || $t('accounts.account') }}
        </h1>
        <p class="text-sm text-muted-foreground">
          {{ account?.platform || $t('accounts.general') }}
        </p>
      </div>
      <Button v-if="account && !$route.meta.accountAction" variant="outline" class="ml-auto" as-child>
        <router-link data-testid="account-analysis-link" :to="{ name: 'AccountAnalysis', params: { id: account.id } }"><ChartNoAxesCombined class="size-4" />{{ $t('portfolio.accountAnalysis') }}</router-link>
      </Button>
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
      <details @toggle="showTags = $event.target.open"><summary class="cursor-pointer text-sm text-primary">{{ $t('wealth.tags') }}</summary><EntityTags v-if="showTags" kind="account" :id="account.id" class="mt-3" /></details>
      <div class="grid min-w-0 grid-cols-2 gap-3 xl:grid-cols-6">
        <Card
          :data-testid="`account-summary-${stat.label}`"
          v-for="stat in summaryCards"
          :key="stat.label"
          class="min-w-0 gap-2 py-4"
        >
          <CardHeader class="px-4">
            <CardDescription class="flex items-center gap-2">
              <span class="inline-flex shrink-0 text-primary/80" aria-hidden="true">
                <component :is="stat.icon" class="size-4" />
              </span>
              <span class="min-w-0 break-words">{{ $t('accountDetail.' + stat.label) }}</span>
            </CardDescription>
          </CardHeader>
          <CardContent
            class="break-all px-4 text-lg font-semibold tabular-nums"
          >
            {{ stat.value }}
          </CardContent>
        </Card>
      </div>
      <Alert v-if="holdingsSummary && !holdingsSummary.complete" variant="warning" :description="holdingsSummary.warnings.join('; ')" />
      <Tabs v-model="activeTab" class="min-w-0 space-y-4">
        <TabsList class="grid w-full grid-cols-2 sm:w-fit">
          <TabsTrigger data-testid="holdings-tab" value="holdings">
            {{ $t('accountDetail.holdings') }}
          </TabsTrigger>
          <TabsTrigger data-testid="activities-tab" value="activities">
            {{ $t('accountDetail.activities') }}
          </TabsTrigger>
        </TabsList>
        <TabsContent
          data-testid="holdings-panel"
          value="holdings"
          class="min-w-0 space-y-4"
        >
          <HoldingsList :account="account" :has-holdings="holdings.length > 0"
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
    <InlineConfirmation
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
  InlineConfirmation,
  FullPageSpinner,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui'
import EntityTags from '@/components/wealth/EntityTags.vue'
import { zen } from '@/composables/useZen'
import TransactionsPanel from '@/components/portfolio/TransactionsPanel.vue'
import HoldingsList from '@/components/holdings/HoldingsList.vue'
import { ArrowLeft, ChartNoAxesCombined } from 'lucide-vue-next'
import { LuArrowLeftRight, LuBanknote, LuCoins, LuLayers, LuPercent, LuWallet } from '@kalimahapps/vue-icons/lu'
import { accountAPI } from '@/services/accountApi'
import { holdingAPI } from '@/services/holdingApi'
import { activityAPI } from '@/services/activityApi'

export default {
  name: 'AccountDetailPage',
  components: {
    TransactionsPanel,
    EntityTags,
    HoldingsList,
    Alert,
    ArrowLeft,
    ChartNoAxesCombined,
    Button,
    Card,
    CardContent,
    CardDescription,
      CardHeader,
      InlineConfirmation,
      FullPageSpinner,
                Tabs,
    TabsContent,
    TabsList,
    TabsTrigger,
  },
  data() {
    return {
      showTags: false,
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
      return [
        {
          label: 'cashBalance',
          icon: LuWallet,
          value: this.formatCurrency(
            this.account?.cash_balance,
            this.account?.currency
          ),
        },
        {
          label: 'equity',
          icon: LuLayers,
          value: this.formatCurrency(this.totalEquity, this.account?.currency),
        },
        {
          label: 'interest',
          icon: LuPercent,
          value: this.formatCurrency(this.holdingsSummary?.interest, this.account?.currency),
        },
        {
          label: 'dividend',
          icon: LuCoins,
          value: this.formatCurrency(this.holdingsSummary?.dividend, this.account?.currency),
        },
        { label: 'activities', icon: LuArrowLeftRight, value: this.activities.length },
        { label: 'currency', icon: LuBanknote, value: this.account?.currency },
      ]
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
        await Promise.all([this.fetchHoldings(), this.fetchActivities()])
      } catch (err) {
        this.error = err.message || 'Failed to load account details'
      } finally {
        this.loading = false
      }
    },
    async fetchHoldings() {
      this.holdingsError = ''
      try {
        const response = await holdingAPI.getHoldings(this.$route.params.id)
        this.holdings = response.data.holdings || []
        this.holdingsSummary = response.data.summary || null
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
    formatCurrency(amount, currency) {
      if (amount == null) return '—'
      if (zen.value) return '••••'
      return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: currency || 'USD',
      }).format(amount || 0)
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
