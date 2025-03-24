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

        // 如果是 /login 或 /register 请求，使用默认的 guest 作为 X-User-ID
        if (config.url.includes('/login') || config.url.includes('/register')) {
            config.headers['X-User-ID'] = 'guest';  // 使用默认的 guest
        } else {
            // 对于其他请求，确保使用真实的 userId 和 authToken
            if (userId) {
                config.headers['X-User-ID'] = userId;  // 使用实际的 userId
            } else {
                // 如果没有 userId，设置为 guest
                config.headers['X-User-ID'] = userId;
            }

            if (authToken) {
                config.headers['Authorization'] = authToken;  // 使用实际的 Authorization token
            }
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
