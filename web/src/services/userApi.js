import apiFetch from './api'

export const userAPI = {
    getMe: () =>
        apiFetch('/user/me', {
            method: 'GET'
        }),

    updateProfile: (data) =>
        apiFetch('/user/profile', {
            method: 'PUT',
            body: JSON.stringify(data)
        }),

    updateBaseCurrency: (baseCurrency) =>
        apiFetch('/user/base-currency', {
            method: 'PUT',
            body: JSON.stringify({ base_currency: baseCurrency })
        }),

    updateTheme: (theme) =>
        apiFetch('/user/theme', {
            method: 'PUT',
            body: JSON.stringify({ theme })
        })
}

export default userAPI
