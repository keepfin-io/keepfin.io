<template>
    <!-- Full Page Spinner -->
    <FullPageSpinner :visible="loading || saving || deleting" :message="spinnerMessage" />

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

        <!-- Inline Notification -->
        <div v-if="notification" class="inline-notification" :class="[`notification-${notification.variant}`]">
            <div class="notification-icon">
                <!-- Success Icon -->
                <svg v-if="notification.variant === 'success'" xmlns="http://www.w3.org/2000/svg" width="20" height="20"
                    viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                    stroke-linejoin="round">
                    <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
                    <polyline points="22 4 12 14.01 9 11.01"></polyline>
                </svg>
                <!-- Error Icon -->
                <svg v-else xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <circle cx="12" cy="12" r="10"></circle>
                    <line x1="15" y1="9" x2="9" y2="15"></line>
                    <line x1="9" y1="9" x2="15" y2="15"></line>
                </svg>
            </div>
            <div class="notification-content">
                <p class="notification-title">{{ notification.title }}</p>
                <p v-if="notification.description" class="notification-description">{{ notification.description }}</p>
            </div>
            <button @click="clearNotification" class="notification-close">
                <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <line x1="18" y1="6" x2="6" y2="18"></line>
                    <line x1="6" y1="6" x2="18" y2="18"></line>
                </svg>
            </button>
        </div>

        <!-- Error State -->
        <div v-if="!loading && error" class="error-container">
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
        <div v-if="!loading && !error && accounts.length === 0" class="empty-state">
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

        <!-- Accounts Table -->
        <div v-if="!loading && !error && accounts.length > 0" class="table-container">
            <table class="data-table">
                <thead>
                    <tr>
                        <th class="th-account">Account</th>
                        <th class="th-platform">Platform</th>
                        <th class="th-currency">Currency</th>
                        <th class="th-balance">Balance</th>
                        <th class="th-status">Status</th>
                        <th class="th-actions">Actions</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="account in accounts" :key="account.id" class="table-row"
                        :class="{ 'is-excluded': account.is_excluded }">
                        <!-- Account -->
                        <td class="td-account">
                            <div class="account-info">
                                <span class="account-name">{{ account.account_name }}</span>
                                <span v-if="account.description" class="account-description">{{
                                    account.description }}</span>
                            </div>
                        </td>

                        <!-- Platform -->
                        <td class="td-platform">
                            <span class="platform-badge">
                                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24"
                                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                    stroke-linejoin="round">
                                    <rect x="2" y="5" width="20" height="14" rx="2"></rect>
                                    <line x1="2" y1="10" x2="22" y2="10"></line>
                                </svg>
                                {{ account.platform || 'General' }}
                            </span>
                        </td>

                        <!-- Currency -->
                        <td class="td-currency">
                            <span class="currency-badge">{{ account.currency }}</span>
                        </td>

                        <!-- Balance -->
                        <td class="td-balance">
                            <span class="balance-amount" :class="{ 'negative': account.initial_balance < 0 }">
                                {{ formatCurrency(account.initial_balance, account.currency) }}
                            </span>
                        </td>

                        <!-- Status -->
                        <td class="td-status">
                            <span v-if="account.is_excluded" class="status-badge status-excluded">
                                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24"
                                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                    stroke-linejoin="round">
                                    <circle cx="12" cy="12" r="10"></circle>
                                    <line x1="4.93" y1="4.93" x2="19.07" y2="19.07"></line>
                                </svg>
                                Excluded
                            </span>
                            <span v-else class="status-badge status-active">
                                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24"
                                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                    stroke-linejoin="round">
                                    <polyline points="20 6 9 17 4 12"></polyline>
                                </svg>
                                Active
                            </span>
                        </td>

                        <!-- Actions -->
                        <td class="td-actions">
                            <div class="action-buttons">
                                <button @click="openEditModal(account)" class="btn-action btn-edit" title="Edit">
                                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"
                                        fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                        stroke-linejoin="round">
                                        <path d="M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"></path>
                                    </svg>
                                </button>
                                <button @click="confirmDelete(account)" class="btn-action btn-delete" title="Delete">
                                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"
                                        fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                        stroke-linejoin="round">
                                        <polyline points="3 6 5 6 21 6"></polyline>
                                        <path
                                            d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2">
                                        </path>
                                    </svg>
                                </button>
                            </div>
                        </td>
                    </tr>
                </tbody>
            </table>

            <!-- Table Footer -->
            <div class="table-footer">
                <span class="table-count">{{ accounts.length }} account{{ accounts.length !== 1 ? 's' : '' }}</span>
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
        <ConfirmModal :visible="showDeleteModal" variant="danger" title="Delete Account"
            :message="`Are you sure you want to delete ${accountToDelete?.account_name}? This action cannot be undone.`"
            confirmText="Delete Account" cancelText="Cancel" loadingText="Deleting..." :loading="deleting"
            @confirm="deleteAccount" @cancel="cancelDelete" />

        <!-- Update Confirmation Modal -->
        <ConfirmModal :visible="showUpdateModal" variant="warning" title="Update Account"
            :message="`Are you sure you want to update ${form.account_name}?`" confirmText="Update Account"
            cancelText="Cancel" loadingText="Updating..." :loading="saving" @confirm="confirmUpdateAccount"
            @cancel="cancelUpdate" />
    </main>
</template>

<script>
import { accountAPI } from '@/services/accountApi'
import FullPageSpinner from '@/components/ui/FullPageSpinner.vue'
import ConfirmModal from '@/components/ui/ConfirmModal.vue'

export default {
    name: 'AccountsPage',
    components: {
        FullPageSpinner,
        ConfirmModal
    },
    computed: {
        spinnerMessage() {
            if (this.loading) return 'Loading accounts...'
            if (this.saving) return 'Saving account...'
            if (this.deleting) return 'Deleting account...'
            return ''
        }
    },
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
            notification: null,
            notificationTimer: null,
            showUpdateModal: false,
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
            // If editing, show confirmation modal first
            if (this.editingAccount) {
                this.showUpdateModal = true
                return
            }
            // For create, proceed directly
            await this.performSave()
        },

        cancelUpdate() {
            this.showUpdateModal = false
        },

        async confirmUpdateAccount() {
            this.showUpdateModal = false
            await this.performSave()
        },

        async performSave() {
            this.saving = true
            this.formError = null
            try {
                if (this.editingAccount) {
                    await accountAPI.updateAccount(this.editingAccount.id, this.form)
                    this.showNotification('success', 'Account updated', `${this.form.account_name} has been updated successfully.`)
                } else {
                    await accountAPI.createAccount(this.form)
                    this.showNotification('success', 'Account created', `${this.form.account_name} has been created successfully.`)
                }
                this.closeModal()
                await this.fetchAccounts()
            } catch (err) {
                this.formError = err.message || 'Failed to save account'
                this.showNotification('error', 'Error', this.formError)
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
            const accountName = this.accountToDelete.account_name
            try {
                await accountAPI.deleteAccount(this.accountToDelete.id)
                this.cancelDelete()
                await this.fetchAccounts()
                this.showNotification('success', 'Account deleted', `${accountName} has been deleted.`)
            } catch (err) {
                this.formError = err.message || 'Failed to delete account'
                this.showNotification('error', 'Error', this.formError)
            } finally {
                this.deleting = false
            }
        },

        showNotification(variant, title, description = '') {
            // Clear existing timer
            if (this.notificationTimer) {
                clearTimeout(this.notificationTimer)
            }
            // Set notification
            this.notification = { variant, title, description }
            // Auto-clear after 5 seconds
            this.notificationTimer = setTimeout(() => {
                this.clearNotification()
            }, 5000)
        },

        clearNotification() {
            this.notification = null
            if (this.notificationTimer) {
                clearTimeout(this.notificationTimer)
                this.notificationTimer = null
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

/* Inline Notification */
.inline-notification {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
    padding: 1rem 1.25rem;
    border-radius: 0.75rem;
    border: 1px solid;
    margin-bottom: 1.5rem;
}

.notification-success {
    background: linear-gradient(135deg, #f0fdf4 0%, #dcfce7 100%);
    border-color: #86efac;
    color: #166534;
}

.notification-success .notification-icon {
    color: #22c55e;
}

.notification-error {
    background: linear-gradient(135deg, #fef2f2 0%, #fee2e2 100%);
    border-color: #fca5a5;
    color: #991b1b;
}

.notification-error .notification-icon {
    color: #ef4444;
}

.notification-icon {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-top: 0.125rem;
}

.notification-content {
    flex: 1;
    min-width: 0;
}

.notification-title {
    font-size: 0.9375rem;
    font-weight: 600;
    margin: 0;
    line-height: 1.4;
}

.notification-description {
    font-size: 0.8125rem;
    margin: 0.25rem 0 0 0;
    opacity: 0.9;
    line-height: 1.4;
}

.notification-close {
    flex-shrink: 0;
    padding: 0.25rem;
    background: transparent;
    border: none;
    border-radius: 0.375rem;
    cursor: pointer;
    opacity: 0.5;
    transition: all 0.2s;
    color: inherit;
}

.notification-close:hover {
    opacity: 1;
    background: rgba(0, 0, 0, 0.05);
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

/* Data Table */
.table-container {
    background: white;
    border-radius: 1rem;
    border: 1px solid #e2e8f0;
    overflow: hidden;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.data-table {
    width: 100%;
    border-collapse: collapse;
}

.data-table thead {
    background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
    border-bottom: 1px solid #e2e8f0;
}

.data-table th {
    padding: 0.875rem 1rem;
    text-align: left;
    font-size: 0.75rem;
    font-weight: 600;
    color: #64748b;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    white-space: nowrap;
}

.data-table th.th-balance,
.data-table th.th-actions {
    text-align: right;
}

.data-table th.th-currency,
.data-table th.th-status {
    text-align: center;
}

.table-row {
    border-bottom: 1px solid #f1f5f9;
    transition: all 0.15s ease;
}

.table-row:last-child {
    border-bottom: none;
}

.table-row:hover {
    background: #f8fafc;
}

.table-row.is-excluded {
    opacity: 0.7;
    background: #fafafa;
}

.data-table td {
    padding: 1rem;
    vertical-align: middle;
}

/* Account Info */
.account-info {
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
    min-width: 0;
}

.account-name {
    font-size: 0.9375rem;
    font-weight: 600;
    color: #1e293b;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.account-description {
    font-size: 0.75rem;
    color: #94a3b8;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 200px;
}

/* Platform Badge */
.platform-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    padding: 0.375rem 0.75rem;
    background: #f1f5f9;
    border-radius: 6px;
    font-size: 0.8125rem;
    font-weight: 500;
    color: #475569;
}

.platform-badge svg {
    opacity: 0.6;
}

/* Currency Badge */
.td-currency {
    text-align: center;
}

.currency-badge {
    display: inline-block;
    padding: 0.25rem 0.625rem;
    background: linear-gradient(135deg, #3b82f6 0%, #6366f1 100%);
    border-radius: 6px;
    font-size: 0.6875rem;
    font-weight: 700;
    color: white;
    letter-spacing: 0.05em;
}

/* Balance */
.td-balance {
    text-align: right;
}

.balance-amount {
    font-size: 1rem;
    font-weight: 700;
    color: #059669;
    font-variant-numeric: tabular-nums;
}

.balance-amount.negative {
    color: #dc2626;
}

/* Status Badge */
.td-status {
    text-align: center;
}

.status-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.25rem 0.625rem;
    border-radius: 20px;
    font-size: 0.6875rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.03em;
}

.status-active {
    background: #dcfce7;
    color: #16a34a;
}

.status-excluded {
    background: #fef3c7;
    color: #d97706;
}

/* Action Buttons */
.td-actions {
    text-align: right;
}

.action-buttons {
    display: flex;
    justify-content: flex-end;
    gap: 0.375rem;
}

.btn-action {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    padding: 0;
    border: none;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.15s ease;
}

.btn-edit {
    background: #f1f5f9;
    color: #64748b;
}

.btn-edit:hover {
    background: #e0f2fe;
    color: #0284c7;
    transform: scale(1.05);
}

.btn-delete {
    background: #f1f5f9;
    color: #64748b;
}

.btn-delete:hover {
    background: #fee2e2;
    color: #dc2626;
    transform: scale(1.05);
}

/* Table Footer */
.table-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.875rem 1rem;
    background: #f8fafc;
    border-top: 1px solid #e2e8f0;
}

.table-count {
    font-size: 0.8125rem;
    color: #64748b;
}

/* Responsive Table */
@media (max-width: 900px) {

    .th-platform,
    .td-platform,
    .th-status,
    .td-status {
        display: none;
    }
}

@media (max-width: 640px) {

    .data-table th,
    .data-table td {
        padding: 0.75rem 0.625rem;
    }

    .account-avatar {
        width: 36px;
        height: 36px;
        font-size: 0.875rem;
    }

    .account-name {
        font-size: 0.875rem;
    }

    .account-description {
        display: none;
    }

    .balance-amount {
        font-size: 0.875rem;
    }

    .th-currency,
    .td-currency {
        display: none;
    }
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
