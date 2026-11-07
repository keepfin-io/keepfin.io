import { createRouter, createWebHistory } from 'vue-router'
import Login from './pages/Login.vue'
import AuthCallback from './pages/AuthCallback.vue'
import { refreshSession } from './composables/useAuth'
import Dashboard from './pages/Dashboard.vue'
import Accounts from './pages/Accounts.vue'
import AccountForm from './pages/AccountForm.vue'
import AccountAction from './pages/AccountAction.vue'
import AccountDetail from './pages/AccountDetail.vue'
import AccountAnalysis from './pages/AccountAnalysis.vue'
import HoldingDetail from './pages/HoldingDetail.vue'
import Projects from './pages/Projects.vue'
import Settings from './pages/Settings.vue'
import Profile from './pages/Profile.vue'

import WealthLayout from './pages/wealth/WealthLayout.vue'
import Watchlist from './pages/wealth/Watchlist.vue'
import ManualAssets from './pages/wealth/ManualAssets.vue'
import NetWorth from './pages/wealth/NetWorth.vue'
import Activities from './pages/wealth/Activities.vue'
import Planning from './pages/wealth/Planning.vue'
import Sharing from './pages/wealth/Sharing.vue'
import MarketData from './pages/wealth/MarketData.vue'
import SharedPortfolio from './pages/wealth/SharedPortfolio.vue'

const routes = [
 {path:'/shared',name:'SharedPortfolio',component:SharedPortfolio},
 {path:'/wealth',component:WealthLayout,meta:{requiresAuth:true},children:[
  {path:'',redirect:'/wealth/watchlist'},
  {path:'watchlist',name:'Watchlist',component:Watchlist},
  {path:'assets',name:'ManualAssets',component:ManualAssets},
  {path:'net-worth',name:'NetWorth',component:NetWorth},
  {path:'activities',name:'Activities',component:Activities},
  {path:'risk',name:'Risk',component:Planning,props:{mode:'risk'}},
  {path:'fire',name:'Fire',component:Planning,props:{mode:'fire'}},
  {path:'sharing',name:'Sharing',component:Sharing},
  {path:'data',name:'MarketData',component:MarketData},
 ]},
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
    path: '/accounts/new',
    name: 'AccountCreate',
    component: AccountForm,
    meta: { requiresAuth: true }
  },
  {
    path: '/accounts/:id',
    name: 'AccountDetail',
    component: AccountDetail,
    meta: { requiresAuth: true },
    children: [
      { path: 'analysis', name: 'AccountAnalysis', component: AccountAnalysis, meta: { accountAction: 'account-analysis' } },
      { path: 'holdings/:holdingId', name: 'HoldingDetail', component: HoldingDetail, meta: { accountAction: 'holding-detail' } },
      { path: 'edit', name: 'AccountEdit', component: AccountAction, meta: { accountAction: 'account-edit' } },
      { path: 'holdings/new', name: 'HoldingCreate', component: AccountAction, meta: { accountAction: 'holding-create' } },
      { path: 'holdings/:holdingId/edit', name: 'HoldingEdit', component: AccountAction, meta: { accountAction: 'holding-edit' } },
      { path: 'activities/new', name: 'ActivityCreate', component: AccountAction, meta: { accountAction: 'activity-create' } },
      { path: 'activities/:activityId/edit', name: 'ActivityEdit', component: AccountAction, meta: { accountAction: 'activity-edit' } },
      { path: 'transfer', name: 'AccountTransfer', component: AccountAction, meta: { accountAction: 'transfer' } },
    ]
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
