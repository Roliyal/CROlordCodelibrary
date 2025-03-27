import axios from 'axios';
import store from './store';

const axiosInstance = axios.create({
    baseURL: 'http://micro.roliyal.com',
    timeout: 10000,
    withCredentials: true,  // 允许跨域携带 Cookie
});

// 获取 Cookie 的辅助函数
function getCookie(name) {
    const cookies = document.cookie.split(';');
    for (let cookie of cookies) {
        const [cookieName, cookieValue] = cookie.trim().split('=');
        if (cookieName === name) return cookieValue;
    }
    return null;
}

// 请求拦截器
axiosInstance.interceptors.request.use((config) => {
    // 优先级：Vuex → localStorage → Cookie
    const userId = store.getters.userId || localStorage.getItem('userId') || getCookie('X-User-ID');
    const authToken = store.getters.authToken || localStorage.getItem('authToken');

    // 设置请求头
    config.headers = config.headers || {};
    if (userId) config.headers['X-User-ID'] = userId;
    if (authToken) config.headers['Authorization'] = `Bearer ${authToken}`;

    // 确保 Content-Type
    if (!config.headers['Content-Type']) {
        config.headers['Content-Type'] = 'application/json';
    }

    console.log('Request headers:', config.headers);
    return config;
}, (error) => {
    console.error('Request error:', error);
    return Promise.reject(error);
});

export default axiosInstance;