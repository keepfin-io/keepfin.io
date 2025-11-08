<template>
  <div class="flex min-h-screen items-center justify-center bg-gray-50">
    <form @submit.prevent="handleRegister" class="w-full max-w-sm bg-white p-8 rounded-lg shadow-md space-y-6">
      <h2 class="text-2xl font-bold text-center mb-4">Register for KeepNode</h2>

      <!-- Error Message -->
      <div v-if="error" class="p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
        {{ error }}
      </div>

      <div>
        <label for="email" class="block text-sm font-medium text-gray-700 mb-1">Email</label>
        <input
          id="email"
          type="email"
          v-model="email"
          required
          :disabled="isLoading"
          class="block w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100"
          placeholder="you@example.com"
        />
      </div>

      <div>
        <label for="password" class="block text-sm font-medium text-gray-700 mb-1">Password</label>
        <input
          id="password"
          type="password"
          v-model="password"
          required
          :disabled="isLoading"
          class="block w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100"
          placeholder="••••••••"
        />
      </div>

      <button
        type="submit"
        :disabled="isLoading"
        class="w-full py-2 px-4 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700 transition disabled:bg-blue-300 disabled:cursor-not-allowed"
      >
        {{ isLoading ? 'Creating account...' : 'Register' }}
      </button>

      <div class="text-center text-sm text-gray-500">
        Already have an account?
        <router-link to="/login" class="text-blue-600 hover:underline">Login</router-link>
      </div>
    </form>
  </div>
</template>

<script>
import { useAuth } from '../composables/useAuth'

export default {
  name: "RegisterPage",
  setup() {
    const { register, isLoading, error } = useAuth()
    return { register, isLoading, error }
  },
  data() {
    return {
      email: "",
      password: ""
    }
  },
  methods: {
    async handleRegister() {
      await this.register(this.email, this.password)
    }
  }
}
</script>

<style scoped>
</style>
