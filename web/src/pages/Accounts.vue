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
      <div data-testid="accounts-summary" class="grid gap-3 sm:grid-cols-2">
        <Card class="gap-0 py-4">
          <CardContent class="space-y-1 px-4">
            <p class="flex items-center gap-2 text-sm text-muted-foreground"><LuLayers class="size-4 text-primary/80" aria-hidden="true" />{{ $t('accounts.title') }}</p>
            <p class="text-2xl font-semibold tabular-nums">{{ accounts.length }}</p>
            <p data-testid="accounts-active-count" class="text-xs text-muted-foreground">{{ $t('accounts.summaryActive', { count: activeCount }) }}</p>
          </CardContent>
        </Card>
        <Card class="gap-0 py-4">
          <CardContent class="space-y-1 px-4">
            <p class="flex items-center gap-2 text-sm text-muted-foreground"><LuBanknote class="size-4 text-primary/80" aria-hidden="true" />{{ $t('accounts.cashBalance') }}<ZenToggle data-testid="accounts-zen-toggle" class="-my-1.5 ml-auto" /></p>
            <div data-testid="accounts-cash-totals" class="flex flex-wrap gap-x-4 gap-y-1">
              <p v-for="[currency, amount] in cashTotals" :key="currency" class="text-2xl font-semibold tabular-nums" :class="amount < 0 && 'text-destructive'">{{ formatCurrency(amount, currency) }}</p>
              <p v-if="!cashTotals.length" class="text-2xl font-semibold">—</p>
            </div>
            <p class="text-xs text-muted-foreground">{{ $t('accounts.summaryCashHint') }}</p>
          </CardContent>
        </Card>
      </div>
      <!-- One list for every width: columns from xl, two stacked lines below.
           The account name link stretches over its row, so clicking anywhere
           opens the account; the status switch and actions sit above it. -->
      <div data-testid="accounts-list" class="overflow-hidden rounded-2xl border bg-card shadow-xs">
        <div class="hidden border-b bg-muted/40 px-5 py-2.5 text-xs font-medium uppercase tracking-wide text-muted-foreground xl:grid xl:grid-cols-[minmax(0,1fr)_10rem_7rem_16rem] xl:items-center xl:gap-6" aria-hidden="true">
          <span>{{ $t('accounts.account') }}</span>
          <span class="text-right">{{ $t('accounts.cashBalance') }}</span>
          <span>{{ $t('accounts.status') }}</span>
          <span class="text-right">{{ $t('common.actions') }}</span>
        </div>
        <ul role="list" class="divide-y">
          <li
            v-for="account in accounts"
            :key="account.id"
            :data-testid="`account-row-${account.id}`"
            class="relative grid grid-cols-[minmax(0,1fr)_auto] items-center gap-x-4 gap-y-3 px-5 py-4 transition-colors focus-within:bg-muted/30 hover:bg-muted/30 xl:grid-cols-[minmax(0,1fr)_10rem_7rem_16rem] xl:gap-6"
          >
            <div class="flex min-w-0 items-center gap-3" :class="!account.is_active && 'opacity-60'">
              <div class="flex size-10 shrink-0 items-center justify-center rounded-xl border bg-background" aria-hidden="true">
                <Favicon v-if="account.platform" :domain="getPlatformDomain(account.platform)" :size="22" />
                <LuWallet v-else class="size-5 text-muted-foreground" />
              </div>
              <div class="min-w-0">
                <div class="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1">
                  <router-link :data-testid="`account-row-${account.id}-name`" :to="`/accounts/${account.id}`" class="truncate font-semibold after:absolute after:inset-0 after:content-[''] hover:underline focus-visible:underline focus-visible:outline-none">{{ account.account_name }}</router-link>
                  <Badge v-if="!account.is_active" variant="outline" class="h-5 px-1.5 text-[11px]">{{ $t('common.inactive') }}</Badge>
                  <Badge v-if="account.is_excluded" variant="outline" class="h-5 px-1.5 text-[11px]">{{ $t('accounts.excludeFromTotals') }}</Badge>
                </div>
                <p class="truncate text-sm text-muted-foreground">
                  {{ account.platform || $t('accounts.general') }} · {{ getCurrencyFlag(account.currency) }} {{ account.currency }}<template v-if="account.description"> · {{ account.description }}</template>
                </p>
                <div v-if="accountTags[account.id]?.length" :data-testid="`account-row-${account.id}-tags`" class="mt-1.5 flex flex-wrap gap-1" :aria-label="$t('wealth.tags')">
                  <TagBadge v-for="tag in accountTags[account.id]" :key="tag.id" :tag="tag" />
                </div>
              </div>
            </div>
            <div class="text-right" :class="!account.is_active && 'opacity-60'">
              <p class="text-xs text-muted-foreground xl:hidden">{{ $t('accounts.cashBalance') }}</p>
              <p class="font-semibold tabular-nums" :class="account.cash_balance < 0 && 'text-destructive'">{{ formatCurrency(account.cash_balance, account.currency) }}</p>
            </div>
            <div class="relative z-10 flex h-8 w-fit items-center gap-2">
              <Switch
                :id="`status-${account.id}`"
                :data-testid="`account-row-${account.id}-status`"
                :model-value="account.is_active"
                @update:model-value="toggleAccountStatus(account)"
              />
              <Label :for="`status-${account.id}`" class="text-sm" :class="!account.is_active && 'text-muted-foreground'">{{ account.is_active ? $t('common.active') : $t('common.inactive') }}</Label>
            </div>
            <div class="relative z-10 flex justify-end">
              <RecordActions
                :data-testid="`account-row-${account.id}-actions`"
                :view-to="`/accounts/${account.id}`"
                @edit="openEditPage(account)"
                @delete="confirmDelete(account)"
              />
            </div>
          </li>
        </ul>
      </div>
    </template>
    <ConfirmModal
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
import { zen } from '@/composables/useZen'
import ZenToggle from '@/components/wealth/ZenToggle.vue'
import {
  RecordActions,
  FeedbackAlert as Alert,
  Badge,
  Button,
  Card,
  CardContent,
  ConfirmModal,
  InlineConfirmation,
  Favicon,
  FullPageSpinner,
  Label,
  Switch,
} from '@/components/ui'
import { CreditCard } from 'lucide-vue-next'
import { Plus } from 'lucide-vue-next'
import { accountAPI } from '@/services/accountApi'
import { platformDomain } from '@/lib/platforms'
import { wealthAPI } from '@/services/wealthApi'
import TagBadge from '@/components/wealth/TagBadge.vue'
import { LuBanknote, LuLayers, LuWallet } from '@kalimahapps/vue-icons/lu'

export default {
  name: 'AccountsPage',
  components: {
    TagBadge,
    ZenToggle,
    LuBanknote,
    LuLayers,
    LuWallet,
    RecordActions,
    Alert,
    Badge,
    Button,
    Card,
    CardContent,
    ConfirmModal,
    InlineConfirmation,
    Label,
    CreditCard,
    Favicon,
    FullPageSpinner,
    Plus,
    Switch,
  },
  computed: {
    activeCount() {
      return this.accounts.filter(account => account.is_active).length
    },
    // Cash is summed per currency; converting would need exchange rates.
    cashTotals() {
      const totals = {}
      for (const account of this.accounts) {
        if (!account.is_active || account.is_excluded) continue
        totals[account.currency] = (totals[account.currency] || 0) + (account.cash_balance || 0)
      }
      return Object.entries(totals).sort(([a], [b]) => a.localeCompare(b))
    },
    spinnerMessage() {
      if (this.loading) return this.$t('accounts.loadingAccounts')
      if (this.deleting) return this.$t('accounts.deletingAccount')
      return ''
    },
  },
  data() {
    return {
      accounts: [],
      accountTags: {},
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
        // Tags are optional context; accounts still load if they are unavailable.
        const [response, tags] = await Promise.all([
          accountAPI.getAccounts(),
          wealthAPI.get('/tags/account').catch(() => ({ data: {} })),
        ])
        this.accounts = response.data || []
        this.accountTags = tags.data || {}
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
        this.accountToToggle.is_active = response.data.is_active
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
      if (zen.value) return '••••'
      return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: currency || 'USD',
      }).format(amount || 0)
    },

    getPlatformDomain(platform) {
      return platformDomain(platform) || 'example.com'
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
