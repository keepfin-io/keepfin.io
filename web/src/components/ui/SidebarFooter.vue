<template>
  <div class="sidebar-footer-inner" :class="{ 'is-collapsed': collapsed }">
    <slot></slot>
  </div>
</template>

<script>
export default {
  name: "SidebarFooter",
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
.sidebar-footer-inner {
  padding: 1rem;
}

.sidebar-footer-inner.is-collapsed {
  padding: 0.5rem;
}
</style>
