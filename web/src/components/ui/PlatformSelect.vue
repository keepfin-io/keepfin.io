<template>
  <Popover v-model:open="open">
    <PopoverTrigger as-child>
      <Button
        v-bind="$attrs"
        :id="id"
        type="button"
        variant="outline"
        role="combobox"
        :aria-expanded="open"
        class="w-full min-w-0 justify-between font-normal"
      >
        <span class="truncate">
          {{ modelValue || $t('accounts.selectPlatform') }}
        </span>
        <ChevronsUpDown class="size-4 shrink-0 opacity-50" />
      </Button>
    </PopoverTrigger>
    <PopoverContent
      :data-testid="$attrs['data-testid'] && `${$attrs['data-testid']}-options`"
      class="w-[var(--reka-popover-trigger-width)] p-0"
      align="start"
    >
      <Command>
        <CommandInput
          :data-testid="
            $attrs['data-testid'] && `${$attrs['data-testid']}-search`
          "
          :placeholder="$t('accounts.searchPlatform')"
        />
        <CommandList>
          <CommandEmpty
            :data-testid="
              $attrs['data-testid'] && `${$attrs['data-testid']}-empty`
            "
          >
            {{ $t('common.noResults') }}
          </CommandEmpty>
          <CommandGroup>
            <CommandItem
              :data-testid="
                $attrs['data-testid'] && `${$attrs['data-testid']}-option-none`
              "
              value="__none__"
              @select="select('')"
            >
              {{ $t('accounts.general') }}
            </CommandItem>
          </CommandGroup>
          <CommandGroup
            v-for="group in platforms"
            :key="group.label"
            :heading="group.label"
          >
            <CommandItem
              :data-testid="
                $attrs['data-testid'] &&
                `${$attrs['data-testid']}-option-${item.domain}`
              "
              v-for="item in group.options"
              :key="item.name"
              :value="item.name"
              @select="select(item.name)"
            >
              <Favicon :domain="item.domain" :size="16" />
              <span class="truncate">{{ item.name }}</span>
              <Check v-if="modelValue === item.name" class="ml-auto size-4" />
            </CommandItem>
          </CommandGroup>
        </CommandList>
      </Command>
    </PopoverContent>
  </Popover>
</template>
<script setup>
defineOptions({ inheritAttrs: false })
import { ref } from 'vue'
import { PLATFORM_GROUPS } from '@/lib/platforms'
import { Button } from '@/components/ui/button'
import {
  Popover,
  PopoverTrigger,
  PopoverContent,
} from '@/components/ui/popover'
import {
  Command,
  CommandInput,
  CommandList,
  CommandEmpty,
  CommandGroup,
  CommandItem,
} from '@/components/ui/command'
import { Check, ChevronsUpDown } from 'lucide-vue-next'
import Favicon from './Favicon.vue'
defineProps({ modelValue: String, id: String })
const emit = defineEmits(['update:modelValue'])
const open = ref(false)
function select(value) {
  emit('update:modelValue', value)
  open.value = false
}
const platforms = PLATFORM_GROUPS
</script>
