<template>
  <Teleport to="body">
    <div
      v-bind="$attrs"
      v-if="visible"
      class="keepfin-loader-screen keepfin-loader-overlay"
    >
      <BrandSpinner :message="message" />
    </div>
  </Teleport>
</template>
<script setup>
defineOptions({ inheritAttrs: false });
import BrandSpinner from "./BrandSpinner.vue";
defineProps({ visible: Boolean, message: String });
</script>
