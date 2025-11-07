<template>
  <div class="sidebar-footer-inner">
    <slot></slot>
  </div>
</template>

<script>
export default {
  name: "SidebarFooter"
};
</script>

<style scoped>
.sidebar-footer-inner {
  padding: 1rem;
}

.collapsed .sidebar-footer-inner {
  padding: 0.5rem;
}
</style>
