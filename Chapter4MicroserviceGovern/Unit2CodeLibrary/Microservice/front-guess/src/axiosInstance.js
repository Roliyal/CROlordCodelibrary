import axios from 'axios';
import store from './store';

const axiosInstance = axios.create({
    baseURL: 'http://micro.roliyal.com',
    timeout: 10000,
    withCredentials: true,
});

// 清除已有 Cookie 中的指定字段（可选，但建议）
function deleteCookie(name) {
    document.cookie = name + '=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';
}

axiosInstance.interceptors.request.use(config => {
    let userId = store.state.userId || localStorage.getItem('userId') || '000020';
    let authToken = store.state.authToken || localStorage.getItem('authToken');

    // 清理旧值（如果有）
    deleteCookie('X-User-ID');
    deleteCookie('x-pre-higress-tag');

    document.cookie = `X-User-ID=${userId}; path=/;`;
    document.cookie = `x-pre-higress-tag=gray; path=/;`;

    if (authToken) {
        config.headers['Authorization'] = authToken;
    }

    // ✅ 设置默认 Content-Type
    if (!config.headers['Content-Type']) {
        config.headers['Content-Type'] = 'application/json';
    }

    return config;
}, error => Promise.reject(error));

export default axiosInstance;
