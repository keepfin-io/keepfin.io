<template>
  <AlertDialog
    :open="visible"
    @update:open="!$event && !loading && $emit('cancel')"
  >
    <AlertDialogContent
      v-bind="$attrs"
      class="gap-0 overflow-hidden rounded-2xl p-0 sm:max-w-md"
      :aria-busy="loading"
    >
      <div class="flex gap-4 p-6">
        <span
          class="flex size-10 shrink-0 items-center justify-center rounded-full"
          :class="danger ? 'bg-destructive/10 text-destructive' : 'bg-primary/10 text-primary'"
          aria-hidden="true"
        >
          <LuTriangleAlert v-if="danger" class="size-5" />
          <LuCircleHelp v-else class="size-5" />
        </span>
        <AlertDialogHeader class="min-w-0 gap-1.5 text-left">
          <AlertDialogTitle
            :data-testid="$attrs['data-testid'] && `${$attrs['data-testid']}-title`"
          >
            {{ title }}
          </AlertDialogTitle>
          <AlertDialogDescription
            :data-testid="$attrs['data-testid'] && `${$attrs['data-testid']}-description`"
            class="leading-relaxed"
          >
            {{ message }}
          </AlertDialogDescription>
          <slot />
        </AlertDialogHeader>
      </div>
      <AlertDialogFooter class="border-t bg-muted/30 px-6 py-4">
        <AlertDialogCancel
          :data-testid="$attrs['data-testid'] && `${$attrs['data-testid']}-cancel`"
          :disabled="loading"
        >
          {{ cancelText }}
        </AlertDialogCancel>
        <Button
          :data-testid="$attrs['data-testid'] && `${$attrs['data-testid']}-confirm`"
          type="button"
          :loading="loading"
          :loading-text="loadingText"
          :variant="danger ? 'destructive' : 'default'"
          @click="$emit('confirm')"
        >
          {{ confirmText }}
        </Button>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>
<script setup>
defineOptions({ inheritAttrs: false })
import { computed } from 'vue'
import { LuCircleHelp, LuTriangleAlert } from '@kalimahapps/vue-icons/lu'
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogCancel,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'

// The app's single confirmation dialog. It blocks the page, focuses Cancel
// first, and cannot be dismissed while the confirmed action is running.
// Its props match the former InlineConfirmation, so call sites can switch.
const props = defineProps({
  visible: Boolean,
  variant: String,
  title: String,
  message: String,
  confirmText: { type: String, default: 'Confirm' },
  cancelText: { type: String, default: 'Cancel' },
  loadingText: { type: String, default: 'Processing...' },
  loading: Boolean,
})
defineEmits(['confirm', 'cancel'])
const danger = computed(() => props.variant === 'danger')
</script>
