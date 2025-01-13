// src/axiosInstance.js
import axios from "axios";
import store from "./store";

// 创建 Axios 实例
const axiosInstance = axios.create({
    baseURL: 'http://47.83.211.8:8083', // 登录服务的基础 URL
    timeout: 10000, // 请求超时时间（毫秒）
    withCredentials: true, // 允许携带凭证（如 cookies）
});

// 定义不需要添加 Headers 的端点
const excludedEndpoints = ['/login', '/register'];

// 添加请求拦截器
axiosInstance.interceptors.request.use(
    config => {
        // 检查当前请求是否在排除列表中
        if (excludedEndpoints.some(endpoint => config.url.startsWith(endpoint))) {
            return config; // 不添加 Headers，直接返回配置
        }

        // 从 store 或 localStorage 获取 userId 和 authToken
        const userId = store.state.userId || localStorage.getItem("id");
        const authToken = localStorage.getItem("authToken");

        console.log('Adding headers:', { userId, authToken });

        if (userId) {
            config.headers['X-User-ID'] = userId; // 添加自定义 Header
            console.log('X-User-ID header added');
        }

        if (authToken) {
            config.headers['Authorization'] = `Bearer ${authToken}`; // 添加 Authorization Header
            console.log('Authorization header added');
        }

        return config;
    },
    error => {
        console.error('Request error:', error);
        return Promise.reject(error);
    }
);

export default axiosInstance;