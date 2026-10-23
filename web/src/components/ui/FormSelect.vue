<template>
  <Select
    :model-value="modelValue || undefined"
    :disabled="disabled"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <SelectTrigger
      v-bind="$attrs"
      :id="id"
      class="w-full"
      :aria-invalid="invalid"
    >
      <SelectValue :placeholder="placeholder" />
    </SelectTrigger>
    <SelectContent
      :data-testid="$attrs['data-testid'] && `${$attrs['data-testid']}-options`"
    >
      <SelectItem
        :data-testid="
          $attrs['data-testid'] &&
          `${$attrs['data-testid']}-option-${
            typeof option === 'string' ? option : option.value
          }`
        "
        v-for="option in options"
        :key="typeof option === 'string' ? option : option.value"
        :value="typeof option === 'string' ? option : option.value"
      >
        {{ typeof option === 'string' ? option : option.label }}
      </SelectItem>
    </SelectContent>
  </Select>
</template>
<script setup>
defineOptions({ inheritAttrs: false })
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
defineProps({
  modelValue: String,
  id: String,
  options: { type: Array, required: true },
  placeholder: String,
  disabled: Boolean,
  invalid: Boolean,
})
defineEmits(['update:modelValue'])
</script>
