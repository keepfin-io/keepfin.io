import { createServer } from 'node:http'
import { toNodeHandler } from 'better-auth/node'
import { auth, config, pool } from './runtime.js'

const handler = toNodeHandler(auth)
const server = createServer(async (req, res) => {
  res.setHeader('Cache-Control', 'no-store')
  res.setHeader('Referrer-Policy', 'no-referrer')
  res.setHeader('X-Content-Type-Options', 'nosniff')
  try {
    const path = new URL(req.url, config.baseURL).pathname
    if (req.method === 'GET' && path === '/api/auth/config') {
      res.setHeader('Content-Type', 'application/json')
      res.end(JSON.stringify({ google: config.google, email: config.mailMode !== 'disabled', developmentMail: config.mailMode === 'file' }))
      return
    }
    if (path.startsWith('/api/auth/')) {
      await handler(req, res)
      return
    }
    res.writeHead(404).end()
  } catch {
    // Do not log request URLs: OAuth codes and Magic Link tokens are secrets.
    console.error('Authentication request failed')
    if (!res.headersSent) res.writeHead(500, { 'Content-Type': 'application/json' })
    res.end(JSON.stringify({ message: 'Authentication service unavailable' }))
  }
})
server.requestTimeout = 15000
server.headersTimeout = 10000
server.listen(Number(process.env.PORT || 3001), '0.0.0.0', () => console.log('Keepfin auth listening'))
for (const signal of ['SIGINT', 'SIGTERM']) {
  process.on(signal, () => server.close(async () => { await pool.end(); process.exit(0) }))
}
