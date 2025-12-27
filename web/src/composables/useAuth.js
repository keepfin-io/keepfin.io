import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { authAPI } from '../services/authApi'

// Load user from sessionStorage on init
const storedUser = sessionStorage.getItem('keepnode_user')
const user = ref(storedUser ? JSON.parse(storedUser) : null)
const isLoading = ref(false)
const error = ref(null)

export function useAuth() {
  const router = useRouter()

  const isAuthenticated = computed(() => !!user.value)

  const setUser = (userData) => {
    user.value = userData
    if (userData) {
      sessionStorage.setItem('keepnode_user', JSON.stringify(userData))
    } else {
      sessionStorage.removeItem('keepnode_user')
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
    try {
      const data = await authAPI.checkAuth()
      if (data.error === false && data.user) {
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
  if (typeof document !== 'undefined') {
    document.addEventListener('visibilitychange', () => {
      if (!document.hidden && user.value) {
        checkAuth()
      }
    })
  }

  return {
    user: computed(() => user.value),
    isAuthenticated,
    isLoading: computed(() => isLoading.value),
    error: computed(() => error.value),
    login,
    register,
    logout,
    checkAuth,
    init
  }
}
