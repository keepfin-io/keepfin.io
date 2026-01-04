import apiFetch from './api'

// Holding API - CRUD operations for account holdings
export const holdingAPI = {
    // Get all holdings for an account with live data
    getHoldings: (accountId) =>
        apiFetch(`/accounts/${accountId}/holdings`, {
            method: 'GET'
        }),

    // Add a new holding to an account
    addHolding: (accountId, holdingData) =>
        apiFetch(`/accounts/${accountId}/holdings`, {
            method: 'POST',
            body: JSON.stringify(holdingData)
        }),

    // Update an existing holding
    updateHolding: (accountId, holdingId, holdingData) =>
        apiFetch(`/accounts/${accountId}/holdings/${holdingId}`, {
            method: 'PUT',
            body: JSON.stringify(holdingData)
        }),

    // Delete a holding
    deleteHolding: (accountId, holdingId) =>
        apiFetch(`/accounts/${accountId}/holdings/${holdingId}`, {
            method: 'DELETE'
        })
}

export default holdingAPI
