<template>
  <section class="min-w-0 space-y-6" data-testid="wealth-page">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">{{ $t('wealth.title') }}</h1>
      <p class="mt-1 text-sm text-muted-foreground">{{ $t('wealth.subtitle') }}</p>
    </div>
    <!-- Grouped so the tabs read as "track", "plan" and "manage" at a glance. -->
    <nav data-testid="wealth-nav" class="-mx-4 overflow-x-auto px-4 sm:mx-0 sm:px-0" :aria-label="$t('wealth.title')">
      <div class="flex w-max items-center gap-1 rounded-xl border bg-muted/40 p-1">
        <template v-for="(group, index) in groups" :key="group.key">
          <span v-if="index" class="mx-1 h-5 w-px bg-border" aria-hidden="true" />
          <router-link
            v-for="page in group.pages"
            :key="page.path"
            :data-testid="`wealth-nav-${page.path}`"
            :to="`/wealth/${page.path}`"
            class="inline-flex h-9 items-center gap-2 whitespace-nowrap rounded-lg px-3 text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            :class="isActive(page) ? 'bg-background text-foreground shadow-xs' : 'text-muted-foreground hover:bg-background/60 hover:text-foreground'"
            :aria-current="isActive(page) ? 'page' : undefined"
          >
            <component :is="page.icon" class="size-4" :class="isActive(page) && 'text-primary'" aria-hidden="true" />
            {{ $t(`wealth.${page.label}`) }}
          </router-link>
        </template>
      </div>
    </nav>
    <router-view />
  </section>
</template>
<script setup>
import { useRoute } from 'vue-router'
import { LuDatabase, LuEye, LuFlame, LuGem, LuListOrdered, LuScale, LuShare2, LuShieldCheck } from '@kalimahapps/vue-icons/lu'
const route = useRoute()
const groups = [
  { key: 'track', pages: [{ path: 'watchlist', label: 'watchlist', icon: LuEye }, { path: 'assets', label: 'manualAssets', icon: LuGem }, { path: 'net-worth', label: 'netWorth', icon: LuScale }, { path: 'activities', label: 'activities', icon: LuListOrdered }] },
  { key: 'plan', pages: [{ path: 'risk', label: 'risk', icon: LuShieldCheck }, { path: 'fire', label: 'fire', icon: LuFlame }] },
  { key: 'manage', pages: [{ path: 'sharing', label: 'sharing', icon: LuShare2 }, { path: 'data', label: 'marketData', icon: LuDatabase }] },
]
const isActive = page => route.path.endsWith(`/${page.path}`)
</script>
