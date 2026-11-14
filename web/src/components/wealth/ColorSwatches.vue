<template>
  <div class="flex flex-wrap gap-1.5" role="radiogroup" :aria-label="t('wealth.tagColor')">
    <button
      v-for="color in TAG_COLORS"
      :key="color"
      type="button"
      role="radio"
      :data-testid="`${testid}-${color}`"
      :aria-checked="modelValue === color"
      :aria-label="t(`wealth.tagColors.${color}`)"
      :title="t(`wealth.tagColors.${color}`)"
      :disabled="disabled"
      class="flex size-6 items-center justify-center rounded-full ring-offset-2 ring-offset-background transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
      :class="[tagStyle(color).dot, modelValue === color && 'ring-2 ring-foreground']"
      @click="$emit('update:modelValue', color)"
    >
      <LuCheck v-if="modelValue === color" class="size-3.5 text-white" aria-hidden="true" />
    </button>
  </div>
</template>
<script setup>
import { useI18n } from 'vue-i18n'
import { LuCheck } from '@kalimahapps/vue-icons/lu'
import { TAG_COLORS, tagStyle } from '@/lib/tags'

defineProps({ modelValue: { type: String, default: 'gray' }, disabled: Boolean, testid: { type: String, default: 'tag-color' } })
defineEmits(['update:modelValue'])
const { t } = useI18n()
</script>
