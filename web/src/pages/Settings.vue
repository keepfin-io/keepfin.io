<template>
  <section class="space-y-6" data-testid="settings-page"><h1 class="text-2xl font-semibold">{{ t('settings.title') }}</h1><PageFeedback :error="error" :busy="busy" :notice="notice" />
    <Card data-testid="settings-card"><CardContent class="space-y-4 pt-6"><h2 class="font-semibold">{{ t('wealth.display') }}</h2><div class="flex items-center gap-3"><ZenToggle /><span class="text-sm">{{ t('wealth.zenHint') }}</span></div><Button v-if="installPrompt" variant="outline" @click="run(installApp)">{{ t('wealth.install') }}</Button><p v-else class="text-xs text-muted-foreground">{{ t('wealth.installHint') }}</p></CardContent></Card>
    <Card><CardContent class="space-y-4 pt-6"><h2 class="font-semibold">{{ t('wealth.tags') }}</h2><form class="flex flex-wrap gap-3" @submit.prevent="add"><label class="flex-1"><span class="sr-only">{{ t('wealth.tagName') }}</span><Input v-model="name" :placeholder="t('wealth.tagName')" required maxlength="40" :disabled="busy" /></label><Button type="submit" :disabled="busy">{{ t('wealth.addTag') }}</Button></form><div v-for="tag in tags" :key="tag.id" class="flex items-center gap-2"><label class="flex-1"><span class="sr-only">{{ t('wealth.tagName') }}</span><Input v-model="names[tag.id]" maxlength="40" :disabled="busy" /></label><Button variant="outline" size="sm" :disabled="busy || !names[tag.id]?.trim()" @click="rename(tag)">{{ t('common.save') }}</Button><Button variant="ghost" size="sm" :disabled="busy" @click="deleting = tag">{{ t('common.delete') }}</Button></div></CardContent></Card>
    <div class="flex flex-wrap gap-3"><Button v-for="page in ['risk','fire','sharing','data']" :key="page" as-child variant="outline"><router-link :to="`/wealth/${page}`">{{ t(`wealth.${page === 'data' ? 'marketData' : page}`) }}</router-link></Button></div>
    <InlineConfirmation :visible="!!deleting" :title="t('common.delete')" :message="t('wealth.deleteTagHint')" :confirm-text="t('common.delete')" :cancel-text="t('common.cancel')" :loading="busy" @cancel="deleting = null" @confirm="remove" />
  </section>
</template>
<script setup>
import { onMounted, ref } from 'vue'
import { Button, Input, Card, CardContent, InlineConfirmation } from '@/components/ui'
import PageFeedback from '@/components/wealth/PageFeedback.vue'
import ZenToggle from '@/components/wealth/ZenToggle.vue'
import { useWealthPage } from '@/composables/useWealthPage'
import { wealthAPI } from '@/services/wealthApi'
import { installPrompt, installApp } from '@/composables/useInstall'
const { t, busy, error, notice, run } = useWealthPage()
const tags = ref([]), name = ref(''), names = ref({}), deleting = ref(null)
async function load() { tags.value=(await wealthAPI.get('/tags')).data;names.value=Object.fromEntries(tags.value.map(t => [t.id,t.name])) }
function add(){run(async()=>{await wealthAPI.save('/tags',{name:name.value});name.value='';await load()},true)}
function rename(tag){run(async()=>{await wealthAPI.save(`/tags/${tag.id}`,{name:names.value[tag.id]},'PUT');await load()},true)}
function remove(){run(async()=>{await wealthAPI.remove(`/tags/${deleting.value.id}`);deleting.value=null;await load()},true)}
onMounted(()=>run(load))
</script>
