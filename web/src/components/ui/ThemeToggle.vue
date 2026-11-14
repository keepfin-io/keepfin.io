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
      @click="saveTheme(isDark ? 'light' : 'dark', $event)"
    >
      <Transition
        mode="out-in"
        enter-active-class="transition duration-300 ease-out motion-reduce:transition-none"
        enter-from-class="-rotate-90 scale-50 opacity-0"
        leave-active-class="transition duration-150 ease-in motion-reduce:transition-none"
        leave-to-class="rotate-90 scale-50 opacity-0"
      >
        <Sun v-if="isDark" key="sun" class="size-4" />
        <Moon v-else key="moon" class="size-4" />
      </Transition>
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
        @click="saveTheme(option, $event)"
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
import { applyTheme, themeOrigin, useTheme } from '@/composables/useTheme'
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

    // Apply at once so the switch animates from the button; restore the
    // previous theme if the preference cannot be saved.
    async function saveTheme(value, event) {
      saveFailed.value = false
      const previous = theme.value
      applyTheme(value, themeOrigin(event))
      try {
        await setTheme(value)
      } catch {
        applyTheme(previous)
        saveFailed.value = true
      }
    }

    return { theme, isDark, isSavingTheme, saveFailed, toggleLabel, saveTheme }
  },
}
</script>
