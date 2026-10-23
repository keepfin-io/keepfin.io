<template>
  <ShadcnAlert
    :variant="variant === 'error' ? 'destructive' : 'default'"
    class="min-w-0"
    :class="dismissible && 'pr-12'"
  >
    <CircleAlert
      v-if="variant === 'error' || variant === 'warning'"
      class="size-4"
    />
    <CircleCheck v-else class="size-4" />
    <AlertTitle
      :data-testid="$attrs['data-testid'] && `${$attrs['data-testid']}-title`"
      v-if="title"
    >
      {{ title }}
    </AlertTitle>
    <AlertDescription
      :data-testid="
        $attrs['data-testid'] && `${$attrs['data-testid']}-description`
      "
      class="break-words"
    >
      {{ description }}
      <slot />
    </AlertDescription>
    <Button
      :data-testid="$attrs['data-testid'] && `${$attrs['data-testid']}-dismiss`"
      v-if="dismissible"
      variant="ghost"
      size="icon"
      class="absolute right-2 top-2 size-8"
      :aria-label="$t('common.dismiss')"
      @click="$emit('dismiss')"
    >
      <X class="size-4" />
    </Button>
  </ShadcnAlert>
</template>
<script setup>
import {
  Alert as ShadcnAlert,
  AlertTitle,
  AlertDescription,
} from '@/components/ui/alert/index.js'
import { Button } from '@/components/ui/button'
import { CircleAlert, CircleCheck, X } from 'lucide-vue-next'
defineProps({
  variant: { type: String, default: 'info' },
  title: String,
  description: String,
  dismissible: Boolean,
})
defineEmits(['dismiss'])
</script>
