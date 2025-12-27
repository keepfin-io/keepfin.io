<template>
    <Teleport to="body">
        <div class="toast-container">
            <div v-for="toast in toasts" :key="toast.id" class="toast" :class="[`toast-${toast.variant}`]">
                <div class="toast-icon">
                    <!-- Success Icon -->
                    <svg v-if="toast.variant === 'success'" xmlns="http://www.w3.org/2000/svg" width="20" height="20"
                        viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                        stroke-linejoin="round">
                        <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
                        <polyline points="22 4 12 14.01 9 11.01"></polyline>
                    </svg>
                    <!-- Error Icon -->
                    <svg v-else-if="toast.variant === 'error'" xmlns="http://www.w3.org/2000/svg" width="20" height="20"
                        viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                        stroke-linejoin="round">
                        <circle cx="12" cy="12" r="10"></circle>
                        <line x1="15" y1="9" x2="9" y2="15"></line>
                        <line x1="9" y1="9" x2="15" y2="15"></line>
                    </svg>
                    <!-- Warning Icon -->
                    <svg v-else-if="toast.variant === 'warning'" xmlns="http://www.w3.org/2000/svg" width="20"
                        height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                        stroke-linecap="round" stroke-linejoin="round">
                        <path
                            d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z">
                        </path>
                        <line x1="12" y1="9" x2="12" y2="13"></line>
                        <line x1="12" y1="17" x2="12.01" y2="17"></line>
                    </svg>
                    <!-- Info Icon -->
                    <svg v-else xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24"
                        fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                        stroke-linejoin="round">
                        <circle cx="12" cy="12" r="10"></circle>
                        <line x1="12" y1="16" x2="12" y2="12"></line>
                        <line x1="12" y1="8" x2="12.01" y2="8"></line>
                    </svg>
                </div>
                <div class="toast-content">
                    <p class="toast-title">{{ toast.title }}</p>
                    <p v-if="toast.description" class="toast-description">{{ toast.description }}</p>
                </div>
                <button @click="removeToast(toast.id)" class="toast-close">
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <line x1="18" y1="6" x2="6" y2="18"></line>
                        <line x1="6" y1="6" x2="18" y2="18"></line>
                    </svg>
                </button>
            </div>
        </div>
    </Teleport>
</template>

<script>
import { ref, provide, readonly } from 'vue'

const toasts = ref([])
let toastId = 0

function addToast({ title, description = '', variant = 'info', duration = 5000 }) {
    const id = ++toastId
    toasts.value.push({ id, title, description, variant })

    if (duration > 0) {
        setTimeout(() => removeToast(id), duration)
    }

    return id
}

function removeToast(id) {
    const index = toasts.value.findIndex(t => t.id === id)
    if (index > -1) {
        toasts.value.splice(index, 1)
    }
}

// Toast helper functions
export const toast = {
    success: (title, description) => addToast({ title, description, variant: 'success' }),
    error: (title, description) => addToast({ title, description, variant: 'error' }),
    warning: (title, description) => addToast({ title, description, variant: 'warning' }),
    info: (title, description) => addToast({ title, description, variant: 'info' }),
    custom: (options) => addToast(options)
}

export default {
    name: 'ToastContainer',
    setup() {
        provide('toast', readonly(toast))
        return {
            toasts,
            removeToast
        }
    }
}
</script>

<style scoped>
.toast-container {
    position: fixed;
    top: 1rem;
    right: 1rem;
    z-index: 9998;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    max-width: 400px;
    width: 100%;
    pointer-events: none;
}

.toast {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
    padding: 1rem;
    background: white;
    border-radius: 0.75rem;
    box-shadow: 0 10px 40px -10px rgba(0, 0, 0, 0.2), 0 0 0 1px rgba(0, 0, 0, 0.05);
    pointer-events: auto;
    position: relative;
    overflow: hidden;
}

.toast::before {
    content: '';
    position: absolute;
    left: 0;
    top: 0;
    bottom: 0;
    width: 4px;
}

/* Variants */
.toast-success::before {
    background: linear-gradient(180deg, #22c55e 0%, #16a34a 100%);
}

.toast-success .toast-icon {
    color: #22c55e;
}

.toast-error::before {
    background: linear-gradient(180deg, #ef4444 0%, #dc2626 100%);
}

.toast-error .toast-icon {
    color: #ef4444;
}

.toast-warning::before {
    background: linear-gradient(180deg, #f59e0b 0%, #d97706 100%);
}

.toast-warning .toast-icon {
    color: #f59e0b;
}

.toast-info::before {
    background: linear-gradient(180deg, #3b82f6 0%, #2563eb 100%);
}

.toast-info .toast-icon {
    color: #3b82f6;
}

/* Icon */
.toast-icon {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-top: 0.125rem;
}

/* Content */
.toast-content {
    flex: 1;
    min-width: 0;
}

.toast-title {
    font-size: 0.9375rem;
    font-weight: 600;
    color: #1e293b;
    margin: 0;
    line-height: 1.4;
}

.toast-description {
    font-size: 0.8125rem;
    color: #64748b;
    margin: 0.25rem 0 0 0;
    line-height: 1.4;
}

/* Close Button */
.toast-close {
    flex-shrink: 0;
    padding: 0.25rem;
    background: transparent;
    border: none;
    border-radius: 0.375rem;
    cursor: pointer;
    color: #94a3b8;
    transition: all 0.2s;
}

.toast-close:hover {
    color: #475569;
    background: #f1f5f9;
}

/* Responsive */
@media (max-width: 480px) {
    .toast-container {
        left: 1rem;
        right: 1rem;
        max-width: none;
    }
}
</style>
