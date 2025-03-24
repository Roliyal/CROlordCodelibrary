// src/axiosInstance.js
import axios from 'axios';
import store from './store';  // 引入 Vuex store

// 创建 Axios 实例
const axiosInstance = axios.create({
    baseURL: 'http://micro.roliyal.com',  // 后端服务的基础 URL
    timeout: 10000,  // 请求超时时间
    withCredentials: true,  // 允许携带凭证（如 cookies）
});

// 请求拦截器
axiosInstance.interceptors.request.use(
    (config) => {
        const userId = store.getters.userId || localStorage.getItem('userId');
        const authToken = store.getters.authToken || localStorage.getItem('authToken');

        console.log('Adding headers:', { userId, authToken });  // 日志输出，检查请求头

        // 在所有请求中都添加 X-User-ID 和 Authorization
        // 只要 userId 存在就添加 X-User-ID
        if (userId) {
            config.headers['X-User-ID'] = userId;  // 添加 X-User-ID 请求头
        }

        // 确保在所有请求中添加 Authorization 请求头
        if (authToken) {
            config.headers['Authorization'] = authToken;  // 添加 Authorization 请求头
        }

        console.log('Request headers:', config.headers);  // 打印出完整的请求头，检查是否正确

        // 设置 Content-Type 为 application/json（如果未设置）
        if (!config.headers['Content-Type']) {
            config.headers['Content-Type'] = 'application/json';
        }

        return config;
    },
    (error) => {
        console.error('Request error:', error);
        return Promise.reject(error);
    }
);

export default axiosInstance;
