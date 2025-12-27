<template>
    <Teleport to="body">
        <div v-if="visible" class="confirm-overlay" @click.self="handleCancel">
            <div class="confirm-modal" :class="[`confirm-${variant}`]">
                <!-- Icon -->
                <div class="confirm-icon-wrapper">
                    <div class="confirm-icon">
                        <!-- Warning Icon (default) -->
                        <svg v-if="variant === 'warning'" xmlns="http://www.w3.org/2000/svg" width="28" height="28"
                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                            stroke-linecap="round" stroke-linejoin="round">
                            <path
                                d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z">
                            </path>
                            <line x1="12" y1="9" x2="12" y2="13"></line>
                            <line x1="12" y1="17" x2="12.01" y2="17"></line>
                        </svg>
                        <!-- Danger Icon -->
                        <svg v-else-if="variant === 'danger'" xmlns="http://www.w3.org/2000/svg" width="28" height="28"
                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                            stroke-linecap="round" stroke-linejoin="round">
                            <circle cx="12" cy="12" r="10"></circle>
                            <line x1="15" y1="9" x2="9" y2="15"></line>
                            <line x1="9" y1="9" x2="15" y2="15"></line>
                        </svg>
                        <!-- Info Icon -->
                        <svg v-else xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24"
                            fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                            stroke-linejoin="round">
                            <circle cx="12" cy="12" r="10"></circle>
                            <line x1="12" y1="16" x2="12" y2="12"></line>
                            <line x1="12" y1="8" x2="12.01" y2="8"></line>
                        </svg>
                    </div>
                </div>

                <!-- Content -->
                <div class="confirm-content">
                    <h3 class="confirm-title">{{ title }}</h3>
                    <p v-if="message" class="confirm-message">{{ message }}</p>
                    <slot></slot>
                </div>

                <!-- Actions -->
                <div class="confirm-actions">
                    <button @click="handleCancel" class="btn-cancel" :disabled="loading">
                        {{ cancelText }}
                    </button>
                    <button @click="handleConfirm" class="btn-confirm" :class="[`btn-${variant}`]" :disabled="loading">
                        <span v-if="loading" class="btn-spinner"></span>
                        {{ loading ? loadingText : confirmText }}
                    </button>
                </div>
            </div>
        </div>
    </Teleport>
</template>

<script>
export default {
    name: 'ConfirmModal',
    props: {
        visible: {
            type: Boolean,
            default: false
        },
        variant: {
            type: String,
            default: 'warning',
            validator: (value) => ['warning', 'danger', 'info'].indexOf(value) !== -1
        },
        title: {
            type: String,
            default: 'Are you sure?'
        },
        message: {
            type: String,
            default: ''
        },
        confirmText: {
            type: String,
            default: 'Confirm'
        },
        cancelText: {
            type: String,
            default: 'Cancel'
        },
        loadingText: {
            type: String,
            default: 'Processing...'
        },
        loading: {
            type: Boolean,
            default: false
        }
    },
    emits: ['confirm', 'cancel'],
    methods: {
        handleConfirm() {
            this.$emit('confirm')
        },
        handleCancel() {
            if (!this.loading) {
                this.$emit('cancel')
            }
        }
    }
}
</script>

<style scoped>
.confirm-overlay {
    position: fixed;
    inset: 0;
    z-index: 9999;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1rem;
    background: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(4px);
}

.confirm-modal {
    background: white;
    border-radius: 1rem;
    width: 100%;
    max-width: 400px;
    padding: 1.75rem;
    text-align: center;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
}

/* Icon Wrapper */
.confirm-icon-wrapper {
    display: flex;
    justify-content: center;
    margin-bottom: 1.25rem;
}

.confirm-icon {
    width: 64px;
    height: 64px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
}

.confirm-warning .confirm-icon {
    background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
    color: #d97706;
}

.confirm-danger .confirm-icon {
    background: linear-gradient(135deg, #fee2e2 0%, #fecaca 100%);
    color: #dc2626;
}

.confirm-info .confirm-icon {
    background: linear-gradient(135deg, #dbeafe 0%, #bfdbfe 100%);
    color: #2563eb;
}

/* Content */
.confirm-content {
    margin-bottom: 1.5rem;
}

.confirm-title {
    font-size: 1.25rem;
    font-weight: 700;
    color: #1e293b;
    margin: 0 0 0.5rem 0;
}

.confirm-message {
    font-size: 0.9375rem;
    color: #64748b;
    margin: 0;
    line-height: 1.5;
}

/* Actions */
.confirm-actions {
    display: flex;
    gap: 0.75rem;
}

.btn-cancel,
.btn-confirm {
    flex: 1;
    padding: 0.75rem 1rem;
    border-radius: 0.5rem;
    font-size: 0.9375rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
}

.btn-cancel {
    background: #f1f5f9;
    border: 1px solid #e2e8f0;
    color: #475569;
}

.btn-cancel:hover:not(:disabled) {
    background: #e2e8f0;
}

.btn-confirm {
    border: none;
    color: white;
}

.btn-warning {
    background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
}

.btn-warning:hover:not(:disabled) {
    box-shadow: 0 4px 12px rgba(245, 158, 11, 0.4);
}

.btn-danger {
    background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
}

.btn-danger:hover:not(:disabled) {
    box-shadow: 0 4px 12px rgba(239, 68, 68, 0.4);
}

.btn-info {
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
}

.btn-info:hover:not(:disabled) {
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.btn-cancel:disabled,
.btn-confirm:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

/* Spinner */
.btn-spinner {
    width: 16px;
    height: 16px;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-top-color: white;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
}

@keyframes spin {
    to {
        transform: rotate(360deg);
    }
}
</style>
