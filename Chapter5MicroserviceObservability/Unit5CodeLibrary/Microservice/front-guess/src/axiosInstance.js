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

// 响应拦截器：获取响应头中的 X-B3-TraceId
axiosInstance.interceptors.response.use(
    response => {
        // 在响应头中获取 X-B3-TraceId
        const traceId = response.headers['x-b3-traceid'] || 'No traceId available';
        console.log('Trace ID from response headers:', traceId); // 打印响应头中的 traceId

        // 将 traceId 存储到 Vuex 状态或其他地方（如 localStorage）
        store.commit('setTraceId', traceId);  // 你需要在 Vuex store 中定义这个 mutation

        return response;
    },
    error => {
        // 如果请求失败，检查请求头中的 X-B3-TraceId
        const traceId = error.response?.headers['x-b3-traceid'] || 'No traceId available';
        console.error('Request failed:', error);
        console.log('Trace ID from response headers (on error):', traceId);

        // 如果你希望在失败时也保存 traceId 可以在这里做
        store.commit('setTraceId', traceId);  // 你需要在 Vuex store 中定义这个 mutation

        return Promise.reject(error);
    }
);

export default axiosInstance;
