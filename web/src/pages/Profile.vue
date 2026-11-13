<template>
  <FullPageSpinner
    data-testid="profile-loading"
    :visible="loading"
    :message="$t('common.loading')"
  />
  <section
    data-testid="profile-page"
    class="mx-auto min-w-0 max-w-3xl space-y-6"
  >
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">
        {{ $t('profile.title') }}
      </h1>
      <p class="mt-1 text-sm text-muted-foreground">
        {{ $t('profile.subtitle') }}
      </p>
    </div>
    <Card>
      <CardContent class="space-y-6 pt-6">
        <div class="flex flex-col items-center gap-3">
          <Button
            variant="ghost"
            class="size-24 rounded-full p-0"
            :aria-label="$t('profile.chooseAvatar')"
            data-testid="profile-avatar-trigger"
            @click="showEmojiPicker = true"
          >
            <Avatar class="size-24">
              <AvatarFallback class="text-5xl">
                {{ selectedEmoji }}
              </AvatarFallback>
            </Avatar>
          </Button>
          <p class="text-xs text-muted-foreground">
            {{ $t('profile.avatarHint') }}
          </p>
        </div>
        <Separator />
        <dl class="grid gap-5 sm:grid-cols-2">
          <div class="min-w-0 sm:col-span-2">
            <dt class="text-sm text-muted-foreground">
              {{ $t('profile.userId') }}
            </dt>
            <dd
              data-testid="profile-user-id"
              class="mt-1 break-all font-mono text-sm"
            >
              {{ user?.id || '-' }}
            </dd>
          </div>
          <div class="min-w-0">
            <dt class="text-sm text-muted-foreground">
              {{ $t('profile.email') }}
            </dt>
            <dd
              data-testid="profile-email"
              class="mt-1 break-all text-sm font-medium"
            >
              {{ user?.email || '-' }}
            </dd>
          </div>
          <div>
            <dt class="text-sm text-muted-foreground">
              {{ $t('profile.memberSince') }}
            </dt>
            <dd
              data-testid="profile-member-since"
              class="mt-1 text-sm font-medium"
            >
              {{ formatDate(user?.created_at) }}
            </dd>
          </div>
        </dl>
      </CardContent>
    </Card>
    <Card data-testid="profile-sessions">
      <CardHeader>
        <CardTitle>{{ $t('profile.sessions.title') }}</CardTitle>
        <p class="text-sm leading-relaxed text-muted-foreground">
          {{ $t('profile.sessions.hint') }}
        </p>
      </CardHeader>
      <CardContent :aria-busy="sessionsLoading">
        <p v-if="sessionsLoading" data-testid="profile-sessions-loading" role="status" class="py-4 text-sm text-muted-foreground">
          {{ $t('profile.sessions.loading') }}
        </p>
        <div v-else-if="sessionsError" data-testid="profile-sessions-error" role="alert" class="space-y-3 text-sm">
          <p class="text-muted-foreground">{{ $t(`profile.sessions.${sessionsError}`) }}</p>
          <p v-if="authError === 'signOutFailed'" class="text-destructive">{{ $t('auth.signOutFailed') }}</p>
          <Button v-if="sessionsError === 'reauthRequired'" data-testid="profile-sessions-sign-out" variant="outline" :disabled="isLoading" @click="logout">
            {{ $t('profile.sessions.signOut') }}
          </Button>
          <Button v-else variant="outline" data-testid="profile-sessions-retry" @click="loadSessions">
            {{ $t('profile.sessions.retry') }}
          </Button>
        </div>
        <ul v-else-if="sessionDevices.length" class="divide-y divide-border/60">
          <li v-for="session in sessionDevices" :key="session.id" data-testid="profile-session" class="flex gap-3 py-5 first:pt-0 last:pb-0">
            <div class="flex size-10 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
              <component :is="session.device.icon" class="size-5" aria-hidden="true" />
            </div>
            <div class="min-w-0 flex-1 space-y-3">
              <div>
                <div class="flex flex-wrap items-center gap-x-3 gap-y-1.5">
                  <h3 class="text-sm font-medium">{{ session.device.browser }} · {{ session.device.os }}</h3>
                  <span v-if="session.id === currentSessionId" data-testid="profile-current-session" class="rounded-full bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">
                    {{ $t('profile.sessions.current') }}
                  </span>
                </div>
                <p class="mt-1 text-xs text-muted-foreground">{{ $t(`profile.sessions.${session.device.kind}`) }}</p>
              </div>
              <dl class="grid gap-x-4 gap-y-3 text-xs sm:grid-cols-3">
                <div class="min-w-0">
                  <dt class="text-muted-foreground">{{ $t('profile.sessions.ipAddress') }}</dt>
                  <dd class="mt-1 break-all font-mono">{{ session.ipAddress || '—' }}</dd>
                </div>
                <div>
                  <dt class="text-muted-foreground">{{ $t('profile.sessions.signedIn') }}</dt>
                  <dd class="mt-1">{{ formatSessionDate(session.createdAt) }}</dd>
                </div>
                <div>
                  <dt class="text-muted-foreground">{{ $t('profile.sessions.expires') }}</dt>
                  <dd class="mt-1">{{ formatSessionDate(session.expiresAt) }}</dd>
                </div>
              </dl>
            </div>
          </li>
        </ul>
        <p v-else data-testid="profile-sessions-empty" class="py-4 text-sm text-muted-foreground">
          {{ $t('profile.sessions.empty') }}
        </p>
      </CardContent>
    </Card>
    <Card>
      <CardHeader>
        <CardTitle>{{ $t('profile.preferences') }}</CardTitle>
      </CardHeader>
      <CardContent class="space-y-6">
        <div class="space-y-2">
          <h3 class="text-sm font-medium">{{ $t('profile.theme') }}</h3>
          <p class="text-sm text-muted-foreground">
            {{ $t('profile.themeHint') }}
          </p>
          <ThemeToggle data-testid="profile-theme" />
        </div>
        <Separator />
        <div class="space-y-2">
          <Label for="baseCurrency">{{ $t('profile.baseCurrency') }}</Label>
          <p class="text-sm text-muted-foreground">
            {{ $t('profile.baseCurrencyHint') }}
          </p>
          <FormSelect
            id="baseCurrency"
            data-testid="profile-currency"
            v-model="baseCurrency"
            :options="
              currencies.map((c) => ({
                value: c.code,
                label: `${c.symbol} ${c.code} - ${c.name}`,
              }))
            "
            :disabled="saving"
            @update:model-value="saveCurrency"
          />
        </div>
        <div class="space-y-2">
          <Label for="language">{{ $t('profile.language') }}</Label>
          <p class="text-sm text-muted-foreground">
            {{ $t('profile.languageHint') }}
          </p>
          <FormSelect
            id="language"
            data-testid="profile-language"
            v-model="language"
            :options="
              languages.map((l) => ({
                value: l.code,
                label: `${l.flag} ${l.name}`,
              }))
            "
            @update:model-value="saveLanguage"
          />
        </div>
        <div class="space-y-2">
          <Label for="locale">{{ $t('profile.dateFormat') }}</Label>
          <p class="text-sm text-muted-foreground">
            {{ $t('profile.dateFormatHint') }}
          </p>
          <FormSelect
            id="locale"
            data-testid="profile-locale"
            v-model="locale"
            :options="
              locales.map((l) => ({
                value: l.code,
                label: `${l.name} (${l.example})`,
              }))
            "
            @update:model-value="saveLocale"
          />
        </div>
      </CardContent>
    </Card>
    <Dialog v-model:open="showEmojiPicker">
      <DialogContent data-testid="profile-avatar-dialog">
        <DialogHeader>
          <DialogTitle>{{ $t('profile.chooseAvatar') }}</DialogTitle>
          <DialogDescription>{{ $t('profile.avatarHint') }}</DialogDescription>
        </DialogHeader>
        <div class="grid grid-cols-5 gap-2 sm:grid-cols-8">
          <Button
            :data-testid="`profile-avatar-option-${Array.from(emoji, (char) =>
              char.codePointAt(0).toString(16)
            ).join('-')}`"
            v-for="emoji in emojiList"
            :key="emoji"
            :variant="emoji === selectedEmoji ? 'secondary' : 'ghost'"
            class="size-10 text-2xl"
            :aria-label="emoji"
            :aria-pressed="emoji === selectedEmoji"
            @click="selectEmoji(emoji)"
          >
            {{ emoji }}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  </section>
</template>
<script>
import {
  Avatar,
  AvatarFallback,
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  FormSelect,
  FullPageSpinner,
  Label,
  Separator,
  ThemeToggle,
} from '@/components/ui'
import { useAuth } from '@/composables/useAuth'
import { authClient } from '@/lib/auth-client'
import { userAPI } from '@/services/userApi'
import { setLanguage } from '@/i18n'
import { LuMonitor, LuSmartphone, LuTablet, LuCircleHelp } from '@kalimahapps/vue-icons/lu'
import { FaBandsApple, FaBandsWindows, FaBandsLinux, FaBandsAndroid, FaBandsChrome } from '@kalimahapps/vue-icons/fa'

export default {
  name: 'ProfilePage',
  components: {
    Avatar,
    AvatarFallback,
    Button,
    Card,
    CardContent,
    CardHeader,
    CardTitle,
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    FormSelect,
    FullPageSpinner,
    Label,
    Separator,
    ThemeToggle,
    LuMonitor,
    LuSmartphone,
    LuTablet,
    LuCircleHelp,
    FaBandsApple,
    FaBandsWindows,
    FaBandsLinux,
    FaBandsAndroid,
    FaBandsChrome,
  },
  setup() {
    const { user, isLoading, currentSessionId, logout, error: authError } = useAuth()
    return { user, isLoading, currentSessionId, logout, authError }
  },
  computed: {
    loading() {
      return this.isLoading || !this.user
    },
    sessionDevices() {
      return this.sessions.map(session => ({ ...session, device: this.describeDevice(session.userAgent) }))
        .sort((a, b) => Number(b.id === this.currentSessionId) - Number(a.id === this.currentSessionId) ||
          (Date.parse(b.createdAt) || 0) - (Date.parse(a.createdAt) || 0))
    },
  },
  data() {
    return {
      showEmojiPicker: false,
      selectedEmoji: localStorage.getItem('user_avatar') || '😊',
      baseCurrency: 'USD',
      language: localStorage.getItem('language') || 'en',
      locale: localStorage.getItem('locale') || 'en-US',
      saving: false,
      sessions: [],
      sessionsLoading: false,
      sessionsError: '',
      sessionsRequest: 0,
      emojiList: [
        '😊',
        '😎',
        '🤩',
        '😇',
        '🥳',
        '😄',
        '🤗',
        '😏',
        '🧑‍💻',
        '👨‍💻',
        '👩‍💻',
        '🧑‍🎨',
        '👨‍🎨',
        '👩‍🎨',
        '🧑‍🚀',
        '👨‍🚀',
        '🦊',
        '🐱',
        '🐶',
        '🐼',
        '🐨',
        '🦁',
        '🐯',
        '🐸',
        '🌟',
        '⭐',
        '✨',
        '💫',
        '🔥',
        '💎',
        '🎯',
        '🚀',
        '🎨',
        '🎭',
        '🎪',
        '🎢',
        '🎡',
        '🏆',
        '🥇',
        '🎖️',
        '💜',
        '💙',
        '💚',
        '💛',
        '🧡',
        '❤️',
        '🖤',
        '🤍',
      ],
      currencies: [
        { code: 'USD', symbol: '$', name: 'US Dollar' },
        { code: 'EUR', symbol: '€', name: 'Euro' },
        { code: 'GBP', symbol: '£', name: 'British Pound' },
        { code: 'TRY', symbol: '₺', name: 'Turkish Lira' },
        { code: 'JPY', symbol: '¥', name: 'Japanese Yen' },
        { code: 'CNY', symbol: '¥', name: 'Chinese Yuan' },
        { code: 'CHF', symbol: 'Fr', name: 'Swiss Franc' },
        { code: 'CAD', symbol: '$', name: 'Canadian Dollar' },
        { code: 'AUD', symbol: '$', name: 'Australian Dollar' },
        { code: 'INR', symbol: '₹', name: 'Indian Rupee' },
      ],
      languages: [
        { code: 'en', name: 'English', flag: '🇺🇸' },
        { code: 'tr', name: 'Türkçe', flag: '🇹🇷' },
        { code: 'de', name: 'Deutsch', flag: '🇩🇪' },
        { code: 'fr', name: 'Français', flag: '🇫🇷' },
        { code: 'es', name: 'Español', flag: '🇪🇸' },
        { code: 'pt', name: 'Português', flag: '🇧🇷' },
        { code: 'ja', name: '日本語', flag: '🇯🇵' },
        { code: 'zh', name: '中文', flag: '🇨🇳' },
      ],
      locales: [
        { code: 'en-US', name: 'United States', example: '12/27/2025' },
        { code: 'en-GB', name: 'United Kingdom', example: '27/12/2025' },
        { code: 'tr-TR', name: 'Turkey', example: '27.12.2025' },
        { code: 'de-DE', name: 'Germany', example: '27.12.2025' },
        { code: 'fr-FR', name: 'France', example: '27/12/2025' },
        { code: 'es-ES', name: 'Spain', example: '27/12/2025' },
        { code: 'pt-BR', name: 'Brazil', example: '27/12/2025' },
        { code: 'ja-JP', name: 'Japan', example: '2025/12/27' },
        { code: 'zh-CN', name: 'China', example: '2025/12/27' },
      ],
    }
  },
  watch: {
    'user.id': {
      immediate: true,
      handler() {
        this.sessionsRequest += 1
        this.sessions = []
        this.sessionsError = ''
        this.sessionsLoading = false
        if (this.user) this.loadSessions()
      },
    },
    user: {
      immediate: true,
      handler(newUser) {
        if (newUser) {
          this.baseCurrency = newUser.base_currency || 'USD'
          this.locale = newUser.locale || 'en-US'
        }
      },
    },
  },
  beforeUnmount() {
    this.sessionsRequest += 1
  },
  methods: {
    async loadSessions() {
      if (!this.user || this.sessionsLoading) return
      const request = ++this.sessionsRequest
      this.sessionsLoading = true
      this.sessionsError = ''
      this.sessions = []
      try {
        const { data, error } = await authClient.listSessions()
        if (request !== this.sessionsRequest) return
        if (error) {
          this.sessionsError = error.code === 'SESSION_NOT_FRESH' ? 'reauthRequired' : 'loadError'
          return
        }
        if (!Array.isArray(data)) throw new Error('Invalid session response')
        // Retain only display fields, never session tokens or full auth responses.
        this.sessions = data.map(({ id, userAgent, ipAddress, createdAt, expiresAt }) => ({
          id, userAgent, ipAddress, createdAt, expiresAt,
        }))
      } catch {
        if (request === this.sessionsRequest) this.sessionsError = 'loadError'
      } finally {
        if (request === this.sessionsRequest) this.sessionsLoading = false
      }
    },
    describeDevice(userAgent) {
      const ua = typeof userAgent === 'string' ? userAgent : ''
      // User-Agent provides an approximate device description, not a hardware identity.
      const browser = [
        [/Edg(e|A|iOS)?\//, 'Edge'], [/OPR\/|Opera\//, 'Opera'],
        [/SamsungBrowser\//, 'Samsung Internet'], [/Firefox\/|FxiOS\//, 'Firefox'],
        [/Chrome\/|CriOS\//, 'Chrome'], [/Version\/.*Safari\//, 'Safari'],
      ].find(([pattern]) => pattern.test(ua))?.[1] || this.$t('profile.sessions.unknownBrowser')
      const os = [
        [/iPad/, 'iPadOS'], [/iPhone|iPod/, 'iOS'], [/Android/, 'Android'],
        [/Windows/, 'Windows'], [/CrOS/, 'ChromeOS'], [/Macintosh|Mac OS X/, 'macOS'], [/Linux/, 'Linux'],
      ].find(([pattern]) => pattern.test(ua))?.[1] || this.$t('profile.sessions.unknownOS')
      let kind = 'unknownDevice'
      if (/iPad|Tablet/i.test(ua) || (/Android/.test(ua) && !/Mobile/.test(ua))) kind = 'tablet'
      else if (/Mobile|iPhone|iPod/i.test(ua)) kind = 'mobile'
      else if (/Windows|Macintosh|Linux|CrOS/.test(ua)) kind = 'desktop'
      const osIcons = {
        macOS: 'FaBandsApple', iOS: 'FaBandsApple', iPadOS: 'FaBandsApple',
        Windows: 'FaBandsWindows', Linux: 'FaBandsLinux',
        Android: 'FaBandsAndroid', ChromeOS: 'FaBandsChrome',
      }
      const deviceIcons = { desktop: 'LuMonitor', mobile: 'LuSmartphone', tablet: 'LuTablet', unknownDevice: 'LuCircleHelp' }
      return { browser, os, kind, icon: osIcons[os] || deviceIcons[kind] }
    },
    formatSessionDate(value) {
      if (!value || Number.isNaN(new Date(value).getTime())) return '—'
      return new Date(value).toLocaleString(this.locale, { dateStyle: 'medium', timeStyle: 'short' })
    },
    async selectEmoji(emoji) {
      this.selectedEmoji = emoji
      localStorage.setItem('user_avatar', emoji)
      this.showEmojiPicker = false
    },
    async saveCurrency() {
      try {
        this.saving = true
        await userAPI.updateBaseCurrency(this.baseCurrency)
        localStorage.setItem('base_currency', this.baseCurrency)
      } catch (error) {
        console.error('Failed to save currency:', error)
      } finally {
        this.saving = false
      }
    },
    async saveLanguage() {
      const languageToLocale = {
        en: 'en-US',
        tr: 'tr-TR',
        de: 'de-DE',
        fr: 'fr-FR',
        es: 'es-ES',
        pt: 'pt-BR',
        ja: 'ja-JP',
        zh: 'zh-CN',
      }

      const newLocale = languageToLocale[this.language] || 'en-US'
      this.locale = newLocale

      // Update i18n locale
      setLanguage(this.language)

      try {
        await userAPI.updateLocale(newLocale)
        localStorage.setItem('language', this.language)
        localStorage.setItem('locale', newLocale)
      } catch (error) {
        console.error('Failed to save language/locale:', error)
      }
    },
    async saveLocale() {
      try {
        await userAPI.updateLocale(this.locale)
        localStorage.setItem('locale', this.locale)
      } catch (error) {
        console.error('Failed to save locale:', error)
      }
    },
    formatDate(dateString) {
      if (!dateString) return '-'
      const date = new Date(dateString)
      return date.toLocaleDateString(this.locale, {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
      })
    },
  },
}
</script>
