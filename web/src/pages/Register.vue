<template>
  <div class="flex min-h-svh items-center justify-center p-4 sm:p-8">
    <Card class="w-full max-w-md">
      <CardHeader class="text-center">
        <CardTitle class="text-2xl">{{ $t('auth.registerTitle') }}</CardTitle>
      </CardHeader>
      <CardContent>
        <form class="space-y-5" @submit.prevent="handleRegister">
          <Alert v-if="error" variant="error" :description="error" />
          <div class="space-y-2">
            <Label for="email">{{ $t('auth.email') }}</Label>
            <Input
              id="email"
              v-model="form.email"
              type="email"
              autocomplete="email"
              :disabled="isLoading"
              :aria-invalid="!!errors.email"
            />
            <p v-if="errors.email" class="text-sm text-destructive">
              {{ errors.email }}
            </p>
          </div>
          <div class="space-y-2">
            <Label for="password">{{ $t('auth.password') }}</Label>
            <Input
              id="password"
              v-model="form.password"
              type="password"
              autocomplete="new-password"
              :disabled="isLoading"
              :aria-invalid="!!errors.password"
            />
            <p v-if="errors.password" class="text-sm text-destructive">
              {{ errors.password }}
            </p>
          </div>
          <div class="space-y-2">
            <Label for="confirmPassword">
              {{ $t('auth.confirmPassword') }}
            </Label>
            <Input
              id="confirmPassword"
              v-model="form.confirmPassword"
              type="password"
              autocomplete="new-password"
              :disabled="isLoading"
              :aria-invalid="!!errors.confirmPassword"
            />
            <p v-if="errors.confirmPassword" class="text-sm text-destructive">
              {{ errors.confirmPassword }}
            </p>
          </div>
          <div class="space-y-1 text-xs text-muted-foreground">
            <p>{{ $t('auth.passwordRequirements') }}</p>
            <ul class="list-inside list-disc">
              <li
                v-for="rule in [
                  { ok: hasMinLength, label: 'minLength' },
                  { ok: hasUppercase, label: 'uppercase' },
                  { ok: hasLowercase, label: 'lowercase' },
                  { ok: hasNumber, label: 'number' },
                ]"
                :key="rule.label"
                :class="rule.ok && 'text-emerald-700 dark:text-emerald-400'"
              >
                {{ $t('auth.' + rule.label) }}
              </li>
            </ul>
          </div>
          <Button
            type="submit"
            class="w-full"
            :loading="isLoading"
            :loading-text="$t('common.registering')"
          >
            {{ $t('auth.register') }}
          </Button>
        </form>
      </CardContent>
      <CardFooter
        class="flex-wrap justify-center gap-1 text-sm text-muted-foreground"
      >
        {{ $t('auth.hasAccount') }}
        <Button variant="link" size="inline" as-child>
          <router-link to="/login">{{ $t('auth.login') }}</router-link>
        </Button>
      </CardFooter>
    </Card>
  </div>
</template>
<script>
import {
  FeedbackAlert as Alert,
  Button,
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
  Input,
  Label,
} from '@/components/ui'
import { useAuth } from '../composables/useAuth'
import { registerSchema } from '@/schemas/auth'

export default {
  components: {
    Alert,
    Button,
    Card,
    CardContent,
    CardFooter,
    CardHeader,
    CardTitle,
    Input,
    Label,
  },
  name: 'RegisterPage',
  setup() {
    const { register, isLoading, error } = useAuth()
    return { register, isLoading, error }
  },
  data() {
    return {
      form: {
        email: '',
        password: '',
        confirmPassword: '',
      },
      errors: {},
    }
  },
  computed: {
    hasMinLength() {
      return this.form.password.length >= 8
    },
    hasUppercase() {
      return /[A-Z]/.test(this.form.password)
    },
    hasLowercase() {
      return /[a-z]/.test(this.form.password)
    },
    hasNumber() {
      return /[0-9]/.test(this.form.password)
    },
  },
  methods: {
    validate() {
      this.errors = {}
      const result = registerSchema.safeParse(this.form)

      if (!result.success) {
        const fieldErrors = result.error.flatten().fieldErrors
        this.errors = {
          email: fieldErrors.email?.[0],
          password: fieldErrors.password?.[0],
          confirmPassword: fieldErrors.confirmPassword?.[0],
        }
        return false
      }
      return true
    },
    async handleRegister() {
      if (!this.validate()) return
      await this.register(this.form.email, this.form.password)
    },
  },
}
</script>
