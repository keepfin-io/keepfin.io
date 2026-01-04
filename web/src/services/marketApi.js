import apiFetch from './api'

export const marketAPI = {
    searchAssets: (query, limit = 10) =>
        apiFetch(`/market/search?q=${encodeURIComponent(query)}&limit=${limit}`, {
            method: 'GET'
        }),


    getLivePrice: (symbol) =>
        apiFetch(`/market/live/${encodeURIComponent(symbol)}`, {
            method: 'GET'
        }),


    getMultipleLivePrices: (symbols) => {
        const symbolsParam = Array.isArray(symbols) ? symbols.join(',') : symbols
        return apiFetch(`/market/live?symbols=${encodeURIComponent(symbolsParam)}`, {
            method: 'GET'
        })
    },

    getEODData: (symbol, options = {}) => {
        const params = new URLSearchParams()
        if (options.from) params.append('from', options.from)
        if (options.to) params.append('to', options.to)
        if (options.period) params.append('period', options.period)
        if (options.order) params.append('order', options.order)

        const queryString = params.toString()
        const url = `/market/eod/${encodeURIComponent(symbol)}${queryString ? `?${queryString}` : ''}`

        return apiFetch(url, {
            method: 'GET'
        })
    }
}

export default marketAPI
