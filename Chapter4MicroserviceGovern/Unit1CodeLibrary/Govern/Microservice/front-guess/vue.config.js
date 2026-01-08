// front-guess/vue.config.js

module.exports = {
  devServer: {
    proxy: {
      '/login': {
        target: 'http://localhost:80',
        changeOrigin: true,
      },
      '/register': {
        target: 'http://localhost:8083',
        changeOrigin: true,
      },
    },
  },
}
