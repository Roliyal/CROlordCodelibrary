import axios from 'axios';
import store from './store';

const axiosInstance = axios.create({
    baseURL: 'http://micro.roliyal.com',
    timeout: 10000,
    withCredentials: true, // 允许浏览器在跨域请求中发送 Cookie
});

axiosInstance.interceptors.request.use(config => {
    let userId = store.state.userId || localStorage.getItem('userId');
    let authToken = store.state.authToken || localStorage.getItem('authToken');

    // 1) X-User-ID Cookie
    if (userId) {
        // 如果是同域HTTP
        document.cookie = `X-User-ID=${userId}; path=/; SameSite=None`;
        // 如果要跨域且HTTPS，需要 `; Secure`
        //document.cookie = `X-User-ID=${userId}; path=/; SameSite=None; Secure`;
    }

    // 2) x-pre-higress-tag=gray
    document.cookie = `x-pre-higress-tag=gray; path=/; SameSite=None`;
    // 如果HTTPS: + Secure

    // 3) Authorization header
    if (authToken) {
        config.headers['Authorization'] = authToken;
    }

    // 4) Content-Type
    if (!config.headers['Content-Type']) {
        config.headers['Content-Type'] = 'application/json';
    }

    return config;
}, error => Promise.reject(error));

export default axiosInstance;
