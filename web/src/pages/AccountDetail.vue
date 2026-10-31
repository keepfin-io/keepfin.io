<template>
  <FullPageSpinner
    data-testid="account-detail-loading"
    :visible="loading || deleting"
    :message="spinnerMessage"
  />
  <section data-testid="account-detail-page" class="min-w-0 space-y-6">
    <div class="flex items-center gap-3">
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
        <component :is="Component" :key="$route.path" :account="account" :holdings="holdings" :activities="activities" @saved="onActionSaved" />
      </router-view>
      <template v-else>
      <div class="grid min-w-0 grid-cols-2 gap-3 xl:grid-cols-6">
        <Card
          :data-testid="`account-summary-${stat.label}`"
          v-for="stat in summaryCards"
          :key="stat.label"
          class="min-w-0 gap-2 py-4"
        >
          <CardHeader class="px-4">
            <CardDescription class="break-words">
              {{ $t('accountDetail.' + stat.label) }}
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
          <Button
            data-testid="holdings-add"
            class="w-full sm:w-auto"
            :disabled="!account.is_active"
            @click="$router.push({ name: 'HoldingCreate', params: { id: account.id } })"
          >
            <Plus class="size-4" />
            {{ $t('accountDetail.addHolding') }}
          </Button>
          <Card data-testid="holdings-empty" v-if="!holdings.length">
            <CardContent
              class="py-10 text-center text-sm text-muted-foreground"
            >
              {{ $t('accountDetail.noHoldingsYet') }}
            </CardContent>
          </Card>
          <template v-else>
            <div data-testid="holdings-cards" class="grid gap-4 md:hidden">
              <Card
                :data-testid="`holding-card-${holding.id}`"
                v-for="holding in holdings"
                :key="holding.id"
                class="min-w-0"
              >
                <CardHeader>
                  <CardTitle class="break-words text-base">
                    {{ holding.name || holding.symbol }}
                  </CardTitle>
                  <CardDescription>{{ holding.symbol }}</CardDescription>
                </CardHeader>
                <CardContent class="grid grid-cols-2 gap-4 text-sm">
                  <div
                    v-for="metric in [
                      { label: 'quantity', value: holding.quantity },
                      {
                        label: 'currentPrice',
                        value: formatCurrency(
                          holding.price_available ? holding.current_price : null,
                          holding.buy_currency
                        ),
                      },
                      {
                        label: 'value',
                        value: formatCurrency(holding.price_available ? holding.value : null, holding.buy_currency),
                      },
                      {
                        label: 'profitLoss',
                        value: formatCurrency(
                          holding.price_available ? holding.profit_loss : null,
                          holding.buy_currency
                        ),
                      },
                      {
                        label: 'performance',
                        value: holding.price_available ? holding.performance?.toFixed(2) + '%' : '—',
                      },
                    ]"
                    :key="metric.label"
                  >
                    <p class="text-xs text-muted-foreground">
                      {{ $t('accountDetail.' + metric.label) }}
                    </p>
                    <p class="mt-1 break-all font-medium tabular-nums">
                      {{ metric.value }}
                    </p>
                  </div>
                </CardContent>
                <CardFooter class="justify-end gap-2 border-t pt-4">
                  <RecordActions
                    :data-testid="`holding-card-${holding.id}-actions`"
                    :compact="false"
                    @edit="openEditHolding(holding)"
                    @delete="confirmDeleteHolding(holding)"
                  />
                </CardFooter>
              </Card>
            </div>
            <Card class="hidden min-w-0 overflow-hidden py-0 md:block">
              <CardContent class="min-w-0 p-0">
                <Table data-testid="holdings-table">
                  <TableHeader>
                    <TableRow>
                      <TableHead
                        v-for="field in [
                          'name',
                          'quantity',
                          'currentPrice',
                          'value',
                          'profitLoss',
                          'performance',
                          'actions',
                        ]"
                        :key="field"
                      >
                        {{ $t('accountDetail.' + field) }}
                      </TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    <TableRow
                      :data-testid="`holding-row-${holding.id}`"
                      v-for="holding in holdings"
                      :key="holding.id"
                    >
                      <TableCell class="max-w-60">
                        <p class="truncate font-medium">
                          {{ holding.name || holding.symbol }}
                        </p>
                        <p class="text-xs text-muted-foreground">
                          {{ holding.symbol }}
                        </p>
                      </TableCell>
                      <TableCell>{{ holding.quantity }}</TableCell>
                      <TableCell>
                        {{
                          formatCurrency(
                            holding.price_available ? holding.current_price : null,
                            holding.buy_currency
                          )
                        }}
                      </TableCell>
                      <TableCell>
                        {{ formatCurrency(holding.price_available ? holding.value : null, holding.buy_currency) }}
                      </TableCell>
                      <TableCell
                        :class="
                          holding.profit_loss < 0
                            ? 'text-destructive'
                            : 'text-brand'
                        "
                      >
                        {{
                          formatCurrency(holding.price_available ? holding.profit_loss : null, holding.buy_currency)
                        }}
                      </TableCell>
                      <TableCell>
                        {{ holding.price_available ? holding.performance?.toFixed(2) + '%' : '—' }}
                      </TableCell>
                      <TableCell>
                        <RecordActions
                          :data-testid="`holding-row-${holding.id}-actions`"
                          @edit="openEditHolding(holding)"
                          @delete="confirmDeleteHolding(holding)"
                        />
                      </TableCell>
                    </TableRow>
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          </template>
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
  CardFooter,
  CardHeader,
  CardTitle,
  RecordActions,
  InlineConfirmation,
  FullPageSpinner,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui'
import TransactionsPanel from '@/components/portfolio/TransactionsPanel.vue'
import { ArrowLeft } from 'lucide-vue-next'
import { Plus } from 'lucide-vue-next'
import { accountAPI } from '@/services/accountApi'
import { holdingAPI } from '@/services/holdingApi'
import { activityAPI } from '@/services/activityApi'

export default {
  name: 'AccountDetailPage',
  components: {
    TransactionsPanel,
    Alert,
    ArrowLeft,
    Button,
    Card,
    CardContent,
    CardDescription,
    CardFooter,
    CardHeader,
    CardTitle,
    InlineConfirmation,
    RecordActions,
    FullPageSpinner,
    Plus,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
    Tabs,
    TabsContent,
    TabsList,
    TabsTrigger,
  },
  data() {
    return {
      account: null,
      holdings: [],
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
          value: this.formatCurrency(
            this.account?.cash_balance,
            this.account?.currency
          ),
        },
        {
          label: 'equity',
          value: this.formatCurrency(this.totalEquity, this.account?.currency),
        },
        {
          label: 'interest',
          value: this.formatCurrency(this.holdingsSummary?.interest, this.account?.currency),
        },
        {
          label: 'dividend',
          value: this.formatCurrency(this.holdingsSummary?.dividend, this.account?.currency),
        },
        { label: 'activities', value: this.activities.length },
        { label: 'currency', value: this.account?.currency },
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
        this.account = response.account
        // Fetch holdings and activities after account is loaded
        await Promise.all([this.fetchHoldings(), this.fetchActivities()])
      } catch (err) {
        this.error = err.message || 'Failed to load account details'
      } finally {
        this.loading = false
      }
    },
    async fetchHoldings() {
      try {
        const response = await holdingAPI.getHoldings(this.$route.params.id)
        this.holdings = response.holdings || []
        this.holdingsSummary = response.summary || null
      } catch (err) {
        console.error('Failed to fetch holdings:', err)
        this.holdings = []
        this.holdingsSummary = { complete: false, warnings: [err.message] }
      }
    },
    async fetchActivities() {
      try {
        const response = await activityAPI.getActivities(this.$route.params.id)
        this.activities = response.activities || []
      } catch (err) {
        console.error('Failed to fetch activities:', err)
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
