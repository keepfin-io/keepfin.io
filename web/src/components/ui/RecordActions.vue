<script setup>
import { ChevronRight, Pencil, Trash2 } from 'lucide-vue-next'

// Row actions as one segmented group with visible labels. iconOnly is for
// rows too narrow for text; the labels then remain as accessible names and
// tooltips.
defineProps({
  viewTo: { type: [String, Object], default: undefined },
  iconOnly: Boolean,
  disabled: Boolean,
})
defineEmits(['edit', 'delete'])

const segment = 'inline-flex h-8 items-center justify-center gap-1.5 whitespace-nowrap text-sm font-medium transition-colors focus-visible:relative focus-visible:z-10 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 [&_svg]:size-3.5 [&_svg]:shrink-0'
</script>

<template>
  <div class="inline-flex shrink-0 items-stretch divide-x overflow-hidden rounded-lg border bg-background shadow-xs">
    <!-- A labelled link says where it goes; an eye icon read as "show/hide". -->
    <router-link
      v-if="viewTo"
      :data-testid="$attrs['data-testid'] && `${$attrs['data-testid']}-view`"
      :to="viewTo"
      :title="$t('common.viewDetails')"
      :class="[segment, 'pl-3 pr-2 text-primary hover:bg-primary/10']"
    >
      {{ $t('common.details') }}
      <ChevronRight aria-hidden="true" />
    </router-link>
    <button
      type="button"
      :disabled="disabled"
      :aria-label="iconOnly ? $t('common.edit') : undefined"
      :title="iconOnly ? $t('common.edit') : undefined"
      :data-testid="$attrs['data-testid'] && `${$attrs['data-testid']}-edit`"
      :class="[segment, iconOnly ? 'w-8' : 'px-3', 'hover:bg-muted']"
      @click="$emit('edit')"
    >
      <Pencil aria-hidden="true" />
      <span v-if="!iconOnly">{{ $t('common.edit') }}</span>
    </button>
    <button
      type="button"
      :disabled="disabled"
      :aria-label="iconOnly ? $t('common.delete') : undefined"
      :title="iconOnly ? $t('common.delete') : undefined"
      :data-testid="$attrs['data-testid'] && `${$attrs['data-testid']}-delete`"
      :class="[segment, iconOnly ? 'w-8' : 'px-3', 'text-destructive hover:bg-destructive/10']"
      @click="$emit('delete')"
    >
      <Trash2 aria-hidden="true" />
      <span v-if="!iconOnly">{{ $t('common.delete') }}</span>
    </button>
  </div>
</template>
