const { defineConfig } = require('@vue/cli-service')
module.exports = defineConfig({
  transpileDependencies: true,
  devServer: {
    host: '0.0.0.0',
    port: 8080,
    hot: true,
    liveReload: true,
    proxy: {
      '/api/auth': {
        target: process.env.AUTH_PROXY_TARGET || 'http://localhost:3001',
        onProxyReq(proxyReq, req) {
          // Never let clients choose the rate-limit identity.
          proxyReq.setHeader('X-Real-IP', req.socket.remoteAddress)
          proxyReq.removeHeader('X-Forwarded-For')
          proxyReq.removeHeader('Forwarded')
        }
      },
      '/api/v1': { target: process.env.API_PROXY_TARGET || 'http://localhost:3000' }
    },
    watchFiles: {
      paths: ['src/**/*', 'public/**/*'],
      options: {
        usePolling: true, // Docker için gerekli
        interval: 1000
      }
    },
    client: {
      webSocketURL: 'auto://0.0.0.0:0/ws',
      logging: 'info',
      overlay: {
        errors: true,
        warnings: false
      }
    }
  }
})
