// vue.config.js
const { createProxyMiddleware } = require("http-proxy-middleware");

module.exports = {
  devServer: {
    proxy: {
      '/api': {
        target: 'http://app-go-backend-service.cicd.svc.cluster.local:8081',  // 指向后端服务的内部 IP 或域名
        changeOrigin: true,
        pathRewrite: { '^/api': '' },  // 重写路径，去掉 /api
      },
    },
  },
  productionSourceMap: false,
};
