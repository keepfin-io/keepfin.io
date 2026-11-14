<template>
  <Alert v-if="error" variant="error" :description="error" />
  <p v-if="busy" role="status" class="flex items-center gap-2 text-sm text-muted-foreground"><LuLoaderCircle class="size-4 animate-spin" aria-hidden="true" />{{ $t('common.loading') }}</p>
  <p v-if="notice" role="status" class="flex items-center gap-2 text-sm text-emerald-700 dark:text-emerald-400"><LuCircleCheck class="size-4" aria-hidden="true" />{{ notice }}</p>
</template>
<script setup>
import { LuCircleCheck, LuLoaderCircle } from '@kalimahapps/vue-icons/lu'
import { FeedbackAlert as Alert } from '@/components/ui'
defineProps({ error: String, busy: Boolean, notice: String })
</script>
