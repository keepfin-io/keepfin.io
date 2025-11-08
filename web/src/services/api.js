const API_URL = process.env.VUE_APP_API_URL || 'http://localhost:3000/api/v1'

// Fetch wrapper with default options
async function apiFetch(endpoint, options = {}) {
  const config = {
    credentials: 'include', // Important for cookies!
    headers: {
      'Content-Type': 'application/json',
      ...options.headers
    },
    ...options
  }

  const response = await fetch(`${API_URL}${endpoint}`, config)

  // Handle 401 Unauthorized
  if (response.status === 401) {
    if (window.location.pathname !== '/login') {
      window.location.href = '/login'
    }
    throw new Error('Unauthorized')
  }

  const data = await response.json()

  if (!response.ok) {
    throw new Error(data.message || 'Request failed')
  }

  return data
}

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

export default apiFetch
