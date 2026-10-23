<template>
  <FullPageSpinner
    :visible="loading || saving || deleting"
    :message="spinnerMessage"
  />
  <section class="min-w-0 space-y-6">
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
      <Button class="w-full sm:w-auto" @click="openCreateModal">
        <Plus class="size-4" />
        {{ $t('accounts.addAccount') }}
      </Button>
    </div>
    <Alert
      v-if="notification"
      v-bind="notification"
      dismissible
      @dismiss="clearNotification"
    />
    <Card v-if="!loading && error">
      <CardContent class="space-y-4 py-6">
        <Alert variant="error" :description="error" />
        <Button variant="outline" @click="fetchAccounts">
          {{ $t('common.retry') }}
        </Button>
      </CardContent>
    </Card>
    <Card v-if="!loading && !error && accounts.length === 0">
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
        <Button class="max-w-full whitespace-normal" @click="openCreateModal">
          {{ $t('accounts.createFirst') }}
        </Button>
      </CardContent>
    </Card>
    <template v-if="!loading && !error && accounts.length">
      <div class="grid gap-4 md:hidden">
        <Card v-for="account in accounts" :key="account.id" class="min-w-0">
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
              :class="account.initial_balance < 0 && 'text-destructive'"
            >
              {{ formatCurrency(account.initial_balance, account.currency) }}
            </p>
          </CardContent>
          <CardFooter class="flex-wrap justify-between gap-3 border-t pt-4">
            <div class="flex items-center gap-2">
              <Switch
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
              :view-to="`/accounts/${account.id}`"
              @edit="openEditModal(account)"
              @delete="confirmDelete(account)"
            />
          </CardFooter>
        </Card>
      </div>
      <Card class="hidden min-w-0 overflow-hidden py-0 md:block">
        <CardContent class="min-w-0 p-0">
          <Table>
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
              <TableRow v-for="account in accounts" :key="account.id">
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
                  :class="account.initial_balance < 0 && 'text-destructive'"
                >
                  {{
                    formatCurrency(account.initial_balance, account.currency)
                  }}
                </TableCell>
                <TableCell>
                  <div class="flex items-center gap-2">
                    <Switch
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
                    class="justify-end"
                    :view-to="`/accounts/${account.id}`"
                    @edit="openEditModal(account)"
                    @delete="confirmDelete(account)"
                  />
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>
      <p class="text-sm text-muted-foreground">
        {{ accounts.length }} {{ $t('accounts.accounts_count') }}
      </p>
    </template>
    <Dialog :open="showModal" @update:open="!$event && !saving && closeModal()">
      <DialogContent
        :show-close-button="!saving"
        @interact-outside="saving && $event.preventDefault()"
        @escape-key-down="saving && $event.preventDefault()"
      >
        <DialogHeader>
          <DialogTitle>
            {{
              editingAccount
                ? $t('accounts.editAccount')
                : $t('accounts.createAccount')
            }}
          </DialogTitle>
          <DialogDescription>{{ $t('accounts.subtitle') }}</DialogDescription>
        </DialogHeader>
        <form class="space-y-4" @submit.prevent="saveAccount">
          <div class="space-y-2">
            <Label for="accountName">{{ $t('accounts.accountName') }} *</Label>
            <Input
              id="accountName"
              v-model="form.account_name"
              maxlength="100"
              :placeholder="$t('accounts.accountNamePlaceholder')"
              :aria-invalid="!!fieldErrors.account_name"
            />
            <p v-if="fieldErrors.account_name" class="text-sm text-destructive">
              {{ fieldErrors.account_name }}
            </p>
          </div>
          <div class="grid gap-4 sm:grid-cols-2">
            <div class="space-y-2">
              <Label for="currency">{{ $t('accounts.currency') }} *</Label>
              <FormSelect
                id="currency"
                v-model="form.currency"
                :options="currencyOptions"
                :placeholder="$t('accounts.selectCurrency')"
                :invalid="!!fieldErrors.currency"
              />
              <p v-if="fieldErrors.currency" class="text-sm text-destructive">
                {{ fieldErrors.currency }}
              </p>
            </div>
            <div class="space-y-2">
              <Label for="initialBalance">
                {{ $t('accounts.cashBalance') }}
              </Label>
              <Input
                id="initialBalance"
                v-model.number="form.initial_balance"
                type="number"
                step="0.01"
                placeholder="0.00"
                :aria-invalid="!!fieldErrors.initial_balance"
              />
              <p
                v-if="fieldErrors.initial_balance"
                class="text-sm text-destructive"
              >
                {{ fieldErrors.initial_balance }}
              </p>
            </div>
          </div>
          <div class="space-y-2">
            <Label for="platform">{{ $t('accounts.platform') }}</Label>
            <PlatformSelect id="platform" v-model="form.platform" />
            <p v-if="fieldErrors.platform" class="text-sm text-destructive">
              {{ fieldErrors.platform }}
            </p>
          </div>
          <div class="space-y-2">
            <Label for="description">{{ $t('accounts.description') }}</Label>
            <Textarea
              id="description"
              v-model="form.description"
              maxlength="255"
              :placeholder="$t('accounts.descriptionPlaceholder')"
            />
            <p v-if="fieldErrors.description" class="text-sm text-destructive">
              {{ fieldErrors.description }}
            </p>
          </div>
          <div class="flex items-start gap-3 rounded-lg border p-3">
            <Checkbox id="excluded" v-model="form.is_excluded" />
            <div class="space-y-1">
              <Label for="excluded" class="leading-snug">
                {{ $t('accounts.excludeFromTotals') }}
              </Label>
              <p class="text-xs text-muted-foreground">
                {{ $t('accounts.excludeHint') }}
              </p>
            </div>
          </div>
          <Alert v-if="formError" variant="error" :description="formError" />
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              :disabled="saving"
              @click="closeModal"
            >
              {{ $t('common.cancel') }}
            </Button>
            <Button
              type="submit"
              :loading="saving"
              :loading-text="$t('common.saving')"
            >
              {{
                editingAccount
                  ? $t('accounts.updateAccount')
                  : $t('accounts.createAccount')
              }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
    <ConfirmModal
      :visible="showDeleteModal"
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
    <ConfirmModal
      :visible="showUpdateModal"
      variant="warning"
      :title="$t('accounts.updateConfirmTitle')"
      :message="`${$t('accounts.updateConfirmMessage')} ${form.account_name}?`"
      :confirm-text="$t('accounts.updateAccount')"
      :cancel-text="$t('common.cancel')"
      :loading-text="$t('common.updating')"
      :loading="saving"
      @confirm="confirmUpdateAccount"
      @cancel="cancelUpdate"
    />
    <ConfirmModal
      :visible="showStatusModal"
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
  Checkbox,
  ConfirmModal,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Favicon,
  FormSelect,
  FullPageSpinner,
  Input,
  Label,
  PlatformSelect,
  Switch,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  Textarea,
} from '@/components/ui'
import { CreditCard } from 'lucide-vue-next'
import { Plus } from 'lucide-vue-next'
import { accountAPI } from '@/services/accountApi'
import { createAccountSchema, updateAccountSchema } from '@/schemas/account'

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
    Checkbox,
    ConfirmModal,
    CreditCard,
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    Favicon,
    FormSelect,
    FullPageSpinner,
    Input,
    Label,
    PlatformSelect,
    Plus,
    Switch,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
    Textarea,
  },
  computed: {
    spinnerMessage() {
      if (this.loading) return this.$t('accounts.loadingAccounts')
      if (this.saving) return this.$t('accounts.savingAccount')
      if (this.deleting) return this.$t('accounts.deletingAccount')
      return ''
    },
  },
  data() {
    return {
      accounts: [],
      currencyOptions: [
        'USD',
        'EUR',
        'GBP',
        'TRY',
        'JPY',
        'CAD',
        'AUD',
        'CHF',
        'CNY',
        'INR',
        'KRW',
        'SGD',
        'HKD',
        'SEK',
        'NOK',
        'DKK',
        'PLN',
        'BRL',
        'MXN',
        'RUB',
      ],
      loading: true,
      error: null,
      showModal: false,
      showDeleteModal: false,
      editingAccount: null,
      accountToDelete: null,
      saving: false,
      deleting: false,
      formError: null,
      notification: null,
      notificationTimer: null,
      showUpdateModal: false,
      showStatusModal: false,
      accountToToggle: null,
      togglingStatus: false,
      form: {
        account_name: '',
        currency: '',
        initial_balance: 0,
        platform: '',
        description: '',
        is_excluded: false,
      },
      fieldErrors: {},
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
      } catch (err) {
        this.error = err.message || 'Failed to load accounts'
      } finally {
        this.loading = false
      }
    },

    openCreateModal() {
      this.editingAccount = null
      this.form = {
        account_name: '',
        currency: '',
        initial_balance: 0,
        platform: '',
        description: '',
        is_excluded: false,
      }
      this.formError = null
      this.fieldErrors = {}
      this.showModal = true
    },

    openEditModal(account) {
      this.editingAccount = account
      this.form = {
        account_name: account.account_name,
        currency: account.currency,
        initial_balance: account.initial_balance,
        platform: account.platform || '',
        description: account.description || '',
        is_excluded: account.is_excluded,
      }
      this.formError = null
      this.fieldErrors = {}
      this.showModal = true
    },

    closeModal() {
      this.showModal = false
      this.editingAccount = null
      this.formError = null
      this.fieldErrors = {}
    },

    validateForm() {
      this.fieldErrors = {}
      const schema = this.editingAccount
        ? updateAccountSchema
        : createAccountSchema
      const result = schema.safeParse(this.form)

      if (!result.success) {
        const errors = result.error.flatten().fieldErrors
        this.fieldErrors = {
          account_name: errors.account_name?.[0],
          currency: errors.currency?.[0],
          initial_balance: errors.initial_balance?.[0],
          platform: errors.platform?.[0],
          description: errors.description?.[0],
        }
        return false
      }
      return true
    },

    async saveAccount() {
      if (!this.validateForm()) return

      // If editing, show confirmation modal first
      if (this.editingAccount) {
        this.showUpdateModal = true
        return
      }
      // For create, proceed directly
      await this.performSave()
    },

    cancelUpdate() {
      this.showUpdateModal = false
    },

    async confirmUpdateAccount() {
      this.showUpdateModal = false
      await this.performSave()
    },

    async performSave() {
      this.saving = true
      this.formError = null
      try {
        if (this.editingAccount) {
          await accountAPI.updateAccount(this.editingAccount.id, this.form)
          this.showNotification(
            'success',
            'Account updated',
            `${this.form.account_name} has been updated successfully.`
          )
        } else {
          await accountAPI.createAccount(this.form)
          this.showNotification(
            'success',
            'Account created',
            `${this.form.account_name} has been created successfully.`
          )
        }
        this.closeModal()
        await this.fetchAccounts()
      } catch (err) {
        this.formError = err.message || 'Failed to save account'
        this.showNotification('error', 'Error', this.formError)
      } finally {
        this.saving = false
      }
    },

    toggleAccountStatus(account) {
      this.accountToToggle = account
      this.showStatusModal = true
    },

    cancelToggleStatus() {
      this.showStatusModal = false
      this.accountToToggle = null
    },

    async confirmToggleStatus() {
      if (!this.accountToToggle) return
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
        this.showStatusModal = false
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
      this.showDeleteModal = true
    },

    cancelDelete() {
      this.showDeleteModal = false
      this.accountToDelete = null
    },

    async deleteAccount() {
      if (!this.accountToDelete) return
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
