import { ref } from 'vue'
export const installPrompt = ref(null)
window.addEventListener('beforeinstallprompt', event => { event.preventDefault(); installPrompt.value = event })
window.addEventListener('appinstalled', () => { installPrompt.value = null })
export async function installApp() {if (!installPrompt.value) return;await installPrompt.value.prompt();await installPrompt.value.userChoice;installPrompt.value=null}
