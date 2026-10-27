import { createRouter, createWebHistory } from 'vue-router'
import Login from './pages/Login.vue'
import AuthCallback from './pages/AuthCallback.vue'
import { refreshSession } from './composables/useAuth'
import Dashboard from './pages/Dashboard.vue'
import Accounts from './pages/Accounts.vue'
import AccountDetail from './pages/AccountDetail.vue'
import Projects from './pages/Projects.vue'
import Settings from './pages/Settings.vue'
import Profile from './pages/Profile.vue'

const routes = [
  { path: '/auth/callback', name: 'AuthCallback', component: AuthCallback },
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { requiresGuest: true }
  },
  {
    path: '/register',
    redirect: '/login'
  },
  {
    path: '/',
    name: 'Dashboard',
    component: Dashboard,
    meta: { requiresAuth: true }
  },
  {
    path: '/accounts',
    name: 'Accounts',
    component: Accounts,
    meta: { requiresAuth: true }
  },
  {
    path: '/accounts/:id',
    name: 'AccountDetail',
    component: AccountDetail,
    meta: { requiresAuth: true }
  },
  {
    path: '/projects',
    name: 'Projects',
    component: Projects,
    meta: { requiresAuth: true }
  },
  {
    path: '/settings',
    name: 'Settings',
    component: Settings,
    meta: { requiresAuth: true }
  },
  {
    path: '/profile',
    name: 'Profile',
    component: Profile,
    meta: { requiresAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// A cached browser profile must never authorize a route.
router.beforeEach(async (to) => {
  if (!to.meta.requiresAuth && !to.meta.requiresGuest) return true
  try {
    const authenticated = await refreshSession()
    if (to.meta.requiresAuth && !authenticated) return '/login'
    if (to.meta.requiresGuest && authenticated) return '/'
  } catch {
    if (to.meta.requiresAuth) return '/login'
  }
  return true
})

export default router
