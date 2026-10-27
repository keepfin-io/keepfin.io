const API_URL = '/api/v1'

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

  if (!response.ok) {
    const error = new Error(data.message || 'Request failed')
    error.status = response.status
    if (response.status === 401 && !['/login', '/register', '/auth/callback'].includes(window.location.pathname)) {
      window.location.assign('/login')
    }
    throw error
  }

  return data
}

export default apiFetch
