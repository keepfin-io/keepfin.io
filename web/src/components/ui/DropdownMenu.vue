<template>
  <div class="relative" @click="dropdownOpen = !dropdownOpen">
    <slot name="trigger"></slot>
    <transition name="fade">
      <div v-if="dropdownOpen" class="absolute left-0 bottom-12 w-full bg-white border rounded shadow-lg z-50">
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
  }
};
</script>

<style scoped>
.fade-enter-active, .fade-leave-active {
  transition: opacity 0.2s;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}
</style>
