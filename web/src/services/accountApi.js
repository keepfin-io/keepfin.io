import apiFetch from './api'

// Account API
export const accountAPI = {
    getAccounts: () =>
        apiFetch('/accounts', {
            method: 'GET'
        }),

    getAccount: (id) =>
        apiFetch(`/accounts/${id}`, {
            method: 'GET'
        }),

    createAccount: (account) =>
        apiFetch('/accounts/create', {
            method: 'POST',
            body: JSON.stringify(account)
        }),

    updateAccount: (id, account) =>
        apiFetch(`/accounts/${id}`, {
            method: 'PUT',
            body: JSON.stringify(account)
        }),

    deleteAccount: (id) =>
        apiFetch(`/accounts/${id}`, {
            method: 'DELETE'
        })
}

export default accountAPI
