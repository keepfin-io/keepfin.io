<template>
  <header class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
    <div class="flex min-w-0 gap-3">
      <span v-if="icon" class="flex size-10 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary" aria-hidden="true"><component :is="icon" class="size-5" /></span>
      <div class="min-w-0">
        <component :is="level" class="font-semibold tracking-tight" :class="level === 'h2' ? 'text-xl' : 'text-base'">{{ title }}</component>
        <p v-if="description" class="mt-1 max-w-3xl text-sm leading-relaxed text-muted-foreground">{{ description }}</p>
      </div>
    </div>
    <div v-if="$slots.default" class="flex shrink-0 flex-wrap items-center gap-2"><slot /></div>
  </header>
</template>
<script setup>
// Title row shared by the wealth pages: what the page is for, and its main action.
defineProps({ icon: { type: [Object, Function], default: null }, title: { type: String, required: true }, description: { type: String, default: '' }, level: { type: String, default: 'h2' } })
</script>
