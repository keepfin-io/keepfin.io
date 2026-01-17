<template>
  <div class="flex min-h-screen items-center justify-center bg-[#0f0f12]">
    <form @submit.prevent="handleRegister"
      class="w-full max-w-sm bg-[#1a1a1f] p-8 rounded-lg shadow-xl border border-[#2a2a30] space-y-6">
      <h2 class="text-2xl font-bold text-center mb-4 text-white">{{ $t('auth.registerTitle') }}</h2>

      <!-- Error Message -->
      <div v-if="error" class="p-3 bg-red-900/30 border border-red-800 rounded-lg text-red-300 text-sm">
        {{ error }}
      </div>

      <div>
        <label for="email" class="block text-sm font-medium text-gray-300 mb-1">{{ $t('auth.email') }}</label>
        <input id="email" type="email" v-model="form.email" :disabled="isLoading"
          class="block w-full px-4 py-2 border border-[#2a2a30] rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 bg-[#0f0f12] text-white placeholder-gray-500 disabled:bg-[#1a1a1f]"
          :class="{ 'border-red-500': errors.email }" placeholder="you@example.com" />
        <p v-if="errors.email" class="mt-1 text-sm text-red-400">{{ errors.email }}</p>
      </div>

      <div>
        <label for="password" class="block text-sm font-medium text-gray-300 mb-1">{{ $t('auth.password') }}</label>
        <input id="password" type="password" v-model="form.password" :disabled="isLoading"
          class="block w-full px-4 py-2 border border-[#2a2a30] rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 bg-[#0f0f12] text-white placeholder-gray-500 disabled:bg-[#1a1a1f]"
          :class="{ 'border-red-500': errors.password }" placeholder="••••••••" />
        <p v-if="errors.password" class="mt-1 text-sm text-red-400">{{ errors.password }}</p>
      </div>

      <div>
        <label for="confirmPassword" class="block text-sm font-medium text-gray-300 mb-1">{{ $t('auth.confirmPassword')
          }}</label>
        <input id="confirmPassword" type="password" v-model="form.confirmPassword" :disabled="isLoading"
          class="block w-full px-4 py-2 border border-[#2a2a30] rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 bg-[#0f0f12] text-white placeholder-gray-500 disabled:bg-[#1a1a1f]"
          :class="{ 'border-red-500': errors.confirmPassword }" placeholder="••••••••" />
        <p v-if="errors.confirmPassword" class="mt-1 text-sm text-red-400">{{ errors.confirmPassword }}</p>
      </div>

      <!-- Password Requirements -->
      <div class="text-xs text-gray-400 space-y-1">
        <p class="font-medium">Password must contain:</p>
        <ul class="list-disc list-inside space-y-0.5">
          <li :class="{ 'text-green-400': hasMinLength }">At least 8 characters</li>
          <li :class="{ 'text-green-400': hasUppercase }">One uppercase letter</li>
          <li :class="{ 'text-green-400': hasLowercase }">One lowercase letter</li>
          <li :class="{ 'text-green-400': hasNumber }">One number</li>
        </ul>
      </div>

      <button type="submit" :disabled="isLoading"
        class="w-full py-2 px-4 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700 transition disabled:bg-blue-300 disabled:cursor-not-allowed">
        {{ isLoading ? $t('common.registering') : $t('auth.register') }}
      </button>

      <div class="text-center text-sm text-gray-400">
        {{ $t('auth.hasAccount') }}
        <router-link to="/login" class="text-blue-400 hover:underline">{{ $t('auth.login') }}</router-link>
      </div>
    </form>
  </div>
</template>

<script>
import { useAuth } from '../composables/useAuth'
import { registerSchema } from '@/schemas/auth'

export default {
  name: "RegisterPage",
  setup() {
    const { register, isLoading, error } = useAuth()
    return { register, isLoading, error }
  },
  data() {
    return {
      form: {
        email: "",
        password: "",
        confirmPassword: ""
      },
      errors: {}
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
    }
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
          confirmPassword: fieldErrors.confirmPassword?.[0]
        }
        return false
      }
      return true
    },
    async handleRegister() {
      if (!this.validate()) return
      await this.register(this.form.email, this.form.password)
    }
  }
}
</script>

<style scoped></style>
