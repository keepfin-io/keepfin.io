<template>
  <main data-testid="login-page" class="login-page">
    <div class="login-topbar">
      <span class="login-wordmark">KEEPFIN <span> / </span> {{ $t('auth.workspaceLabel') }}</span>
      <Button data-testid="login-theme" variant="ghost" size="icon" class="rounded-full" :aria-label="$t(isDark ? 'profile.switchToLight' : 'profile.switchToDark')" @click="applyTheme(isDark ? 'light' : 'dark')">
        <Sun v-if="isDark" class="size-4" /><Moon v-else class="size-4" />
      </Button>
    </div>

    <div class="login-shell">
      <aside class="login-story" :aria-label="$t('auth.storyTitle')">
        <BrandLogo data-testid="login-brand" class="h-11! w-40!" />
        <div class="story-copy">
          <p class="login-eyebrow"><span /> {{ $t('auth.storyEyebrow') }}</p>
          <h2>{{ $t('auth.storyTitle') }}<br><span>{{ $t('auth.storyAccent') }}</span></h2>
          <p class="story-description">{{ $t('auth.storyDescription') }}</p>
        </div>

        <div class="portfolio-art" aria-hidden="true">
          <div class="art-orbit art-orbit-outer" />
          <div class="art-orbit art-orbit-inner" />
          <div class="art-cross art-cross-horizontal" /><div class="art-cross art-cross-vertical" />
          <div class="art-center"><BrandLogo compact class="h-12! w-12!" /></div>
          <div class="art-tile art-tile-investments"><span class="art-icon"><ChartNoAxesCombined /></span><span>{{ $t('auth.artInvestments') }}<small>{{ $t('auth.artOverview') }}</small></span><span class="art-dot" /></div>
          <div class="art-tile art-tile-cash"><span class="art-icon"><WalletCards /></span><span>{{ $t('auth.artCash') }}<small>USD <span class="currency-dot">·</span> EUR <span class="currency-dot">·</span> TRY</small></span></div>
          <div class="art-tile art-tile-accounts"><Layers3 class="size-4" /><span>{{ $t('auth.artAccounts') }}</span></div>
        </div>

        <div class="story-footer"><span class="story-footer-line" /><p>{{ $t('auth.storyFooter') }}</p></div>
      </aside>

      <section class="login-form-panel" aria-labelledby="login-heading">
        <BrandLogo data-testid="login-mobile-brand" class="mb-8 h-10! w-36! lg:hidden!" />
        <div class="login-form-content">
          <div class="login-heading">
            <span class="welcome-icon"><ArrowRight class="size-5" /></span>
            <h1 id="login-heading">{{ $t('auth.loginTitle') }}</h1>
            <p>{{ $t('auth.loginHint') }}</p>
          </div>

          <Alert v-if="displayError" data-testid="login-error" variant="error" :description="$t(`auth.${displayError}`)" />
          <p v-if="loadingConfig" role="status" class="text-sm text-muted-foreground">{{ $t('common.loading') }}</p>

          <div class="social-buttons">
            <Button data-testid="login-google" class="social-button" variant="outline" :loading="busy === 'google'" :disabled="loadingConfig || !config.google || !!busy" @click="signInGoogle">
              <svg aria-hidden="true" viewBox="0 0 24 24" class="size-4 shrink-0"><path fill="#4285F4" d="M21.6 12.23c0-.71-.06-1.39-.18-2.05H12v3.88h5.38a4.6 4.6 0 0 1-1.99 3.02v2.51h3.23c1.89-1.74 2.98-4.3 2.98-7.36Z"/><path fill="#34A853" d="M12 22c2.7 0 4.96-.9 6.62-2.41l-3.23-2.51c-.9.6-2.05.96-3.39.96-2.61 0-4.82-1.76-5.61-4.12H3.05v2.59A10 10 0 0 0 12 22Z"/><path fill="#FBBC05" d="M6.39 13.92a6 6 0 0 1 0-3.84V7.49H3.05a10 10 0 0 0 0 9.02l3.34-2.59Z"/><path fill="#EA4335" d="M12 5.96c1.47 0 2.79.51 3.83 1.51L18.7 4.6A9.6 9.6 0 0 0 12 2a10 10 0 0 0-8.95 5.49l3.34 2.59C7.18 7.72 9.39 5.96 12 5.96Z"/></svg>
              <span>{{ $t('auth.continueGoogle') }}</span>
              <span v-if="!loadingConfig && !config.google" class="coming-soon">{{ $t('auth.comingSoon') }}</span>
            </Button>
            <Button data-testid="login-apple" class="social-button" variant="outline" disabled>
              <svg aria-hidden="true" viewBox="0 0 24 24" class="size-4 shrink-0" fill="currentColor"><path d="M17.05 12.54c.03 3.28 2.88 4.37 2.91 4.38-.02.08-.46 1.56-1.51 3.09-.91 1.32-1.85 2.63-3.34 2.66-1.47.03-1.94-.86-3.62-.86-1.67 0-2.19.83-3.58.89-1.44.05-2.53-1.43-3.45-2.75-1.88-2.71-3.32-7.65-1.39-11  .96-1.66 2.67-2.72 4.53-2.75 1.42-.03 2.75.95 3.61.95.87 0 2.49-1.17 4.2-1 .72.03 2.73.29 4.02 2.18-.1.06-2.4 1.4-2.38 4.21ZM14.31 4.36C15.07 3.44 15.58 2.17 15.44.9c-1.1.05-2.43.73-3.22 1.65-.71.81-1.33 2.1-1.16 3.34 1.23.1 2.49-.63 3.25-1.53Z"/></svg>
              <span>{{ $t('auth.continueApple') }}</span>
              <span class="coming-soon">{{ $t('auth.comingSoon') }}</span>
            </Button>
          </div>

          <template v-if="config.email">
            <div class="login-divider"><span />{{ $t('auth.orEmail') }}<span /></div>
            <form data-testid="login-form" class="email-form" @submit.prevent="sendLink">
              <div class="space-y-2.5">
                <Label for="login-email">{{ $t('auth.email') }}</Label>
                <div class="input-with-icon">
                  <Mail class="input-icon" aria-hidden="true" />
                  <Input id="login-email" data-testid="login-email" v-model="email" class="login-input" type="email" autocomplete="email" :placeholder="$t('auth.emailPlaceholder')" required maxlength="254" :disabled="!!busy || sent" />
                </div>
              </div>
              <Button data-testid="login-submit" type="submit" class="login-primary" :loading="busy === 'email'" :disabled="!!busy || sent" :loading-text="$t('common.loading')">
                {{ $t('auth.sendLink') }}<ArrowRight class="ml-auto size-4" aria-hidden="true" />
              </Button>
              <div v-if="sent" data-testid="login-link-sent" role="status" class="login-success">
                <CircleCheck class="mt-0.5 size-4 shrink-0" aria-hidden="true" />
                <p>{{ $t(config.developmentMail ? 'auth.devLinkCreated' : 'auth.linkSent') }}</p>
              </div>
              <p v-else class="login-form-note"><LockKeyhole class="size-3.5 shrink-0" aria-hidden="true" />{{ $t('auth.emailNote') }}</p>
              <Button v-if="sent" data-testid="login-use-another-email" type="button" variant="link" class="w-full" @click="sent = false">{{ $t('auth.tryAnotherEmail') }}</Button>
            </form>
          </template>

          <details v-if="config.internalTesting" data-testid="login-internal" class="internal-login">
            <summary><FlaskConical class="size-4" aria-hidden="true" /><span>{{ $t('auth.internalTesting') }}</span><ChevronDown class="internal-chevron ml-auto size-4" aria-hidden="true" /></summary>
            <form data-testid="login-internal-form" class="space-y-4 px-4 pb-4 pt-1" @submit.prevent="signInInternal">
              <div class="space-y-2"><Label for="login-username">{{ $t('auth.username') }}</Label><Input id="login-username" data-testid="login-username" v-model="username" class="h-11 rounded-xl" autocomplete="username" required maxlength="254" :disabled="!!busy" /></div>
              <div class="space-y-2"><Label for="login-password">{{ $t('auth.password') }}</Label><div class="relative"><Input id="login-password" data-testid="login-password" v-model="password" class="h-11 rounded-xl pr-11" :type="showPassword ? 'text' : 'password'" autocomplete="current-password" required maxlength="128" :disabled="!!busy" /><Button data-testid="login-password-toggle" type="button" variant="ghost" size="icon" class="absolute right-1 top-1 size-9 rounded-lg" :aria-label="$t(showPassword ? 'auth.hidePassword' : 'auth.showPassword')" :aria-pressed="showPassword" @click="showPassword = !showPassword"><EyeOff v-if="showPassword" class="size-4" /><Eye v-else class="size-4" /></Button></div></div>
              <Button data-testid="login-internal-submit" class="h-11 w-full rounded-xl" type="submit" :loading="busy === 'internal'" :disabled="!!busy">{{ $t('auth.signIn') }}<ArrowRight class="size-4" aria-hidden="true" /></Button>
            </form>
          </details>

          <p v-if="!loadingConfig && !config.email && !config.google && !config.internalTesting && !localError" data-testid="login-unconfigured" class="text-sm text-muted-foreground">{{ $t('auth.notConfigured') }}</p>
          <Button v-if="hasSession" data-testid="login-sign-out" variant="outline" class="w-full" :loading="isLoading" @click="logout">{{ $t('common.signOut') }}</Button>
          <Button v-if="displayError === 'sessionUnavailable'" data-testid="login-retry" variant="outline" class="w-full" :disabled="loadingConfig" @click="retrySession">{{ $t('common.retry') }}</Button>
        </div>
      </section>
    </div>
    <footer class="login-page-footer"><span>keepfin.io</span><span></span></footer>
  </main>
</template>
<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { FeedbackAlert as Alert, Button, Input, Label } from '@/components/ui'
import BrandLogo from '@/components/BrandLogo.vue'
import { ArrowRight, ChartNoAxesCombined, ChevronDown, CircleCheck, Eye, EyeOff, FlaskConical, Layers3, LockKeyhole, Mail, Moon, Sun, WalletCards } from 'lucide-vue-next'
import { applyTheme, useTheme } from '@/composables/useTheme'
import { authClient } from '@/lib/auth-client'
import { useAuth, refreshSession } from '@/composables/useAuth'

const route = useRoute()
const router = useRouter()
const { error, hasSession, logout, isLoading } = useAuth()
const email = ref('')
const username = ref('internal-testing')
const password = ref('')
const showPassword = ref(false)
const { isDark } = useTheme()
const busy = ref('')
const sent = ref(false)
const localError = ref(null)
const config = ref({ google: false, email: false, internalTesting: false })
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
  if (busy.value || !config.value.google) return
  busy.value = 'google'
  localError.value = null
  try {
    const result = await authClient.signIn.social({ provider: 'google', callbackURL, errorCallbackURL: '/login?error=google' })
    if (result.error) localError.value = result.error.status === 429 ? 'tooManyAttempts' : 'signInFailed'
  } catch { localError.value = 'signInFailed' }
  finally { busy.value = '' }
}
async function signInInternal() {
  if (busy.value || !config.value.internalTesting) return
  busy.value = 'internal'
  localError.value = null
  try {
    const name = username.value.trim().toLowerCase()
    const result = await authClient.signIn.email({
      email: name.includes('@') ? name : `${name}@keepfin.test`,
      password: password.value,
    })
    if (result.error) { localError.value = result.error.status === 429 ? 'tooManyAttempts' : 'invalidCredentials'; return }
    password.value = ''
    if (await refreshSession()) await router.replace('/')
    else localError.value = 'signInFailed'
  } catch { localError.value = 'signInFailed' }
  finally { busy.value = '' }
}
onMounted(loadConfig)
</script>

<style scoped>
.login-page { --login-accent: #007e78; --login-accent-soft: #e6f5f2; min-height: 100svh; display: flex; flex-direction: column; justify-content: center; padding: 24px 40px; background: radial-gradient(ellipse at 15% 40%, #d8eee833, transparent 60%), #f7faf9; }
.dark .login-page { --login-accent: #68e1d4; --login-accent-soft: #14332f; background: radial-gradient(ellipse at 12% 45%, #174a4024, transparent 55%), #0f0f12; }
.login-topbar, .login-page-footer { width: 100%; max-width: 1100px; margin-inline: auto; display: flex; align-items: center; justify-content: space-between; gap: 16px; }
.login-topbar { margin-bottom: 20px; }
.login-wordmark { color: var(--muted-foreground); font-size: 10px; font-weight: 600; letter-spacing: .17em; text-transform: uppercase; }
.login-wordmark span { margin-inline: 12px; opacity: .4; }
.login-shell { width: 100%; max-width: 1100px; margin-inline: auto; display: grid; grid-template-columns: 1fr 1fr; overflow: hidden; border: 1px solid var(--border); border-radius: 28px; background: var(--card); box-shadow: 0 24px 80px -48px #001c243d; }
.login-story { position: relative; display: flex; min-width: 0; flex-direction: column; overflow: hidden; padding: 40px; background: linear-gradient(145deg, #f1f8f5 0%, #e4f2ed 60%, #d7ece5 100%); border-right: 1px solid #0e453d0c; }
.dark .login-story { background: radial-gradient(ellipse at 40% 80%, #1a524a66, transparent 70%), linear-gradient(145deg, #14231f, #111b19); border-right-color: #ffffff09; }
.story-copy { position: relative; z-index: 1; margin-top: 50px; }
.login-eyebrow { display: flex; align-items: center; gap: 9px; color: var(--login-accent); font-size: 10px; font-weight: 600; letter-spacing: .13em; text-transform: uppercase; }
.login-eyebrow span { width: 5px; height: 5px; border-radius: 50%; background: currentColor; }
.story-copy h2 { margin-top: 18px; font-size: clamp(32px, 3.4vw, 46px); font-weight: 500; line-height: 1.14; letter-spacing: -.055em; }
.story-copy h2 span { color: var(--login-accent); }
.story-description { max-width: 340px; margin-top: 18px; color: var(--muted-foreground); font-size: 14px; line-height: 1.8; }
.portfolio-art { position: relative; flex: 1; min-height: 222px; margin-top: 25px; isolation: isolate; }
.art-orbit { position: absolute; left: 50%; top: 50%; border: 1px solid #168a7330; border-radius: 50%; transform: translate(-50%, -50%); }
.art-orbit-outer { width: 310px; height: 310px; border-style: dashed; opacity: .6; }
.art-orbit-inner { width: 205px; height: 205px; }
.art-cross { position: absolute; left: 50%; top: 50%; background: #168a7314; transform: translate(-50%, -50%); z-index: -1; }
.art-cross-horizontal { width: 100%; height: 1px; }.art-cross-vertical { width: 1px; height: 100%; }
.art-center { position: absolute; top: 50%; left: 50%; display: grid; place-items: center; width: 88px; height: 88px; border: 1px solid #328d7926; border-radius: 26px; background: #f5fcf8; transform: translate(-50%, -50%) rotate(-7deg); box-shadow: 0 12px 35px #174e3512; }
.art-center > * { transform: rotate(7deg); }
.dark .art-center { background: #1a2e28; border-color: #69e0c23b; box-shadow: 0 0 70px #34c4a512; }
.art-tile { position: absolute; display: flex; align-items: center; gap: 10px; padding: 12px; border: 1px solid #246b5420; border-radius: 14px; background: #ffffffed; font-size: 11px; font-weight: 500; box-shadow: 0 10px 25px -16px #001d2b66; }
.dark .art-tile { background: #1b302a; border-color: #8cdcc121; }
.art-tile small { display: block; margin-top: 4px; color: var(--muted-foreground); font-size: 9px; font-weight: 400; }
.art-tile-investments { top: 15px; left: 0; transform: rotate(-5deg); }
.art-tile-cash { right: -8px; bottom: 24px; transform: rotate(5deg); }
.art-tile-accounts { left: 12px; bottom: 2px; gap: 8px; color: var(--login-accent); font-size: 10px; }
.art-icon { display: grid; place-items: center; width: 32px; height: 32px; border-radius: 9px; color: var(--login-accent); background: var(--login-accent-soft); }
.art-icon svg { width: 16px; height: 16px; }
.art-dot { width: 5px; height: 5px; margin-left: 6px; border-radius: 50%; background: #13b995; }.currency-dot { margin-inline: 3px; opacity: .4; }
.story-footer { position: relative; display: flex; align-items: center; gap: 12px; margin-top: 42px; color: var(--muted-foreground); font-size: 10px; }
.story-footer-line { height: 1px; width: 26px; background: var(--login-accent); }
.login-form-panel { min-width: 0; display: flex; flex-direction: column; justify-content: center; padding: 44px clamp(32px, 4vw, 64px); }
.login-form-content { width: 100%; max-width: 380px; margin-inline: auto; display: flex; flex-direction: column; gap: 20px; }
.login-heading { margin-bottom: 7px; }
.welcome-icon { display: inline-flex; align-items: center; justify-content: center; width: 38px; height: 38px; margin-bottom: 20px; border: 1px solid #148f7826; border-radius: 12px; color: var(--login-accent); background: var(--login-accent-soft); }
.login-heading h1 { font-size: 25px; font-weight: 600; letter-spacing: -.045em; line-height: 1.3; }
.login-heading > p { margin-top: 10px; font-size: 13px; line-height: 1.7; color: var(--muted-foreground); }
.social-buttons { display: grid; gap: 10px; }
.social-button { width: 100%; height: 46px; justify-content: flex-start; gap: 12px; padding-inline: 14px; border-radius: 12px; font-size: 12px; box-shadow: none; }
.social-button:disabled { opacity: 1; color: var(--muted-foreground); background: transparent; cursor: not-allowed; }
.social-button:disabled > svg { opacity: .65; }
.coming-soon { margin-left: auto; padding: 3px 6px; border: 1px solid var(--border); border-radius: 5px; font-size: 9px; line-height: 1.2; font-weight: 400; }
.login-divider { display: flex; align-items: center; gap: 13px; color: var(--muted-foreground); font-size: 10px; }
.login-divider span { flex: 1; height: 1px; background: var(--border); }
.email-form { display: flex; flex-direction: column; gap: 14px; }
.input-with-icon { position: relative; }
.input-icon { position: absolute; top: 16px; left: 14px; width: 16px; height: 16px; color: var(--muted-foreground); pointer-events: none; }
.login-input { height: 48px; padding-left: 42px; border-radius: 12px; font-size: 13px; box-shadow: none; }
.login-input:focus-visible { border-color: var(--login-accent); outline-color: var(--login-accent); --tw-ring-color: #18a28a24; }
.login-primary { width: 100%; height: 48px; justify-content: space-between; padding-inline: 18px; border-radius: 12px; background: #58d9c8; color: #092b24; font-size: 13px; font-weight: 600; box-shadow: 0 4px 12px #0a936519; }
.login-primary:hover { background: #77e6d7; }.login-primary:disabled { opacity: .55; }
.login-form-note { display: flex; align-items: center; justify-content: center; gap: 7px; color: var(--muted-foreground); font-size: 11px; line-height: 1.6; }
.login-success { display: flex; align-items: flex-start; gap: 10px; padding: 12px; border: 1px solid #138b712e; border-radius: 10px; color: var(--login-accent); background: var(--login-accent-soft); font-size: 12px; line-height: 1.7; }
.internal-login { margin-top: 4px; border: 1px solid var(--border); border-radius: 12px; background: color-mix(in srgb, var(--muted) 15%, transparent); }
.internal-login summary { display: flex; align-items: center; gap: 9px; padding: 13px 14px; color: var(--muted-foreground); cursor: pointer; list-style: none; font-size: 11px; }
.internal-login summary::-webkit-details-marker { display: none; }.internal-login summary:focus-visible { outline: 2px solid var(--login-accent); outline-offset: 3px; border-radius: 10px; }
.internal-login[open] .internal-chevron { transform: rotate(180deg); }
.login-page-footer { margin-top: 20px; color: var(--muted-foreground); font-size: 10px; }
.login-page-footer > span:first-child { font-weight: 600; letter-spacing: -.02em; }
@media (max-width: 1023px) {
  .login-page { padding: 24px; }.login-topbar, .login-page-footer, .login-shell { max-width: 480px; }
  .login-shell { display: block; border-radius: 24px; }.login-story { display: none; }.login-form-panel { padding: 32px; }
  .welcome-icon { display: none; }.login-heading { margin-bottom: 4px; }.login-form-content { max-width: none; }
}
@media (max-width: 380px) {
  .login-page { padding: 16px 12px; }.login-form-panel { padding: 24px 20px; }.login-wordmark { font-size: 9px; letter-spacing: .08em; }.login-wordmark span { margin-inline: 6px; }
  .login-heading h1 { font-size: 23px; }.social-button { gap: 8px; font-size: 11px; }.login-page-footer { flex-wrap: wrap; gap: 6px; }
}
</style>
