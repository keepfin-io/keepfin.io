<template>
  <Teleport to="body">
    <div
      v-bind="$attrs"
      v-if="visible"
      class="fixed inset-0 z-[100] flex items-center justify-center bg-background/80 p-4 backdrop-blur-sm"
      role="status"
      aria-live="polite"
    >
      <Card class="w-full max-w-xs">
        <CardContent class="flex flex-col items-center gap-4 py-6 text-center">
          <LoaderCircle class="size-7 animate-spin text-muted-foreground" />
          <p class="text-sm">
            {{ message || $t('common.loading') }}
          </p>
        </CardContent>
      </Card>
    </div>
  </Teleport>
</template>
<script setup>
defineOptions({ inheritAttrs: false })
import { Card, CardContent } from '@/components/ui/card'
import { LoaderCircle } from 'lucide-vue-next'
defineProps({ visible: Boolean, message: String })
</script>
