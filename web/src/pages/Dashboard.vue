<template>
  <section data-testid="dashboard-page" class="space-y-6">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">
        {{ $t('dashboard.title') }}
      </h1>
      <p class="mt-1 text-sm text-muted-foreground">
        {{ $t('dashboard.welcome') }}
      </p>
    </div>
    <Card data-testid="dashboard-card">
      <CardHeader>
        <Wallet class="mb-2 size-8 text-muted-foreground" />
        <CardTitle>{{ $t('accounts.title') }}</CardTitle>
        <CardDescription>{{ $t('accounts.subtitle') }}</CardDescription>
      </CardHeader>
      <CardFooter>
        <Button as-child class="w-full sm:w-auto">
          <router-link data-testid="dashboard-accounts-link" to="/accounts">
            {{ $t('nav.accounts') }}
            <ArrowRight class="size-4" />
          </router-link>
        </Button>
      </CardFooter>
    </Card>
  </section>
</template>
<script>
import {
  Button,
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui'
import { ArrowRight } from 'lucide-vue-next'
import { Wallet } from 'lucide-vue-next'
export default {
  components: {
    ArrowRight,
    Button,
    Card,
    CardDescription,
    CardFooter,
    CardHeader,
    CardTitle,
    Wallet,
  },
  name: 'DashboardPage',
}
</script>
