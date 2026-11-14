<template>
  <span
    :data-testid="`tag-badge-${tag.id}`"
    class="inline-flex max-w-full items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium ring-1 ring-inset"
    :class="style.badge"
  >
    <span class="size-1.5 shrink-0 rounded-full" :class="style.dot" aria-hidden="true" />
    <span class="truncate">{{ tag.name }}</span>
  </span>
</template>
<script setup>
import { computed } from 'vue'
import { tagStyle } from '@/lib/tags'

const props = defineProps({ tag: { type: Object, required: true } })
const style = computed(() => tagStyle(props.tag.color))
</script>
