<template>
  <AlertDialog
    :open="visible"
    @update:open="!$event && !loading && $emit('cancel')"
  >
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>{{ title }}</AlertDialogTitle>
        <AlertDialogDescription>{{ message }}</AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel :disabled="loading">
          {{ cancelText }}
        </AlertDialogCancel>
        <Button
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
