<template>
  <div
    data-testid="auth-layout"
    v-if="isAuthPage"
    class="min-h-svh bg-background text-foreground"
  >
    <router-view />
  </div>
  <SidebarProvider v-else>
    <AppSidebar />
    <SidebarInset data-testid="app-layout" class="min-w-0">
      <header
        data-testid="app-header"
        class="sticky top-0 z-20 flex h-14 shrink-0 items-center gap-3 border-b bg-background/95 px-4 backdrop-blur sm:px-6"
      >
        <SidebarTrigger
          data-testid="sidebar-toggle"
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
        <ThemeToggle data-testid="header-theme" compact />
      </header>
      <div
        data-testid="app-content"
        class="mx-auto w-full min-w-0 max-w-7xl p-4 sm:p-6 lg:p-8"
      >
        <router-view />
      </div>
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
import { computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppSidebar from '@/components/AppSidebar.vue'
import BrandLogo from '@/components/BrandLogo.vue'
import { useAuth } from '@/composables/useAuth'
const route = useRoute()
const router = useRouter()
const { checkAuth } = useAuth()
async function revalidateSession() {
  if (document.hidden || !route.meta.requiresAuth) return
  try { if (!await checkAuth()) await router.replace('/login') }
  catch { await router.replace('/login') }
}
onMounted(() => document.addEventListener('visibilitychange', revalidateSession))
onUnmounted(() => document.removeEventListener('visibilitychange', revalidateSession))
const isAuthPage = computed(() => ['/login', '/register', '/auth/callback'].includes(route.path))
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
const pageTitle = computed(() =>
  t(
    'nav.' +
      ({
        Dashboard: 'dashboard',
        Accounts: 'accounts',
        AccountDetail: 'accounts',
        Profile: 'profile',
        Projects: 'projects',
        Settings: 'settings',
      }[route.name] || 'dashboard')
  )
)
</script>
