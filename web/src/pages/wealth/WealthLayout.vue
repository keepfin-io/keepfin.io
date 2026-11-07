<template>
  <section class="space-y-6" data-testid="wealth-page">
    <div><h1 class="text-2xl font-semibold tracking-tight">{{ $t('wealth.title') }}</h1><p class="mt-2 text-sm text-muted-foreground">{{ $t('wealth.subtitle') }}</p></div>
    <nav class="flex flex-wrap gap-1 border-b pb-3" :aria-label="$t('wealth.title')">
      <Button v-for="page in pages" :key="page.path" as-child size="sm" :variant="$route.path.endsWith(page.path) ? 'secondary' : 'ghost'"><router-link :to="`/wealth/${page.path}`">{{ $t(`wealth.${page.label}`) }}</router-link></Button>
    </nav>
    <router-view />
  </section>
</template>
<script setup>
import { Button } from '@/components/ui'
const pages = [{ path: 'watchlist', label: 'watchlist' }, { path: 'assets', label: 'manualAssets' }, { path: 'net-worth', label: 'netWorth' }, { path: 'activities', label: 'activities' }, { path: 'risk', label: 'risk' }, { path: 'fire', label: 'fire' }, { path: 'sharing', label: 'sharing' }, { path: 'data', label: 'marketData' }]
</script>
