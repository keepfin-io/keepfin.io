<template>
  <Dialog :open="open" @update:open="!$event && !busy && $emit('close')">
    <DialogContent v-bind="$attrs" class="max-h-[90vh] gap-0 overflow-y-auto rounded-2xl p-0 sm:max-w-lg">
      <DialogHeader class="flex-row gap-3 border-b p-6 pr-12 text-left">
        <span v-if="icon" class="flex size-10 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary" aria-hidden="true"><component :is="icon" class="size-5" /></span>
        <div class="min-w-0 space-y-1">
          <DialogTitle>{{ title }}</DialogTitle>
          <DialogDescription class="leading-relaxed">{{ description }}</DialogDescription>
        </div>
      </DialogHeader>
      <div class="space-y-4 p-6"><Alert v-if="error" variant="error" :description="error" /><slot /></div>
      <DialogFooter v-if="$slots.footer" class="border-t bg-muted/30 px-6 py-4"><slot name="footer" /></DialogFooter>
    </DialogContent>
  </Dialog>
</template>
<script setup>
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, FeedbackAlert as Alert } from '@/components/ui'
// Forms for one record open on the same page; a running save cannot be dismissed.
defineOptions({ inheritAttrs: false })
defineProps({ open: Boolean, busy: Boolean, error: { type: String, default: '' }, icon: { type: [Object, Function], default: null }, title: { type: String, required: true }, description: { type: String, default: '' } })
defineEmits(['close'])
</script>
