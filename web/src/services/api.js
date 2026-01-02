const API_URL = process.env.VUE_APP_API_URL || 'http://localhost:3000/api/v1'

console.log('[API] Base URL:', API_URL)

// Fetch wrapper with default options
async function apiFetch(endpoint, options = {}) {
  const fullUrl = `${API_URL}${endpoint}`
  console.log('[API] Request:', fullUrl)

  const config = {
    credentials: 'include', // Important for cookies!
    headers: {
      'Content-Type': 'application/json',
      ...options.headers
    },
    ...options
  }

  const response = await fetch(fullUrl, config)

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

export default apiFetch
