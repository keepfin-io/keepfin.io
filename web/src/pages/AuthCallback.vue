<template>
  <div data-testid="auth-callback" class="flex min-h-full items-center justify-center p-4">
    <Card class="w-full max-w-md">
      <CardHeader class="items-center"><BrandLogo /><CardTitle>{{ $t('auth.completingSignIn') }}</CardTitle></CardHeader>
      <CardContent class="space-y-4">
        <p v-if="!failed" role="status">{{ $t('common.loading') }}</p>
        <Alert v-else data-testid="auth-callback-error" variant="error" :description="$t(`auth.${error || 'invalidLink'}`)" />
        <Button v-if="failed" as-child class="w-full"><router-link data-testid="auth-callback-login" to="/login">{{ $t('auth.backToLogin') }}</router-link></Button>
      </CardContent>
    </Card>
  </div>
</template>
<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Card, CardHeader, CardTitle, CardContent, Button, FeedbackAlert as Alert } from '@/components/ui'
import BrandLogo from '@/components/BrandLogo.vue'
import { refreshSession, useAuth } from '@/composables/useAuth'
const router = useRouter()
const { error } = useAuth()
const failed = ref(false)
onMounted(async () => {
  try {
    if (await refreshSession()) await router.replace('/')
    else failed.value = true
  } catch { failed.value = true }
})
</script>
