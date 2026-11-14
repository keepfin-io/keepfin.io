<template>
  <span class="inline-flex items-center gap-2" :class="$attrs.class">
    <Button
      v-bind="buttonAttrs"
      type="button"
      variant="ghost"
      size="icon"
      class="size-7 shrink-0 text-muted-foreground hover:text-foreground"
      :disabled="busy"
      :aria-pressed="zen"
      :aria-label="$t(zen ? 'wealth.showValues' : 'wealth.hideValues')"
      :title="$t(zen ? 'wealth.showValues' : 'wealth.hideValues')"
      @click="toggle"
    >
      <LuEyeOff v-if="zen" class="size-4" aria-hidden="true" /><LuEye v-else class="size-4" aria-hidden="true" />
    </Button>
    <span v-if="error" role="alert" class="max-w-48 text-xs text-destructive">{{ error }}</span>
  </span>
</template>
<script setup>
import { computed, ref, useAttrs } from 'vue'
import { LuEye, LuEyeOff } from '@kalimahapps/vue-icons/lu'
import { Button } from '@/components/ui'
import { zen, toggleZen } from '@/composables/useZen'
// Hides money amounts everywhere. It sits next to the values it hides.
defineOptions({ inheritAttrs: false })
const attrs = useAttrs()
// Layout classes place the toggle; everything else (test ids, labels) is the button's.
const buttonAttrs = computed(() => Object.fromEntries(Object.entries(attrs).filter(([key]) => key !== 'class')))
const busy = ref(false), error = ref('')
async function toggle() {
  if (busy.value) return
  busy.value = true; error.value = ''
  try { await toggleZen() } catch (e) { error.value = e.message } finally { busy.value = false }
}
</script>
