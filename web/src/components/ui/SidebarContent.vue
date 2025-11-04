<template>
  <nav class="flex-1 p-4">
    <ul class="space-y-2">
      <li>
        <router-link to="/" class="flex items-center gap-2 px-4 py-2 w-full text-left rounded hover:bg-gray-100">
          <span class="material-icons">dashboard</span>
          <span>Dashboard</span>
        </router-link>
      </li>
      <li>
        <router-link to="/projects" class="flex items-center gap-2 px-4 py-2 w-full text-left rounded hover:bg-gray-100">
          <span class="material-icons">folder</span>
          <span>Projects</span>
        </router-link>
      </li>
      <li>
        <router-link to="/settings" class="flex items-center gap-2 px-4 py-2 w-full text-left rounded hover:bg-gray-100">
          <span class="material-icons">settings</span>
          <span>Settings</span>
        </router-link>
      </li>
    </ul>
  </nav>
</template>

<script>
export default {
  name: "SidebarContent"
};
</script>
