<template>
  <div :data-testid="testid" class="space-y-5">
    <form :data-testid="`${testid}-create`" class="space-y-3" @submit.prevent="create">
      <Label :for="`${testid}-name`">{{ t('wealth.newTag') }}</Label>
      <div class="flex flex-wrap gap-2">
        <Input
          :id="`${testid}-name`"
          ref="nameInput"
          v-model="name"
          :data-testid="`${testid}-name`"
          :placeholder="t('wealth.tagName')"
          maxlength="40"
          class="min-w-40 flex-1"
          :disabled="busy"
        />
        <Button type="submit" :data-testid="`${testid}-add`" :disabled="busy || !name.trim()">
          <LuPlus class="size-4" aria-hidden="true" />{{ t('wealth.addTag') }}
        </Button>
      </div>
      <ColorSwatches v-model="color" :disabled="busy" :testid="`${testid}-color`" />
      <p v-if="name.trim()" class="flex items-center gap-2 text-xs text-muted-foreground">
        {{ t('wealth.tagPreview') }} <TagBadge :tag="{ id: 'preview', name: name.trim(), color }" />
      </p>
    </form>
    <p v-if="error" :data-testid="`${testid}-error`" role="alert" class="text-sm text-destructive">{{ error }}</p>
    <section class="space-y-3" :aria-label="t('wealth.existingTags')">
      <h3 class="text-sm font-medium">{{ t('wealth.existingTags') }}</h3>
      <p v-if="loading" class="text-sm text-muted-foreground">{{ t('common.loading') }}</p>
      <p v-else-if="!tags.length" class="text-sm text-muted-foreground">{{ t('wealth.noTagsYet') }}</p>
      <template v-else>
      <div v-for="tag in tags" :key="tag.id" :data-testid="`${testid}-${tag.id}`" class="space-y-2 rounded-xl border p-3">
        <div class="flex flex-wrap items-center gap-2">
          <Input v-model="drafts[tag.id].name" maxlength="40" class="min-w-32 flex-1" :aria-label="t('wealth.tagName')" :disabled="busy" />
          <Button variant="outline" size="sm" :disabled="busy || !changed(tag) || !drafts[tag.id].name.trim()" @click="save(tag)">{{ t('common.save') }}</Button>
          <template v-if="deleting === tag.id">
            <Button :data-testid="`${testid}-${tag.id}-confirm-delete`" variant="destructive" size="sm" :disabled="busy" @click="remove(tag)">{{ t('common.delete') }}</Button>
            <Button variant="ghost" size="sm" :disabled="busy" @click="deleting = ''">{{ t('common.cancel') }}</Button>
          </template>
          <Button v-else :data-testid="`${testid}-${tag.id}-delete`" variant="ghost" size="icon" class="size-8" :disabled="busy" :aria-label="t('common.delete')" :title="t('common.delete')" @click="deleting = tag.id">
            <LuTrash2 class="size-4" aria-hidden="true" />
          </Button>
        </div>
        <ColorSwatches v-model="drafts[tag.id].color" :disabled="busy" :testid="`${testid}-${tag.id}-color`" />
        <p v-if="deleting === tag.id" class="text-xs text-muted-foreground">{{ t('wealth.deleteTagHint') }}</p>
      </div>
      </template>
    </section>
  </div>
</template>
<script setup>
import { nextTick, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { LuPlus, LuTrash2 } from '@kalimahapps/vue-icons/lu'
import { Button, Input, Label } from '@/components/ui'
import ColorSwatches from './ColorSwatches.vue'
import TagBadge from './TagBadge.vue'
import { wealthAPI } from '@/services/wealthApi'

// Creates, renames, recolors and deletes tags in place. Parents listen to the
// events to keep their own tag lists in step.
const props = defineProps({
  initialName: { type: String, default: '' },
  autofocus: Boolean,
  testid: { type: String, default: 'tag-manager' },
})
const emit = defineEmits(['created', 'updated', 'deleted'])
const { t } = useI18n()
const tags = ref([]), drafts = ref({}), name = ref(props.initialName), color = ref('blue')
const loading = ref(true), busy = ref(false), error = ref(''), deleting = ref(''), nameInput = ref(null)

const byName = (a, b) => a.name.localeCompare(b.name)
function setTags(rows) {
  tags.value = [...rows].sort(byName)
  drafts.value = Object.fromEntries(tags.value.map(tag => [tag.id, { name: tag.name, color: tag.color || 'gray' }]))
}
function changed(tag) {
  const draft = drafts.value[tag.id]
  return draft.name.trim() !== tag.name || draft.color !== (tag.color || 'gray')
}
async function attempt(action) {
  busy.value = true
  error.value = ''
  try {
    await action()
  } catch (e) {
    error.value = e.message
  } finally {
    busy.value = false
  }
}
function create() {
  if (!name.value.trim() || busy.value) return
  return attempt(async () => {
    const created = (await wealthAPI.save('/tags', { name: name.value.trim(), color: color.value })).data
    setTags([...tags.value, created])
    name.value = ''
    emit('created', created)
  })
}
function save(tag) {
  return attempt(async () => {
    const updated = (await wealthAPI.save(`/tags/${tag.id}`, { name: drafts.value[tag.id].name.trim(), color: drafts.value[tag.id].color }, 'PUT')).data
    setTags(tags.value.map(item => (item.id === tag.id ? updated : item)))
    emit('updated', updated)
  })
}
function remove(tag) {
  return attempt(async () => {
    await wealthAPI.remove(`/tags/${tag.id}`)
    deleting.value = ''
    setTags(tags.value.filter(item => item.id !== tag.id))
    emit('deleted', tag.id)
  })
}
onMounted(async () => {
  try {
    setTags((await wealthAPI.get('/tags')).data || [])
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
  if (props.autofocus) {
    await nextTick()
    nameInput.value?.$el?.focus?.()
  }
})
</script>
