<template>
  <div class="flex min-h-screen items-center justify-center bg-gray-50">
    <form class="w-full max-w-sm bg-white p-8 rounded-lg shadow-md space-y-6">
      <h2 class="text-2xl font-bold text-center mb-4">Login</h2>
      <div>
        <label for="email" class="block text-sm font-medium text-gray-700 mb-1">Email</label>
        <input id="email" type="email" v-model="email" class="block w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="you@example.com" />
      </div>
      <div>
        <label for="password" class="block text-sm font-medium text-gray-700 mb-1">Password</label>
        <input id="password" type="password" v-model="password" class="block w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="••••••••" />
      </div>
      <button type="submit" class="w-full py-2 px-4 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700 transition">Login</button>
      <div class="text-center text-sm text-gray-500">
        Don't have an account?
        <router-link to="/register" class="text-blue-600 hover:underline">Register</router-link>
      </div>
    </form>
  </div>
</template>

<script>
export default {
  name: "LoginPage",
  data() {
    return {
      email: "",
      password: ""
    };
  },
  methods: {
    // Add login logic here
  }
};
</script>

<style scoped>
</style>
