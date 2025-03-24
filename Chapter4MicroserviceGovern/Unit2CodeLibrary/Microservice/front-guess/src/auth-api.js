// src/auth-api.js
import axiosInstance from './axiosInstance'; // 引入自定义的 axios 实例

export default {
    // 用户登录
    async authenticate(username, password) {
        try {
            const response = await axiosInstance.post('/login', {
                username,
                password,
            });

            console.log('Login response:', response.data);

            if (response.data && response.data.success && response.data.id && response.data.authToken) {
                return {
                    id: response.data.id,
                    authToken: response.data.authToken,
                };
            }

            return null;
        } catch (error) {
            console.error('Login failed:', error);
            return null;
        }
    },

    // 用户注册
    async register(username, password) {
        try {
            const response = await axiosInstance.post('/register', { username, password });

            console.log('Register response:', response.data);

            if (response.status === 201) {
                return response.data;
            }

            return null;
        } catch (error) {
            console.error('Registration failed:', error);
            return null;
        }
    },
}
