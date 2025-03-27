import axios from 'axios';
import store from './store';

const axiosInstance = axios.create({
    baseURL: 'http://micro.roliyal.com',
    timeout: 10000,
    withCredentials: true // 关键：允许跨域携带凭证
});

// 优化的cookie获取方法
function getCookie(name) {
    const cookies = document.cookie.split(';');
    for (let cookie of cookies) {
        const [cookieName, cookieValue] = cookie.trim().split('=');
        if (cookieName === name) return cookieValue;
    }
    return null;
}

axiosInstance.interceptors.request.use((config) => {
    // 获取用户凭证的优先级：Vuex > localStorage > cookie
    const userId = store.getters.userId || localStorage.getItem('userId') || getCookie('X-User-ID');
    const authToken = store.getters.authToken || localStorage.getItem('authToken');

    // 确保headers对象存在
    config.headers = config.headers || {};

    // 设置标识头
    if (userId) config.headers['X-User-ID'] = userId;
    if (authToken) config.headers['Authorization'] = `Bearer ${authToken}`;

    // 默认Content-Type
    if (!config.headers['Content-Type']) {
        config.headers['Content-Type'] = 'application/json';
    }

    console.debug('Request headers:', config.headers);
    return config;
}, (error) => {
    console.error('Request error:', error);
    return Promise.reject(error);
});

export default axiosInstance;