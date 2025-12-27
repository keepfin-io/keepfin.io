import apiFetch from './api'

// Auth API
export const authAPI = {
    signup: (email, password) =>
        apiFetch('/user/signup', {
            method: 'POST',
            body: JSON.stringify({ email, password })
        }),

    signin: (email, password) =>
        apiFetch('/user/signin', {
            method: 'POST',
            body: JSON.stringify({ email, password })
        }),

    signout: () =>
        apiFetch('/user/signout', {
            method: 'POST'
        }),

    checkAuth: () =>
        apiFetch('/user/me', {
            method: 'GET'
        })
}

export default authAPI
