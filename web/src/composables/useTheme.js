import { computed, readonly, ref } from 'vue'

const savedTheme = localStorage.getItem('theme')
const theme = ref(savedTheme === 'light' ? 'light' : 'dark')

export function applyTheme(value) {
  if (value !== 'light' && value !== 'dark') return

  theme.value = value
  document.documentElement.classList.toggle('dark', value === 'dark')
  document.documentElement.style.colorScheme = value
  localStorage.setItem('theme', value)
}

export function useTheme() {
  return {
    theme: readonly(theme),
    isDark: computed(() => theme.value === 'dark')
  }
}

applyTheme(theme.value)
