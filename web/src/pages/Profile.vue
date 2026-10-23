<template>
  <FullPageSpinner :visible="loading" :message="$t('common.loading')" />
  <section class="mx-auto min-w-0 max-w-3xl space-y-6">
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
            <dd class="mt-1 break-all font-mono text-sm">
              {{ user?.id || '-' }}
            </dd>
          </div>
          <div class="min-w-0">
            <dt class="text-sm text-muted-foreground">
              {{ $t('profile.email') }}
            </dt>
            <dd class="mt-1 break-all text-sm font-medium">
              {{ user?.email || '-' }}
            </dd>
          </div>
          <div>
            <dt class="text-sm text-muted-foreground">
              {{ $t('profile.memberSince') }}
            </dt>
            <dd class="mt-1 text-sm font-medium">
              {{ formatDate(user?.created_at) }}
            </dd>
          </div>
        </dl>
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
          <ThemeToggle />
        </div>
        <Separator />
        <div class="space-y-2">
          <Label for="baseCurrency">{{ $t('profile.baseCurrency') }}</Label>
          <p class="text-sm text-muted-foreground">
            {{ $t('profile.baseCurrencyHint') }}
          </p>
          <FormSelect
            id="baseCurrency"
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
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ $t('profile.chooseAvatar') }}</DialogTitle>
          <DialogDescription>{{ $t('profile.avatarHint') }}</DialogDescription>
        </DialogHeader>
        <div class="grid grid-cols-5 gap-2 sm:grid-cols-8">
          <Button
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
import { userAPI } from '@/services/userApi'
import { setLanguage } from '@/i18n'

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
  },
  setup() {
    const { user, isLoading } = useAuth()
    return { user, isLoading }
  },
  computed: {
    loading() {
      return this.isLoading || !this.user
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
  methods: {
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
