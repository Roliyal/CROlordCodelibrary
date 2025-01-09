// src/axiosInstance.js
import axios from "axios";
import store from "./store";

const axiosInstance = axios.create({
    baseURL: 'http://47.83.211.8', // 确保与您的后端 URL 一致
    timeout: 10000, // 请求超时时间（毫秒）
});

// 添加请求拦截器
axiosInstance.interceptors.request.use(
    config => {
        // 从 store 或 localStorage 获取 userId
        const userId = store.state.userId || localStorage.getItem("id");
        if (userId) {
            config.headers['X-User-ID'] = userId; // 添加自定义 Header
        }

        // 从 localStorage 获取 authToken 并设置 Authorization Header
        const authToken = localStorage.getItem("authToken");
        if (authToken) {
            config.headers['Authorization'] = `Bearer ${authToken}`;
        }

        return config;
    },
    error => {
        return Promise.reject(error);
    }
);

export default axiosInstance;
