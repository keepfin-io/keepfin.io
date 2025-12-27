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
            <DropdownMenu>
              <template #trigger>
                <button class="user-button">
                  <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="flex-shrink-0">
                    <path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"></path>
                    <circle cx="12" cy="7" r="4"></circle>
                  </svg>
                  <div class="user-info">
                    <span class="user-text">{{ user?.email || 'User' }}</span>
                    <span class="build-version">v2025.12.27</span>
                  </div>
                  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="chevron-icon">
                    <polyline points="6 9 12 15 18 9"></polyline>
                  </svg>
                </button>
              </template>
              <template #content>
                <button class="dropdown-item">
                  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"></path>
                    <circle cx="12" cy="7" r="4"></circle>
                  </svg>
                  Account
                </button>
                <button class="dropdown-item">
                  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <rect x="2" y="5" width="20" height="14" rx="2"></rect>
                    <line x1="2" y1="10" x2="22" y2="10"></line>
                  </svg>
                  Billing
                </button>
                <button @click="handleLogout" class="dropdown-item text-red-600 hover:bg-red-50">
                  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
                    <polyline points="16 17 21 12 16 7"></polyline>
                    <line x1="21" y1="12" x2="9" y2="12"></line>
                  </svg>
                  Sign out
                </button>
              </template>
            </DropdownMenu>
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
import DropdownMenu from './components/ui/DropdownMenu.vue'
import { useAuth } from './composables/useAuth'

export default {
  name: 'App',
  components: {
    Sidebar,
    SidebarHeader,
    SidebarContent,
    SidebarFooter,
    DropdownMenu
  },
  setup() {
    const { user, logout, init } = useAuth()

    // Initialize auth check on app mount
    init()

    return { user, logout }
  },
  computed: {
    isAuthPage() {
      const authPages = ['/login', '/register'];
      return authPages.includes(this.$route.path);
    }
  },
  methods: {
    async handleLogout() {
      await this.logout()
    }
  }
}
</script>

<style>
.user-button {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  width: 100%;
  padding: 0.625rem 0.75rem;
  border-radius: 0.5rem;
  transition: all 0.2s;
  background: transparent;
  border: none;
  cursor: pointer;
  color: #4b5563;
  font-size: 0.875rem;
  font-weight: 500;
}

.user-button:hover {
  background: #f3f4f6;
}

.user-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  text-align: left;
}

.user-text {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 0.875rem;
}

.build-version {
  font-size: 0.625rem;
  color: #9ca3af;
  font-weight: 400;
  letter-spacing: 0.02em;
  margin-top: 0.125rem;
}

.chevron-icon {
  transition: transform 0.2s;
  flex-shrink: 0;
}

/* Collapsed state için */
.is-collapsed .user-info,
.is-collapsed .chevron-icon {
  display: none;
}

.is-collapsed .user-button {
  justify-content: center;
  padding: 0.625rem;
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  width: 100%;
  padding: 0.625rem 0.875rem;
  text-align: left;
  background: transparent;
  border: none;
  cursor: pointer;
  transition: background 0.2s;
  font-size: 0.875rem;
  color: #374151;
}

.dropdown-item:hover {
  background: #f3f4f6;
}
</style>