<template><Button data-testid="zen-toggle" variant="ghost" size="icon" :disabled="busy" :aria-pressed="zen" :aria-label="$t('wealth.zen')" @click="toggle"><EyeOff v-if="zen" class="size-4" /><Eye v-else class="size-4" /></Button><span v-if="error" role="alert" class="max-w-48 text-xs text-destructive">{{ error }}</span></template>
<script setup>
import { ref } from 'vue'
import { Eye, EyeOff } from 'lucide-vue-next'
import { Button } from '@/components/ui'
import { zen, toggleZen } from '@/composables/useZen'
const busy = ref(false), error = ref('')
async function toggle() {busy.value=true;error.value='';try {await toggleZen()}catch(e){error.value=e.message}finally{busy.value=false}}
</script>
