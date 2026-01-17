import { createI18n } from 'vue-i18n'
import en from './locales/en.json'
import tr from './locales/tr.json'

// Get saved locale from localStorage, default to 'en-US'
const savedLocale = localStorage.getItem('locale') || 'en-US'
// Extract language code from locale (e.g., 'en-US' -> 'en', 'tr-TR' -> 'tr')
const defaultLanguage = savedLocale.split('-')[0] || 'en'

const i18n = createI18n({
    legacy: false, // Use Composition API mode
    locale: defaultLanguage,
    fallbackLocale: 'en',
    messages: {
        en,
        tr
    }
})

export function setLanguage(lang) {
    i18n.global.locale.value = lang
    localStorage.setItem('language', lang)
}

export function getLanguage() {
    return i18n.global.locale.value
}

export default i18n
