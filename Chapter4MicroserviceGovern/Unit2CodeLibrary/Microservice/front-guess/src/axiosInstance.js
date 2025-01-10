// src/axiosInstance.js
import axios from "axios";
import store from "./store";

const axiosInstance = axios.create({
    baseURL: 'http://47.83.211.8', // 后端API的基础URL
    timeout: 10000, // 请求超时时间（毫秒）
});

// 添加请求拦截器
axiosInstance.interceptors.request.use(
    config => {
        // 定义不添加Headers的端点
        const excludedEndpoints = ['/login', '/register'];

        // 检查当前请求是否在排除列表中
        if (excludedEndpoints.some(endpoint => config.url.startsWith(endpoint))) {
            return config; // 不添加Headers，直接返回配置
        }

        // 从store或localStorage中获取userId和authToken
        const userId = store.state.userId || localStorage.getItem("id");
        const authToken = localStorage.getItem("authToken");

        console.log('Adding headers:', { userId, authToken });

        if (userId) {
            config.headers['X-User-ID'] = userId; // 添加自定义 Header
            console.log('X-User-ID header added');
        }

        if (authToken) {
            config.headers['Authorization'] = `Bearer ${authToken}`; // 添加Authorization Header
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
