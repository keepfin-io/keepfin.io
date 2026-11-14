import { ref } from 'vue'

// The emoji chosen on the profile page. One ref, so the sidebar avatar
// changes as soon as a new emoji is picked.
const KEY = 'user_avatar'
export const DEFAULT_AVATAR = '😊'
function read() {
  try { return localStorage.getItem(KEY) || DEFAULT_AVATAR } catch { return DEFAULT_AVATAR }
}
export const avatar = ref(read())
export function setAvatar(emoji) {
  avatar.value = emoji
  try { localStorage.setItem(KEY, emoji) } catch { /* The choice still applies to this session. */ }
}
