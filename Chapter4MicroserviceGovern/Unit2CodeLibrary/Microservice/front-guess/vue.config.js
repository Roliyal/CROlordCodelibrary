// front-guess/vue.config.js

module.exports = {
  devServer: {
    proxy: {
      '/login': {
        target: 'http://47.238.211.214:80',
        changeOrigin: true,
      },
      '/register': {
        target: 'http://47.238.211.214:8083',
        changeOrigin: true,
      },
    },
  },
}
