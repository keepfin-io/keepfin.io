<template>
  <div class="flex flex-wrap items-center justify-between gap-3 border-t bg-muted/20 px-4 py-3 text-sm text-muted-foreground">
    <p aria-live="polite">{{ $t('holdings.showing', { from: total ? (page - 1) * size + 1 : 0, to: Math.min(page * size, total), total }) }}</p>
    <div class="flex items-center gap-2">
      <slot />
      <Button variant="outline" size="sm" :disabled="disabled || page <= 1" @click="$emit('update:page', page - 1)"><LuChevronLeft class="size-4" aria-hidden="true" />{{ $t('wealth.previous') }}</Button>
      <span class="min-w-12 text-center tabular-nums text-foreground">{{ page }} / {{ pages }}</span>
      <Button variant="outline" size="sm" :disabled="disabled || page >= pages" @click="$emit('update:page', page + 1)">{{ $t('wealth.next') }}<LuChevronRight class="size-4" aria-hidden="true" /></Button>
    </div>
  </div>
</template>
<script setup>
import { LuChevronLeft, LuChevronRight } from '@kalimahapps/vue-icons/lu'
import { Button } from '@/components/ui'
defineProps({ page: { type: Number, required: true }, pages: { type: Number, required: true }, total: { type: Number, required: true }, size: { type: Number, required: true }, disabled: Boolean })
defineEmits(['update:page'])
</script>
