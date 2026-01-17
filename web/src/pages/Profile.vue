<template>
    <!-- Loading Spinner -->
    <FullPageSpinner :visible="loading" message="Loading profile..." />

    <main class="profile-page">
        <!-- Header -->
        <div class="page-header">
            <div class="header-content">
                <h1 class="page-title">{{ $t('profile.title') }}</h1>
                <p class="page-subtitle">{{ $t('profile.subtitle') }}</p>
            </div>
        </div>

        <!-- Profile Card -->
        <div class="profile-card">
            <!-- Avatar Section -->
            <div class="avatar-section">
                <div class="avatar-display" @click="showEmojiPicker = true">
                    <span class="avatar-emoji">{{ selectedEmoji }}</span>
                    <div class="avatar-edit-hint">
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
                            <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
                        </svg>
                    </div>
                </div>
                <p class="avatar-hint">{{ $t('profile.avatarHint') }}</p>
            </div>

            <!-- User Info -->
            <div class="info-section">
                <div class="info-group">
                    <label class="info-label">{{ $t('profile.userId') }}</label>
                    <p class="info-value info-uuid">{{ user?.id || '-' }}</p>
                </div>
                <div class="info-group">
                    <label class="info-label">{{ $t('profile.email') }}</label>
                    <p class="info-value">{{ user?.email || '-' }}</p>
                </div>
                <div class="info-group">
                    <label class="info-label">{{ $t('profile.memberSince') }}</label>
                    <p class="info-value">{{ formatDate(user?.created_at) }}</p>
                </div>
            </div>
        </div>

        <!-- Preferences Card -->
        <div class="profile-card preferences-card">
            <h3 class="card-title">{{ $t('profile.preferences') }}</h3>

            <!-- Base Currency -->
            <div class="preference-group">
                <label class="preference-label">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <circle cx="12" cy="12" r="10"></circle>
                        <line x1="12" y1="6" x2="12" y2="12"></line>
                        <line x1="12" y1="18" x2="12" y2="18.01"></line>
                    </svg>
                    {{ $t('profile.baseCurrency') }}
                </label>
                <p class="preference-hint">{{ $t('profile.baseCurrencyHint') }}</p>
                <select v-model="baseCurrency" @change="saveCurrency" class="preference-select">
                    <option v-for="currency in currencies" :key="currency.code" :value="currency.code">
                        {{ currency.symbol }} {{ currency.code }} - {{ currency.name }}
                    </option>
                </select>
            </div>

            <!-- Language -->
            <div class="preference-group">
                <label class="preference-label">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <circle cx="12" cy="12" r="10"></circle>
                        <line x1="2" y1="12" x2="22" y2="12"></line>
                        <path
                            d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z">
                        </path>
                    </svg>
                    {{ $t('profile.language') }}
                </label>
                <p class="preference-hint">{{ $t('profile.languageHint') }}</p>
                <select v-model="language" @change="saveLanguage" class="preference-select">
                    <option v-for="lang in languages" :key="lang.code" :value="lang.code">
                        {{ lang.flag }} {{ lang.name }}
                    </option>
                </select>
            </div>

            <!-- Locale -->
            <div class="preference-group">
                <label class="preference-label">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
                        <line x1="16" y1="2" x2="16" y2="6"></line>
                        <line x1="8" y1="2" x2="8" y2="6"></line>
                        <line x1="3" y1="10" x2="21" y2="10"></line>
                    </svg>
                    {{ $t('profile.dateFormat') }}
                </label>
                <p class="preference-hint">{{ $t('profile.dateFormatHint') }}</p>
                <select v-model="locale" @change="saveLocale" class="preference-select">
                    <option v-for="loc in locales" :key="loc.code" :value="loc.code">
                        {{ loc.name }} ({{ loc.example }})
                    </option>
                </select>
            </div>
        </div>

        <!-- Emoji Picker Modal -->
        <div v-if="showEmojiPicker" class="modal-overlay" @click.self="showEmojiPicker = false">
            <div class="emoji-picker-modal">
                <div class="modal-header">
                    <h2>{{ $t('profile.chooseAvatar') }}</h2>
                    <button @click="showEmojiPicker = false" class="btn-close">
                        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <line x1="18" y1="6" x2="6" y2="18"></line>
                            <line x1="6" y1="6" x2="18" y2="18"></line>
                        </svg>
                    </button>
                </div>
                <div class="emoji-grid">
                    <button v-for="emoji in emojiList" :key="emoji" @click="selectEmoji(emoji)" class="emoji-option"
                        :class="{ 'is-selected': emoji === selectedEmoji }">
                        {{ emoji }}
                    </button>
                </div>
            </div>
        </div>
    </main>
</template>

<script>
import { useAuth } from '@/composables/useAuth'
import { userAPI } from '@/services/userApi'
import { setLanguage } from '@/i18n'
import FullPageSpinner from '@/components/ui/FullPageSpinner.vue'

export default {
    name: 'ProfilePage',
    components: {
        FullPageSpinner
    },
    setup() {
        const { user, isLoading } = useAuth()
        return { user, isLoading }
    },
    computed: {
        loading() {
            return this.isLoading || !this.user
        }
    },
    data() {
        return {
            showEmojiPicker: false,
            selectedEmoji: localStorage.getItem('user_avatar') || '😊',
            baseCurrency: 'USD',
            language: localStorage.getItem('language') || 'en',
            locale: localStorage.getItem('locale') || 'en-US',
            theme: 'light',
            saving: false,
            emojiList: [
                '😊', '😎', '🤩', '😇', '🥳', '😄', '🤗', '😏',
                '🧑‍💻', '👨‍💻', '👩‍💻', '🧑‍🎨', '👨‍🎨', '👩‍🎨', '🧑‍🚀', '👨‍🚀',
                '🦊', '🐱', '🐶', '🐼', '🐨', '🦁', '🐯', '🐸',
                '🌟', '⭐', '✨', '💫', '🔥', '💎', '🎯', '🚀',
                '🎨', '🎭', '🎪', '🎢', '🎡', '🏆', '🥇', '🎖️',
                '💜', '💙', '💚', '💛', '🧡', '❤️', '🖤', '🤍'
            ],
            currencies: [
                { code: 'USD', symbol: '$', name: 'US Dollar' },
                { code: 'EUR', symbol: '€', name: 'Euro' },
                { code: 'GBP', symbol: '£', name: 'British Pound' },
                { code: 'TRY', symbol: '₺', name: 'Turkish Lira' },
                { code: 'JPY', symbol: '¥', name: 'Japanese Yen' },
                { code: 'CNY', symbol: '¥', name: 'Chinese Yuan' },
                { code: 'CHF', symbol: 'Fr', name: 'Swiss Franc' },
                { code: 'CAD', symbol: '$', name: 'Canadian Dollar' },
                { code: 'AUD', symbol: '$', name: 'Australian Dollar' },
                { code: 'INR', symbol: '₹', name: 'Indian Rupee' }
            ],
            languages: [
                { code: 'en', name: 'English', flag: '🇺🇸' },
                { code: 'tr', name: 'Türkçe', flag: '🇹🇷' },
                { code: 'de', name: 'Deutsch', flag: '🇩🇪' },
                { code: 'fr', name: 'Français', flag: '🇫🇷' },
                { code: 'es', name: 'Español', flag: '🇪🇸' },
                { code: 'pt', name: 'Português', flag: '🇧🇷' },
                { code: 'ja', name: '日本語', flag: '🇯🇵' },
                { code: 'zh', name: '中文', flag: '🇨🇳' }
            ],
            locales: [
                { code: 'en-US', name: 'United States', example: '12/27/2025' },
                { code: 'en-GB', name: 'United Kingdom', example: '27/12/2025' },
                { code: 'tr-TR', name: 'Turkey', example: '27.12.2025' },
                { code: 'de-DE', name: 'Germany', example: '27.12.2025' },
                { code: 'fr-FR', name: 'France', example: '27/12/2025' },
                { code: 'es-ES', name: 'Spain', example: '27/12/2025' },
                { code: 'pt-BR', name: 'Brazil', example: '27/12/2025' },
                { code: 'ja-JP', name: 'Japan', example: '2025/12/27' },
                { code: 'zh-CN', name: 'China', example: '2025/12/27' }
            ]
        }
    },
    watch: {
        user: {
            immediate: true,
            handler(newUser) {
                if (newUser) {
                    this.baseCurrency = newUser.base_currency || 'USD'
                    this.theme = newUser.theme || 'light'
                    this.locale = newUser.locale || 'en-US'
                    this.applyTheme()
                }
            }
        }
    },
    mounted() {
        this.applyTheme()
    },
    methods: {
        async selectEmoji(emoji) {
            this.selectedEmoji = emoji
            localStorage.setItem('user_avatar', emoji)
            this.showEmojiPicker = false
        },
        async saveCurrency() {
            try {
                this.saving = true
                await userAPI.updateBaseCurrency(this.baseCurrency)
                localStorage.setItem('base_currency', this.baseCurrency)
            } catch (error) {
                console.error('Failed to save currency:', error)
            } finally {
                this.saving = false
            }
        },
        async saveLanguage() {
            const languageToLocale = {
                'en': 'en-US',
                'tr': 'tr-TR',
                'de': 'de-DE',
                'fr': 'fr-FR',
                'es': 'es-ES',
                'pt': 'pt-BR',
                'ja': 'ja-JP',
                'zh': 'zh-CN'
            }

            const newLocale = languageToLocale[this.language] || 'en-US'
            this.locale = newLocale

            // Update i18n locale
            setLanguage(this.language)

            try {
                await userAPI.updateLocale(newLocale)
                localStorage.setItem('language', this.language)
                localStorage.setItem('locale', newLocale)
            } catch (error) {
                console.error('Failed to save language/locale:', error)
            }
        },
        async saveLocale() {
            try {
                await userAPI.updateLocale(this.locale)
                localStorage.setItem('locale', this.locale)
            } catch (error) {
                console.error('Failed to save locale:', error)
            }
        },
        async setTheme(theme) {
            this.theme = theme
            this.applyTheme()

            if (theme === 'light' || theme === 'dark') {
                try {
                    await userAPI.updateTheme(theme)
                } catch (error) {
                    console.error('Failed to save theme:', error)
                }
            }
            localStorage.setItem('theme', theme)
        },
        async applyTheme() {
            const root = document.documentElement
            let isDark = this.theme === 'dark' ||
                (this.theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches)

            root.classList.toggle('dark', isDark)
        },
        formatDate(dateString) {
            if (!dateString) return '-'
            const date = new Date(dateString)
            return date.toLocaleDateString(this.locale, {
                year: 'numeric',
                month: 'long',
                day: 'numeric'
            })
        }
    }
}
</script>

<style scoped>
.profile-page {
    padding: 1.5rem;
    max-width: 800px;
    margin: 0 auto;
}

/* Header */
.page-header {
    margin-bottom: 2rem;
}

.page-title {
    font-size: 1.75rem;
    font-weight: 700;
    color: #1f2937;
    margin: 0;
}

.page-subtitle {
    color: #6b7280;
    margin: 0.25rem 0 0 0;
    font-size: 0.875rem;
}

/* Profile Card */
.profile-card {
    background: white;
    border-radius: 1.25rem;
    border: 1px solid #e2e8f0;
    padding: 2rem;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

/* Avatar Section */
.avatar-section {
    display: flex;
    flex-direction: column;
    align-items: center;
    margin-bottom: 2rem;
    padding-bottom: 2rem;
    border-bottom: 1px solid #f1f5f9;
}

.avatar-display {
    width: 120px;
    height: 120px;
    background: linear-gradient(135deg, #f0f4ff 0%, #e5edff 100%);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.2s;
    position: relative;
    border: 3px solid #e2e8f0;
}

.avatar-display:hover {
    transform: scale(1.05);
    border-color: #3b82f6;
    box-shadow: 0 8px 24px rgba(59, 130, 246, 0.2);
}

.avatar-emoji {
    font-size: 3.5rem;
    line-height: 1;
}

.avatar-edit-hint {
    position: absolute;
    bottom: 4px;
    right: 4px;
    width: 32px;
    height: 32px;
    background: white;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
    color: #64748b;
}

.avatar-hint {
    margin-top: 0.75rem;
    font-size: 0.8125rem;
    color: #94a3b8;
}

/* Info Section */
.info-section {
    display: grid;
    gap: 1.5rem;
}

.info-group {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
}

.info-label {
    font-size: 0.75rem;
    font-weight: 600;
    color: #64748b;
    text-transform: uppercase;
    letter-spacing: 0.05em;
}

.info-value {
    font-size: 1rem;
    color: #1e293b;
    margin: 0;
    font-weight: 500;
}

.info-uuid {
    font-family: 'SF Mono', 'Consolas', 'Monaco', monospace;
    font-size: 0.8125rem;
    color: #64748b;
    background: #f8fafc;
    padding: 0.375rem 0.625rem;
    border-radius: 0.375rem;
    display: inline-block;
    border: 1px solid #e2e8f0;
}

/* Preferences Card */
.preferences-card {
    margin-top: 1.5rem;
}

.card-title {
    font-size: 1.125rem;
    font-weight: 600;
    color: #1e293b;
    margin: 0 0 1.5rem 0;
    padding-bottom: 1rem;
    border-bottom: 1px solid #f1f5f9;
}

.preference-group {
    padding: 1rem 0;
    border-bottom: 1px solid #f1f5f9;
}

.preference-group:last-child {
    border-bottom: none;
    padding-bottom: 0;
}

/* Theme Toggle */
.theme-toggle {
    display: flex;
    gap: 0.5rem;
}

.theme-btn {
    flex: 1;
    padding: 0.75rem 1rem;
    background: #f8fafc;
    border: 2px solid #e2e8f0;
    border-radius: 0.5rem;
    cursor: pointer;
    font-size: 0.875rem;
    font-weight: 500;
    color: #64748b;
    transition: all 0.2s;
}

.theme-btn:hover {
    background: #f1f5f9;
    border-color: #cbd5e1;
}

.theme-btn.is-active {
    background: #eff6ff;
    border-color: #3b82f6;
    color: #3b82f6;
}

.preference-label {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.9375rem;
    font-weight: 600;
    color: #1e293b;
    margin-bottom: 0.25rem;
}

.preference-label svg {
    color: #64748b;
}

.preference-hint {
    font-size: 0.8125rem;
    color: #94a3b8;
    margin: 0 0 0.75rem 0;
}

.preference-select {
    width: 100%;
    padding: 0.75rem 1rem;
    font-size: 0.9375rem;
    border: 1px solid #e2e8f0;
    border-radius: 0.5rem;
    background: white;
    color: #1e293b;
    cursor: pointer;
    transition: all 0.2s;
    appearance: none;
    background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%236b7280' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3e%3c/svg%3e");
    background-position: right 0.75rem center;
    background-repeat: no-repeat;
    background-size: 1.25rem;
    padding-right: 2.5rem;
}

.preference-select:hover {
    border-color: #cbd5e1;
}

.preference-select:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

/* Modal */
.modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 1rem;
}

.emoji-picker-modal {
    background: white;
    border-radius: 1rem;
    width: 100%;
    max-width: 400px;
    max-height: 80vh;
    overflow: hidden;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
}

.modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.25rem 1.5rem;
    border-bottom: 1px solid #f1f5f9;
}

.modal-header h2 {
    font-size: 1.125rem;
    font-weight: 600;
    color: #1e293b;
    margin: 0;
}

.btn-close {
    padding: 0.375rem;
    background: transparent;
    border: none;
    border-radius: 0.375rem;
    cursor: pointer;
    color: #94a3b8;
    transition: all 0.2s;
}

.btn-close:hover {
    background: #f1f5f9;
    color: #475569;
}

/* Emoji Grid */
.emoji-grid {
    display: grid;
    grid-template-columns: repeat(8, 1fr);
    gap: 0.25rem;
    padding: 1rem;
    max-height: 300px;
    overflow-y: auto;
}

.emoji-option {
    width: 100%;
    aspect-ratio: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.5rem;
    background: transparent;
    border: 2px solid transparent;
    border-radius: 0.5rem;
    cursor: pointer;
    transition: all 0.15s;
}

.emoji-option:hover {
    background: #f0f4ff;
    transform: scale(1.1);
}

.emoji-option.is-selected {
    background: #dbeafe;
    border-color: #3b82f6;
}

/* Responsive */
@media (max-width: 480px) {
    .emoji-grid {
        grid-template-columns: repeat(6, 1fr);
    }

    .avatar-display {
        width: 100px;
        height: 100px;
    }

    .avatar-emoji {
        font-size: 3rem;
    }
}
</style>

<!-- Dark Mode Styles (unscoped to work with .dark on html element) -->
<style>
/* ==================== PROFILE DARK MODE ==================== */
/* Using pure black/gray tones instead of blue-tinted slate */

/* Page Background */
.dark .profile-page {
    background: #0f0f12;
}

/* Header */
.dark .page-title {
    color: #fafafa;
}

.dark .page-subtitle {
    color: #a1a1aa;
}

/* Profile Card */
.dark .profile-card {
    background: #1a1a1f !important;
    border-color: #2a2a30 !important;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.4);
}

/* Avatar Section */
.dark .avatar-section {
    border-bottom-color: #2a2a30;
}

.dark .avatar-display {
    background: linear-gradient(135deg, #2a2a30 0%, #1a1a1f 100%);
    border-color: #42424a;
}

.dark .avatar-display:hover {
    border-color: #a1a1aa;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
}

.dark .avatar-edit-hint {
    background: #2a2a30;
    color: #a1a1aa;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.4);
}

.dark .avatar-hint {
    color: #71717a;
}

/* Info Section */
.dark .info-label {
    color: #a1a1aa;
}

.dark .info-value {
    color: #fafafa;
}

.dark .info-uuid {
    background: #0f0f12;
    color: #a1a1aa;
    border-color: #2a2a30;
}

/* Preferences Card */
.dark .card-title {
    color: #fafafa;
    border-bottom-color: #2a2a30;
}

.dark .preference-group {
    border-bottom-color: #2a2a30;
}

.dark .preference-label {
    color: #fafafa;
}

.dark .preference-label svg {
    color: #a1a1aa;
}

.dark .preference-hint {
    color: #71717a;
}

.dark .preference-select {
    background-color: #0f0f12;
    border-color: #2a2a30;
    color: #fafafa;
    background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%23a1a1aa' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3e%3c/svg%3e");
    background-position: right 0.75rem center;
    background-repeat: no-repeat;
    background-size: 1.25rem;
}

.dark .preference-select option {
    background: #1a1a1f;
    color: #fafafa;
}

.dark .preference-select:hover {
    border-color: #42424a;
}

.dark .preference-select:focus {
    border-color: #555560;
    box-shadow: 0 0 0 3px rgba(82, 82, 91, 0.3);
}

/* Theme Toggle */
.dark .theme-btn {
    background: #0f0f12;
    border-color: #2a2a30;
    color: #a1a1aa;
}

.dark .theme-btn:hover {
    background: #1a1a1f;
    border-color: #42424a;
}

.dark .theme-btn.is-active {
    background: #2a2a30;
    border-color: #555560;
    color: #fafafa;
}

/* Modal */
.dark .modal-overlay {
    background: rgba(0, 0, 0, 0.8);
}

.dark .emoji-picker-modal {
    background: #1a1a1f;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.6);
}

.dark .modal-header {
    border-bottom-color: #2a2a30;
}

.dark .modal-header h2 {
    color: #fafafa;
}

.dark .btn-close {
    color: #71717a;
}

.dark .btn-close:hover {
    background: #2a2a30;
    color: #a1a1aa;
}

/* Emoji Grid */
.dark .emoji-option:hover {
    background: #2a2a30;
}

.dark .emoji-option.is-selected {
    background: #42424a;
    border-color: #555560;
}
</style>
