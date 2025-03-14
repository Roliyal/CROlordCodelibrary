import axios from "axios";
import store from "./store";  // 引入 Vuex store

// 创建 Axios 实例
const axiosInstance = axios.create({
    baseURL: 'http://micro.roliyal.com',  // 后端服务的基础 URL
    timeout: 10000,  // 请求超时时间
    withCredentials: true,  // 允许携带凭证（如 cookies）
});

// 请求拦截器
axiosInstance.interceptors.request.use(
    (config) => {
        // 从 Vuex 获取 userId 和 authToken
        const userId = store.state.userId || localStorage.getItem("userId");
        const authToken = store.state.authToken || localStorage.getItem("authToken");

        // 如果是登录请求，也带上 X-User-ID
        if (config.url === "/login" && userId) {
            config.headers['X-User-ID'] = userId;  // 将 X-User-ID 加入请求头
        }

        if (authToken) {
            config.headers['Authorization'] = `Bearer ${authToken}`;  // Bearer 认证模式
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

export default axiosInstance;
