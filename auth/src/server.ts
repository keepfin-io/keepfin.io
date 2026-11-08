import { createServer } from 'node:http'
import { toNodeHandler } from 'better-auth/node'
import { auth, config, pool } from './runtime.ts'
import { createClientIPResolver } from './proxy.ts'
import { authHTTPHandler } from './http-handler.ts'

const handler = toNodeHandler(authHTTPHandler(auth))
const clientIP = createClientIPResolver(config.trustedProxyCIDRs)
const server = createServer(async (req, res) => {
  req.headers['x-real-ip'] = clientIP(req.socket.remoteAddress, req.headers['x-real-ip'])
  delete req.headers['x-forwarded-for']
  delete req.headers.forwarded
  res.setHeader('Cache-Control', 'no-store')
  res.setHeader('Referrer-Policy', 'no-referrer')
  res.setHeader('X-Content-Type-Options', 'nosniff')
  try {
    const path = new URL(req.url ?? '/', config.baseURL).pathname
    if (req.method === 'GET' && path === '/api/auth/config') {
      res.setHeader('Content-Type', 'application/json')
      res.end(JSON.stringify({ google: config.google, apple: false, internalTesting: config.internalTesting, email: config.mailMode !== 'disabled', developmentMail: config.mailMode === 'file' }))
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
