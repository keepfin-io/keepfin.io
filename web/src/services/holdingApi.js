import apiFetch from './api'

export const holdingAPI = {
    getHoldings: (accountId) =>
        apiFetch(`/accounts/${accountId}/holdings`, {
            method: 'GET'
        }),

    addHolding: (accountId, holdingData) =>
        apiFetch(`/accounts/${accountId}/holdings`, {
            method: 'POST',
            body: JSON.stringify(holdingData)
        }),

    updateHolding: (accountId, holdingId, holdingData) =>
        apiFetch(`/accounts/${accountId}/holdings/${holdingId}`, {
            method: 'PUT',
            body: JSON.stringify(holdingData)
        }),

    deleteHolding: (accountId, holdingId) =>
        apiFetch(`/accounts/${accountId}/holdings/${holdingId}`, {
            method: 'DELETE'
        })
}

export default holdingAPI
