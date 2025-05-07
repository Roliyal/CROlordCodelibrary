import axios from 'axios';
import store from './store';

const axiosInstance = axios.create({
    baseURL: 'http://micro.roliyal.com',
    timeout: 10000,
    withCredentials: true,
});

function deleteCookie(name) {
    document.cookie = name + '=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';
}

axiosInstance.interceptors.request.use(config => {
    const userId = store.state.userId || localStorage.getItem('userId');
    const authToken = store.state.authToken || localStorage.getItem('authToken');

    // 仅在 userId 存在时设置 Cookie
    if (userId) {
        deleteCookie('X-User-ID');
        document.cookie = `X-User-ID=${userId}; path=/;`;
    }

    // 固定设置版本灰度标签
    deleteCookie('x-pre-higress-tag');
    document.cookie = `x-pre-higress-tag=gray; path=/;`;

    if (authToken) {
        config.headers['Authorization'] = authToken;
    }

    if (!config.headers['Content-Type']) {
        config.headers['Content-Type'] = 'application/json';
    }

    return config;
}, error => Promise.reject(error));

// 响应拦截器：获取 traceId
axiosInstance.interceptors.response.use(response => {
    // 从响应头中获取 traceId
    const traceId = response.headers['x-b3-traceid'];

    if (traceId) {
        // 存储 traceId 到 Vuex 或 localStorage
        store.commit('setTraceId', traceId);
        localStorage.setItem('traceId', traceId);
    }

    return response;
}, error => {
    // 出错时从响应头中获取 traceId 并打印
    const traceId = error.response?.headers['x-b3-traceid'];
    if (traceId) {
        console.error(`Error occurred, Trace ID: ${traceId}`);
    }
    return Promise.reject(error);
});

export default axiosInstance;
