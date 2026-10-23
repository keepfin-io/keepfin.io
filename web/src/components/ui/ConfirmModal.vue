<template>
  <AlertDialog
    :open="visible"
    @update:open="!$event && !loading && $emit('cancel')"
  >
    <AlertDialogContent v-bind="$attrs">
      <AlertDialogHeader>
        <AlertDialogTitle
          :data-testid="
            $attrs['data-testid'] && `${$attrs['data-testid']}-title`
          "
        >
          {{ title }}
        </AlertDialogTitle>
        <AlertDialogDescription
          :data-testid="
            $attrs['data-testid'] && `${$attrs['data-testid']}-description`
          "
        >
          {{ message }}
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel
          :data-testid="
            $attrs['data-testid'] && `${$attrs['data-testid']}-cancel`
          "
          :disabled="loading"
        >
          {{ cancelText }}
        </AlertDialogCancel>
        <Button
          :data-testid="
            $attrs['data-testid'] && `${$attrs['data-testid']}-confirm`
          "
          type="button"
          :loading="loading"
          :loading-text="loadingText"
          :variant="variant === 'danger' ? 'destructive' : 'default'"
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
defineProps({
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
</script>
