import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import i18n from './i18n'

import './styles/globals.css'

const app = createApp(App).use(router).use(i18n)

// Initial navigation includes the session/profile guards. Do not mount the
// application shell against the router's temporary, unmatched '/' location.
router.isReady().then(() => app.mount('#app'))

if (process.env.NODE_ENV === 'production' && 'serviceWorker' in navigator) {
  window.addEventListener('load', () => { navigator.serviceWorker.register('/sw.js').catch(() => {}) })
}
