import apiFetch from './api'

export const portfolioAPI = {
  report(options = {}) {
    const query = new URLSearchParams(Object.entries(options).filter(([, value]) => value))
    return apiFetch(`/portfolio?${query}`)
  },
  transfer(data) { return apiFetch('/transfers', { method: 'POST', body: JSON.stringify(data) }) },
  importCSV(accountId, csv, commit = false) {
    return apiFetch(`/accounts/${accountId}/activities/import`, { method: 'POST', body: JSON.stringify({ csv, commit }) })
  },
  importJSON(accountId, json, commit = false) {return apiFetch(`/accounts/${accountId}/activities/import`, {method:'POST',body:JSON.stringify({json,commit})})},
  async exportCSV(accountId, format = 'csv') {
    const response = await fetch(`/api/v1/accounts/${accountId}/activities/export?format=${format}`, { credentials: 'same-origin' })
    if (!response.ok) throw new Error('Unable to export activities')
    const url = URL.createObjectURL(await response.blob())
    const link = document.createElement('a')
    link.href = url; link.download = `keepfin-activities.${format}`; link.click()
    setTimeout(() => URL.revokeObjectURL(url), 1000)
  },
}
