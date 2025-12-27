<template>
    <main class="accounts-page">
        <!-- Header -->
        <div class="page-header">
            <div class="header-content">
                <h1 class="page-title">Accounts</h1>
                <p class="page-subtitle">Manage your financial accounts</p>
            </div>
            <button @click="openCreateModal" class="btn-primary">
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <line x1="12" y1="5" x2="12" y2="19"></line>
                    <line x1="5" y1="12" x2="19" y2="12"></line>
                </svg>
                Add Account
            </button>
        </div>

        <!-- Loading State -->
        <div v-if="loading" class="loading-container">
            <div class="spinner"></div>
            <p>Loading accounts...</p>
        </div>

        <!-- Error State -->
        <div v-else-if="error" class="error-container">
            <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none"
                stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="10"></circle>
                <line x1="12" y1="8" x2="12" y2="12"></line>
                <line x1="12" y1="16" x2="12.01" y2="16"></line>
            </svg>
            <p>{{ error }}</p>
            <button @click="fetchAccounts" class="btn-secondary">Retry</button>
        </div>

        <!-- Empty State -->
        <div v-else-if="accounts.length === 0" class="empty-state">
            <div class="empty-state-illustration">
                <div class="illustration-bg"></div>
                <div class="illustration-icon">
                    <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                        <rect x="2" y="5" width="20" height="14" rx="2"></rect>
                        <line x1="2" y1="10" x2="22" y2="10"></line>
                    </svg>
                </div>
            </div>
            <h3 class="empty-state-title">No accounts yet</h3>
            <p class="empty-state-description">Start managing your finances by adding your first account.<br>Track
                balances across multiple currencies and platforms.</p>
            <button @click="openCreateModal" class="btn-primary btn-lg">
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <line x1="12" y1="5" x2="12" y2="19"></line>
                    <line x1="5" y1="12" x2="19" y2="12"></line>
                </svg>
                Create Your First Account
            </button>
        </div>

        <!-- Accounts Grid -->
        <div v-else class="accounts-grid">
            <div v-for="account in accounts" :key="account.id" class="account-card">
                <div class="card-header">
                    <div class="account-icon" :style="{ background: getAccountColor(account.platform) }">
                        {{ getAccountInitial(account.account_name) }}
                    </div>
                    <div class="account-info">
                        <h3 class="account-name">{{ account.account_name }}</h3>
                        <span class="account-platform">{{ account.platform || 'General' }}</span>
                    </div>
                    <div class="card-actions">
                        <button @click="openEditModal(account)" class="btn-icon" title="Edit">
                            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round">
                                <path d="M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"></path>
                            </svg>
                        </button>
                        <button @click="confirmDelete(account)" class="btn-icon btn-danger" title="Delete">
                            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round">
                                <polyline points="3 6 5 6 21 6"></polyline>
                                <path
                                    d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2">
                                </path>
                            </svg>
                        </button>
                    </div>
                </div>
                <div class="card-body">
                    <div class="balance-section">
                        <span class="balance-label">Initial Balance</span>
                        <span class="balance-amount" :class="{ 'negative': account.initial_balance < 0 }">
                            {{ formatCurrency(account.initial_balance, account.currency) }}
                        </span>
                    </div>
                    <div class="currency-badge">{{ account.currency }}</div>
                </div>
                <div v-if="account.description" class="card-footer">
                    <p class="account-description">{{ account.description }}</p>
                </div>
                <div v-if="account.is_excluded" class="excluded-badge">
                    Excluded from totals
                </div>
            </div>
        </div>

        <!-- Create/Edit Modal -->
        <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
            <div class="modal">
                <div class="modal-header">
                    <h2>{{ editingAccount ? 'Edit Account' : 'Create Account' }}</h2>
                    <button @click="closeModal" class="btn-close">
                        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <line x1="18" y1="6" x2="6" y2="18"></line>
                            <line x1="6" y1="6" x2="18" y2="18"></line>
                        </svg>
                    </button>
                </div>
                <form @submit.prevent="saveAccount" class="modal-body">
                    <div class="form-group">
                        <label for="accountName">Account Name *</label>
                        <input type="text" id="accountName" v-model="form.account_name" required maxlength="100"
                            placeholder="e.g., Main Checking Account">
                    </div>

                    <div class="form-row">
                        <div class="form-group">
                            <label for="currency">Currency *</label>
                            <select id="currency" v-model="form.currency" required>
                                <option value="">Select currency</option>
                                <option value="USD">USD - US Dollar</option>
                                <option value="EUR">EUR - Euro</option>
                                <option value="GBP">GBP - British Pound</option>
                                <option value="TRY">TRY - Turkish Lira</option>
                                <option value="JPY">JPY - Japanese Yen</option>
                                <option value="CAD">CAD - Canadian Dollar</option>
                                <option value="AUD">AUD - Australian Dollar</option>
                                <option value="CHF">CHF - Swiss Franc</option>
                            </select>
                        </div>
                        <div class="form-group">
                            <label for="initialBalance">Initial Balance</label>
                            <input type="number" id="initialBalance" v-model.number="form.initial_balance" step="0.01"
                                placeholder="0.00">
                        </div>
                    </div>

                    <div class="form-group">
                        <label for="platform">Platform</label>
                        <input type="text" id="platform" v-model="form.platform" maxlength="100"
                            placeholder="e.g., Bank of America, PayPal">
                    </div>

                    <div class="form-group">
                        <label for="description">Description</label>
                        <textarea id="description" v-model="form.description" maxlength="255" rows="3"
                            placeholder="Optional notes about this account"></textarea>
                    </div>

                    <div class="form-group checkbox-group">
                        <label class="checkbox-label">
                            <input type="checkbox" v-model="form.is_excluded">
                            <span class="checkmark"></span>
                            Exclude from totals
                        </label>
                        <span class="checkbox-hint">This account won't be included in total balance calculations</span>
                    </div>

                    <div v-if="formError" class="form-error">{{ formError }}</div>

                    <div class="modal-footer">
                        <button type="button" @click="closeModal" class="btn-secondary">Cancel</button>
                        <button type="submit" class="btn-primary" :disabled="saving">
                            {{ saving ? 'Saving...' : (editingAccount ? 'Update Account' : 'Create Account') }}
                        </button>
                    </div>
                </form>
            </div>
        </div>

        <!-- Delete Confirmation Modal -->
        <div v-if="showDeleteModal" class="modal-overlay" @click.self="cancelDelete">
            <div class="modal modal-delete">
                <div class="modal-header">
                    <h2>Delete Account</h2>
                </div>
                <div class="modal-body">
                    <p>Are you sure you want to delete <strong>{{ accountToDelete?.account_name }}</strong>?</p>
                    <p class="delete-warning">This action cannot be undone.</p>
                </div>
                <div class="modal-footer">
                    <button @click="cancelDelete" class="btn-secondary">Cancel</button>
                    <button @click="deleteAccount" class="btn-danger" :disabled="deleting">
                        {{ deleting ? 'Deleting...' : 'Delete Account' }}
                    </button>
                </div>
            </div>
        </div>
    </main>
</template>

<script>
import { accountAPI } from '@/services/accountApi'

export default {
    name: 'AccountsPage',
    data() {
        return {
            accounts: [],
            loading: true,
            error: null,
            showModal: false,
            showDeleteModal: false,
            editingAccount: null,
            accountToDelete: null,
            saving: false,
            deleting: false,
            formError: null,
            form: {
                account_name: '',
                currency: '',
                initial_balance: 0,
                platform: '',
                description: '',
                is_excluded: false
            }
        }
    },
    mounted() {
        this.fetchAccounts()
    },
    methods: {
        async fetchAccounts() {
            this.loading = true
            this.error = null
            try {
                const response = await accountAPI.getAccounts()
                this.accounts = response.accounts || []
            } catch (err) {
                this.error = err.message || 'Failed to load accounts'
            } finally {
                this.loading = false
            }
        },

        openCreateModal() {
            this.editingAccount = null
            this.form = {
                account_name: '',
                currency: '',
                initial_balance: 0,
                platform: '',
                description: '',
                is_excluded: false
            }
            this.formError = null
            this.showModal = true
        },

        openEditModal(account) {
            this.editingAccount = account
            this.form = {
                account_name: account.account_name,
                currency: account.currency,
                initial_balance: account.initial_balance,
                platform: account.platform || '',
                description: account.description || '',
                is_excluded: account.is_excluded
            }
            this.formError = null
            this.showModal = true
        },

        closeModal() {
            this.showModal = false
            this.editingAccount = null
            this.formError = null
        },

        async saveAccount() {
            this.saving = true
            this.formError = null
            try {
                if (this.editingAccount) {
                    await accountAPI.updateAccount(this.editingAccount.id, this.form)
                } else {
                    await accountAPI.createAccount(this.form)
                }
                this.closeModal()
                await this.fetchAccounts()
            } catch (err) {
                this.formError = err.message || 'Failed to save account'
            } finally {
                this.saving = false
            }
        },

        confirmDelete(account) {
            this.accountToDelete = account
            this.showDeleteModal = true
        },

        cancelDelete() {
            this.showDeleteModal = false
            this.accountToDelete = null
        },

        async deleteAccount() {
            if (!this.accountToDelete) return
            this.deleting = true
            try {
                await accountAPI.deleteAccount(this.accountToDelete.id)
                this.cancelDelete()
                await this.fetchAccounts()
            } catch (err) {
                this.formError = err.message || 'Failed to delete account'
            } finally {
                this.deleting = false
            }
        },

        formatCurrency(amount, currency) {
            return new Intl.NumberFormat('en-US', {
                style: 'currency',
                currency: currency || 'USD'
            }).format(amount || 0)
        },

        getAccountInitial(name) {
            return name ? name.charAt(0).toUpperCase() : '?'
        },

        getAccountColor(platform) {
            const colors = {
                'paypal': 'linear-gradient(135deg, #003087 0%, #009cde 100%)',
                'stripe': 'linear-gradient(135deg, #635bff 0%, #a960ee 100%)',
                'wise': 'linear-gradient(135deg, #9fe870 0%, #37517e 100%)',
                'revolut': 'linear-gradient(135deg, #0075eb 0%, #191c1f 100%)',
                'default': 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)'
            }
            const key = (platform || '').toLowerCase()
            return colors[key] || colors.default
        }
    }
}
</script>

<style scoped>
.accounts-page {
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
    flex-wrap: wrap;
    gap: 1rem;
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
    transition: all 0.2s;
    box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);
}

.btn-primary:hover:not(:disabled) {
    transform: translateY(-1px);
    box-shadow: 0 4px 8px rgba(59, 130, 246, 0.4);
}

.btn-primary:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

.btn-secondary {
    padding: 0.625rem 1.25rem;
    background: #f3f4f6;
    color: #374151;
    border: 1px solid #e5e7eb;
    border-radius: 0.5rem;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
}

.btn-secondary:hover {
    background: #e5e7eb;
}

.btn-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 2rem;
    height: 2rem;
    padding: 0;
    background: transparent;
    border: none;
    border-radius: 0.375rem;
    color: #6b7280;
    cursor: pointer;
    transition: all 0.2s;
}

.btn-icon:hover {
    background: #f3f4f6;
    color: #1f2937;
}

.btn-icon.btn-danger:hover {
    background: #fef2f2;
    color: #dc2626;
}

.btn-danger {
    background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
    color: white;
    border: none;
    padding: 0.625rem 1.25rem;
    border-radius: 0.5rem;
    font-size: 0.875rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
}

.btn-danger:hover:not(:disabled) {
    transform: translateY(-1px);
}

.btn-danger:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

.btn-close {
    background: transparent;
    border: none;
    color: #6b7280;
    cursor: pointer;
    padding: 0.25rem;
    border-radius: 0.375rem;
    transition: all 0.2s;
}

.btn-close:hover {
    background: #f3f4f6;
    color: #1f2937;
}

/* Loading State */
.loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 4rem;
    color: #6b7280;
}

.spinner {
    width: 40px;
    height: 40px;
    border: 3px solid #e5e7eb;
    border-top-color: #3b82f6;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin-bottom: 1rem;
}

@keyframes spin {
    to {
        transform: rotate(360deg);
    }
}

/* Error State */
.error-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 4rem;
    color: #dc2626;
    text-align: center;
}

.error-container svg {
    margin-bottom: 1rem;
}

.error-container p {
    margin-bottom: 1rem;
}

/* Empty State */
.empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 4rem 2rem;
    text-align: center;
    background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
    border-radius: 1.5rem;
    border: 1px solid #e2e8f0;
    position: relative;
    overflow: hidden;
}

.empty-state::before {
    content: '';
    position: absolute;
    top: -50%;
    left: -50%;
    width: 200%;
    height: 200%;
    background: radial-gradient(circle at center, rgba(59, 130, 246, 0.03) 0%, transparent 50%);
    pointer-events: none;
}

.empty-state-illustration {
    position: relative;
    margin-bottom: 1.5rem;
}

.illustration-bg {
    width: 120px;
    height: 120px;
    background: linear-gradient(135deg, #dbeafe 0%, #e0e7ff 100%);
    border-radius: 50%;
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    animation: pulse-soft 3s ease-in-out infinite;
}

.illustration-icon {
    position: relative;
    width: 100px;
    height: 100px;
    background: linear-gradient(135deg, #3b82f6 0%, #6366f1 100%);
    border-radius: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    box-shadow: 0 20px 40px -10px rgba(59, 130, 246, 0.4);
    transform: rotate(-6deg);
    transition: transform 0.3s ease;
}

.empty-state:hover .illustration-icon {
    transform: rotate(0deg) scale(1.05);
}

@keyframes pulse-soft {

    0%,
    100% {
        transform: translate(-50%, -50%) scale(1);
        opacity: 1;
    }

    50% {
        transform: translate(-50%, -50%) scale(1.1);
        opacity: 0.7;
    }
}

.empty-state-title {
    font-size: 1.5rem;
    font-weight: 700;
    color: #1e293b;
    margin: 0 0 0.75rem 0;
}

.empty-state-description {
    color: #64748b;
    margin: 0 0 2rem 0;
    font-size: 1rem;
    line-height: 1.6;
    max-width: 400px;
}

.btn-lg {
    padding: 0.875rem 1.75rem;
    font-size: 1rem;
}

/* Accounts Grid */
.accounts-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 1.5rem;
}

.account-card {
    background: white;
    border-radius: 1rem;
    border: 1px solid #e5e7eb;
    overflow: hidden;
    transition: all 0.2s;
    position: relative;
}

.account-card:hover {
    border-color: #d1d5db;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.card-header {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 1.25rem;
    border-bottom: 1px solid #f3f4f6;
}

.account-icon {
    width: 48px;
    height: 48px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    font-size: 1.25rem;
    font-weight: 700;
    flex-shrink: 0;
}

.account-info {
    flex: 1;
    min-width: 0;
}

.account-name {
    font-size: 1rem;
    font-weight: 600;
    color: #1f2937;
    margin: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.account-platform {
    font-size: 0.75rem;
    color: #9ca3af;
    text-transform: uppercase;
    letter-spacing: 0.05em;
}

.card-actions {
    display: flex;
    gap: 0.25rem;
}

.card-body {
    padding: 1.25rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.balance-section {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
}

.balance-label {
    font-size: 0.75rem;
    color: #6b7280;
    text-transform: uppercase;
    letter-spacing: 0.05em;
}

.balance-amount {
    font-size: 1.5rem;
    font-weight: 700;
    color: #059669;
}

.balance-amount.negative {
    color: #dc2626;
}

.currency-badge {
    padding: 0.25rem 0.75rem;
    background: #f3f4f6;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 600;
    color: #6b7280;
}

.card-footer {
    padding: 0 1.25rem 1.25rem;
}

.account-description {
    font-size: 0.875rem;
    color: #6b7280;
    margin: 0;
    line-height: 1.5;
}

.excluded-badge {
    position: absolute;
    top: 0.75rem;
    right: 0.75rem;
    padding: 0.125rem 0.5rem;
    background: #fef3c7;
    color: #92400e;
    border-radius: 0.25rem;
    font-size: 0.625rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
}

/* Modal */
.modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1rem;
    z-index: 1000;
    backdrop-filter: blur(4px);
}

.modal {
    background: white;
    border-radius: 1rem;
    width: 100%;
    max-width: 560px;
    max-height: 90vh;
    overflow-y: auto;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
}

.modal-delete {
    max-width: 400px;
}

.modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.25rem 1.5rem;
    border-bottom: 1px solid #e5e7eb;
}

.modal-header h2 {
    font-size: 1.125rem;
    font-weight: 600;
    color: #1f2937;
    margin: 0;
}

.modal-body {
    padding: 1.5rem;
}

.modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 0.75rem;
    padding-top: 1.5rem;
    margin-top: 1rem;
    border-top: 1px solid #e5e7eb;
}

/* Form */
.form-group {
    margin-bottom: 1.25rem;
}

.form-group:last-of-type {
    margin-bottom: 0;
}

.form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
}

.form-group label {
    display: block;
    font-size: 0.875rem;
    font-weight: 500;
    color: #374151;
    margin-bottom: 0.5rem;
}

.form-group input,
.form-group select,
.form-group textarea {
    width: 100%;
    padding: 0.625rem 0.875rem;
    border: 1px solid #d1d5db;
    border-radius: 0.5rem;
    font-size: 0.875rem;
    color: #1f2937;
    background: white;
    transition: all 0.2s;
}

.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-group input::placeholder,
.form-group textarea::placeholder {
    color: #9ca3af;
}

.form-group textarea {
    resize: vertical;
    min-height: 80px;
}

.checkbox-group {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
}

.checkbox-label {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    cursor: pointer;
    font-weight: 400 !important;
}

.checkbox-label input[type="checkbox"] {
    width: 1rem;
    height: 1rem;
    accent-color: #3b82f6;
}

.checkbox-hint {
    font-size: 0.75rem;
    color: #6b7280;
    margin-left: 1.5rem;
}

.form-error {
    padding: 0.75rem;
    background: #fef2f2;
    border: 1px solid #fecaca;
    border-radius: 0.5rem;
    color: #dc2626;
    font-size: 0.875rem;
    margin-top: 1rem;
}

.delete-warning {
    color: #dc2626;
    font-size: 0.875rem;
    margin-top: 0.5rem;
}

/* Responsive */
@media (max-width: 640px) {
    .accounts-page {
        padding: 1rem;
    }

    .page-header {
        flex-direction: column;
        align-items: flex-start;
    }

    .accounts-grid {
        grid-template-columns: 1fr;
    }

    .form-row {
        grid-template-columns: 1fr;
    }
}
</style>
