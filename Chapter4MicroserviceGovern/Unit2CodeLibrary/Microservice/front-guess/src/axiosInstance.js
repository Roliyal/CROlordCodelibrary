// src/axiosInstance.js
import axios from 'axios';
import store from './store';

const axiosInstance = axios.create({
    baseURL: 'http://micro.roliyal.com', // 这里替换成你的后端地址
    timeout: 10000,
    withCredentials: true, // 允许携带 Cookie
});

// 请求拦截器
axiosInstance.interceptors.request.use(
    (config) => {
        // 优先从 Vuex 里取 userId、authToken
        let userId = store.state.userId;
        let authToken = store.state.authToken;

        // 如果没有，就再看看 localStorage
        if (!userId) {
            userId = localStorage.getItem('userId');
        }
        if (!authToken) {
            authToken = localStorage.getItem('authToken');
        }

        // 如果拿到了，就加到请求头
        if (userId) {
            config.headers['X-User-ID'] = userId;
            // 同时让浏览器带 Cookie（如果还没有，也可以写 Cookie）
            document.cookie = `X-User-ID=${userId}; path=/;`;
        }
        if (authToken) {
            config.headers['Authorization'] = authToken;
        }

        // 确保 Content-Type
        if (!config.headers['Content-Type']) {
            config.headers['Content-Type'] = 'application/json';
        }

        return config;
    },
    (error) => {
        return Promise.reject(error);
    }
);

export default axiosInstance;
