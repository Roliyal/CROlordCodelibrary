import axios from 'axios';

// 创建 axios 实例
const instance = axios.create({
    baseURL: process.env.VUE_APP_API_URL || '',
    headers: {
        'X-Version': process.env.VUE_APP_VERSION, // 添加版本信息到 header 中
    },
});

// 请求拦截器
instance.interceptors.request.use(config => {
    return config;
}, error => {
    return Promise.reject(error);
});

// 响应拦截器
instance.interceptors.response.use(response => {
    return response;
}, error => {
    return Promise.reject(error);
});

export default instance;
