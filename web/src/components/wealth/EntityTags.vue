<template>
  <div class="space-y-3 rounded-xl border p-4" data-testid="entity-tags">
    <PageFeedback :error="error" :busy="busy" :notice="notice" />
    <div class="flex flex-wrap gap-3"><label v-for="tag in tags" :key="tag.id" class="flex items-center gap-2 text-sm"><input v-model="selected" type="checkbox" :value="tag.id" :disabled="busy">{{ tag.name }}</label></div>
    <div class="flex flex-wrap gap-2"><Button size="sm" :disabled="busy" @click="save">{{ t('wealth.saveTags') }}</Button><Button variant="ghost" size="sm" as-child><router-link to="/settings">{{ t('wealth.manageTags') }}</router-link></Button></div>
  </div>
</template>
<script setup>
import { onMounted, ref } from 'vue'
import { Button } from '@/components/ui'
import PageFeedback from './PageFeedback.vue'
import { wealthAPI } from '@/services/wealthApi'
import { useWealthPage } from '@/composables/useWealthPage'
const props = defineProps({ kind: { type: String, required: true }, id: { type: String, required: true } })
const emit = defineEmits(['saved'])
const { t, busy, error, notice, run } = useWealthPage()
const tags = ref([]), selected = ref([])
onMounted(() => run(async () => { const [all, current] = await Promise.all([wealthAPI.get('/tags'), wealthAPI.get(`/tags/${props.kind}/${props.id}`)]); tags.value = all.data; selected.value = current.data.map(t => t.id) }))
function save() { run(async () => { await wealthAPI.save(`/tags/${props.kind}/${props.id}`, { tag_ids: selected.value }, 'PUT'); emit('saved') }, true) }
</script>
