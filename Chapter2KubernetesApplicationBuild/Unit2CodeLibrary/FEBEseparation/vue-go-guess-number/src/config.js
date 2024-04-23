// config.js
const isProduction = process.env.NODE_ENV === "production";

const config = {
    apiUrl: isProduction
        ? "http://47.76.196.15/check-guess"
        : "http://47.76.196.15:8081/check-guess",
};

console.log(config.apiUrl); // 调试

export default config;
