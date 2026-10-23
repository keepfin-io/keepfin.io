<template>
  <Teleport to="body">
    <div
      data-testid="toast-container"
      class="pointer-events-none fixed right-4 top-4 z-[110] grid w-[calc(100vw-2rem)] max-w-sm gap-3"
    >
      <Alert
        :data-testid="`toast-${item.id}`"
        v-for="item in toasts"
        :key="item.id"
        class="pointer-events-auto"
        v-bind="item"
        dismissible
        @dismiss="removeToast(item.id)"
      />
    </div>
  </Teleport>
</template>
<script>
import { ref, provide, readonly } from 'vue'
import Alert from './FeedbackAlert.vue'

const toasts = ref([])
let toastId = 0

function addToast({
  title,
  description = '',
  variant = 'info',
  duration = 5000,
}) {
  const id = ++toastId
  toasts.value.push({ id, title, description, variant })

  if (duration > 0) {
    setTimeout(() => removeToast(id), duration)
  }

  return id
}

function removeToast(id) {
  const index = toasts.value.findIndex((t) => t.id === id)
  if (index > -1) {
    toasts.value.splice(index, 1)
  }
}

// Toast helper functions
export const toast = {
  success: (title, description) =>
    addToast({ title, description, variant: 'success' }),
  error: (title, description) =>
    addToast({ title, description, variant: 'error' }),
  warning: (title, description) =>
    addToast({ title, description, variant: 'warning' }),
  info: (title, description) =>
    addToast({ title, description, variant: 'info' }),
  custom: (options) => addToast(options),
}

export default {
  name: 'ToastContainer',
  components: { Alert },
  setup() {
    provide('toast', readonly(toast))
    return {
      toasts,
      removeToast,
    }
  },
}
</script>
