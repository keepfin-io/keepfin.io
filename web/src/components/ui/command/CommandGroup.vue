<script setup>
import 'core-js/modules/es.set.difference.v2.js'
import 'core-js/modules/es.set.intersection.v2.js'
import 'core-js/modules/es.set.is-disjoint-from.v2.js'
import 'core-js/modules/es.set.is-subset-of.v2.js'
import 'core-js/modules/es.set.is-superset-of.v2.js'
import 'core-js/modules/es.set.symmetric-difference.v2.js'
import 'core-js/modules/es.set.union.v2.js'
import { reactiveOmit } from '@vueuse/core'
import { cn } from 'cn'
import { ListboxGroup, ListboxGroupLabel, useId } from 'reka-ui'
import { computed, onMounted, onUnmounted } from 'vue'
import { provideCommandGroupContext, useCommand } from '.'

const props = defineProps({
  asChild: { type: Boolean, required: false },
  as: { type: null, required: false },
  class: { type: null, required: false },
  heading: { type: String, required: false },
})

const delegatedProps = reactiveOmit(props, 'class')

const { allGroups, filterState } = useCommand()
const id = useId()

const isRender = computed(() =>
  !filterState.search ? true : filterState.filtered.groups.has(id)
)

provideCommandGroupContext({ id })
onMounted(() => {
  if (!allGroups.value.has(id)) allGroups.value.set(id, new Set())
})
onUnmounted(() => {
  allGroups.value.delete(id)
})
</script>

<template>
  <ListboxGroup
    v-bind="delegatedProps"
    :id="id"
    data-slot="command-group"
    :class="cn('text-foreground overflow-hidden p-1', props.class)"
    :hidden="isRender ? undefined : true"
  >
    <ListboxGroupLabel
      v-if="heading"
      data-slot="command-group-heading"
      class="px-2 py-1.5 text-xs font-medium text-muted-foreground"
    >
      {{ heading }}
    </ListboxGroupLabel>
    <slot />
  </ListboxGroup>
</template>
