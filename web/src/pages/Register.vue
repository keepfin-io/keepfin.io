<template>
  <div class="flex min-h-screen items-center justify-center bg-gray-50">
    <form class="w-full max-w-sm bg-white p-8 rounded-lg shadow-md space-y-6">
      <h2 class="text-2xl font-bold text-center mb-4">Register</h2>
      <div>
        <label for="name" class="block text-sm font-medium text-gray-700 mb-1">Name</label>
        <input id="name" type="text" v-model="name" class="block w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="Your Name" />
      </div>
      <div>
        <label for="email" class="block text-sm font-medium text-gray-700 mb-1">Email</label>
        <input id="email" type="email" v-model="email" class="block w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="you@example.com" />
      </div>
      <div>
        <label for="password" class="block text-sm font-medium text-gray-700 mb-1">Password</label>
        <input id="password" type="password" v-model="password" class="block w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="••••••••" />
      </div>
      <button type="submit" class="w-full py-2 px-4 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700 transition">Register</button>
      <div class="text-center text-sm text-gray-500">
        Already have an account?
        <router-link to="/login" class="text-blue-600 hover:underline">Login</router-link>
      </div>
    </form>
  </div>
</template>

<script>
export default {
  name: "RegisterPage",
  data() {
    return {
      name: "",
      email: "",
      password: ""
    };
  },
  methods: {
    // Add register logic here
  }
};
</script>

<style scoped>
</style>
