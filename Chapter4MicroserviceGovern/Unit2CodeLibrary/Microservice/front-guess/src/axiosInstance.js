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

        // 如果是 /login 或 /register 请求，强制携带 X-User-ID
        if (config.url.includes('/login') || config.url.includes('/register')) {
            if (userId) {
                config.headers['X-User-ID'] = userId;  // 添加 X-User-ID 请求头
            } else {
                // 如果 userId 为 null，使用 'guest' 或其他默认值
                config.headers['X-User-ID'] = 'guest';
            }
        }

        // 添加 Authorization 请求头
        if (authToken) {
            config.headers['Authorization'] = `authToken`;  // 添加 Authorization 请求头
        }

        // 打印请求头，确保正确设置
        console.log('Request headers:', config.headers);

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
