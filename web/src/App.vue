<template>
  <div v-if="isAuthPage" class="min-h-svh bg-background text-foreground">
    <router-view />
  </div>
  <SidebarProvider v-else>
    <AppSidebar />
    <SidebarInset class="min-w-0">
      <header
        class="sticky top-0 z-20 flex h-14 shrink-0 items-center gap-3 border-b bg-background/95 px-4 backdrop-blur sm:px-6"
      >
        <SidebarTrigger :aria-label="$t('common.toggleMenu')" />
        <Separator orientation="vertical" class="h-5!" />
        <span class="min-w-0 flex-1 truncate text-sm font-medium">
          {{ pageTitle }}
        </span>
        <ThemeToggle compact />
      </header>
      <div class="mx-auto w-full min-w-0 max-w-7xl p-4 sm:p-6 lg:p-8">
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
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import AppSidebar from '@/components/AppSidebar.vue'
import { useAuth } from '@/composables/useAuth'
const route = useRoute()
const { init } = useAuth()
init()
const isAuthPage = computed(() => ['/login', '/register'].includes(route.path))
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
