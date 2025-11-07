<template>
  <div class="sidebar-header-content" :class="{ 'is-collapsed': collapsed }">
    <img src="/favicon.ico" alt="Logo" class="header-logo" />
    <transition name="fade-text">
      <span v-if="!collapsed || mobile" class="header-text">KeepNode</span>
    </transition>
  </div>
</template>

<script>
export default {
  name: "SidebarHeader",
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
.sidebar-header-content {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  transition: all 0.3s;
}

.sidebar-header-content.is-collapsed {
  justify-content: center;
}

.header-logo {
  width: 2rem;
  height: 2rem;
  border-radius: 0.375rem;
  flex-shrink: 0;
  transition: all 0.3s;
}

.header-text {
  font-size: 1.125rem;
  font-weight: 700;
  color: #1f2937;
  white-space: nowrap;
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
</style>
