<template>
  <main class="flex items-center justify-center min-h-screen">
    <div class="text-center">
      <h1 class="text-3xl font-bold mb-4">Dashboard</h1>
      <p class="text-gray-600 mb-8">Welcome to your dashboard!</p>
      <router-link to="/login" class="px-4 py-2 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700 transition">Logout</router-link>
    </div>
  </main>
</template>

<script>
export default {
  name: "DashboardPage"
};
</script>

<style scoped>
</style>
