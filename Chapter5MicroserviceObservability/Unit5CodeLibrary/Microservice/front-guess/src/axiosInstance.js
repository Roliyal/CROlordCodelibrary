// src/axiosInstance.js
import axios from 'axios';
import store from './store';

const axiosInstance = axios.create({
    baseURL: 'http://micro.roliyal.com',
    timeout: 10000,
    withCredentials: true,
});

// 请求拦截器
axiosInstance.interceptors.request.use(config => {
    const userId = store.state.userId || localStorage.getItem('userId');
    const authToken = store.state.authToken || localStorage.getItem('authToken');

    if (userId) {
        document.cookie = `X-User-ID=${userId}; path=/;`;
    }

    if (authToken) {
        config.headers['Authorization'] = authToken;
    }

    if (!config.headers['Content-Type']) {
        config.headers['Content-Type'] = 'application/json';
    }

    return config;
}, error => Promise.reject(error));

// 响应拦截器
axiosInstance.interceptors.response.use(response => {
    // 从响应头中提取 Trace ID
    const traceId = response.headers['x-b3-traceid'];
    if (traceId) {
        // 存储 Trace ID 到 Vuex 中
        store.commit('setTraceId', traceId);
    }
    return response;
}, error => {
    // 如果发生错误，尝试从错误的响应中提取 Trace ID
    const traceId = error.response?.headers['x-b3-traceid'];
    if (traceId) {
        store.commit('setTraceId', traceId);
    }
    return Promise.reject(error);
});

export default axiosInstance;
