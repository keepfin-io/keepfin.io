<template>
  <div data-testid="login-page" class="flex min-h-svh items-center justify-center p-4 sm:p-8">
    <Card class="w-full max-w-md">
      <CardHeader class="items-center gap-4 text-center">
        <BrandLogo data-testid="login-brand" class="h-16! w-56! max-w-full" />
        <CardTitle class="text-2xl">{{ $t('auth.loginTitle') }}</CardTitle>
        <p class="text-sm text-muted-foreground">{{ $t('auth.passwordlessHint') }}</p>
      </CardHeader>
      <CardContent class="space-y-5">
        <Alert v-if="displayError" data-testid="login-error" variant="error" :description="$t(`auth.${displayError}`)" />
        <p v-if="loadingConfig" role="status" class="text-sm text-muted-foreground">{{ $t('common.loading') }}</p>
        <Button v-if="config.google" data-testid="login-google" class="w-full" variant="outline" :loading="busy === 'google'" :disabled="!!busy" @click="signInGoogle">
          {{ $t('auth.continueGoogle') }}
        </Button>
        <Separator v-if="config.google && config.email" />
        <form v-if="config.email" data-testid="login-form" class="space-y-4" @submit.prevent="sendLink">
          <div class="space-y-2">
            <Label for="login-email">{{ $t('auth.email') }}</Label>
            <Input id="login-email" data-testid="login-email" v-model="email" type="email" autocomplete="email" required maxlength="254" :disabled="!!busy" />
          </div>
          <Button data-testid="login-submit" type="submit" class="w-full" :loading="busy === 'email'" :disabled="!!busy || sent" :loading-text="$t('common.loading')">
            {{ $t('auth.sendLink') }}
          </Button>
          <p v-if="sent" data-testid="login-link-sent" role="status" class="text-sm text-muted-foreground">{{ $t(config.developmentMail ? 'auth.devLinkCreated' : 'auth.linkSent') }}</p>
          <Button v-if="sent" data-testid="login-use-another-email" type="button" variant="link" class="w-full" @click="sent = false">{{ $t('auth.tryAnotherEmail') }}</Button>
        </form>
        <p v-if="!loadingConfig && !config.email && !config.google && !localError" data-testid="login-unconfigured" class="text-sm text-muted-foreground">{{ $t('auth.notConfigured') }}</p>
        <Button v-if="hasSession" data-testid="login-sign-out" variant="outline" class="w-full" :loading="isLoading" @click="logout">{{ $t('common.signOut') }}</Button>
        <Button v-if="displayError === 'sessionUnavailable'" data-testid="login-retry" variant="outline" class="w-full" :disabled="loadingConfig" @click="retrySession">{{ $t('common.retry') }}</Button>
      </CardContent>
    </Card>
  </div>
</template>
<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { FeedbackAlert as Alert, Button, Card, CardContent, CardHeader, CardTitle, Input, Label, Separator } from '@/components/ui'
import BrandLogo from '@/components/BrandLogo.vue'
import { authClient } from '@/lib/auth-client'
import { useAuth, refreshSession } from '@/composables/useAuth'

const route = useRoute()
const router = useRouter()
const { error, hasSession, logout, isLoading } = useAuth()
const email = ref('')
const busy = ref('')
const sent = ref(false)
const localError = ref(null)
const config = ref({ google: false, email: false })
const loadingConfig = ref(true)
const displayError = computed(() => localError.value || error.value || (route.query.error ? 'invalidLink' : null))
const callbackURL = '/auth/callback'

async function loadConfig() {
  loadingConfig.value = true
  localError.value = null
  try {
    const response = await fetch('/api/auth/config', { credentials: 'same-origin' })
    if (!response.ok) throw new Error()
    config.value = await response.json()
  } catch { localError.value = 'sessionUnavailable' }
  finally { loadingConfig.value = false }
}
async function retrySession() {
  await loadConfig()
  try { if (await refreshSession()) await router.replace('/') }
  catch { /* The shared session error is displayed above. */ }
}
async function sendLink() {
  if (busy.value || sent.value) return
  busy.value = 'email'
  localError.value = null
  try {
    const result = await authClient.signIn.magicLink({
      email: email.value.trim(), callbackURL, newUserCallbackURL: callbackURL,
      errorCallbackURL: '/login?error=magic-link',
    })
    if (result.error) { localError.value = result.error.status === 429 ? 'tooManyAttempts' : 'signInFailed'; return }
    sent.value = true
  } catch { localError.value = 'signInFailed' }
  finally { busy.value = '' }
}
async function signInGoogle() {
  if (busy.value) return
  busy.value = 'google'
  localError.value = null
  try {
    const result = await authClient.signIn.social({ provider: 'google', callbackURL, errorCallbackURL: '/login?error=google' })
    if (result.error) localError.value = result.error.status === 429 ? 'tooManyAttempts' : 'signInFailed'
  } catch { localError.value = 'signInFailed' }
  finally { busy.value = '' }
}
onMounted(loadConfig)
</script>
