<template>
  <Sidebar collapsible="icon" class="workspace-sidebar">
    <SidebarHeader>
      <div class="flex items-center">
        <SidebarMenu class="min-w-0">
          <SidebarMenuItem>
            <SidebarMenuButton
              size="lg"
              class="workspace-brand group-data-[collapsible=icon]:justify-center"
              as-child
            >
              <router-link
                data-testid="sidebar-brand-link"
                to="/"
                class="font-semibold"
              >
                <span class="flex min-w-0 flex-col gap-2 group-data-[collapsible=icon]:hidden">
                  <BrandLogo class="workspace-brand-logo h-12! w-40! max-w-full" />
                  <span class="workspace-brand-version text-[11px] leading-none tracking-widest text-muted-foreground tabular-nums">v2026.10.30</span>
                </span>
                <BrandLogo compact class="workspace-brand-logo hidden group-data-[collapsible=icon]:inline-flex group-data-[collapsible=icon]:size-6" />
              </router-link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
        <Button
          data-testid="sidebar-close"
          variant="ghost"
          size="icon"
          class="mr-2 size-11 shrink-0 md:hidden"
          :aria-label="$t('common.closeMenu')"
          @click="setOpenMobile(false)"
        >
          <X class="size-4" />
        </Button>
      </div>
    </SidebarHeader>
    <SidebarContent data-testid="sidebar-navigation">
      <SidebarGroup>
        <SidebarGroupContent>
          <SidebarMenu>
            <SidebarMenuItem v-for="item in navigation" :key="item.path">
              <SidebarMenuButton
                class="min-h-11 rounded-xl group-data-[collapsible=icon]:min-h-8"
                as-child
                :tooltip="$t(item.label)"
              >
                <router-link
                  :data-testid="`sidebar-nav-${
                    item.path === '/' ? 'dashboard' : item.path.slice(1)
                  }`"
                  :to="item.path"
                >
                  <component :is="item.icon" />
                  <span>{{ $t(item.label) }}</span>
                </router-link>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarGroupContent>
      </SidebarGroup>
    </SidebarContent>
    <SidebarFooter>
      <p v-if="error" data-testid="sidebar-auth-error" role="alert" class="px-2 text-xs text-destructive">{{ $t(`auth.${error}`) }}</p>
      <SidebarMenu>
        <SidebarMenuItem>
          <DropdownMenu>
            <DropdownMenuTrigger as-child>
              <SidebarMenuButton
                data-testid="sidebar-user-menu-trigger"
                size="lg"
              >
                <Avatar class="size-8 rounded-lg">
                  <AvatarFallback class="rounded-lg">
                    {{ user?.email?.slice(0, 2).toUpperCase() }}
                  </AvatarFallback>
                </Avatar>
                <span class="min-w-0 flex-1 truncate">{{ user?.email }}</span>
                <ChevronsUpDown class="ml-auto size-4" />
              </SidebarMenuButton>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              data-testid="sidebar-user-menu"
              side="top"
              align="start"
              class="w-56 max-w-[calc(100vw-2rem)]"
            >
              <DropdownMenuLabel class="truncate">
                {{ user?.email }}
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem as-child>
                <router-link data-testid="sidebar-profile-link" to="/profile">
                  <User />
                  {{ $t('nav.profile') }}
                </router-link>
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem data-testid="sidebar-sign-out" @select="logout">
                <LogOut />
                {{ $t('common.signOut') }}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarFooter>
    <SidebarRail data-testid="sidebar-rail" />
  </Sidebar>
</template>
<script setup>
import {
  Button,
  Sidebar,
  SidebarHeader,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
  SidebarRail,
  useSidebar,
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuItem,
  Avatar,
  AvatarFallback,
} from '@/components/ui'
import { watch } from 'vue'
import { useRoute } from 'vue-router'
import {
  LayoutDashboard,
  CreditCard,
  Folder,
  Settings,
  User,
  LogOut,
  ChevronsUpDown,
  X,
} from 'lucide-vue-next'
import { useAuth } from '@/composables/useAuth'
import BrandLogo from '@/components/BrandLogo.vue'
const route = useRoute()
const { user, logout, error } = useAuth()
const { setOpenMobile } = useSidebar()
watch(
  () => route.fullPath,
  () => setOpenMobile(false)
)
const navigation = [
  { path: '/', label: 'nav.dashboard', icon: LayoutDashboard },
  { path: '/accounts', label: 'nav.accounts', icon: CreditCard },
  { path: '/wealth', label: 'wealth.title', icon: Folder },
  { path: '/projects', label: 'nav.projects', icon: Folder },
  { path: '/settings', label: 'nav.settings', icon: Settings },
]
</script>
