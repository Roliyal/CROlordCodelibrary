const { createProxyMiddleware } = require("http-proxy-middleware");
module.exports = {
  productionSourceMap: false,
};
module.exports = {
  devServer: {
    proxy: {
      "/api": {
        target: "http://47.76.196.15:8081",
        changeOrigin: true,
        pathRewrite: {
          "^/api": "",
        },
      },
    },
  },
};

