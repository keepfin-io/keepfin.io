<template>
  <div
    data-testid="auth-layout"
    v-if="isAuthPage"
    class="auth-shell bg-background text-foreground"
  >
    <div data-testid="auth-content" class="auth-content" tabindex="0" role="region" :aria-label="$t('common.mainContent')">
      <router-view />
    </div>
    <AppFooter />
  </div>
  <SidebarProvider v-else-if="isAuthenticated && route.meta.requiresAuth" class="workspace-shell">
    <AppSidebar />
    <SidebarInset data-testid="app-layout" class="workspace-main min-w-0">
      <header
        data-testid="app-header"
        class="workspace-header sticky top-0 z-20 flex h-16 shrink-0 items-center gap-3 border-b px-4 backdrop-blur sm:px-6"
      >
        <SidebarTrigger
          data-testid="sidebar-toggle"
          class="size-11 rounded-xl"
          :aria-label="$t('common.toggleMenu')"
        />
        <Separator orientation="vertical" class="h-5!" />
        <router-link to="/" data-testid="header-brand-link" aria-label="Keepfin">
          <BrandLogo compact />
        </router-link>
        <span
          data-testid="app-page-title"
          class="min-w-0 flex-1 truncate text-sm font-medium"
        >
          {{ pageTitle }}
        </span>
        <ZenToggle />
        <ThemeToggle data-testid="header-theme" class="workspace-theme" compact />
      </header>
      <div
        data-testid="app-content"
        class="workspace-content w-full min-w-0"
        tabindex="0"
        role="region"
        :aria-label="pageTitle"
      >
        <div class="mx-auto w-full min-w-0 max-w-7xl p-4 sm:p-5 lg:p-8">
          <router-view :key="$route.params.id || $route.name" />
        </div>
      </div>
      <AppFooter />
    </SidebarInset>
  </SidebarProvider>
</template>
<script setup>
import {
  SidebarProvider,
  SidebarInset,
  SidebarTrigger,
  Separator,
  ThemeToggle,
} from '@/components/ui'
import { computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ZenToggle from '@/components/wealth/ZenToggle.vue'
import { loadZen } from '@/composables/useZen'
import AppSidebar from '@/components/AppSidebar.vue'
import AppFooter from '@/components/AppFooter.vue'
import BrandLogo from '@/components/BrandLogo.vue'
import { useAuth } from '@/composables/useAuth'
import '@/styles/workspace.css'
const route = useRoute()
const router = useRouter()
const { checkAuth, isAuthenticated, user } = useAuth()
watch(() => user.value?.id, id => { loadZen(id).catch(() => {}) }, { immediate: true })
async function revalidateSession() {
  if (document.hidden || !route.meta.requiresAuth) return
  try { if (!await checkAuth()) await router.replace('/login') }
  catch { await router.replace('/login') }
}
onMounted(() => document.addEventListener('visibilitychange', revalidateSession))
onUnmounted(() => document.removeEventListener('visibilitychange', revalidateSession))
const isAuthPage = computed(() => ['/login', '/register', '/auth/callback', '/shared'].includes(route.path))
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
const pageTitle = computed(() =>
  t(
    'nav.' +
      ({
        Dashboard: 'dashboard',
        Accounts: 'accounts',
        AccountCreate: 'accounts',
        AccountEdit: 'accounts',
        HoldingCreate: 'accounts',
        HoldingEdit: 'accounts',
        ActivityCreate: 'accounts',
        ActivityEdit: 'accounts',
        AccountTransfer: 'accounts',
        AccountDetail: 'accounts',
        AccountAnalysis: 'accounts',
        HoldingDetail: 'accounts',
        Profile: 'profile',
        Projects: 'projects',
        Settings: 'settings',
      }[route.name] || 'dashboard')
  )
)
</script>
