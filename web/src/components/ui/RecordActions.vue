<script setup>
import { Button } from './button'
import { Eye, Pencil, Trash2 } from 'lucide-vue-next'

defineProps({
  viewTo: { type: [String, Object], default: undefined },
  compact: { type: Boolean, default: true },
  disabled: Boolean,
})
defineEmits(['edit', 'delete'])
</script>

<template>
  <div class="flex shrink-0 items-center gap-1">
    <Button
      v-if="viewTo"
      variant="ghost"
      :size="compact ? 'icon' : 'sm'"
      as-child
    >
      <router-link
        :data-testid="$attrs['data-testid'] && `${$attrs['data-testid']}-view`"
        :to="viewTo"
        :aria-label="$t('common.viewDetails')"
      >
        <Eye aria-hidden="true" />
        <span v-if="!compact">{{ $t('common.viewDetails') }}</span>
      </router-link>
    </Button>
    <Button
      type="button"
      :variant="compact ? 'ghost' : 'outline'"
      :size="compact ? 'icon' : 'sm'"
      :disabled="disabled"
      :aria-label="$t('common.edit')"
      :data-testid="$attrs['data-testid'] && `${$attrs['data-testid']}-edit`"
      @click="$emit('edit')"
    >
      <Pencil aria-hidden="true" />
      <span v-if="!compact">{{ $t('common.edit') }}</span>
    </Button>
    <Button
      type="button"
      :variant="compact ? 'destructive-ghost' : 'destructive-outline'"
      :size="compact ? 'icon' : 'sm'"
      :disabled="disabled"
      :aria-label="$t('common.delete')"
      :data-testid="$attrs['data-testid'] && `${$attrs['data-testid']}-delete`"
      @click="$emit('delete')"
    >
      <Trash2 aria-hidden="true" />
      <span v-if="!compact">{{ $t('common.delete') }}</span>
    </Button>
  </div>
</template>
