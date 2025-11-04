<template>
  <div class="flex w-full max-w-sm">
    <slot name="input" />
    <slot name="addon" />
  </div>
</template>

<script>
export default {
  name: "InputGroup"
}
</script>
