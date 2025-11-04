<template>
  <div class="flex min-h-screen bg-gray-50">
    <!-- Material Icons font link -->
    <link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet">
    <template v-if="!isAuthPage">
      <Sidebar>
        <template #header>
          <SidebarHeader />
        </template>
        <template #content>
          <SidebarContent />
        </template>
        <template #footer>
          <SidebarFooter>
            <SidebarMenu>
              <SidebarMenuItem>
                <DropdownMenu>
                  <template #trigger>
                    <button class="flex items-center gap-2 w-full px-4 py-2 rounded hover:bg-gray-100">
                      <span class="material-icons">person</span>
                      <span>Username</span>
                      <span class="material-icons ml-auto">expand_more</span>
                    </button>
                  </template>
                  <template #content>
                    <button class="block w-full text-left px-4 py-2 hover:bg-gray-100">Account</button>
                    <button class="block w-full text-left px-4 py-2 hover:bg-gray-100">Billing</button>
                    <button class="block w-full text-left px-4 py-2 hover:bg-gray-100">Sign out</button>
                  </template>
                </DropdownMenu>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarFooter>
        </template>
      </Sidebar>
    </template>
    <div class="flex-1">
      <router-view />
    </div>
  </div>
</template>

<script>
import Sidebar from './components/ui/Sidebar.vue'
import SidebarHeader from './components/ui/SidebarHeader.vue'
import SidebarContent from './components/ui/SidebarContent.vue'
import SidebarFooter from './components/ui/SidebarFooter.vue'
import SidebarMenu from './components/ui/SidebarMenu.vue'
import SidebarMenuItem from './components/ui/SidebarMenuItem.vue'
import DropdownMenu from './components/ui/DropdownMenu.vue'

export default {
  name: 'App',
  components: {
    Sidebar,
    SidebarHeader,
    SidebarContent,
    SidebarFooter,
    SidebarMenu,
    SidebarMenuItem,
    DropdownMenu
  },
  computed: {
    isAuthPage() {
      const authPages = ['/login', '/register'];
      return authPages.includes(this.$route.path);
    }
  }
}
</script>