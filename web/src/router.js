import { createRouter, createWebHistory } from 'vue-router'
import Login from './pages/Login.vue'
import Register from './pages/Register.vue'
import Dashboard from './pages/Dashboard.vue'
import Accounts from './pages/Accounts.vue'
import AccountDetail from './pages/AccountDetail.vue'
import Projects from './pages/Projects.vue'
import Settings from './pages/Settings.vue'
import Profile from './pages/Profile.vue'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { requiresGuest: true }
  },
  {
    path: '/register',
    name: 'Register',
    component: Register,
    meta: { requiresGuest: true }
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

// Route guard
router.beforeEach((to, from, next) => {
  // Check if user is authenticated by checking if user data exists
  // Since we're using HTTP-only cookies, we can't directly check the cookie
  // We'll rely on the API to validate authentication
  const userDataString = sessionStorage.getItem('keepnode_user')
  const isAuthenticated = !!userDataString

  // If route requires authentication and user is not authenticated
  if (to.meta.requiresAuth && !isAuthenticated) {
    next('/login')
  }
  // If route requires guest (login/register) and user is authenticated
  else if (to.meta.requiresGuest && isAuthenticated) {
    next('/')
  }
  // Otherwise, proceed
  else {
    next()
  }
})

export default router
