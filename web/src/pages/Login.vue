<template>
  <div class="flex min-h-screen items-center justify-center bg-gray-50">
    <form @submit.prevent="handleLogin" class="w-full max-w-sm bg-white p-8 rounded-lg shadow-md space-y-6">
      <h2 class="text-2xl font-bold text-center mb-4">Login to KeepNode</h2>

      <!-- Error Message -->
      <div v-if="error" class="p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
        {{ error }}
      </div>

      <div>
        <label for="email" class="block text-sm font-medium text-gray-700 mb-1">Email</label>
        <input id="email" type="email" v-model="form.email" :disabled="isLoading"
          class="block w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100"
          :class="{ 'border-red-500': errors.email }" placeholder="you@example.com" />
        <p v-if="errors.email" class="mt-1 text-sm text-red-600">{{ errors.email }}</p>
      </div>

      <div>
        <label for="password" class="block text-sm font-medium text-gray-700 mb-1">Password</label>
        <input id="password" type="password" v-model="form.password" :disabled="isLoading"
          class="block w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100"
          :class="{ 'border-red-500': errors.password }" placeholder="••••••••" />
        <p v-if="errors.password" class="mt-1 text-sm text-red-600">{{ errors.password }}</p>
      </div>

      <button type="submit" :disabled="isLoading"
        class="w-full py-2 px-4 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700 transition disabled:bg-blue-300 disabled:cursor-not-allowed">
        {{ isLoading ? 'Logging in...' : 'Login' }}
      </button>

      <div class="text-center text-sm text-gray-500">
        Don't have an account?
        <router-link to="/register" class="text-blue-600 hover:underline">Register</router-link>
      </div>
    </form>
  </div>
</template>

<script>
import { useAuth } from '../composables/useAuth'
import { loginSchema } from '@/schemas/auth'

export default {
  name: "LoginPage",
  setup() {
    const { login, isLoading, error } = useAuth()
    return { login, isLoading, error }
  },
  data() {
    return {
      form: {
        email: "",
        password: ""
      },
      errors: {}
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
          password: fieldErrors.password?.[0]
        }
        return false
      }
      return true
    },
    async handleLogin() {
      if (!this.validate()) return
      await this.login(this.form.email, this.form.password)
    }
  }
}
</script>

<style scoped></style>
