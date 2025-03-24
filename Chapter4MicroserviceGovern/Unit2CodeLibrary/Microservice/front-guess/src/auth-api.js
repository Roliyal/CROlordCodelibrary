// src/auth-api.js
import axiosInstance from './axiosInstance'; // 引入自定义的 axios 实例
import store from './store';  // 引入 Vuex store

export default {
    // 用户登录
    async authenticate(username, password) {
        try {
            // 获取 Vuex 或 localStorage 中的 userId，如果没有则使用 guest 或 null
            const userId = store.getters.userId || localStorage.getItem('userId') || 'guest';
            const headers = {
                'X-User-ID': userId,  // 显式添加 X-User-ID 头部
            };

            const response = await axiosInstance.post('/login', {
                username,
                password,
            }, { headers });  // 将 headers 传递给请求

            console.log('Login response:', response.data);

            if (response.data && response.data.success && response.data.id && response.data.authToken) {
                // 登录成功后更新 Vuex 和 localStorage
                store.commit('setUserId', response.data.id);
                store.commit('setAuthToken', response.data.authToken);
                store.commit('setIsLoggedIn', true);
                localStorage.setItem('userId', response.data.id);
                localStorage.setItem('authToken', response.data.authToken);

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
