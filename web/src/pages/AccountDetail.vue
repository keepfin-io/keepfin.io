<template>
    <!-- Loading Spinner -->
    <FullPageSpinner :visible="loading" :message="$t('accountDetail.loadingDetails')" />

    <main class="account-detail-page">
        <!-- Header -->
        <div class="page-header">
            <div class="header-left">
                <router-link to="/accounts" class="back-button">
                    <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <line x1="19" y1="12" x2="5" y2="12"></line>
                        <polyline points="12 19 5 12 12 5"></polyline>
                    </svg>
                </router-link>
                <div class="header-content">
                    <h1 class="page-title">{{ account?.account_name || 'Account' }}</h1>
                    <p class="page-subtitle">{{ account?.platform || $t('accounts.general') }}</p>
                </div>
            </div>
        </div>

        <!-- Error State -->
        <div v-if="error" class="error-container">
            <p>{{ error }}</p>
            <router-link to="/accounts" class="btn-primary">{{ $t('accountDetail.backToAccounts') }}</router-link>
        </div>

        <!-- Account Content -->
        <template v-if="!loading && !error && account">
            <!-- Summary Cards -->
            <div class="summary-cards">
                <div class="summary-card">
                    <span class="summary-label">{{ $t('accountDetail.cashBalance') }}</span>
                    <span class="summary-value">{{ formatCurrency(account.initial_balance, account.currency) }}</span>
                </div>
                <div class="summary-card">
                    <span class="summary-label">{{ $t('accountDetail.equity') }}</span>
                    <span class="summary-value">{{ formatCurrency(totalEquity, account.currency) }}</span>
                </div>
                <div class="summary-card">
                    <span class="summary-label">{{ $t('accountDetail.interest') }}</span>
                    <span class="summary-value">{{ formatCurrency(0, account.currency) }}</span>
                </div>
                <div class="summary-card">
                    <span class="summary-label">{{ $t('accountDetail.dividend') }}</span>
                    <span class="summary-value">{{ formatCurrency(0, account.currency) }}</span>
                </div>
                <div class="summary-card">
                    <span class="summary-label">{{ $t('accountDetail.activities') }}</span>
                    <span class="summary-value">{{ activities.length }}</span>
                </div>
                <div class="summary-card">
                    <span class="summary-label">{{ $t('accountDetail.currency') }}</span>
                    <span class="summary-value">{{ account.currency }}</span>
                </div>
            </div>

            <!-- Tabs -->
            <div class="tabs-container">
                <div class="tabs">
                    <button @click="activeTab = 'holdings'" class="tab"
                        :class="{ 'is-active': activeTab === 'holdings' }">
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <path
                                d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z">
                            </path>
                        </svg>
                        {{ $t('accountDetail.holdings') }}
                    </button>
                    <button @click="activeTab = 'activities'" class="tab"
                        :class="{ 'is-active': activeTab === 'activities' }">
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"></polyline>
                        </svg>
                        {{ $t('accountDetail.activities') }}
                    </button>
                </div>
            </div>

            <!-- Holdings Tab -->
            <div v-if="activeTab === 'holdings'" class="tab-content">
                <div class="tab-header">
                    <button @click="showAddHoldingModal = true" class="btn-primary btn-sm">
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <line x1="12" y1="5" x2="12" y2="19"></line>
                            <line x1="5" y1="12" x2="19" y2="12"></line>
                        </svg>
                        {{ $t('accountDetail.addHolding') }}
                    </button>
                </div>
                <div v-if="holdings.length === 0" class="empty-state-sm">
                    <p>{{ $t('accountDetail.noHoldingsYet') }}</p>
                </div>
                <div v-else class="table-container">
                    <table class="data-table">
                        <thead>
                            <tr>
                                <th>{{ $t('accountDetail.name') }}</th>
                                <th>{{ $t('accountDetail.quantity') }}</th>
                                <th>{{ $t('accountDetail.currentPrice') }}</th>
                                <th>{{ $t('accountDetail.value') }}</th>
                                <th>{{ $t('accountDetail.profitLoss') }}</th>
                                <th>{{ $t('accountDetail.performance') }}</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="holding in holdings" :key="holding.id">
                                <td class="holding-name">
                                    <span class="name">{{ holding.name || holding.symbol }}</span>
                                    <span class="symbol">{{ holding.symbol }}</span>
                                </td>
                                <td>{{ holding.quantity }}</td>
                                <td>{{ formatCurrency(holding.current_price, account.currency) }}</td>
                                <td>{{ formatCurrency(holding.value, account.currency) }}</td>
                                <td :class="holding.profit_loss >= 0 ? 'text-green' : 'text-red'">
                                    {{ formatCurrency(holding.profit_loss, account.currency) }}
                                </td>
                                <td :class="holding.performance >= 0 ? 'text-green' : 'text-red'">
                                    {{ holding.performance?.toFixed(2) }}%
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>

            <!-- Activities Tab -->
            <div v-if="activeTab === 'activities'" class="tab-content">
                <div class="tab-header">
                    <button @click="showAddActivityModal = true" class="btn-primary btn-sm">
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <line x1="12" y1="5" x2="12" y2="19"></line>
                            <line x1="5" y1="12" x2="19" y2="12"></line>
                        </svg>
                        {{ $t('accountDetail.addActivity') }}
                    </button>
                </div>
                <div v-if="activities.length === 0" class="empty-state-sm">
                    <p>{{ $t('accountDetail.noActivitiesRecorded') }}</p>
                </div>
                <div v-else class="table-container">
                    <table class="data-table">
                        <thead>
                            <tr>
                                <th>{{ $t('accountDetail.date') }}</th>
                                <th>{{ $t('accountDetail.type') }}</th>
                                <th>{{ $t('accountDetail.symbol') }}</th>
                                <th>{{ $t('accountDetail.quantity') }}</th>
                                <th>{{ $t('accountDetail.price') }}</th>
                                <th>{{ $t('accountDetail.total') }}</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="activity in activities" :key="activity.id">
                                <td>{{ formatDate(activity.date) }}</td>
                                <td>
                                    <span class="activity-type" :class="`type-${activity.type}`">
                                        {{ activity.type }}
                                    </span>
                                </td>
                                <td>{{ activity.symbol }}</td>
                                <td>{{ activity.quantity }}</td>
                                <td>{{ formatCurrency(activity.price, account.currency) }}</td>
                                <td>{{ formatCurrency(activity.total, account.currency) }}</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </template>

        <!-- Add Holding Modal -->
        <AddHoldingModal :show="showAddHoldingModal" :accountId="$route.params.id" @close="showAddHoldingModal = false"
            @success="onHoldingAdded" />
    </main>
</template>

<script>
import { accountAPI } from '@/services/accountApi'
import { holdingAPI } from '@/services/holdingApi'
import FullPageSpinner from '@/components/ui/FullPageSpinner.vue'
import AddHoldingModal from '@/components/holdings/AddHoldingModal.vue'

export default {
    name: 'AccountDetailPage',
    components: {
        FullPageSpinner,
        AddHoldingModal
    },
    data() {
        return {
            account: null,
            holdings: [],
            holdingsSummary: null,
            activities: [],
            loading: true,
            error: null,
            activeTab: 'holdings',
            showAddHoldingModal: false,
            showAddActivityModal: false
        }
    },
    computed: {
        totalEquity() {
            return this.holdings.reduce((sum, h) => sum + (h.value || 0), 0)
        }
    },
    mounted() {
        this.fetchAccountDetails()
    },
    methods: {
        async fetchAccountDetails() {
            this.loading = true
            this.error = null
            try {
                const response = await accountAPI.getAccount(this.$route.params.id)
                this.account = response.account
                // Fetch holdings after account is loaded
                await this.fetchHoldings()
                this.activities = []
            } catch (err) {
                this.error = err.message || 'Failed to load account details'
            } finally {
                this.loading = false
            }
        },
        async fetchHoldings() {
            try {
                const response = await holdingAPI.getHoldings(this.$route.params.id)
                this.holdings = response.holdings || []
                this.holdingsSummary = response.summary || null
            } catch (err) {
                console.error('Failed to fetch holdings:', err)
                this.holdings = []
            }
        },
        onHoldingAdded() {
            this.fetchHoldings()
        },
        formatCurrency(amount, currency) {
            return new Intl.NumberFormat('en-US', {
                style: 'currency',
                currency: currency || 'USD'
            }).format(amount || 0)
        },
        formatDate(dateString) {
            if (!dateString) return '-'
            const date = new Date(dateString)
            return date.toLocaleDateString('en-US', {
                year: 'numeric',
                month: 'short',
                day: 'numeric'
            })
        }
    }
}
</script>

<style scoped>
.account-detail-page {
    padding: 1.5rem;
    max-width: 1400px;
    margin: 0 auto;
}

/* Header */
.page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;
}

.header-left {
    display: flex;
    align-items: center;
    gap: 1rem;
}

.back-button {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 40px;
    height: 40px;
    border-radius: 10px;
    background: #f1f5f9;
    color: #64748b;
    text-decoration: none;
    transition: all 0.2s;
}

.back-button:hover {
    background: #e2e8f0;
    color: #1e293b;
}

.page-title {
    font-size: 1.5rem;
    font-weight: 700;
    color: #1f2937;
    margin: 0;
}

.page-subtitle {
    color: #6b7280;
    margin: 0.25rem 0 0 0;
    font-size: 0.875rem;
}

/* Error */
.error-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 4rem;
    text-align: center;
    color: #dc2626;
}

/* Summary Cards */
.summary-cards {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
    gap: 1rem;
    margin-bottom: 2rem;
}

.summary-card {
    background: white;
    border-radius: 1rem;
    border: 1px solid #e2e8f0;
    padding: 1.25rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}

.summary-label {
    font-size: 0.75rem;
    font-weight: 600;
    color: #64748b;
    text-transform: uppercase;
    letter-spacing: 0.05em;
}

.summary-value {
    font-size: 1.25rem;
    font-weight: 700;
    color: #1e293b;
}

/* Tabs */
.tabs-container {
    background: white;
    border-radius: 1rem 1rem 0 0;
    border: 1px solid #e2e8f0;
    border-bottom: none;
    padding: 0 1rem;
}

.tabs {
    display: flex;
    gap: 0.5rem;
}

.tab {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 1rem 1.25rem;
    background: transparent;
    border: none;
    border-bottom: 2px solid transparent;
    font-size: 0.875rem;
    font-weight: 500;
    color: #64748b;
    cursor: pointer;
    transition: all 0.2s;
}

.tab:hover {
    color: #3b82f6;
}

.tab.is-active {
    color: #3b82f6;
    border-bottom-color: #3b82f6;
}

/* Tab Content */
.tab-content {
    background: white;
    border-radius: 0 0 1rem 1rem;
    border: 1px solid #e2e8f0;
    border-top: none;
    padding: 1.5rem;
}

.tab-header {
    display: flex;
    justify-content: flex-end;
    margin-bottom: 1rem;
}

/* Empty State */
.empty-state-sm {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem;
    color: #64748b;
    text-align: center;
    gap: 1rem;
}

/* Table */
.table-container {
    overflow-x: auto;
}

.data-table {
    width: 100%;
    border-collapse: collapse;
}

.data-table th {
    padding: 0.75rem 1rem;
    text-align: left;
    font-size: 0.75rem;
    font-weight: 600;
    color: #64748b;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    border-bottom: 1px solid #e2e8f0;
}

.data-table td {
    padding: 1rem;
    font-size: 0.875rem;
    color: #1e293b;
    border-bottom: 1px solid #f1f5f9;
}

.data-table tbody tr:hover {
    background: #f8fafc;
}

/* Holding Name */
.holding-name {
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
}

.holding-name .name {
    font-weight: 600;
}

.holding-name .symbol {
    font-size: 0.75rem;
    color: #94a3b8;
}

/* Activity Type */
.activity-type {
    display: inline-block;
    padding: 0.25rem 0.5rem;
    border-radius: 0.25rem;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
}

.type-buy {
    background: #dcfce7;
    color: #166534;
}

.type-sell {
    background: #fee2e2;
    color: #991b1b;
}

.type-dividend {
    background: #dbeafe;
    color: #1e40af;
}

/* Colors */
.text-green {
    color: #16a34a;
}

.text-red {
    color: #dc2626;
}

/* Buttons */
.btn-primary {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.625rem 1.25rem;
    background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
    color: white;
    border: none;
    border-radius: 0.5rem;
    font-size: 0.875rem;
    font-weight: 600;
    cursor: pointer;
    text-decoration: none;
    transition: all 0.2s;
}

.btn-primary:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.btn-sm {
    padding: 0.5rem 1rem;
    font-size: 0.8125rem;
}

/* Responsive */
@media (max-width: 768px) {
    .summary-cards {
        grid-template-columns: repeat(2, 1fr);
    }
}
</style>

<!-- Dark Mode Styles (unscoped) -->
<style>
/* ==================== ACCOUNT DETAIL DARK MODE ==================== */

/* Page Background */
.dark .account-detail-page {
    background: #0f0f12;
}

/* Header */
.dark .page-title {
    color: #fafafa;
}

.dark .page-subtitle {
    color: #a1a1aa;
}

.dark .back-button {
    background: #2a2a30;
    color: #a1a1aa;
}

.dark .back-button:hover {
    background: #42424a;
    color: #fafafa;
}

/* Summary Cards */
.dark .summary-card {
    background: #1a1a1f;
    border-color: #2a2a30;
}

.dark .summary-label {
    color: #71717a;
}

.dark .summary-value {
    color: #fafafa;
}

/* Tabs */
.dark .tabs-container {
    background: #1a1a1f;
    border-color: #2a2a30;
}

.dark .tab {
    color: #a1a1aa;
}

.dark .tab:hover {
    color: #fafafa;
}

.dark .tab.is-active {
    color: #fafafa;
    border-bottom-color: #fafafa;
}

/* Tab Content */
.dark .tab-content {
    background: #1a1a1f;
    border-color: #2a2a30;
}

/* Empty State */
.dark .empty-state-sm {
    color: #71717a;
}

/* Table */
.dark .data-table th {
    color: #71717a;
    border-bottom-color: #2a2a30;
}

.dark .data-table td {
    color: #fafafa;
    border-bottom-color: #2a2a30;
}

.dark .data-table tbody tr:hover {
    background: #2a2a30;
}

/* Holding Name */
.dark .holding-name .symbol {
    color: #71717a;
}

/* Activity Type */
.dark .type-buy {
    background: #14532d;
    color: #4ade80;
}

.dark .type-sell {
    background: #7f1d1d;
    color: #fca5a5;
}

.dark .type-dividend {
    background: #1e3a5f;
    color: #93c5fd;
}
</style>
