<template>
  <div class="dropdown-container" ref="dropdown">
    <div @click="toggleDropdown">
      <slot name="trigger"></slot>
    </div>
    <transition name="dropdown-fade">
      <div
        v-if="dropdownOpen"
        class="dropdown-menu"
        @click.stop
      >
        <slot name="content"></slot>
      </div>
    </transition>
  </div>
</template>

<script>
export default {
  name: "DropdownMenu",
  data() {
    return {
      dropdownOpen: false
    };
  },
  mounted() {
    document.addEventListener('click', this.handleClickOutside);
  },
  beforeUnmount() {
    document.removeEventListener('click', this.handleClickOutside);
  },
  methods: {
    toggleDropdown() {
      this.dropdownOpen = !this.dropdownOpen;
    },
    handleClickOutside(event) {
      if (this.$refs.dropdown && !this.$refs.dropdown.contains(event.target)) {
        this.dropdownOpen = false;
      }
    }
  }
};
</script>

<style scoped>
.dropdown-container {
  position: relative;
  width: 100%;
}

.dropdown-menu {
  position: absolute;
  left: 0;
  bottom: calc(100% + 0.5rem);
  width: 100%;
  min-width: 200px;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 0.5rem;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
  z-index: 50;
  overflow: hidden;
}

/* Collapsed state - show dropdown on the right */
.collapsed .dropdown-menu {
  left: calc(100% + 0.5rem);
  bottom: 0;
}

.dropdown-fade-enter-active,
.dropdown-fade-leave-active {
  transition: opacity 0.2s, transform 0.2s;
}

.dropdown-fade-enter-from,
.dropdown-fade-leave-to {
  opacity: 0;
  transform: translateY(10px);
}

.collapsed .dropdown-fade-enter-from,
.collapsed .dropdown-fade-leave-to {
  transform: translateX(-10px);
}
</style>
