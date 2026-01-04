import apiFetch from './api'

// Market API - Asset search and live price data
export const marketAPI = {
    // Search for assets by query string
    // Returns: { error: false, query: string, count: number, results: SearchResult[] }
    searchAssets: (query, limit = 10) =>
        apiFetch(`/market/search?q=${encodeURIComponent(query)}&limit=${limit}`, {
            method: 'GET'
        }),

    // Get live price for a single symbol
    // Symbol format: "AAPL" or "AAPL.US"
    getLivePrice: (symbol) =>
        apiFetch(`/market/live/${encodeURIComponent(symbol)}`, {
            method: 'GET'
        }),

    // Get live prices for multiple symbols
    // Symbols should be comma-separated
    getMultipleLivePrices: (symbols) => {
        const symbolsParam = Array.isArray(symbols) ? symbols.join(',') : symbols
        return apiFetch(`/market/live?symbols=${encodeURIComponent(symbolsParam)}`, {
            method: 'GET'
        })
    },

    // Get End-of-Day historical data for a symbol
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
