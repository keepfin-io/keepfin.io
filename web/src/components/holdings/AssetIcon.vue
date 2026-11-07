<template>
  <span data-testid="asset-icon" class="relative inline-flex shrink-0 items-center justify-center overflow-hidden border border-border/60 bg-primary/10 text-primary" :class="compact ? 'size-8 rounded-lg' : 'size-10 rounded-xl'" aria-hidden="true">
    <component v-if="fallbackIcon" :is="fallbackIcon" :class="compact ? 'size-4' : 'size-5'" />
    <span v-else class="text-xs font-semibold tracking-wide">{{ initials }}</span>
    <img v-if="logoUrl" :key="logoUrl" :src="logoUrl" alt="" :width="compact ? 32 : 40" :height="compact ? 32 : 40" loading="lazy" decoding="async" referrerpolicy="no-referrer" class="absolute inset-0 size-full bg-white object-contain p-1.5 transition-opacity" :class="loaded ? 'opacity-100' : 'opacity-0'" @load="onLoad" @error="onError" />
  </span>
</template>
<script setup>
import { computed, ref, watch } from 'vue'
import { Banknote, Coins, Layers } from 'lucide-vue-next'
import { assetCategory } from '@/lib/asset-search'

const props = defineProps({ asset: { type: Object, required: true }, compact: Boolean })
const loaded = ref(false)
const candidateIndex = ref(0)
const logoCandidates = computed(() => {
  const code = props.asset.Code?.trim(), exchange = props.asset.Exchange?.trim()
  if (!code || !exchange || !/^[\w.^=-]+$/.test(code) || !/^[\w-]+$/.test(exchange)) return []
  // EODHD's filenames are case-sensitive: QQQ.png exists, but Apple uses aapl.png.
  return [...new Set([code.toUpperCase(), code.toLowerCase()])].map(symbol =>
    `https://eodhd.com/img/logos/${encodeURIComponent(exchange.toUpperCase())}/${encodeURIComponent(symbol)}.png`
  )
})
const logoUrl = computed(() => logoCandidates.value[candidateIndex.value] || '')
const initials = computed(() => (props.asset.Code || '?').replace(/[^\p{L}\p{N}]/gu, '').slice(0, 2).toUpperCase() || '?')
const fallbackIcon = computed(() => {
  return { fund: Layers, crypto: Coins, forex: Banknote }[assetCategory(props.asset)] || null
})
watch(() => logoCandidates.value.join('|'), () => { loaded.value = false; candidateIndex.value = 0 })

function onLoad(event) {
  if (event.currentTarget.getAttribute('src') === logoUrl.value) loaded.value = true
}

function onError(event) {
  if (event.currentTarget.getAttribute('src') !== logoUrl.value) return
  loaded.value = false
  candidateIndex.value += 1
}
</script>
