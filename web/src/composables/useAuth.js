import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { authClient } from '@/lib/auth-client'
import { userAPI } from '@/services/userApi'
import { applyTheme } from './useTheme'

const user = ref(null)
const authSession = ref(null)
const error = ref(null)
const isLoading = ref(false)
const isSavingTheme = ref(false)
let themeRevision = 0
let sessionRevision = 0
let pendingSession
sessionStorage.removeItem('keepnode_user')

function setUser(value) {
  user.value = value
  if (value) applyTheme(value.theme)
}

export function refreshSession() {
  if (pendingSession) return pendingSession
  const revision = themeRevision
  const sessionVersion = sessionRevision
  pendingSession = (async () => {
    error.value = null
    const result = await authClient.getSession({ query: { disableCookieCache: true } })
    if (result.error) throw new Error('sessionUnavailable')
    if (sessionVersion !== sessionRevision) return false
    authSession.value = result.data
    if (!result.data) { setUser(null); return false }
    const data = await userAPI.getMe()
    if (sessionVersion !== sessionRevision) return false
    if (!data.user || data.error !== false) throw new Error('sessionUnavailable')
    if (revision !== themeRevision && data.user.id === user.value?.id) data.user.theme = user.value.theme
    setUser(data.user)
    return true
  })().catch((err) => {
    if (sessionVersion !== sessionRevision) return false
    setUser(null)
    if (err.status === 401) { authSession.value = null; return false }
    error.value = err.status === 409 ? 'migrationRequired' : err.status === 403 ? 'accountUnavailable' : 'sessionUnavailable'
    throw err
  }).finally(() => { pendingSession = null })
  return pendingSession
}

export function useAuth() {
  const router = useRouter()
  const setTheme = async (theme) => {
    if (!user.value || isSavingTheme.value || !['light', 'dark'].includes(theme)) return
    const userID = user.value.id
    isSavingTheme.value = true
    try {
      const data = await userAPI.updateTheme(theme)
      if (data.error !== false) throw new Error('Failed to save theme')
      if (user.value?.id === userID) {
        themeRevision += 1
        setUser({ ...user.value, theme })
      }
    } finally { isSavingTheme.value = false }
  }
  const logout = async () => {
    if (isLoading.value) return
    isLoading.value = true
    error.value = null
    try {
      const result = await authClient.signOut()
      if (result.error) throw new Error('signOutFailed')
      sessionRevision += 1
      setUser(null)
      authSession.value = null
      await router.replace('/login')
    } catch { error.value = 'signOutFailed' }
    finally { isLoading.value = false }
  }
  return {
    user: computed(() => user.value),
    hasSession: computed(() => !!authSession.value),
    isAuthenticated: computed(() => !!user.value),
    isLoading: computed(() => isLoading.value),
    error: computed(() => error.value),
    isSavingTheme: computed(() => isSavingTheme.value),
    setTheme, logout, checkAuth: refreshSession,
  }
}
