<template>
    <Teleport to="body">
        <Transition name="fade">
            <div v-if="visible" class="spinner-overlay">
                <div class="spinner-container">
                    <div class="spinner">
                        <div class="spinner-ring"></div>
                        <div class="spinner-ring"></div>
                        <div class="spinner-ring"></div>
                    </div>
                    <p v-if="message" class="spinner-message">{{ message }}</p>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<script>
export default {
    name: 'FullPageSpinner',
    props: {
        visible: {
            type: Boolean,
            default: false
        },
        message: {
            type: String,
            default: ''
        }
    }
}
</script>

<style scoped>
.spinner-overlay {
    position: fixed;
    inset: 0;
    z-index: 9999;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(255, 255, 255, 0.7);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
}

.spinner-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1.5rem;
}

.spinner {
    position: relative;
    width: 64px;
    height: 64px;
}

.spinner-ring {
    position: absolute;
    inset: 0;
    border-radius: 50%;
    border: 3px solid transparent;
    animation: spin 1.2s cubic-bezier(0.5, 0, 0.5, 1) infinite;
}

.spinner-ring:nth-child(1) {
    border-top-color: #3b82f6;
    animation-delay: -0.45s;
}

.spinner-ring:nth-child(2) {
    border-right-color: #6366f1;
    animation-delay: -0.3s;
}

.spinner-ring:nth-child(3) {
    border-bottom-color: #8b5cf6;
    animation-delay: -0.15s;
}

@keyframes spin {
    0% {
        transform: rotate(0deg);
    }

    100% {
        transform: rotate(360deg);
    }
}

.spinner-message {
    font-size: 0.9375rem;
    font-weight: 500;
    color: #475569;
    margin: 0;
    text-align: center;
    animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {

    0%,
    100% {
        opacity: 1;
    }

    50% {
        opacity: 0.6;
    }
}

/* Fade Transition */
.fade-enter-active,
.fade-leave-active {
    transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
    opacity: 0;
}
</style>
