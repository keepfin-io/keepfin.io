<template>
  <div class="flex flex-col items-center gap-3 px-6 py-12 text-center">
    <span class="flex size-12 items-center justify-center rounded-2xl bg-muted text-muted-foreground" aria-hidden="true"><component :is="icon" class="size-6" /></span>
    <p class="font-medium">{{ title }}</p>
    <p v-if="description" class="max-w-md text-sm text-muted-foreground">{{ description }}</p>
    <slot />
  </div>
</template>
<script setup>
defineProps({ icon: { type: [Object, Function], required: true }, title: { type: String, required: true }, description: { type: String, default: '' } })
</script>
