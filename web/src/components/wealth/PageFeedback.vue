<template>
  <Alert v-if="error" variant="error" :description="error" />
  <p v-if="busy" role="status" class="text-sm text-muted-foreground">{{ $t('common.loading') }}</p>
  <p v-if="notice" role="status" class="text-sm text-primary">{{ notice }}</p>
</template>
<script setup>
import { FeedbackAlert as Alert } from '@/components/ui'
defineProps({ error: String, busy: Boolean, notice: String })
</script>
