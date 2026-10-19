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

  // -----> Server can return non-JSON text, so parse safely
  const text = await response.text()
  let data = {}
  try {
    data = text ? JSON.parse(text) : {}
  } catch {
    data = { message: text }
  }

  // Handle 401 Unauthorized
  if (response.status === 401) {
    const path = window.location.pathname
    if (path !== '/login' && path !== '/register') {
      window.location.href = '/login'
    }
    throw new Error(data.message || 'Unauthorized')
  }

  if (!response.ok) {
    throw new Error(data.message || 'Request failed')
  }

  return data
}

export default apiFetch
