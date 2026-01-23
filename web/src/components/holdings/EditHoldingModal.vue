<template>
    <div v-if="show" class="modal-overlay">
        <div class="modal-content">
            <div class="modal-header">
                <h2>Edit Holding</h2>
                <button class="close-btn" @click="close">&times;</button>
            </div>

            <div class="modal-body">
                <!-- Selected Asset Display -->
                <div class="selected-asset">
                    <div class="asset-info">
                        <span class="asset-symbol">{{ holding?.symbol }}</span>
                        <span class="asset-name">{{ holding?.name }}</span>
                    </div>
                </div>

                <!-- Form -->
                <form @submit.prevent="submitUpdate">
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

                    <!-- Total Value Preview -->
                    <div v-if="form.quantity && form.buy_price" class="total-preview">
                        <span>Total Cost:</span>
                        <strong>{{ formatCurrency(form.quantity * form.buy_price, form.buy_currency) }}</strong>
                    </div>
                </form>
            </div>

            <div class="modal-footer">
                <button type="button" class="btn-secondary" @click="close">Cancel</button>
                <button type="button" class="btn-primary" @click="submitUpdate" :disabled="!canSubmit || submitting">
                    {{ submitting ? 'Updating...' : 'Update Holding' }}
                </button>
            </div>
        </div>
    </div>
</template>

<script>
import { holdingAPI } from '@/services/holdingApi'

export default {
    name: 'EditHoldingModal',
    props: {
        show: {
            type: Boolean,
            default: false
        },
        accountId: {
            type: String,
            required: true
        },
        holding: {
            type: Object,
            default: null
        }
    },
    emits: ['close', 'success'],
    data() {
        return {
            submitting: false,
            form: {
                quantity: null,
                buy_price: null,
                buy_currency: 'USD',
                notes: ''
            }
        }
    },
    computed: {
        canSubmit() {
            return this.form.quantity > 0 && this.form.buy_price >= 0
        }
    },
    watch: {
        show(newVal) {
            if (newVal && this.holding) {
                this.populateForm()
            }
        },
        holding(newVal) {
            if (newVal && this.show) {
                this.populateForm()
            }
        }
    },
    methods: {
        populateForm() {
            this.form = {
                quantity: this.holding.quantity,
                buy_price: this.holding.buy_price,
                buy_currency: this.holding.buy_currency || 'USD',
                notes: this.holding.notes || ''
            }
        },

        async submitUpdate() {
            if (!this.canSubmit || this.submitting) return

            this.submitting = true
            try {
                await holdingAPI.updateHolding(this.accountId, this.holding.id, {
                    quantity: this.form.quantity,
                    buy_price: this.form.buy_price,
                    buy_currency: this.form.buy_currency,
                    notes: this.form.notes
                })
                this.$emit('success')
                this.close()
            } catch (err) {
                console.error('Failed to update holding:', err)
                alert(err.message || 'Failed to update holding')
            } finally {
                this.submitting = false
            }
        },

        close() {
            this.$emit('close')
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

/* Selected Asset */
.selected-asset {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    background: #f8fafc;
    border: 1px solid #e2e8f0;
    border-radius: 0.5rem;
    margin-bottom: 1.5rem;
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
