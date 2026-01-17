<template>
    <div class="platform-select" :class="{ 'is-open': isOpen }" ref="container">
        <!-- Selected Value / Trigger -->
        <button type="button" class="select-trigger" @click="toggle">
            <div class="selected-value">
                <template v-if="modelValue">
                    <Favicon :domain="getDomain(modelValue)" :size="20" />
                    <span>{{ modelValue }}</span>
                </template>
                <span v-else class="placeholder">Select platform (optional)</span>
            </div>
            <svg class="chevron" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"
                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="6 9 12 15 18 9"></polyline>
            </svg>
        </button>

        <!-- Dropdown -->
        <div v-if="isOpen" class="dropdown">
            <!-- Search -->
            <div class="search-box">
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <circle cx="11" cy="11" r="8"></circle>
                    <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
                </svg>
                <input type="text" v-model="searchQuery" placeholder="Search platforms..." ref="searchInput"
                    @keydown.esc="close" />
            </div>

            <!-- Options -->
            <div class="options-list">
                <div v-if="filteredPlatforms.length === 0" class="no-results">
                    No platforms found
                </div>
                <template v-for="group in filteredPlatforms" :key="group.label">
                    <div class="option-group-label">{{ group.label }}</div>
                    <button v-for="platform in group.options" :key="platform.name" type="button" class="option"
                        :class="{ 'is-selected': modelValue === platform.name }" @click="selectPlatform(platform.name)">
                        <Favicon :domain="platform.domain" :size="20" />
                        <span>{{ platform.name }}</span>
                        <svg v-if="modelValue === platform.name" class="check-icon" xmlns="http://www.w3.org/2000/svg"
                            width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                            stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <polyline points="20 6 9 17 4 12"></polyline>
                        </svg>
                    </button>
                </template>
            </div>

            <!-- Clear Option -->
            <div v-if="modelValue" class="clear-section">
                <button type="button" class="clear-btn" @click="selectPlatform('')">
                    Clear selection
                </button>
            </div>
        </div>
    </div>
</template>

<script>
import Favicon from './Favicon.vue'

export default {
    name: 'PlatformSelect',
    components: {
        Favicon
    },
    props: {
        modelValue: {
            type: String,
            default: ''
        }
    },
    emits: ['update:modelValue'],
    data() {
        return {
            isOpen: false,
            searchQuery: '',
            platforms: [
                {
                    label: 'Banks',
                    options: [
                        { name: 'Bank of America', domain: 'bankofamerica.com' },
                        { name: 'Chase', domain: 'chase.com' },
                        { name: 'Wells Fargo', domain: 'wellsfargo.com' },
                        { name: 'Citibank', domain: 'citibank.com' },
                        { name: 'HSBC', domain: 'hsbc.com' },
                        { name: 'Barclays', domain: 'barclays.co.uk' },
                        { name: 'Deutsche Bank', domain: 'db.com' },
                        { name: 'ING', domain: 'ing.com' },
                        { name: 'N26', domain: 'n26.com' },
                        { name: 'Revolut', domain: 'revolut.com' }
                    ]
                },
                {
                    label: 'Brokerages',
                    options: [
                        { name: 'Interactive Brokers', domain: 'interactivebrokers.com' },
                        { name: 'Fidelity', domain: 'fidelity.com' },
                        { name: 'Charles Schwab', domain: 'schwab.com' },
                        { name: 'TD Ameritrade', domain: 'tdameritrade.com' },
                        { name: 'E*TRADE', domain: 'etrade.com' },
                        { name: 'Robinhood', domain: 'robinhood.com' },
                        { name: 'Webull', domain: 'webull.com' },
                        { name: 'Trading 212', domain: 'trading212.com' },
                        { name: 'eToro', domain: 'etoro.com' },
                        { name: 'Degiro', domain: 'degiro.com' }
                    ]
                },
                {
                    label: 'Crypto',
                    options: [
                        { name: 'Coinbase', domain: 'coinbase.com' },
                        { name: 'Binance', domain: 'binance.com' },
                        { name: 'Kraken', domain: 'kraken.com' },
                        { name: 'Crypto.com', domain: 'crypto.com' },
                        { name: 'Gemini', domain: 'gemini.com' },
                        { name: 'Bybit', domain: 'bybit.com' },
                        { name: 'OKX', domain: 'okx.com' }
                    ]
                },
                {
                    label: 'Payment Services',
                    options: [
                        { name: 'PayPal', domain: 'paypal.com' },
                        { name: 'Wise', domain: 'wise.com' },
                        { name: 'Stripe', domain: 'stripe.com' },
                        { name: 'Venmo', domain: 'venmo.com' },
                        { name: 'Cash App', domain: 'cash.app' }
                    ]
                },
                {
                    label: 'Turkish Banks',
                    options: [
                        { name: 'Garanti BBVA', domain: 'garantibbva.com.tr' },
                        { name: 'İş Bankası', domain: 'isbank.com.tr' },
                        { name: 'Yapı Kredi', domain: 'yapikredi.com.tr' },
                        { name: 'Akbank', domain: 'akbank.com' },
                        { name: 'QNB Finansbank', domain: 'qnbfinansbank.com' },
                        { name: 'Ziraat Bankası', domain: 'ziraatbank.com.tr' },
                        { name: 'Midas', domain: 'getmidas.com' }
                    ]
                }
            ]
        }
    },
    computed: {
        filteredPlatforms() {
            if (!this.searchQuery) return this.platforms

            const query = this.searchQuery.toLowerCase()
            return this.platforms
                .map(group => ({
                    label: group.label,
                    options: group.options.filter(p =>
                        p.name.toLowerCase().includes(query)
                    )
                }))
                .filter(group => group.options.length > 0)
        }
    },
    methods: {
        toggle() {
            this.isOpen = !this.isOpen
            if (this.isOpen) {
                this.searchQuery = ''
                this.$nextTick(() => {
                    this.$refs.searchInput?.focus()
                })
            }
        },
        close() {
            this.isOpen = false
        },
        selectPlatform(name) {
            this.$emit('update:modelValue', name)
            this.close()
        },
        getDomain(platformName) {
            for (const group of this.platforms) {
                const found = group.options.find(p => p.name === platformName)
                if (found) return found.domain
            }
            return 'example.com'
        },
        handleClickOutside(e) {
            if (this.$refs.container && !this.$refs.container.contains(e.target)) {
                this.close()
            }
        }
    },
    mounted() {
        document.addEventListener('click', this.handleClickOutside)
    },
    beforeUnmount() {
        document.removeEventListener('click', this.handleClickOutside)
    }
}
</script>

<style scoped>
.platform-select {
    position: relative;
    width: 100%;
}

.select-trigger {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    padding: 0.75rem 1rem;
    background: white;
    border: 1px solid #e2e8f0;
    border-radius: 0.5rem;
    cursor: pointer;
    transition: all 0.2s;
    text-align: left;
    font-size: 0.9375rem;
}

.select-trigger:hover {
    border-color: #cbd5e1;
}

.select-trigger:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.is-open .select-trigger {
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.selected-value {
    display: flex;
    align-items: center;
    gap: 0.625rem;
}

.placeholder {
    color: #94a3b8;
}

.chevron {
    color: #64748b;
    transition: transform 0.2s;
}

.is-open .chevron {
    transform: rotate(180deg);
}

/* Dropdown */
.dropdown {
    position: absolute;
    top: calc(100% + 0.5rem);
    left: 0;
    right: 0;
    background: white;
    border: 1px solid #e2e8f0;
    border-radius: 0.75rem;
    box-shadow: 0 10px 40px -10px rgba(0, 0, 0, 0.2);
    z-index: 100;
    overflow: hidden;
}

/* Search */
.search-box {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    border-bottom: 1px solid #f1f5f9;
}

.search-box svg {
    color: #94a3b8;
    flex-shrink: 0;
}

.search-box input {
    flex: 1;
    border: none;
    outline: none;
    font-size: 0.875rem;
    color: #1e293b;
}

.search-box input::placeholder {
    color: #94a3b8;
}

/* Options */
.options-list {
    max-height: 280px;
    overflow-y: auto;
    padding: 0.5rem;
}

.option-group-label {
    padding: 0.5rem 0.75rem 0.375rem;
    font-size: 0.6875rem;
    font-weight: 600;
    color: #94a3b8;
    text-transform: uppercase;
    letter-spacing: 0.05em;
}

.option {
    display: flex;
    align-items: center;
    gap: 0.625rem;
    width: 100%;
    padding: 0.625rem 0.75rem;
    background: transparent;
    border: none;
    border-radius: 0.375rem;
    cursor: pointer;
    transition: all 0.15s;
    font-size: 0.875rem;
    color: #1e293b;
    text-align: left;
}

.option:hover {
    background: #f1f5f9;
}

.option.is-selected {
    background: #eff6ff;
    color: #2563eb;
}

.check-icon {
    margin-left: auto;
    color: #2563eb;
}

.no-results {
    padding: 2rem;
    text-align: center;
    color: #94a3b8;
    font-size: 0.875rem;
}

/* Clear Section */
.clear-section {
    padding: 0.5rem;
    border-top: 1px solid #f1f5f9;
}

.clear-btn {
    width: 100%;
    padding: 0.5rem;
    background: transparent;
    border: none;
    border-radius: 0.375rem;
    cursor: pointer;
    font-size: 0.8125rem;
    color: #64748b;
    transition: all 0.15s;
}

.clear-btn:hover {
    background: #fee2e2;
    color: #dc2626;
}
</style>

<!-- Dark Mode Styles (unscoped) -->
<style>
.dark .select-trigger {
    background: #0f0f12;
    border-color: #2a2a30;
    color: #fafafa;
}

.dark .select-trigger:hover {
    border-color: #42424a;
}

.dark .select-trigger:focus,
.dark .is-open .select-trigger {
    border-color: #555560;
    box-shadow: 0 0 0 3px rgba(85, 85, 96, 0.3);
}

.dark .placeholder {
    color: #71717a;
}

.dark .chevron {
    color: #71717a;
}

/* Dropdown */
.dark .dropdown {
    background: #1a1a1f;
    border-color: #2a2a30;
    box-shadow: 0 10px 40px -10px rgba(0, 0, 0, 0.5);
}

/* Search */
.dark .search-box {
    border-bottom-color: #2a2a30;
}

.dark .search-box svg {
    color: #71717a;
}

.dark .search-box input {
    background: transparent;
    color: #fafafa;
}

.dark .search-box input::placeholder {
    color: #555560;
}

/* Options */
.dark .option-group-label {
    color: #71717a;
}

.dark .option {
    color: #fafafa;
}

.dark .option:hover {
    background: #2a2a30;
}

.dark .option.is-selected {
    background: #2a2a30;
    color: #fafafa;
}

.dark .check-icon {
    color: #4ade80;
}

.dark .no-results {
    color: #71717a;
}

/* Clear Section */
.dark .clear-section {
    border-top-color: #2a2a30;
}

.dark .clear-btn {
    color: #a1a1aa;
}

.dark .clear-btn:hover {
    background: #7f1d1d;
    color: #fca5a5;
}
</style>
