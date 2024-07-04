// front-guess/vue.config.js

module.exports = {
  devServer: {
    proxy: {
      '/login': {
        target: 'http://8.217.105.200:80',
        changeOrigin: true,
      },
      '/register': {
        target: 'http://8.217.105.200:8083',
        changeOrigin: true,
      },
    },
  },
}
