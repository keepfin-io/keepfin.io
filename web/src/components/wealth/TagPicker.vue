<template>
  <Popover v-model:open="open">
    <PopoverTrigger as-child>
      <Button :data-testid="`${testid}-toggle`" :variant="variant" :size="size" :disabled="disabled">
        <LuTag class="size-4" aria-hidden="true" />{{ t('wealth.tags') }}
      </Button>
    </PopoverTrigger>
    <PopoverContent :data-testid="testid" :align="align" class="w-72 p-0">
      <div class="border-b p-2">
        <Input
          v-model="query"
          :data-testid="`${testid}-search`"
          :aria-label="t('wealth.tagSearch')"
          :placeholder="t('wealth.tagSearch')"
          maxlength="40"
          class="h-9"
          @keydown.enter.prevent="submitQuery"
        />
      </div>
      <div class="max-h-60 overflow-y-auto p-1" role="group" :aria-label="t('wealth.tags')">
        <p v-if="loading" class="px-2 py-3 text-center text-xs text-muted-foreground">{{ t('common.loading') }}</p>
        <template v-else>
        <button
          v-for="tag in filtered"
          :key="tag.id"
          type="button"
          role="checkbox"
          :aria-checked="selected.has(tag.id)"
          :data-testid="`${testid}-option-${tag.id}`"
          :disabled="busy"
          class="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm hover:bg-muted focus-visible:bg-muted focus-visible:outline-none disabled:opacity-50"
          @click="toggle(tag)"
        >
          <span class="size-2.5 shrink-0 rounded-full" :class="tagStyle(tag.color).dot" aria-hidden="true" />
          <span class="min-w-0 flex-1 truncate">{{ tag.name }}</span>
          <LuCheck v-if="selected.has(tag.id)" class="size-4 shrink-0 text-primary" aria-hidden="true" />
        </button>
        </template>
        <p v-if="!loading && !filtered.length && !canCreate" class="px-2 py-3 text-center text-xs text-muted-foreground">{{ t('wealth.noTagsYet') }}</p>
      </div>
      <div v-if="canCreate" class="border-t p-2">
        <Button :data-testid="`${testid}-create`" size="sm" class="w-full" :disabled="busy" @click="openManager(query.trim())">
          <LuPlus class="size-4" aria-hidden="true" /><span class="truncate">{{ t('wealth.createTag', { name: query.trim() }) }}</span>
        </Button>
      </div>
      <p v-if="error" :data-testid="`${testid}-error`" role="alert" class="border-t px-3 py-2 text-xs text-destructive">{{ error }}</p>
      <div class="border-t p-1">
        <Button :data-testid="`${testid}-manage`" variant="ghost" size="sm" class="w-full justify-start text-xs text-muted-foreground" @click="openManager('')">
          <LuSettings2 class="size-3.5" aria-hidden="true" />{{ t('wealth.manageTags') }}
        </Button>
      </div>
    </PopoverContent>
  </Popover>
  <Dialog v-model:open="managerOpen">
    <DialogContent :data-testid="`${testid}-manager-dialog`" class="max-h-[90vh] overflow-y-auto rounded-2xl sm:max-w-lg">
      <DialogHeader class="pr-6 text-left">
        <DialogTitle>{{ t('wealth.manageTags') }}</DialogTitle>
        <DialogDescription>{{ t('wealth.manageTagsHint') }}</DialogDescription>
      </DialogHeader>
      <TagManager
        v-if="managerOpen"
        :initial-name="pendingName"
        autofocus
        :testid="`${testid}-manager`"
        @created="onCreated"
        @updated="onUpdated"
        @deleted="onDeleted"
      />
    </DialogContent>
  </Dialog>
</template>
<script setup>
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { LuCheck, LuPlus, LuSettings2, LuTag } from '@kalimahapps/vue-icons/lu'
import { Button, Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, Input, Popover, PopoverContent, PopoverTrigger } from '@/components/ui'
import TagManager from './TagManager.vue'
import { tagStyle } from '@/lib/tags'
import { wealthAPI } from '@/services/wealthApi'

// Assigns tags to one account or activity. Every change is saved at once.
// Creating and editing tags happens in a dialog on the same page; a tag
// created here is assigned to this account or activity right away.
const props = defineProps({
  kind: { type: String, required: true },
  id: { type: String, required: true },
  modelValue: { type: Array, default: () => [] },
  testid: { type: String, default: 'tag-picker' },
  variant: { type: String, default: 'ghost' },
  size: { type: String, default: 'default' },
  align: { type: String, default: 'end' },
  disabled: Boolean,
})
const emit = defineEmits(['update:modelValue'])
const { t } = useI18n()
const open = ref(false), query = ref(''), all = ref([]), loading = ref(false), busy = ref(false), error = ref('')
const managerOpen = ref(false), pendingName = ref('')
let loaded = false

const byName = (a, b) => a.name.localeCompare(b.name)
const normalized = computed(() => query.value.trim().toLocaleLowerCase())
const selected = computed(() => new Set(props.modelValue.map(tag => tag.id)))
const filtered = computed(() => all.value.filter(tag => !normalized.value || tag.name.toLocaleLowerCase().includes(normalized.value)))
const exact = computed(() => all.value.find(tag => tag.name.toLocaleLowerCase() === normalized.value))
const canCreate = computed(() => !!normalized.value && !exact.value && !loading.value)

watch(open, value => {
  if (value && !loaded) loadTags()
  if (!value) { query.value = ''; error.value = '' }
})

async function loadTags() {
  loading.value = true
  error.value = ''
  try {
    all.value = [...(await wealthAPI.get('/tags')).data].sort(byName)
    loaded = true
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
async function assign(tags) {
  busy.value = true
  error.value = ''
  try {
    await wealthAPI.save(`/tags/${props.kind}/${props.id}`, { tag_ids: tags.map(tag => tag.id) }, 'PUT')
    emit('update:modelValue', [...tags].sort(byName))
  } catch (e) {
    error.value = e.message
  } finally {
    busy.value = false
  }
}
function toggle(tag) {
  if (busy.value) return
  return assign(selected.value.has(tag.id) ? props.modelValue.filter(item => item.id !== tag.id) : [...props.modelValue, tag])
}
function openManager(name) {
  pendingName.value = name
  open.value = false
  managerOpen.value = true
}
async function onCreated(tag) {
  all.value = [...all.value, tag].sort(byName)
  managerOpen.value = false
  await assign([...props.modelValue, tag])
}
function onUpdated(tag) {
  all.value = all.value.map(item => (item.id === tag.id ? tag : item)).sort(byName)
  if (selected.value.has(tag.id)) emit('update:modelValue', props.modelValue.map(item => (item.id === tag.id ? tag : item)).sort(byName))
}
function onDeleted(id) {
  all.value = all.value.filter(item => item.id !== id)
  // Deleting a tag removes its assignments on the server as well.
  if (selected.value.has(id)) emit('update:modelValue', props.modelValue.filter(item => item.id !== id))
}
function submitQuery() {
  if (exact.value) toggle(exact.value)
  else if (canCreate.value) openManager(query.value.trim())
}
</script>
