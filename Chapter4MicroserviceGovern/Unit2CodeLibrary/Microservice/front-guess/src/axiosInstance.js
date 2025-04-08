import axios from 'axios';
import store from './store';  // Vuex store

const axiosInstance = axios.create({
    baseURL: 'http://micro.roliyal.com',
    timeout: 10000,
    withCredentials: true, // 允许跨域带 Cookie
});

axiosInstance.interceptors.request.use((config) => {
    let userId = store.state.userId || localStorage.getItem('userId');
    let authToken = store.state.authToken || localStorage.getItem('authToken');

    // 分成两个 Cookie
    // 1) x-pre-higress-tag=gray
    document.cookie = `x-pre-higress-tag=gray; path=/; SameSite=None; Secure`;

    // 2) X-User-ID=xxx
    if (userId) {
        document.cookie = `X-User-ID=${userId}; path=/; SameSite=None; Secure`;
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
}, (error) => Promise.reject(error));

export default axiosInstance;
