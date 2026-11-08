import { ref } from 'vue'
import { wealthAPI } from '@/services/wealthApi'

export const zen = ref(false)
let owner = '', revision = 0
export async function loadZen(userID) {
  owner = userID || ''; const current = ++revision
  try { zen.value = !!owner && localStorage.getItem(`keepfin:zen:${owner}`) === 'true' } catch { zen.value = false }
  if (!owner) return
  const data = await wealthAPI.get('/preferences')
  if (current !== revision) return
  zen.value = data.data.zen
  try { localStorage.setItem(`keepfin:zen:${owner}`, String(zen.value)) } catch { /* Optional display cache. */ }
}
export async function toggleZen() {
  if (!owner) return
  const user = owner, current = revision, value = !zen.value
  const data = await wealthAPI.get('/preferences')
  if (current !== revision) return
  await wealthAPI.save('/preferences', { ...data.data, zen: value }, 'PUT')
  if (owner !== user || current !== revision) return
  zen.value = value
  try { localStorage.setItem(`keepfin:zen:${owner}`, String(value)) } catch { /* Server preference is persisted. */ }
}
