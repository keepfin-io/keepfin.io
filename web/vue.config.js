const { defineConfig } = require('@vue/cli-service')
module.exports = defineConfig({
  transpileDependencies: true,
  devServer: {
    host: '0.0.0.0',
    port: 8080,
    hot: true,
    liveReload: true,
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
