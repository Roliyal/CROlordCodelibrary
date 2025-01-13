const axiosInstance = axios.create({
    baseURL: 'http://47.83.211.8', // 后端 API 的基础 URL，请根据实际情况调整
    timeout: 10000, // 请求超时时间（毫秒）
    withCredentials: true, // 允许携带凭证（如 cookies）
});

// 请求拦截器
axiosInstance.interceptors.request.use(
    config => {
        const userId = localStorage.getItem("id");  // 从 localStorage 获取 userId
        const authToken = localStorage.getItem("authToken");  // 获取 authToken

        if (userId) {
            config.headers['X-User-ID'] = userId;  // 设置 X-User-ID
        }

        if (authToken) {
            config.headers['Authorization'] = `Bearer ${authToken}`;  // 设置 Authorization
        }

        return config;
    },
    error => {
        return Promise.reject(error);
    }
);
