<template>
    <div v-if="show" class="modal-overlay">
        <div class="modal-content">
            <div class="modal-header">
                <h2>Add Holding</h2>
                <button class="close-btn" @click="close">&times;</button>
            </div>

            <div class="modal-body">
                <!-- Step 1: Search Asset -->
                <div v-if="!selectedAsset" class="search-step">
                    <div class="form-group">
                        <label>Search Asset</label>
                        <div class="search-input-container">
                            <input type="text" v-model="searchQuery" @input="onSearchInput"
                                placeholder="Search by symbol or name (e.g., AAPL, Tesla)" class="search-input"
                                :disabled="searchLoading" />
                            <div v-if="searchLoading" class="search-spinner"></div>
                        </div>
                    </div>

                    <!-- Search Results -->
                    <div v-if="searchResults.length > 0" class="search-results">
                        <div v-for="result in searchResults" :key="`${result.Code}-${result.Exchange}`"
                            class="search-result-item" @click="selectAsset(result)">
                            <div class="result-main">
                                <span class="result-symbol">{{ result.Code }}</span>
                                <span class="result-name">{{ result.Name }}</span>
                            </div>
                            <div class="result-meta">
                                <span class="result-exchange">{{ result.Exchange }}</span>
                                <span class="result-type">{{ result.Type }}</span>
                            </div>
                        </div>
                    </div>

                    <div v-if="searchQuery.length >= 2 && !searchLoading && searchResults.length === 0"
                        class="no-results">
                        No assets found for "{{ searchQuery }}"
                    </div>
                </div>

                <!-- Step 2: Configure Holding -->
                <div v-else class="configure-step">
                    <!-- Selected Asset Display -->
                    <div class="selected-asset">
                        <div class="asset-info">
                            <span class="asset-symbol">{{ selectedAsset.Code }}</span>
                            <span class="asset-name">{{ selectedAsset.Name }}</span>
                            <span class="asset-exchange">{{ selectedAsset.Exchange }}</span>
                        </div>
                        <button class="change-btn" @click="clearSelection">Change</button>
                    </div>

                    <!-- Current Price Display -->
                    <div v-if="livePrice" class="current-price">
                        <span class="price-label">Current Price:</span>
                        <span class="price-value">{{ formatCurrency(livePrice.close, selectedAsset.Currency) }}</span>
                        <span :class="['price-change', livePrice.change >= 0 ? 'positive' : 'negative']">
                            {{ livePrice.change >= 0 ? '+' : '' }}{{ livePrice.change_p?.toFixed(2) }}%
                        </span>
                    </div>

                    <!-- Form -->
                    <form @submit.prevent="submitHolding">
                        <div class="form-row">
                            <div class="form-group">
                                <label for="quantity">Quantity *</label>
                                <input id="quantity" type="number" v-model.number="form.quantity" step="any"
                                    min="0.00000001" required placeholder="e.g., 10" />
                            </div>
                            <div class="form-group">
                                <label for="buyPrice">Buy Price *</label>
                                <input id="buyPrice" type="number" v-model.number="form.buy_price" step="any" min="0"
                                    required placeholder="e.g., 150.00" />
                            </div>
                        </div>

                        <div class="form-row">
                            <div class="form-group">
                                <label for="fee">Fee</label>
                                <input id="fee" type="number" v-model.number="form.fee" step="any" min="0"
                                    placeholder="e.g., 1.50" />
                            </div>
                            <div class="form-group">
                                <label for="buyCurrency">Currency</label>
                                <select id="buyCurrency" v-model="form.buy_currency">
                                    <option value="USD">USD</option>
                                    <option value="EUR">EUR</option>
                                    <option value="GBP">GBP</option>
                                    <option value="TRY">TRY</option>
                                </select>
                            </div>
                        </div>

                        <div class="form-group">
                            <label for="notes">Notes</label>
                            <textarea id="notes" v-model="form.notes" rows="2"
                                placeholder="Optional notes about this holding"></textarea>
                        </div>

                        <div class="form-group">
                            <label for="notes">Notes</label>
                            <textarea id="notes" v-model="form.notes" rows="2"
                                placeholder="Optional notes about this holding"></textarea>
                        </div>

                        <!-- Total Value Preview -->
                        <div v-if="form.quantity && form.buy_price" class="total-preview">
                            <span>Total Cost:</span>
                            <strong>{{ formatCurrency((form.quantity * form.buy_price) + (form.fee || 0),
                                form.buy_currency) }}</strong>
                        </div>
                    </form>
                </div>
            </div>

            <div class="modal-footer">
                <button type="button" class="btn-secondary" @click="close">Cancel</button>
                <button type="button" class="btn-primary" @click="submitHolding" :disabled="!canSubmit || submitting">
                    {{ submitting ? 'Adding...' : 'Add Holding' }}
                </button>
            </div>
        </div>
    </div>
</template>

<script>
import { marketAPI } from '@/services/marketApi'
import { holdingAPI } from '@/services/holdingApi'

export default {
    name: 'AddHoldingModal',
    props: {
        show: {
            type: Boolean,
            default: false
        },
        accountId: {
            type: String,
            required: true
        }
    },
    emits: ['close', 'success'],
    data() {
        return {
            searchQuery: '',
            searchResults: [],
            searchLoading: false,
            searchTimeout: null,
            selectedAsset: null,
            livePrice: null,
            priceLoading: false,
            submitting: false,
            form: {
                symbol: '',
                quantity: null,
                buy_price: null,
                fee: 0,
                buy_currency: 'USD',
                notes: ''
            }
        }
    },
    computed: {
        canSubmit() {
            return this.selectedAsset &&
                this.form.quantity > 0 &&
                this.form.buy_price >= 0
        }
    },
    watch: {
        show(newVal) {
            if (!newVal) {
                this.resetForm()
            }
        }
    },
    methods: {
        onSearchInput() {
            clearTimeout(this.searchTimeout)

            if (this.searchQuery.length < 2) {
                this.searchResults = []
                return
            }

            this.searchTimeout = setTimeout(() => {
                this.performSearch()
            }, 300) // 300ms debounce
        },

        async performSearch() {
            this.searchLoading = true
            try {
                const response = await marketAPI.searchAssets(this.searchQuery, 10)
                this.searchResults = response.results || []
            } catch (err) {
                console.error('Search failed:', err)
                this.searchResults = []
            } finally {
                this.searchLoading = false
            }
        },

        async selectAsset(asset) {
            this.selectedAsset = asset
            this.searchResults = []
            this.form.symbol = `${asset.Code}.${asset.Exchange}`
            this.form.buy_currency = asset.Currency || 'USD'

            // Fetch live price
            this.priceLoading = true
            try {
                const response = await marketAPI.getLivePrice(`${asset.Code}.${asset.Exchange}`)
                this.livePrice = response.data
                // Pre-fill buy price with current price
                if (this.livePrice?.close) {
                    this.form.buy_price = this.livePrice.close
                }
            } catch (err) {
                console.error('Failed to fetch live price:', err)
            } finally {
                this.priceLoading = false
            }
        },

        clearSelection() {
            this.selectedAsset = null
            this.livePrice = null
            this.searchQuery = ''
            this.form.symbol = ''
            this.form.buy_price = null
        },

        async submitHolding() {
            if (!this.canSubmit || this.submitting) return

            this.submitting = true
            try {
                await holdingAPI.addHolding(this.accountId, {
                    symbol: this.form.symbol,
                    quantity: this.form.quantity,
                    buy_price: this.form.buy_price,
                    fee: this.form.fee || 0,
                    buy_currency: this.form.buy_currency,
                    notes: this.form.notes
                })
                this.$emit('success')
                this.close()
            } catch (err) {
                console.error('Failed to add holding:', err)
                alert(err.message || 'Failed to add holding')
            } finally {
                this.submitting = false
            }
        },

        close() {
            this.$emit('close')
        },

        resetForm() {
            this.searchQuery = ''
            this.searchResults = []
            this.selectedAsset = null
            this.livePrice = null
            this.form = {
                symbol: '',
                quantity: null,
                buy_price: null,
                fee: 0,
                buy_currency: 'USD',
                notes: ''
            }
        },

        formatCurrency(amount, currency) {
            return new Intl.NumberFormat('en-US', {
                style: 'currency',
                currency: currency || 'USD'
            }).format(amount || 0)
        }
    }
}
</script>

<style scoped>
.modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    backdrop-filter: blur(4px);
}

.modal-content {
    background: white;
    border-radius: 1rem;
    width: 100%;
    max-width: 500px;
    max-height: 90vh;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
}

.modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.25rem 1.5rem;
    border-bottom: 1px solid #e2e8f0;
}

.modal-header h2 {
    font-size: 1.25rem;
    font-weight: 600;
    color: #1e293b;
    margin: 0;
}

.close-btn {
    background: none;
    border: none;
    font-size: 1.5rem;
    cursor: pointer;
    color: #64748b;
    padding: 0;
    line-height: 1;
}

.close-btn:hover {
    color: #1e293b;
}

.modal-body {
    padding: 1.5rem;
    overflow-y: auto;
    flex: 1;
}

.modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 0.75rem;
    padding: 1.25rem 1.5rem;
    border-top: 1px solid #e2e8f0;
}

/* Form Styles */
.form-group {
    margin-bottom: 1rem;
}

.form-group label {
    display: block;
    margin-bottom: 0.5rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: #374151;
}

.form-group input,
.form-group select,
.form-group textarea {
    width: 100%;
    padding: 0.625rem 0.875rem;
    border: 1px solid #d1d5db;
    border-radius: 0.5rem;
    font-size: 0.875rem;
    transition: border-color 0.2s, box-shadow 0.2s;
}

.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
}

/* Search Styles */
.search-input-container {
    position: relative;
}

.search-input {
    width: 100%;
    padding: 0.75rem 1rem;
    padding-right: 2.5rem;
    border: 1px solid #d1d5db;
    border-radius: 0.5rem;
    font-size: 0.875rem;
}

.search-spinner {
    position: absolute;
    right: 0.75rem;
    top: 50%;
    transform: translateY(-50%);
    width: 1rem;
    height: 1rem;
    border: 2px solid #e2e8f0;
    border-top-color: #3b82f6;
    border-radius: 50%;
    animation: spin 0.6s linear infinite;
}

@keyframes spin {
    to {
        transform: translateY(-50%) rotate(360deg);
    }
}

.search-results {
    margin-top: 0.5rem;
    border: 1px solid #e2e8f0;
    border-radius: 0.5rem;
    max-height: 250px;
    overflow-y: auto;
}

.search-result-item {
    padding: 0.75rem 1rem;
    cursor: pointer;
    border-bottom: 1px solid #f1f5f9;
    transition: background 0.15s;
}

.search-result-item:last-child {
    border-bottom: none;
}

.search-result-item:hover {
    background: #f8fafc;
}

.result-main {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.25rem;
}

.result-symbol {
    font-weight: 600;
    color: #1e293b;
}

.result-name {
    font-size: 0.875rem;
    color: #64748b;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.result-meta {
    display: flex;
    gap: 0.5rem;
    font-size: 0.75rem;
}

.result-exchange {
    color: #3b82f6;
    font-weight: 500;
}

.result-type {
    color: #94a3b8;
}

.no-results {
    text-align: center;
    padding: 1.5rem;
    color: #64748b;
}

/* Selected Asset */
.selected-asset {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    background: #f8fafc;
    border: 1px solid #e2e8f0;
    border-radius: 0.5rem;
    margin-bottom: 1rem;
}

.asset-info {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    align-items: center;
}

.asset-symbol {
    font-size: 1.125rem;
    font-weight: 600;
    color: #1e293b;
}

.asset-name {
    color: #64748b;
    font-size: 0.875rem;
}

.asset-exchange {
    font-size: 0.75rem;
    background: #dbeafe;
    color: #1d4ed8;
    padding: 0.125rem 0.5rem;
    border-radius: 0.25rem;
}

.change-btn {
    padding: 0.375rem 0.75rem;
    font-size: 0.75rem;
    background: white;
    border: 1px solid #d1d5db;
    border-radius: 0.25rem;
    cursor: pointer;
    color: #64748b;
}

.change-btn:hover {
    background: #f1f5f9;
}

/* Current Price */
.current-price {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    background: #f0fdf4;
    border: 1px solid #bbf7d0;
    border-radius: 0.5rem;
    margin-bottom: 1.5rem;
}

.price-label {
    font-size: 0.875rem;
    color: #374151;
}

.price-value {
    font-weight: 600;
    color: #1e293b;
}

.price-change {
    font-size: 0.75rem;
    font-weight: 500;
    padding: 0.125rem 0.375rem;
    border-radius: 0.25rem;
}

.price-change.positive {
    background: #dcfce7;
    color: #166534;
}

.price-change.negative {
    background: #fee2e2;
    color: #991b1b;
}

/* Total Preview */
.total-preview {
    display: flex;
    justify-content: space-between;
    padding: 0.75rem 1rem;
    background: #eff6ff;
    border-radius: 0.5rem;
    margin-top: 1rem;
    color: #1e40af;
}

/* Buttons */
.btn-primary {
    padding: 0.625rem 1.25rem;
    background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
    color: white;
    border: none;
    border-radius: 0.5rem;
    font-size: 0.875rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
}

.btn-primary:hover:not(:disabled) {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

.btn-secondary {
    padding: 0.625rem 1.25rem;
    background: white;
    border: 1px solid #d1d5db;
    border-radius: 0.5rem;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    color: #374151;
    transition: all 0.2s;
}

.btn-secondary:hover {
    background: #f9fafb;
}
</style>

<style>
.dark .modal-overlay {
    background: rgba(0, 0, 0, 0.7);
}

.dark .modal-content {
    background: #1a1a1f;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
}

.dark .modal-header {
    border-bottom-color: #2a2a30;
}

.dark .modal-header h2 {
    color: #fafafa;
}

.dark .close-btn {
    color: #71717a;
}

.dark .close-btn:hover {
    color: #fafafa;
}

.dark .modal-footer {
    border-top-color: #2a2a30;
}

/* Form */
.dark .form-group label {
    color: #a1a1aa;
}

.dark .form-group input,
.dark .form-group select,
.dark .form-group textarea {
    background: #0f0f12;
    border-color: #2a2a30;
    color: #fafafa;
}

.dark .form-group input:focus,
.dark .form-group select:focus,
.dark .form-group textarea:focus {
    border-color: #555560;
    box-shadow: 0 0 0 3px rgba(85, 85, 96, 0.3);
}

.dark .form-group input::placeholder,
.dark .form-group textarea::placeholder {
    color: #555560;
}

.dark .form-group select option {
    background: #1a1a1f;
    color: #fafafa;
}

/* Search */
.dark .search-input {
    background: #0f0f12;
    border-color: #2a2a30;
    color: #fafafa;
}

.dark .search-results {
    background: #1a1a1f;
    border-color: #2a2a30;
}

.dark .search-result-item {
    border-bottom-color: #2a2a30;
}

.dark .search-result-item:hover {
    background: #2a2a30;
}

.dark .result-symbol {
    color: #fafafa;
}

.dark .result-name {
    color: #a1a1aa;
}

.dark .no-results {
    color: #71717a;
}

/* Selected Asset */
.dark .selected-asset {
    background: #2a2a30;
    border-color: #42424a;
}

.dark .asset-symbol {
    color: #fafafa;
}

.dark .asset-name {
    color: #a1a1aa;
}

.dark .asset-exchange {
    background: #1e3a5f;
    color: #93c5fd;
}

.dark .change-btn {
    background: #1a1a1f;
    border-color: #42424a;
    color: #a1a1aa;
}

.dark .change-btn:hover {
    background: #42424a;
}

/* Current Price */
.dark .current-price {
    background: #14532d;
    border-color: #22c55e;
}

.dark .price-label {
    color: #dcfce7;
}

.dark .price-value {
    color: #fafafa;
}

.dark .price-change.positive {
    background: #166534;
    color: #4ade80;
}

.dark .price-change.negative {
    background: #7f1d1d;
    color: #fca5a5;
}

/* Total Preview */
.dark .total-preview {
    background: #1e3a5f;
    color: #93c5fd;
}

/* Buttons */
.dark .btn-secondary {
    background: #2a2a30;
    border-color: #42424a;
    color: #fafafa;
}

.dark .btn-secondary:hover {
    background: #42424a;
}
</style>
