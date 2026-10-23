<template>
  <div class="flex min-h-svh items-center justify-center p-4 sm:p-8">
    <Card class="w-full max-w-md">
      <CardHeader class="text-center">
        <CardTitle class="text-2xl">{{ $t('auth.loginTitle') }}</CardTitle>
      </CardHeader>
      <CardContent>
        <form class="space-y-5" @submit.prevent="handleLogin">
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
              autocomplete="current-password"
              :disabled="isLoading"
              :aria-invalid="!!errors.password"
            />
            <p v-if="errors.password" class="text-sm text-destructive">
              {{ errors.password }}
            </p>
          </div>
          <Button
            type="submit"
            class="w-full"
            :loading="isLoading"
            :loading-text="$t('common.loggingIn')"
          >
            {{ $t('auth.login') }}
          </Button>
        </form>
      </CardContent>
      <CardFooter
        class="flex-wrap justify-center gap-1 text-sm text-muted-foreground"
      >
        {{ $t('auth.noAccount') }}
        <Button variant="link" size="inline" as-child>
          <router-link to="/register">{{ $t('auth.register') }}</router-link>
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
import { loginSchema } from '@/schemas/auth'

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
  name: 'LoginPage',
  setup() {
    const { login, isLoading, error } = useAuth()
    return { login, isLoading, error }
  },
  data() {
    return {
      form: {
        email: '',
        password: '',
      },
      errors: {},
    }
  },
  methods: {
    validate() {
      this.errors = {}
      const result = loginSchema.safeParse(this.form)

      if (!result.success) {
        const fieldErrors = result.error.flatten().fieldErrors
        this.errors = {
          email: fieldErrors.email?.[0],
          password: fieldErrors.password?.[0],
        }
        return false
      }
      return true
    },
    async handleLogin() {
      if (!this.validate()) return
      await this.login(this.form.email, this.form.password)
    },
  },
}
</script>
