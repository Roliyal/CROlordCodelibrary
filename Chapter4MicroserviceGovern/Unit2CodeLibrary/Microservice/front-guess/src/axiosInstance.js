// src/axiosInstance.js
import axios from 'axios';
import store from './store';  // Vuex

const axiosInstance = axios.create({
    baseURL: 'http://micro.roliyal.com',
    timeout: 10000,
    withCredentials: true, // 允许携带Cookie
});

axiosInstance.interceptors.request.use(
    (config) => {
        // 1. 取 userId / authToken
        let userId = store.state.userId || localStorage.getItem('userId');
        let authToken = store.state.authToken || localStorage.getItem('authToken');

        // 2. 不再加头 X-User-ID
        //    只设置一个 Cookie： x-pre-higress-tag=gray,X-User-ID=<userId>
        if (userId) {
            document.cookie = `x-pre-higress-tag=gray,X-User-ID=${userId}; path=/;`;
        } else {
            // 如果没登录没有 userId，你也可以不写Cookie或写一个默认
        }

        // 如果后端要 Authorization
        if (authToken) {
            config.headers['Authorization'] = authToken;
        }

        // Content-Type
        if (!config.headers['Content-Type']) {
            config.headers['Content-Type'] = 'application/json';
        }

        return config;
    },
    (error) => Promise.reject(error)
);

export default axiosInstance;
