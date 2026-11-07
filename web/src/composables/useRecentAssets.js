import { computed, onBeforeUnmount, onMounted, readonly, ref, watch } from 'vue'

const LIMIT = 6

function normalize(asset) {
  if (!asset || typeof asset !== 'object') return null
  const { Code, Exchange, Name } = asset
  if (typeof Code !== 'string' || typeof Exchange !== 'string' || typeof Name !== 'string') return null
  if (!/^[\w.^=-]{1,40}$/.test(Code) || !/^[\w-]{1,20}$/.test(Exchange) || !Name.trim()) return null
  return {
    Code: Code.toUpperCase(), Exchange: Exchange.toUpperCase(), Name: Name.slice(0, 200),
    Currency: typeof asset.Currency === 'string' ? asset.Currency.slice(0, 10) : '',
    Type: typeof asset.Type === 'string' ? asset.Type.slice(0, 50) : '',
    Country: typeof asset.Country === 'string' ? asset.Country.slice(0, 80) : '',
  }
}

function unique(assets) {
  const seen = new Set()
  return assets.map(normalize).filter(asset => {
    if (!asset) return false
    const key = `${asset.Code}.${asset.Exchange}`
    if (seen.has(key)) return false
    seen.add(key)
    return true
  }).slice(0, LIMIT)
}

export function useRecentAssets(user) {
  const recentAssets = ref([])
  const storageKey = computed(() => user.value?.id ? `keepfin:recent-assets:v1:${user.value.id}` : null)

  function load() {
    recentAssets.value = []
    if (!storageKey.value) return
    try {
      const stored = JSON.parse(localStorage.getItem(storageKey.value) || '[]')
      if (Array.isArray(stored)) recentAssets.value = unique(stored)
    } catch { /* A missing or unreadable history must not block asset selection. */ }
  }

  function rememberAsset(asset) {
    if (!storageKey.value) return
    recentAssets.value = unique([asset, ...recentAssets.value])
    try { localStorage.setItem(storageKey.value, JSON.stringify(recentAssets.value)) } catch { /* Keep the current page usable when storage is unavailable. */ }
  }

  function clearRecentAssets() {
    recentAssets.value = []
    if (!storageKey.value) return
    try { localStorage.removeItem(storageKey.value) } catch { /* Clearing the visible list remains available. */ }
  }

  function sync(event) { if (event.key === storageKey.value || event.key === null) load() }
  watch(storageKey, load, { immediate: true })
  onMounted(() => window.addEventListener('storage', sync))
  onBeforeUnmount(() => window.removeEventListener('storage', sync))
  return { recentAssets: readonly(recentAssets), rememberAsset, clearRecentAssets }
}
