import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { authAPI } from '../services/authApi'
import { userAPI } from '../services/userApi'
import { applyTheme } from './useTheme'

// Load user from sessionStorage on init
const storedUser = sessionStorage.getItem('keepnode_user')
const user = ref(storedUser ? JSON.parse(storedUser) : null)
const isLoading = ref(false)
const error = ref(null)
const isSavingTheme = ref(false)
let themeRevision = 0

if (user.value) applyTheme(user.value.theme)

export function useAuth() {
  const router = useRouter()

  const isAuthenticated = computed(() => !!user.value)

  const setUser = (userData) => {
    user.value = userData
    if (userData) {
      sessionStorage.setItem('keepnode_user', JSON.stringify(userData))
      applyTheme(userData.theme)
    } else {
      sessionStorage.removeItem('keepnode_user')
    }
  }

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
    } finally {
      isSavingTheme.value = false
    }
  }

  const login = async (email, password) => {
    isLoading.value = true
    error.value = null

    try {
      const data = await authAPI.signin(email, password)

      if (data.error === false) {
        setUser(data.user)
        router.push('/')
        return { success: true }
      } else {
        error.value = data.message || 'Login failed'
        return { success: false, error: error.value }
      }
    } catch (err) {
      error.value = err.message || 'An error occurred during login'
      return { success: false, error: error.value }
    } finally {
      isLoading.value = false
    }
  }

  const register = async (email, password) => {
    isLoading.value = true
    error.value = null

    try {
      const data = await authAPI.signup(email, password)

      if (data.error === false) {
        // After successful registration, log in automatically
        return await login(email, password)
      } else {
        error.value = data.message || 'Registration failed'
        return { success: false, error: error.value }
      }
    } catch (err) {
      error.value = err.message || 'An error occurred during registration'
      return { success: false, error: error.value }
    } finally {
      isLoading.value = false
    }
  }

  const logout = async () => {
    try {
      await authAPI.signout()
    } catch (err) {
      console.error('Logout error:', err)
    } finally {
      setUser(null)
      router.push('/login')
    }
  }

  const checkAuth = async () => {
    const revision = themeRevision
    try {
      const data = await authAPI.checkAuth()
      if (data.error === false && data.user) {
        // A session check started before a theme save must not restore the old theme.
        if (revision !== themeRevision && data.user.id === user.value?.id) {
          data.user.theme = user.value.theme
        }
        setUser(data.user)
        return true
      } else {
        setUser(null)
        return false
      }
    } catch (err) {
      // Cookie invalid or expired
      setUser(null)
      if (window.location.pathname !== '/login' && window.location.pathname !== '/register') {
        router.push('/login')
      }
      return false
    }
  }

  // Initialize auth check on mount
  const init = async () => {
    if (user.value) {
      // Verify the session is still valid
      await checkAuth()
    }
  }

  // Visibility API - check auth when tab becomes visible
  const handleVisibilityChange = () => {
    if (!document.hidden && user.value) {
      checkAuth()
    }
  }
  onMounted(() => document.addEventListener('visibilitychange', handleVisibilityChange))
  onUnmounted(() => document.removeEventListener('visibilitychange', handleVisibilityChange))

  return {
    user: computed(() => user.value),
    isAuthenticated,
    isLoading: computed(() => isLoading.value),
    error: computed(() => error.value),
    isSavingTheme: computed(() => isSavingTheme.value),
    setTheme,
    login,
    register,
    logout,
    checkAuth,
    init
  }
}
