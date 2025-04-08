import axios from 'axios';
import store from './store';  // Vuex store

const axiosInstance = axios.create({
    baseURL: 'http://micro.roliyal.com', // 你的后端地址
    timeout: 10000,
    withCredentials: true,  // 允许跨域带 Cookie
});

axiosInstance.interceptors.request.use(
    (config) => {
        // 从 Vuex/localStorage 里拿 userId 和 token
        let userId = store.state.userId || localStorage.getItem('userId');
        let authToken = store.state.authToken || localStorage.getItem('authToken');

        // 设置 Cookie: x-pre-higress-tag=gray,X-User-ID=xxx
        if (userId) {
            document.cookie = `x-pre-higress-tag=gray,X-User-ID=${userId}; path=/; SameSite=None; Secure`;
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
