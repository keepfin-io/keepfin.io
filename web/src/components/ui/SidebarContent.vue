<template>
  <nav class="sidebar-nav">
    <ul class="space-y-1">
      <li>
        <router-link to="/accounts" :class="['sidebar-link', 'group', { 'is-collapsed': collapsed && !mobile }]"
          active-class="active">
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="flex-shrink-0">
            <rect x="2" y="5" width="20" height="14" rx="2"></rect>
            <line x1="2" y1="10" x2="22" y2="10"></line>
          </svg>
          <span v-if="!collapsed || mobile" class="sidebar-text">Accounts</span>
        </router-link>
      </li>
      <li>
        <router-link to="/" :class="['sidebar-link', 'group', { 'is-collapsed': collapsed && !mobile }]"
          active-class="active">
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="flex-shrink-0">
            <rect x="3" y="3" width="7" height="7"></rect>
            <rect x="14" y="3" width="7" height="7"></rect>
            <rect x="14" y="14" width="7" height="7"></rect>
            <rect x="3" y="14" width="7" height="7"></rect>
          </svg>
          <span v-if="!collapsed || mobile" class="sidebar-text">Dashboard</span>
        </router-link>
      </li>
      <li>
        <router-link to="/projects" :class="['sidebar-link', 'group', { 'is-collapsed': collapsed && !mobile }]"
          active-class="active">
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="flex-shrink-0">
            <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
          </svg>
          <span v-if="!collapsed || mobile" class="sidebar-text">Projects</span>
        </router-link>
      </li>
      <li>
        <router-link to="/settings" :class="['sidebar-link', 'group', { 'is-collapsed': collapsed && !mobile }]"
          active-class="active">
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="flex-shrink-0">
            <circle cx="12" cy="12" r="3"></circle>
            <path
              d="M12 1v6m0 6v6m5.196-15.196l-4.243 4.243m-2.828 2.828l-4.243 4.243m15.314 0l-4.243-4.243m-2.828-2.828l-4.243-4.243">
            </path>
          </svg>
          <span v-if="!collapsed || mobile" class="sidebar-text">Settings</span>
        </router-link>
      </li>
    </ul>
  </nav>
</template>

<script>
export default {
  name: "SidebarContent",
  inject: {
    isCollapsed: { default: () => () => false },
    isMobile: { default: () => () => false }
  },
  computed: {
    collapsed() {
      return this.isCollapsed();
    },
    mobile() {
      return this.isMobile();
    }
  }
};
</script>

<style scoped>
.sidebar-nav {
  width: 100%;
}

.sidebar-link {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.625rem 0.75rem;
  width: 100%;
  text-align: left;
  border-radius: 0.5rem;
  color: #4b5563;
  transition: all 0.2s;
  text-decoration: none;
  font-size: 0.875rem;
  font-weight: 500;
  position: relative;
}

.sidebar-link.is-collapsed {
  justify-content: center;
  padding: 0.625rem;
}

.sidebar-link:hover {
  background: #f3f4f6;
  color: #1f2937;
}

.sidebar-link.active {
  background: #eff6ff;
  color: #2563eb;
}

.sidebar-text {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  transition: opacity 0.2s, width 0.2s;
}
</style>
