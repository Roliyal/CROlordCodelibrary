// front-guess/vue.config.js

module.exports = {
  devServer: {
    proxy: {
      '/login': {
        target: 'http://47.83.211.8:80',
        changeOrigin: true,
      },
      '/register': {
        target: 'http://47.83.211.8:8083',
        changeOrigin: true,
      },
    },
  },
}
