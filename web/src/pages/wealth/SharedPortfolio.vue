<template>
  <main class="mx-auto w-full max-w-7xl space-y-6 p-5 sm:p-10" data-testid="shared-portfolio-page"><div class="flex items-center justify-between gap-4"><BrandLogo /><Button as-child variant="outline"><router-link to="/login">{{ $t('wealth.signInForPrivate') }}</router-link></Button></div><p class="text-sm text-muted-foreground">{{ $t('wealth.readOnlyShare') }}</p><PortfolioReport :report-loader="load" readonly /></main>
</template>
<script setup>
import { Button } from '@/components/ui'
import BrandLogo from '@/components/BrandLogo.vue'
import PortfolioReport from '@/components/portfolio/PortfolioReport.vue'
import { refreshSession } from '@/composables/useAuth'
import { wealthAPI } from '@/services/wealthApi'
import { readAPIResponse } from '@/services/api'
async function load(options) {
  const token = window.location.hash.slice(1)
  let authenticated = false
  try { authenticated = await refreshSession() } catch { /* Public links do not require an auth service. */ }
  if (authenticated) return wealthAPI.save('/shared/report', { token, range:options.range })
  const response = await fetch('/api/public/portfolio', { method:'POST', credentials:'same-origin', headers:{'Content-Type':'application/json'}, body:JSON.stringify({ token, range:options.range }) })
  return readAPIResponse(response)
}
</script>
