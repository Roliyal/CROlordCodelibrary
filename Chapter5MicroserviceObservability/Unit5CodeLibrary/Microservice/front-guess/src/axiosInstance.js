import axios from 'axios';
import store from './store';

const axiosInstance = axios.create({
    baseURL: 'http://micro.roliyal.com',
    timeout: 10000,
    withCredentials: true, // 如果需要携带 Cookies 或认证信息
});

// 删除 Cookie 函数
function deleteCookie(name) {
    document.cookie = name + '=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';
}

// 请求拦截器：设置请求头
axiosInstance.interceptors.request.use(config => {
    const userId = store.state.userId || localStorage.getItem('userId');
    const authToken = store.state.authToken || localStorage.getItem('authToken');

    // 设置请求头
    if (userId) {
        deleteCookie('X-User-ID');
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

// 响应拦截器：从请求配置中获取 X-B3-TraceId
axiosInstance.interceptors.response.use(
    response => {
        // 获取请求头中的 X-B3-TraceId（从 request 配置中）
        const traceId = response.config.headers['X-B3-TraceId'] || 'No traceId available';
        console.log('Trace ID from request headers (success):', traceId); // 打印请求头中的 traceId

        return response;
    },
    error => {
        // 如果请求失败，检查请求头中的 X-B3-TraceId
        const traceId = error.config?.headers['X-B3-TraceId'] || 'No traceId available';
        console.error('Request failed:', error);
        console.log('Trace ID from request headers (error):', traceId); // 打印请求头中的 traceId

        return Promise.reject(error);
    }
);

export default axiosInstance;