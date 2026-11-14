import { computed, readonly, ref } from 'vue'

const savedTheme = localStorage.getItem('theme')
const theme = ref(savedTheme === 'light' ? 'light' : 'dark')

function setDocumentTheme(value) {
  theme.value = value
  document.documentElement.classList.toggle('dark', value === 'dark')
  document.documentElement.style.colorScheme = value
  localStorage.setItem('theme', value)
}

function prefersReducedMotion() {
  try {
    return window.matchMedia?.('(prefers-reduced-motion: reduce)').matches ?? false
  } catch {
    return false
  }
}

// Center of the control that requested the change, so the new theme can
// spread out from it.
export function themeOrigin(event) {
  const rect = event?.currentTarget?.getBoundingClientRect?.()
  return rect ? { x: rect.left + rect.width / 2, y: rect.top + rect.height / 2 } : null
}

// Applies a theme. With an origin, the new theme is revealed as a circle
// growing from that point where the browser supports view transitions.
export function applyTheme(value, origin = null) {
  if (value !== 'light' && value !== 'dark') return
  if (!origin || value === theme.value || !document.startViewTransition || prefersReducedMotion()) {
    setDocumentTheme(value)
    return
  }
  const { x, y } = origin
  const radius = Math.hypot(Math.max(x, window.innerWidth - x), Math.max(y, window.innerHeight - y))
  const transition = document.startViewTransition(() => setDocumentTheme(value))
  transition.ready.then(() => {
    document.documentElement.animate(
      { clipPath: [`circle(0px at ${x}px ${y}px)`, `circle(${radius}px at ${x}px ${y}px)`] },
      { duration: 500, easing: 'cubic-bezier(0.4, 0, 0.2, 1)', pseudoElement: '::view-transition-new(root)' },
    )
  }).catch(() => { /* The theme is already applied; only the animation was skipped. */ })
}

export function useTheme() {
  return {
    theme: readonly(theme),
    isDark: computed(() => theme.value === 'dark')
  }
}

applyTheme(theme.value)
