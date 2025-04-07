// src/axiosInstance.js
import axios from 'axios';
import store from './store';

const axiosInstance = axios.create({
    baseURL: 'http://micro.roliyal.com', // 你的后端地址
    timeout: 10000,
    withCredentials: true,              // 允许携带 Cookie
});

axiosInstance.interceptors.request.use(
    (config) => {
        // 1. 优先从 Vuex 获取
        let userId = store.state.userId;
        let authToken = store.state.authToken;

        // 2. 如果 Vuex 没值，就从 localStorage 获取
        if (!userId) {
            userId = localStorage.getItem('userId');
        }
        if (!authToken) {
            authToken = localStorage.getItem('authToken');
        }

        // 3. 拿到就加到请求头里
        if (userId) {
            config.headers['X-User-ID'] = userId;
            // 也写到Cookie（如果需要）
            document.cookie = `X-User-ID=${userId}; path=/;`;
        }
        if (authToken) {
            config.headers['Authorization'] = authToken;
        }

        // 默认 Content-Type
        if (!config.headers['Content-Type']) {
            config.headers['Content-Type'] = 'application/json';
        }

        return config;
    },
    (error) => Promise.reject(error)
);

export default axiosInstance;
