<template>
  <section data-testid="settings-page" class="space-y-6">
    <h1 class="text-2xl font-semibold tracking-tight">
      {{ $t('settings.title') }}
    </h1>
    <Card data-testid="settings-card">
      <CardContent class="flex flex-col items-center gap-4 py-12 text-center">
        <Settings class="size-10 text-muted-foreground" />
        <p class="text-sm text-muted-foreground">
          {{ $t('settings.comingSoon') }}
        </p>
      </CardContent>
    </Card>
  </section>
</template>
<script>
import { Card, CardContent } from '@/components/ui'
import { Settings } from 'lucide-vue-next'
export default {
  components: { Card, CardContent, Settings },
  name: 'SettingsPage',
}
</script>
