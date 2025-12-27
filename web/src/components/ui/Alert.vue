<template>
    <div class="alert" :class="[`alert-${variant}`]" role="alert">
        <div class="alert-icon">
            <!-- Success Icon -->
            <svg v-if="variant === 'success'" xmlns="http://www.w3.org/2000/svg" width="20" height="20"
                viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                stroke-linejoin="round">
                <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
                <polyline points="22 4 12 14.01 9 11.01"></polyline>
            </svg>
            <!-- Error Icon -->
            <svg v-else-if="variant === 'error'" xmlns="http://www.w3.org/2000/svg" width="20" height="20"
                viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                stroke-linejoin="round">
                <circle cx="12" cy="12" r="10"></circle>
                <line x1="15" y1="9" x2="9" y2="15"></line>
                <line x1="9" y1="9" x2="15" y2="15"></line>
            </svg>
            <!-- Warning Icon -->
            <svg v-else-if="variant === 'warning'" xmlns="http://www.w3.org/2000/svg" width="20" height="20"
                viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                stroke-linejoin="round">
                <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z">
                </path>
                <line x1="12" y1="9" x2="12" y2="13"></line>
                <line x1="12" y1="17" x2="12.01" y2="17"></line>
            </svg>
            <!-- Info Icon -->
            <svg v-else xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
                stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="10"></circle>
                <line x1="12" y1="16" x2="12" y2="12"></line>
                <line x1="12" y1="8" x2="12.01" y2="8"></line>
            </svg>
        </div>
        <div class="alert-content">
            <h5 v-if="title" class="alert-title">{{ title }}</h5>
            <p v-if="description" class="alert-description">{{ description }}</p>
            <slot></slot>
        </div>
        <button v-if="dismissible" @click="$emit('dismiss')" class="alert-dismiss" aria-label="Dismiss">
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <line x1="18" y1="6" x2="6" y2="18"></line>
                <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
        </button>
    </div>
</template>

<script>
export default {
    name: 'Alert',
    props: {
        variant: {
            type: String,
            default: 'info',
            validator: (value) => ['success', 'error', 'warning', 'info'].indexOf(value) !== -1
        },
        title: {
            type: String,
            default: ''
        },
        description: {
            type: String,
            default: ''
        },
        dismissible: {
            type: Boolean,
            default: false
        }
    },
    emits: ['dismiss']
}
</script>

<style scoped>
.alert {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
    padding: 1rem;
    border-radius: 0.75rem;
    border: 1px solid;
    position: relative;
}

/* Variants */
.alert-success {
    background: linear-gradient(135deg, #f0fdf4 0%, #dcfce7 100%);
    border-color: #86efac;
    color: #166534;
}

.alert-success .alert-icon {
    color: #22c55e;
}

.alert-error {
    background: linear-gradient(135deg, #fef2f2 0%, #fee2e2 100%);
    border-color: #fca5a5;
    color: #991b1b;
}

.alert-error .alert-icon {
    color: #ef4444;
}

.alert-warning {
    background: linear-gradient(135deg, #fffbeb 0%, #fef3c7 100%);
    border-color: #fcd34d;
    color: #92400e;
}

.alert-warning .alert-icon {
    color: #f59e0b;
}

.alert-info {
    background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
    border-color: #93c5fd;
    color: #1e40af;
}

.alert-info .alert-icon {
    color: #3b82f6;
}

/* Icon */
.alert-icon {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-top: 0.125rem;
}

/* Content */
.alert-content {
    flex: 1;
    min-width: 0;
}

.alert-title {
    font-size: 0.9375rem;
    font-weight: 600;
    margin: 0 0 0.25rem 0;
    line-height: 1.4;
}

.alert-description {
    font-size: 0.875rem;
    margin: 0;
    opacity: 0.9;
    line-height: 1.5;
}

.alert-content :slotted(ul) {
    margin: 0.5rem 0 0 0;
    padding-left: 1.25rem;
    font-size: 0.8125rem;
    opacity: 0.9;
}

.alert-content :slotted(li) {
    margin-top: 0.25rem;
}

/* Dismiss Button */
.alert-dismiss {
    position: absolute;
    top: 0.75rem;
    right: 0.75rem;
    padding: 0.25rem;
    background: transparent;
    border: none;
    border-radius: 0.375rem;
    cursor: pointer;
    opacity: 0.5;
    transition: all 0.2s;
    color: inherit;
}

.alert-dismiss:hover {
    opacity: 1;
    background: rgba(0, 0, 0, 0.05);
}
</style>
