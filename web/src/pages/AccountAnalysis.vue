<template>
  <section data-testid="account-analysis-page" class="min-w-0 space-y-6">
    <Button as-child variant="ghost" class="px-0 text-muted-foreground">
      <router-link data-testid="account-analysis-back" :to="{ name: 'AccountDetail', params: { id: account.id } }"><ArrowLeft class="size-4" />{{ $t('accountDetail.backToAccount') }}</router-link>
    </Button>
    <PortfolioReport :account-id="account.id" :account-name="account.account_name" />
  </section>
</template>
<script setup>
import { ArrowLeft } from 'lucide-vue-next'
import { Button } from '@/components/ui'
import PortfolioReport from '@/components/portfolio/PortfolioReport.vue'
defineOptions({ inheritAttrs: false })
defineProps({ account: { type: Object, required: true } })
</script>
