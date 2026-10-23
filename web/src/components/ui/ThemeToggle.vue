<template>
  <div class="flex min-w-0 flex-col gap-2">
    <Button
      :data-testid="$attrs['data-testid'] && `${$attrs['data-testid']}-toggle`"
      v-if="compact"
      variant="ghost"
      size="icon"
      :disabled="isSavingTheme"
      :aria-label="toggleLabel"
      :title="toggleLabel"
      @click="saveTheme(isDark ? 'light' : 'dark')"
    >
      <Sun v-if="isDark" class="size-4" />
      <Moon v-else class="size-4" />
    </Button>
    <div
      v-else
      class="grid grid-cols-2 gap-2"
      role="group"
      :aria-label="$t('profile.theme')"
    >
      <Button
        :data-testid="
          $attrs['data-testid'] && `${$attrs['data-testid']}-${option}`
        "
        v-for="option in ['light', 'dark']"
        :key="option"
        :variant="theme === option ? 'default' : 'outline'"
        :aria-pressed="theme === option"
        :disabled="isSavingTheme"
        @click="saveTheme(option)"
      >
        <Sun v-if="option === 'light'" class="size-4" />
        <Moon v-else class="size-4" />
        {{
          $t(option === 'light' ? 'profile.themeLight' : 'profile.themeDark')
        }}
      </Button>
    </div>
    <Alert
      :data-testid="$attrs['data-testid'] && `${$attrs['data-testid']}-error`"
      v-if="saveFailed"
      :class="
        compact && 'fixed right-4 top-16 z-50 w-[calc(100vw-2rem)] max-w-sm'
      "
      variant="error"
      :description="$t('profile.themeSaveError')"
      dismissible
      @dismiss="saveFailed = false"
    />
  </div>
</template>
<script>
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Sun, Moon } from 'lucide-vue-next'
import { useAuth } from '@/composables/useAuth'
import { useTheme } from '@/composables/useTheme'
import Alert from './FeedbackAlert.vue'
import { Button } from '@/components/ui/button'

export default {
  name: 'ThemeToggle',
  components: { Sun, Moon, Alert, Button },
  props: { compact: { type: Boolean, default: false } },
  setup() {
    const { t } = useI18n()
    const { theme, isDark } = useTheme()
    const { setTheme, isSavingTheme } = useAuth()
    const saveFailed = ref(false)
    const toggleLabel = computed(() =>
      t(isDark.value ? 'profile.switchToLight' : 'profile.switchToDark')
    )

    async function saveTheme(value) {
      saveFailed.value = false
      try {
        await setTheme(value)
      } catch {
        saveFailed.value = true
      }
    }

    return { theme, isDark, isSavingTheme, saveFailed, toggleLabel, saveTheme }
  },
}
</script>
