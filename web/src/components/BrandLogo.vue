<template>
  <span :class="cn('inline-flex shrink-0', compact ? 'h-8 w-8' : 'h-12 w-40', props.class)">
    <img :src="`/brand/keepfin${compact ? '-mark' : ''}.svg`" alt="Keepfin" class="h-full w-full object-contain dark:hidden" />
    <img :src="`/brand/keepfin${compact ? '-mark' : ''}-dark.svg`" alt="Keepfin" class="hidden h-full w-full object-contain dark:block" />
  </span>
</template>
<script setup>
import { cn } from '@/lib/utils'
const props = defineProps({ compact: Boolean, class: { type: null, required: false } })
</script>
