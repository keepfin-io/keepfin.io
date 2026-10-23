<template>
  <Sidebar collapsible="icon">
    <SidebarHeader>
      <div class="flex items-center">
        <SidebarMenu class="min-w-0">
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" as-child>
              <router-link to="/" class="font-semibold">
                <Wallet class="size-5" />
                <span>Keepfin</span>
              </router-link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
        <Button
          variant="ghost"
          size="icon"
          class="mr-2 shrink-0 md:hidden"
          :aria-label="$t('common.closeMenu')"
          @click="setOpenMobile(false)"
        >
          <X class="size-4" />
        </Button>
      </div>
    </SidebarHeader>
    <SidebarContent>
      <SidebarGroup>
        <SidebarGroupContent>
          <SidebarMenu>
            <SidebarMenuItem v-for="item in navigation" :key="item.path">
              <SidebarMenuButton
                as-child
                :is-active="
                  item.path === '/'
                    ? route.path === '/'
                    : route.path.startsWith(item.path)
                "
                :tooltip="$t(item.label)"
              >
                <router-link :to="item.path">
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
      <SidebarMenu>
        <SidebarMenuItem>
          <DropdownMenu>
            <DropdownMenuTrigger as-child>
              <SidebarMenuButton size="lg">
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
              side="top"
              align="start"
              class="w-56 max-w-[calc(100vw-2rem)]"
            >
              <DropdownMenuLabel class="truncate">
                {{ user?.email }}
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem as-child>
                <router-link to="/profile">
                  <User />
                  {{ $t('nav.profile') }}
                </router-link>
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem @select="logout">
                <LogOut />
                {{ $t('common.signOut') }}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarFooter>
    <SidebarRail />
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
  Wallet,
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
const route = useRoute()
const { user, logout } = useAuth()
const { setOpenMobile } = useSidebar()
watch(
  () => route.fullPath,
  () => setOpenMobile(false)
)
const navigation = [
  { path: '/', label: 'nav.dashboard', icon: LayoutDashboard },
  { path: '/accounts', label: 'nav.accounts', icon: CreditCard },
  { path: '/projects', label: 'nav.projects', icon: Folder },
  { path: '/settings', label: 'nav.settings', icon: Settings },
  { path: '/profile', label: 'nav.profile', icon: User },
]
</script>
