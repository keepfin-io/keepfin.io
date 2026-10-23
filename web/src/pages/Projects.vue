<template>
  <section data-testid="projects-page" class="space-y-6">
    <h1 class="text-2xl font-semibold tracking-tight">
      {{ $t('projects.title') }}
    </h1>
    <Card data-testid="projects-card">
      <CardContent class="flex flex-col items-center gap-4 py-12 text-center">
        <Folder class="size-10 text-muted-foreground" />
        <p class="text-sm text-muted-foreground">
          {{ $t('projects.comingSoon') }}
        </p>
      </CardContent>
    </Card>
  </section>
</template>
<script>
import { Card, CardContent } from '@/components/ui'
import { Folder } from 'lucide-vue-next'
export default {
  components: { Card, CardContent, Folder },
  name: 'ProjectsPage',
}
</script>
