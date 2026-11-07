import apiFetch from './api'

export const wealthAPI = {
  get(path, filters = {}) {
    const query = new URLSearchParams(Object.entries(filters).filter(([, v]) => v !== '' && v != null))
    return apiFetch(`${path}${query.size ? `?${query}` : ''}`)
  },
  save(path, body, method = 'POST') { return apiFetch(path, { method, body: JSON.stringify(body) }) },
  remove(path) { return apiFetch(path, { method: 'DELETE' }) },
}
