// src/axiosInstance.js

import axios from "axios";
import store from "./store";

// 创建 Axios 实例
const axiosInstance = axios.create({
    baseURL: 'http://micro.roliyal.com', // 后端服务的基础服务 URL
    timeout: 10000, // 请求超时时间（毫秒）
    withCredentials: true, // 允许携带凭证（如 cookies）
});

// 添加请求拦截器
axiosInstance.interceptors.request.use(
    config => {
        // 从 store 或 localStorage 获取 userId 和 authToken
        const userId = store.state.userId || localStorage.getItem("userId");
        const authToken = store.state.authToken || localStorage.getItem("authToken");

        console.log('Adding headers:', { userId, authToken });

        // 始终添加 X-User-ID 头
        if (userId) {
            config.headers['X-User-ID'] = userId; // 添加自定义 Header
            console.log('X-User-ID header added');
        } else {
            console.log('X-User-ID header NOT added');
        }

        // 始终添加 Authorization 头
        if (authToken) {
            config.headers['Authorization'] = authToken; // 添加 Authorization Header
            console.log('Authorization header added');
        } else {
            console.log('Authorization header NOT added');
        }

        // 确保 Content-Type 设置为 application/json
        if (!config.headers['Content-Type']) { // 仅在未设置时添加
            config.headers['Content-Type'] = 'application/json';
        }

        console.log('Request Headers:', config.headers);

        return config;
    },
    error => {
        console.error('Request error:', error);
        return Promise.reject(error);
    }
);

export default axiosInstance;
