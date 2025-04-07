// src/axiosInstance.js
import axios from 'axios';
import store from './store';  // 引入 Vuex store

// 创建 Axios 实例
const axiosInstance = axios.create({
    baseURL: 'http://micro.roliyal.com',  // 后端服务的基础 URL
    timeout: 10000,                      // 请求超时时间
    withCredentials: true,               // 允许携带凭证（如 cookies）
});

// 请求拦截器
axiosInstance.interceptors.request.use(
    (config) => {
        // 尝试从 Vuex 或 localStorage 获取 userId 和 authToken
        let userId = store.getters.userId || localStorage.getItem('userId');
        let authToken = store.getters.authToken || localStorage.getItem('authToken');

        // 如果没有从 Vuex 或 localStorage 获取到用户信息，则尝试从 Cookie 获取
        if (!userId) {
            userId = getCookie('X-User-ID'); // 尝试从 cookie 获取 X-User-ID
        }

        // 在请求头中加入 X-User-ID 和 Authorization
        if (userId) {
            config.headers['X-User-ID'] = userId;  // 添加 X-User-ID 请求头
        }
        if (authToken) {
            config.headers['Authorization'] = authToken;  // 添加 Authorization（如果需要）
        }

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

// 获取 cookie 中的值
function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
    return null;
}

export default axiosInstance;
