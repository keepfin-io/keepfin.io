import apiFetch from './api'

export const activityAPI = {

    getActivities(accountId) {
        return apiFetch(`/accounts/${accountId}/activities`, {
            method: 'GET'
        })
    },

    createActivity(accountId, data) {
        return apiFetch(`/accounts/${accountId}/activities`, {
            method: 'POST',
            body: JSON.stringify(data)
        })
    },

    deleteActivity(accountId, activityId) {
        return apiFetch(`/accounts/${accountId}/activities/${activityId}`, {
            method: 'DELETE'
        })
    }
}

export default activityAPI
