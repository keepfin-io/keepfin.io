<template>
  <section class="space-y-6" data-testid="settings-page"><h1 class="text-2xl font-semibold">{{ t('settings.title') }}</h1><PageFeedback :error="error" :busy="busy" :notice="notice" />
    <Card data-testid="settings-card"><CardContent class="space-y-4 pt-6"><h2 class="font-semibold">{{ t('wealth.display') }}</h2><div class="flex items-center gap-3"><ZenToggle /><span class="text-sm">{{ t('wealth.zenHint') }}</span></div><Button v-if="installPrompt" variant="outline" @click="run(installApp)">{{ t('wealth.install') }}</Button><p v-else class="text-xs text-muted-foreground">{{ t('wealth.installHint') }}</p></CardContent></Card>
    <Card data-testid="settings-tags"><CardContent class="space-y-4 pt-6"><h2 class="font-semibold">{{ t('wealth.tags') }}</h2><TagManager testid="settings-tag" /></CardContent></Card>
    <div class="flex flex-wrap gap-3"><Button v-for="page in ['risk','fire','sharing','data']" :key="page" as-child variant="outline"><router-link :to="`/wealth/${page}`">{{ t(`wealth.${page === 'data' ? 'marketData' : page}`) }}</router-link></Button></div>
  </section>
</template>
<script setup>
import { Button, Card, CardContent } from '@/components/ui'
import PageFeedback from '@/components/wealth/PageFeedback.vue'
import ZenToggle from '@/components/wealth/ZenToggle.vue'
import TagManager from '@/components/wealth/TagManager.vue'
import { useWealthPage } from '@/composables/useWealthPage'
import { installPrompt, installApp } from '@/composables/useInstall'
const { t, busy, error, notice, run } = useWealthPage()
</script>
