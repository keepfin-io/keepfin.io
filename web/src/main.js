import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import i18n from './i18n'

import './styles/globals.css'

// Always use dark mode
document.documentElement.classList.add('dark')

createApp(App).use(router).use(i18n).mount('#app')
