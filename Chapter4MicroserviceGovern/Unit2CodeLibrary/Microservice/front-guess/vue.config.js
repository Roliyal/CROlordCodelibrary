// front-guess/vue.config.js

module.exports = {
  devServer: {
    proxy: {
      '/login': {
        target: 'http://47.243.164.40:80',
        changeOrigin: true,
      },
      '/register': {
        target: 'http://47.243.164.40:8083',
        changeOrigin: true,
      },
    },
  },
}
