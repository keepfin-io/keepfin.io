<template>
  <div class="keepfin-loader" role="status" aria-live="polite">
    <BrandLogo compact class="keepfin-loader-mark h-11! w-11!" aria-hidden="true" />
    <span class="keepfin-loader-track" aria-hidden="true"></span>
    <p class="keepfin-loader-message">{{ message || $t('common.loading') }}</p>
  </div>
</template>
<script setup>
import BrandLogo from '@/components/BrandLogo.vue'

// public/styles/loading.css is loaded before Vue so startup uses this design too.
defineProps({ message: String })
</script>
