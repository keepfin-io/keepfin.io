<template>
  <div class="sidebar-container">
    <!-- Mobile overlay -->
    <transition name="overlay">
      <div v-if="isMobileOpen" class="fixed inset-0 bg-black/50 z-40 md:hidden" @click="closeMobile" />
    </transition>

    <!-- Sidebar -->
    <aside :class="[
      'sidebar',
      {
        'collapsed': isCollapsed && !isMobile,
        'mobile-open': isMobileOpen,
        'mobile-closed': !isMobileOpen && isMobile
      }
    ]">
      <!-- Header -->
      <div class="sidebar-header">
        <slot name="header">
          <div class="flex items-center gap-3">
            <img src="/favicon.ico" alt="Logo" class="w-8 h-8 rounded flex-shrink-0" />
            <transition name="fade-text">
              <span v-if="!isCollapsed || isMobile" class="text-lg font-bold text-gray-800 whitespace-nowrap">
                KeepNode
              </span>
            </transition>
          </div>
        </slot>

        <!-- Close button for mobile -->
        <button v-if="isMobile" @click="closeMobile"
          class="md:hidden p-2 hover:bg-gray-100 rounded-lg transition-colors">
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>
      </div>

      <!-- Content -->
      <div class="sidebar-content">
        <slot name="content" />
      </div>

      <!-- Footer -->
      <div class="sidebar-footer">
        <slot name="footer" />
      </div>

      <!-- Collapse Toggle Button (Desktop only) -->
      <button @click="toggleCollapse" class="collapse-toggle hidden md:flex"
        :title="isCollapsed ? 'Genişlet' : 'Daralt'">
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
          stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
          :class="{ 'rotate-180': isCollapsed }" class="transition-transform duration-200">
          <polyline points="15 18 9 12 15 6"></polyline>
        </svg>
      </button>
    </aside>

    <!-- Mobile Menu Button -->
    <button @click="toggleMobile" class="mobile-menu-button md:hidden">
      <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <line x1="3" y1="12" x2="21" y2="12"></line>
        <line x1="3" y1="6" x2="21" y2="6"></line>
        <line x1="3" y1="18" x2="21" y2="18"></line>
      </svg>
    </button>
  </div>
</template>

<script>
export default {
  name: "Sidebar",
  provide() {
    return {
      isCollapsed: () => this.isCollapsed,
      isMobile: () => this.isMobile
    };
  },
  data() {
    return {
      isCollapsed: false,
      isMobileOpen: false,
      isMobile: false
    };
  },
  mounted() {
    this.checkMobile();
    window.addEventListener('resize', this.checkMobile);
  },
  beforeUnmount() {
    window.removeEventListener('resize', this.checkMobile);
  },
  methods: {
    toggleCollapse() {
      this.isCollapsed = !this.isCollapsed;
      this.$emit('collapse-change', this.isCollapsed);
    },
    toggleMobile() {
      this.isMobileOpen = !this.isMobileOpen;
    },
    closeMobile() {
      this.isMobileOpen = false;
    },
    checkMobile() {
      this.isMobile = window.innerWidth < 768;
      if (!this.isMobile) {
        this.isMobileOpen = false;
      }
    }
  }
};
</script>

<style scoped>
.sidebar-container {
  position: relative;
}

/* Sidebar Base Styles */
.sidebar {
  position: fixed;
  top: 0;
  left: 0;
  height: 100vh;
  background: white;
  border-right: 1px solid #e5e7eb;
  display: flex;
  flex-direction: column;
  z-index: 50;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  width: 280px;
}

/* Desktop - Expanded */
@media (min-width: 768px) {
  .sidebar {
    position: sticky;
    width: 280px;
  }
}

/* Desktop - Collapsed */
.sidebar.collapsed {
  width: 80px;
}

/* Mobile - Closed */
.sidebar.mobile-closed {
  transform: translateX(-100%);
}

/* Mobile - Open */
.sidebar.mobile-open {
  transform: translateX(0);
}

/* Header */
.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid #e5e7eb;
  min-height: 64px;
  background: #f9fafb;
}

.collapsed .sidebar-header {
  padding: 1rem;
  justify-content: center;
}

/* Content */
.sidebar-content {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 1rem;
}

.collapsed .sidebar-content {
  padding: 1rem 0.5rem;
}

/* Footer */
.sidebar-footer {
  border-top: 1px solid #e5e7eb;
  margin-top: auto;
}

/* Collapse Toggle Button */
.collapse-toggle {
  position: absolute;
  right: -12px;
  top: 80px;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: white;
  border: 1px solid #e5e7eb;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.collapse-toggle:hover {
  background: #f3f4f6;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

/* Mobile Menu Button */
.mobile-menu-button {
  position: fixed;
  top: 1rem;
  left: 1rem;
  z-index: 40;
  background: white;
  padding: 0.5rem;
  border-radius: 0.5rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  border: 1px solid #e5e7eb;
  cursor: pointer;
  transition: all 0.2s;
}

.mobile-menu-button:hover {
  background: #f3f4f6;
}

/* Transitions */
.overlay-enter-active,
.overlay-leave-active {
  transition: opacity 0.3s;
}

.overlay-enter-from,
.overlay-leave-to {
  opacity: 0;
}

.fade-text-enter-active,
.fade-text-leave-active {
  transition: opacity 0.2s, transform 0.2s;
}

.fade-text-enter-from,
.fade-text-leave-to {
  opacity: 0;
  transform: translateX(-10px);
}

/* Scrollbar Styling */
.sidebar-content::-webkit-scrollbar {
  width: 6px;
}

.sidebar-content::-webkit-scrollbar-track {
  background: transparent;
}

.sidebar-content::-webkit-scrollbar-thumb {
  background: #d1d5db;
  border-radius: 3px;
}

.sidebar-content::-webkit-scrollbar-thumb:hover {
  background: #9ca3af;
}
</style>

<!-- Dark Mode Styles (unscoped) -->
<style>
/* ==================== SIDEBAR DARK MODE ==================== */

/* Sidebar Base */
.dark .sidebar {
  background: #0f0f12;
  border-right-color: #2a2a30;
}

/* Header */
.dark .sidebar-header {
  background: #0f0f12;
  border-bottom-color: #2a2a30;
}

.dark .sidebar-header span {
  color: #fafafa;
}

/* Footer */
.dark .sidebar-footer {
  border-top-color: #2a2a30;
}

/* Collapse Toggle Button */
.dark .collapse-toggle {
  background: #1a1a1f;
  border-color: #2a2a30;
  color: #a1a1aa;
}

.dark .collapse-toggle:hover {
  background: #2a2a30;
}

/* Mobile Menu Button */
.dark .mobile-menu-button {
  background: #1a1a1f;
  border-color: #2a2a30;
  color: #fafafa;
}

.dark .mobile-menu-button:hover {
  background: #2a2a30;
}

/* Scrollbar */
.dark .sidebar-content::-webkit-scrollbar-thumb {
  background: #42424a;
}

.dark .sidebar-content::-webkit-scrollbar-thumb:hover {
  background: #555560;
}
</style>
