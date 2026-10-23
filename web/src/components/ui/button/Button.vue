<script setup>
import { cn } from 'cn'
import { Primitive } from 'reka-ui'
import { LoaderCircle } from 'lucide-vue-next'
import { buttonVariants } from '.'

const props = defineProps({
  variant: { type: null, required: false },
  size: { type: null, required: false },
  class: { type: null, required: false },
  asChild: { type: Boolean, required: false },
  as: { type: null, required: false, default: 'button' },
  disabled: Boolean,
  loading: Boolean,
  loadingText: String,
})
</script>

<template>
  <Primitive
    data-slot="button"
    :data-variant="variant"
    :data-size="size"
    :as="as"
    :as-child="asChild"
    :disabled="disabled || loading"
    :aria-busy="loading || undefined"
    :class="cn(buttonVariants({ variant, size }), props.class)"
  >
    <slot v-if="asChild" />
    <template v-else>
      <LoaderCircle
        v-if="loading"
        class="size-4 animate-spin"
        aria-hidden="true"
      />
      <template v-if="loading && loadingText">{{ loadingText }}</template>
      <slot v-else />
    </template>
  </Primitive>
</template>
