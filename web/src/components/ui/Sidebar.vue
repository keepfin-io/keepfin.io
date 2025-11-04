<template>
  <div>
    <!-- Mobile toggle button -->
    <button @click="toggleSidebar" class="md:hidden fixed top-4 left-4 z-50 bg-blue-600 text-white p-2 rounded-lg shadow">
      <span class="material-icons">menu</span>
    </button>
    <transition name="slide">
      <aside v-if="isOpen" class="fixed md:static top-0 left-0 h-screen w-64 bg-white shadow-md flex flex-col z-40">
        <div class="px-4 py-2 border-b flex justify-between items-center bg-gray-100 h-14">
          <div class="flex items-center gap-2">
            <img src="/favicon.ico" alt="Logo" class="w-7 h-7 rounded" />
            <span class="text-base font-bold text-gray-700">KeepNode</span>
          </div>
          <button @click="toggleSidebar" class="md:hidden text-gray-500 hover:text-gray-700">
            <span class="material-icons">close</span>
          </button>
        </div>
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
        <!-- Sidebar Footer with User Dropdown -->
        <div class="p-4 border-t mt-auto">
          <div class="relative" @click="dropdownOpen = !dropdownOpen">
            <button class="flex items-center gap-2 w-full px-4 py-2 rounded hover:bg-gray-100">
              <span class="material-icons">person</span>
              <span>Username</span>
              <span class="material-icons ml-auto">expand_more</span>
            </button>
            <transition name="fade">
              <div v-if="dropdownOpen" class="absolute left-0 bottom-12 w-full bg-white border rounded shadow-lg z-50">
                  <div class="px-4 py-2 border-b flex justify-between items-center bg-gray-100 h-14">
                    <div class="flex items-center gap-2">
                      <img src="/favicon.ico" alt="Logo" class="w-7 h-7 rounded" />
                      <span class="text-base font-bold text-gray-700">KeepNode</span>
                    </div>
                    <button @click="toggleSidebar" class="md:hidden text-gray-500 hover:text-gray-700">
                      <span class="material-icons">close</span>
                    </button>
                  </div>
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
                  <!-- Sidebar Footer with User Dropdown -->
                  <div class="p-4 border-t mt-auto">
                    <div class="relative" @click="dropdownOpen = !dropdownOpen">
                      <button class="flex items-center gap-2 w-full px-4 py-2 rounded hover:bg-gray-100">
                        <span class="material-icons">person</span>
                        <span>Username</span>
                        <span class="material-icons ml-auto">expand_more</span>
                      </button>
                      <transition name="fade">
                        <div v-if="dropdownOpen" class="absolute left-0 bottom-12 w-full bg-white border rounded shadow-lg z-50">
                          <button class="block w-full text-left px-4 py-2 hover:bg-gray-100">Account</button>
                          <button class="block w-full text-left px-4 py-2 hover:bg-gray-100">Billing</button>
                          <button class="block w-full text-left px-4 py-2 hover:bg-gray-100">Sign out</button>
                        </div>
                      </transition>
                    </div>
                  </div>
              </div>
            </transition>
          </div>
        </div>
      </aside>
    </transition>
  </div>
</template>


<script>
export default {
  name: "SidebarMenu",
  data() {
    return {
      isOpen: window.innerWidth >= 768,
      dropdownOpen: false
    };
  },
  mounted() {
    window.addEventListener('resize', this.handleResize);
    this.handleResize();
  },
  beforeUnmount() {
    window.removeEventListener('resize', this.handleResize);
  },
  methods: {
    toggleSidebar() {
      this.isOpen = !this.isOpen;
    },
    handleResize() {
      if (window.innerWidth >= 768) {
        this.isOpen = true;
      } else {
        // Do not force open on mobile, keep current state
      }
    }
  }
};
</script>

<style scoped>
/* Material icons size */
.material-icons {
  font-size: 20px;
}
.slide-enter-active, .slide-leave-active {
  transition: transform 0.3s;
}
.slide-enter-from {
  transform: translateX(-100%);
}
.slide-leave-to {
  transform: translateX(-100%);
}
.fade-enter-active, .fade-leave-active {
  transition: opacity 0.2s;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}
@media (min-width: 768px) {
  aside {
    position: static !important;
    transform: none !important;
    z-index: 0 !important;
  }
}
</style>
