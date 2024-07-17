/*
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
*/
const HtmlWebpackPlugin = require('html-webpack-plugin');

module.exports = {
  devServer: {
    proxy: {
      '/api': {
        target: 'http://app-go-backend-service.cicd.svc.cluster.local:8081',
        changeOrigin: true,
        pathRewrite: { '^/api': '' },
      },
    },
  },
  productionSourceMap: false,
  configureWebpack: config => {
    config.plugins.forEach(plugin => {
      if (plugin instanceof HtmlWebpackPlugin) {
        plugin.userOptions.templateParameters = {
          ...plugin.userOptions.templateParameters,
          VUE_APP_ENV: process.env.VUE_APP_ENV
        };
      }
    });
  }
};
