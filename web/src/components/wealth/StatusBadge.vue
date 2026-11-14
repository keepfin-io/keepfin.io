<template>
  <span class="inline-flex shrink-0 items-center gap-1.5 whitespace-nowrap rounded-full px-2 py-0.5 text-xs font-medium ring-1 ring-inset" :class="tones[tone] || tones.neutral">
    <span class="size-1.5 rounded-full bg-current" :class="pulse && 'animate-pulse'" aria-hidden="true" />
    <slot />
  </span>
</template>
<script setup>
// A status is spelled out in text; the color only reinforces it.
defineProps({ tone: { type: String, default: 'neutral' }, pulse: Boolean })
const tones = {
  success: 'bg-emerald-500/10 text-emerald-700 ring-emerald-600/20 dark:text-emerald-400',
  warning: 'bg-amber-500/10 text-amber-700 ring-amber-600/20 dark:text-amber-400',
  danger: 'bg-destructive/10 text-destructive ring-destructive/20',
  info: 'bg-blue-500/10 text-blue-700 ring-blue-600/20 dark:text-blue-400',
  neutral: 'bg-muted text-muted-foreground ring-border',
}
</script>
