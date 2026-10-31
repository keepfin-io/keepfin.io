<template>
  <section v-if="visible" ref="panel" tabindex="-1" :aria-label="title" class="rounded-2xl border bg-card p-5 space-y-3 outline-none" :aria-busy="loading">
    <h2 :data-testid="testId && `${testId}-title`" class="text-base font-semibold">{{ title }}</h2>
    <p :data-testid="testId && `${testId}-description`" class="text-sm leading-relaxed text-muted-foreground">{{ message }}</p>
    <div class="flex flex-wrap justify-end gap-3 pt-2">
      <Button :data-testid="testId && `${testId}-cancel`" type="button" variant="outline" :disabled="loading" @click="$emit('cancel')">{{ cancelText }}</Button>
      <Button :data-testid="testId && `${testId}-confirm`" type="button" :variant="variant === 'danger' ? 'destructive' : 'default'" :loading="loading" :loading-text="loadingText" @click="$emit('confirm')">{{ confirmText }}</Button>
    </div>
  </section>
</template>
<script setup>
import { computed, nextTick, ref, useAttrs, watch } from 'vue'
import { Button } from '@/components/ui/button'
const props = defineProps({ visible: Boolean, title: String, message: String, variant: String, confirmText: String, cancelText: String, loadingText: String, loading: Boolean })
defineEmits(['confirm', 'cancel'])
const attrs = useAttrs()
const testId = computed(() => attrs['data-testid'])
const panel = ref(null)
watch(() => props.visible, async visible => {
  if (!visible) return
  await nextTick()
  panel.value?.focus({ preventScroll: true })
  panel.value?.scrollIntoView({ block: 'nearest', behavior: 'smooth' })
}, { immediate: true })
</script>
