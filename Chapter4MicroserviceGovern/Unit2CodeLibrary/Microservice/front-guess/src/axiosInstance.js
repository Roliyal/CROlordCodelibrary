// src/axiosInstance.js
import axios from 'axios';
import store from './store';

const axiosInstance = axios.create({
    baseURL: 'http://micro.roliyal.com',
    timeout: 10000,
    withCredentials: true, // 允许携带 Cookie
});

axiosInstance.interceptors.request.use(
    (config) => {
        // 从 Vuex / localStorage 拿到 userId、authToken
        let userId = store.state.userId || localStorage.getItem('userId');
        let authToken = store.state.authToken || localStorage.getItem('authToken');

        if (userId) {
            document.cookie = `x-pre-higress-tag=gray,X-User-ID=${userId}; path=/;`;
        } else {
            // 如果 userId 为空也可以写一个默认 Cookie
            // document.cookie = 'x-pre-higress-tag=gray,X-User-ID=UNKNOWN; path=/;';
        }

        // ② 如果后端要鉴权，保留 Authorization 头
        if (authToken) {
            config.headers['Authorization'] = authToken;
        }

        // ③ Content-Type
        if (!config.headers['Content-Type']) {
            config.headers['Content-Type'] = 'application/json';
        }

        return config;
    },
    (error) => Promise.reject(error)
);

export default axiosInstance;
