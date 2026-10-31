<template>
  <FullPageSpinner
    data-testid="accounts-loading"
    :visible="loading || deleting"
    :message="spinnerMessage"
  />
  <section data-testid="accounts-page" class="min-w-0 space-y-6">
    <div
      class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"
    >
      <div class="min-w-0">
        <h1 class="text-2xl font-semibold tracking-tight">
          {{ $t('accounts.title') }}
        </h1>
        <p class="mt-1 text-sm text-muted-foreground">
          {{ $t('accounts.subtitle') }}
        </p>
      </div>
      <Button
        data-testid="accounts-add"
        class="w-full sm:w-auto"
        as-child
      >
        <router-link :to="{ name: 'AccountCreate' }">
          <Plus class="size-4" />
          {{ $t('accounts.addAccount') }}
        </router-link>
      </Button>
    </div>
    <Alert
      data-testid="accounts-notification"
      v-if="notification"
      v-bind="notification"
      dismissible
      @dismiss="clearNotification"
    />
    <Card v-if="!loading && error">
      <CardContent class="space-y-4 py-6">
        <Alert
          data-testid="accounts-error"
          variant="error"
          :description="error"
        />
        <Button
          data-testid="accounts-retry"
          variant="outline"
          @click="fetchAccounts"
        >
          {{ $t('common.retry') }}
        </Button>
      </CardContent>
    </Card>
    <Card
      data-testid="accounts-empty"
      v-if="!loading && !error && accounts.length === 0"
    >
      <CardContent
        class="flex flex-col items-center gap-4 px-6 py-12 text-center"
      >
        <CreditCard class="size-10 text-muted-foreground" />
        <h2 class="text-lg font-semibold">
          {{ $t('accounts.noAccountsYet') }}
        </h2>
        <p class="max-w-md text-sm text-muted-foreground">
          {{ $t('accounts.noAccountsDescription') }}
        </p>
        <Button
          data-testid="accounts-create-first"
          class="max-w-full whitespace-normal"
          as-child
        >
          <router-link :to="{ name: 'AccountCreate' }">
            {{ $t('accounts.createFirst') }}
          </router-link>
        </Button>
      </CardContent>
    </Card>
    <template v-if="!loading && !error && accounts.length">
      <div data-testid="accounts-cards" class="grid gap-4 md:hidden">
        <Card
          :data-testid="`account-card-${account.id}`"
          v-for="account in accounts"
          :key="account.id"
          class="min-w-0"
        >
          <CardHeader class="gap-3">
            <CardTitle class="break-words text-base">
              {{ account.account_name }}
            </CardTitle>
            <CardDescription v-if="account.description" class="break-words">
              {{ account.description }}
            </CardDescription>
            <div class="flex flex-wrap gap-2">
              <Badge variant="secondary">
                {{ account.platform || $t('accounts.general') }}
              </Badge>
              <Badge variant="outline">{{ account.currency }}</Badge>
              <Badge v-if="account.is_excluded" variant="outline">
                {{ $t('accounts.excludeFromTotals') }}
              </Badge>
            </div>
          </CardHeader>
          <CardContent>
            <p class="text-xs text-muted-foreground">
              {{ $t('accounts.balance') }}
            </p>
            <p
              class="mt-1 break-all text-2xl font-semibold tabular-nums"
              :class="account.cash_balance < 0 && 'text-destructive'"
            >
              {{ formatCurrency(account.cash_balance, account.currency) }}
            </p>
          </CardContent>
          <CardFooter class="flex-wrap justify-between gap-3 border-t pt-4">
            <div class="flex items-center gap-2">
              <Switch
                :data-testid="`account-card-${account.id}-status`"
                :id="`mobile-status-${account.id}`"
                :model-value="account.is_active"
                @update:model-value="toggleAccountStatus(account)"
              />
              <Label :for="`mobile-status-${account.id}`" class="text-xs">
                {{
                  account.is_active
                    ? $t('common.active')
                    : $t('common.inactive')
                }}
              </Label>
            </div>
            <RecordActions
              :data-testid="`account-card-${account.id}-actions`"
              :view-to="`/accounts/${account.id}`"
              @edit="openEditPage(account)"
              @delete="confirmDelete(account)"
            />
          </CardFooter>
        </Card>
      </div>
      <Card class="hidden min-w-0 overflow-hidden py-0 md:block">
        <CardContent class="min-w-0 p-0">
          <Table data-testid="accounts-table">
            <TableHeader>
              <TableRow>
                <TableHead>{{ $t('accounts.account') }}</TableHead>
                <TableHead>{{ $t('accounts.platform') }}</TableHead>
                <TableHead>{{ $t('accounts.currency') }}</TableHead>
                <TableHead class="text-right">
                  {{ $t('accounts.balance') }}
                </TableHead>
                <TableHead>{{ $t('accounts.status') }}</TableHead>
                <TableHead class="text-right">
                  {{ $t('common.actions') }}
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow
                :data-testid="`account-row-${account.id}`"
                v-for="account in accounts"
                :key="account.id"
              >
                <TableCell class="max-w-64">
                  <p class="truncate font-medium">
                    {{ account.account_name }}
                  </p>
                  <p
                    v-if="account.description"
                    class="truncate text-xs text-muted-foreground"
                  >
                    {{ account.description }}
                  </p>
                  <Badge
                    v-if="account.is_excluded"
                    variant="outline"
                    class="mt-1"
                  >
                    {{ $t('accounts.excludeFromTotals') }}
                  </Badge>
                </TableCell>
                <TableCell>
                  <span class="flex items-center gap-2">
                    <Favicon
                      v-if="account.platform"
                      :domain="getPlatformDomain(account.platform)"
                      :size="16"
                    />
                    {{ account.platform || $t('accounts.general') }}
                  </span>
                </TableCell>
                <TableCell>
                  <Badge variant="secondary">
                    {{ getCurrencyFlag(account.currency) }}
                    {{ account.currency }}
                  </Badge>
                </TableCell>
                <TableCell
                  class="text-right font-medium tabular-nums"
                  :class="account.cash_balance < 0 && 'text-destructive'"
                >
                  {{
                    formatCurrency(account.cash_balance, account.currency)
                  }}
                </TableCell>
                <TableCell>
                  <div class="flex items-center gap-2">
                    <Switch
                      :data-testid="`account-row-${account.id}-status`"
                      :id="`status-${account.id}`"
                      :model-value="account.is_active"
                      @update:model-value="toggleAccountStatus(account)"
                    />
                    <Label :for="`status-${account.id}`" class="text-xs">
                      {{
                        account.is_active
                          ? $t('common.active')
                          : $t('common.inactive')
                      }}
                    </Label>
                  </div>
                </TableCell>
                <TableCell>
                  <RecordActions
                    :data-testid="`account-row-${account.id}-actions`"
                    class="justify-end"
                    :view-to="`/accounts/${account.id}`"
                    @edit="openEditPage(account)"
                    @delete="confirmDelete(account)"
                  />
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>
      <p data-testid="accounts-count" class="text-sm text-muted-foreground">
        {{ accounts.length }} {{ $t('accounts.accounts_count') }}
      </p>
    </template>
    <InlineConfirmation
      data-testid="account-delete-confirmation"
      :visible="showDeleteConfirm"
      variant="danger"
      :title="$t('accounts.deleteConfirmTitle')"
      :message="`${$t('accounts.deleteConfirmMessage')} ${
        accountToDelete?.account_name
      }? ${$t('accounts.deleteConfirmWarning')}`"
      :confirm-text="$t('accounts.deleteAccount')"
      :cancel-text="$t('common.cancel')"
      :loading-text="$t('common.deleting')"
      :loading="deleting"
      @confirm="deleteAccount"
      @cancel="cancelDelete"
    />
    <InlineConfirmation
      data-testid="account-status-confirmation"
      :visible="showStatusConfirm"
      :title="
        accountToToggle?.is_active
          ? $t('accounts.deactivateAccount')
          : $t('accounts.activateAccount')
      "
      :message="`${
        accountToToggle?.is_active
          ? $t('accounts.deactivateConfirm')
          : $t('accounts.activateConfirm')
      } ${accountToToggle?.account_name}?`"
      :confirm-text="
        accountToToggle?.is_active ? $t('common.inactive') : $t('common.active')
      "
      :cancel-text="$t('common.cancel')"
      :loading-text="$t('common.updating')"
      :loading="togglingStatus"
      @confirm="confirmToggleStatus"
      @cancel="cancelToggleStatus"
    />
  </section>
</template>
<script>
import {
  RecordActions,
  FeedbackAlert as Alert,
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
  InlineConfirmation,
  Favicon,
  FullPageSpinner,
  Switch,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui'
import { CreditCard } from 'lucide-vue-next'
import { Plus } from 'lucide-vue-next'
import { accountAPI } from '@/services/accountApi'

export default {
  name: 'AccountsPage',
  components: {
    RecordActions,
    Alert,
    Badge,
    Button,
    Card,
    CardContent,
    CardDescription,
    CardFooter,
    CardHeader,
    CardTitle,
    InlineConfirmation,
    CreditCard,
    Favicon,
    FullPageSpinner,
    Plus,
    Switch,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
  },
  computed: {
    spinnerMessage() {
      if (this.loading) return this.$t('accounts.loadingAccounts')
      if (this.deleting) return this.$t('accounts.deletingAccount')
      return ''
    },
  },
  data() {
    return {
      accounts: [],
      loading: true,
      error: null,
      showDeleteConfirm: false,
      accountToDelete: null,
      deleting: false,
      formError: null,
      notification: null,
      notificationTimer: null,
      showStatusConfirm: false,
      accountToToggle: null,
      togglingStatus: false,
    }
  },
  mounted() {
    this.fetchAccounts()
  },
  methods: {
    async fetchAccounts() {
      this.loading = true
      this.error = null
      try {
        const response = await accountAPI.getAccounts()
        this.accounts = response.accounts || []
        const created = this.accounts.find(account => account.id === this.$route.query.created)
        if (created) {
          this.showNotification('success', this.$t('accounts.accountCreated'), `${created.account_name} ${this.$t('accounts.accountCreatedDesc')}`)
          const query = { ...this.$route.query }
          delete query.created
          await this.$router.replace({ name: 'Accounts', query })
        }
      } catch (err) {
        this.error = err.message || 'Failed to load accounts'
      } finally {
        this.loading = false
      }
    },

    openEditPage(account) {
      this.$router.push({ name: 'AccountEdit', params: { id: account.id } })
    },

    toggleAccountStatus(account) {
      this.accountToToggle = account
      this.showStatusConfirm = true
    },

    cancelToggleStatus() {
      this.showStatusConfirm = false
      this.accountToToggle = null
    },

    async confirmToggleStatus() {
      if (!this.accountToToggle || this.togglingStatus) return
      this.togglingStatus = true
      try {
        const response = await accountAPI.toggleAccountStatus(
          this.accountToToggle.id
        )
        this.accountToToggle.is_active = response.account.is_active
        this.showNotification(
          'success',
          'Status updated',
          `${this.accountToToggle.account_name} is now ${
            this.accountToToggle.is_active ? 'active' : 'inactive'
          }.`
        )
        this.showStatusConfirm = false
        this.accountToToggle = null
      } catch (err) {
        this.showNotification(
          'error',
          'Error',
          err.message || 'Failed to update status'
        )
      } finally {
        this.togglingStatus = false
      }
    },

    confirmDelete(account) {
      this.accountToDelete = account
      this.showDeleteConfirm = true
    },

    cancelDelete() {
      this.showDeleteConfirm = false
      this.accountToDelete = null
    },

    async deleteAccount() {
      if (!this.accountToDelete || this.deleting) return
      this.deleting = true
      const accountName = this.accountToDelete.account_name
      try {
        await accountAPI.deleteAccount(this.accountToDelete.id)
        this.cancelDelete()
        await this.fetchAccounts()
        this.showNotification(
          'success',
          'Account deleted',
          `${accountName} has been deleted.`
        )
      } catch (err) {
        this.formError = err.message || 'Failed to delete account'
        this.showNotification('error', 'Error', this.formError)
      } finally {
        this.deleting = false
      }
    },

    showNotification(variant, title, description = '') {
      // Clear existing timer
      if (this.notificationTimer) {
        clearTimeout(this.notificationTimer)
      }
      // Set notification
      this.notification = { variant, title, description }
      // Auto-clear after 5 seconds
      this.notificationTimer = setTimeout(() => {
        this.clearNotification()
      }, 5000)
    },

    clearNotification() {
      this.notification = null
      if (this.notificationTimer) {
        clearTimeout(this.notificationTimer)
        this.notificationTimer = null
      }
    },

    formatCurrency(amount, currency) {
      return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: currency || 'USD',
      }).format(amount || 0)
    },

    getPlatformDomain(platform) {
      const domains = {
        'Bank of America': 'bankofamerica.com',
        Chase: 'chase.com',
        'Wells Fargo': 'wellsfargo.com',
        Citibank: 'citibank.com',
        HSBC: 'hsbc.com',
        Barclays: 'barclays.co.uk',
        'Deutsche Bank': 'db.com',
        ING: 'ing.com',
        N26: 'n26.com',
        Revolut: 'revolut.com',
        'Interactive Brokers': 'interactivebrokers.com',
        Fidelity: 'fidelity.com',
        'Charles Schwab': 'schwab.com',
        'TD Ameritrade': 'tdameritrade.com',
        'E*TRADE': 'etrade.com',
        Robinhood: 'robinhood.com',
        Webull: 'webull.com',
        'Trading 212': 'trading212.com',
        eToro: 'etoro.com',
        Degiro: 'degiro.com',
        Coinbase: 'coinbase.com',
        Binance: 'binance.com',
        Kraken: 'kraken.com',
        'Crypto.com': 'crypto.com',
        Gemini: 'gemini.com',
        Bybit: 'bybit.com',
        OKX: 'okx.com',
        PayPal: 'paypal.com',
        Wise: 'wise.com',
        Stripe: 'stripe.com',
        Venmo: 'venmo.com',
        'Cash App': 'cash.app',
        'Garanti BBVA': 'garantibbva.com.tr',
        'İş Bankası': 'isbank.com.tr',
        'Yapı Kredi': 'yapikredi.com.tr',
        Akbank: 'akbank.com',
        'QNB Finansbank': 'qnbfinansbank.com',
        'Ziraat Bankası': 'ziraatbank.com.tr',
        Midas: 'getmidas.com',
      }
      return domains[platform] || 'example.com'
    },

    getCurrencyFlag(currency) {
      const flags = {
        USD: '🇺🇸',
        EUR: '🇪🇺',
        GBP: '🇬🇧',
        TRY: '🇹🇷',
        JPY: '🇯🇵',
        CAD: '🇨🇦',
        AUD: '🇦🇺',
        CHF: '🇨🇭',
        CNY: '🇨🇳',
        INR: '🇮🇳',
        KRW: '🇰🇷',
        SGD: '🇸🇬',
        HKD: '🇭🇰',
        SEK: '🇸🇪',
        NOK: '🇳🇴',
        DKK: '🇩🇰',
        PLN: '🇵🇱',
        BRL: '🇧🇷',
        MXN: '🇲🇽',
        RUB: '🇷🇺',
        NZD: '🇳🇿',
        ZAR: '🇿🇦',
        THB: '🇹🇭',
        MYR: '🇲🇾',
        IDR: '🇮🇩',
        PHP: '🇵🇭',
        VND: '🇻🇳',
        CZK: '🇨🇿',
        HUF: '🇭🇺',
        RON: '🇷🇴',
        UAH: '🇺🇦',
        SAR: '🇸🇦',
        AED: '🇦🇪',
        QAR: '🇶🇦',
        KWD: '🇰🇼',
        ILS: '🇮🇱',
        EGP: '🇪🇬',
        ARS: '🇦🇷',
        CLP: '🇨🇱',
        COP: '🇨🇴',
        PEN: '🇵🇪',
        NGN: '🇳🇬',
        KES: '🇰🇪',
        MAD: '🇲🇦',
        BTC: '₿',
        ETH: 'Ξ',
      }
      return flags[currency] || '💰'
    },

    getAccountInitial(name) {
      return name ? name.charAt(0).toUpperCase() : '?'
    },

    getAccountColor(platform) {
      const colors = {
        paypal: 'linear-gradient(135deg, #003087 0%, #009cde 100%)',
        stripe: 'linear-gradient(135deg, #635bff 0%, #a960ee 100%)',
        wise: 'linear-gradient(135deg, #9fe870 0%, #37517e 100%)',
        revolut: 'linear-gradient(135deg, #0075eb 0%, #191c1f 100%)',
        default: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      }
      const key = (platform || '').toLowerCase()
      return colors[key] || colors.default
    },
  },
}
</script>
