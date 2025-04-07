// src/auth-api.js
import axiosInstance from './axiosInstance';
import store from './store';

export default {
    // 登录
    async login(username, password) {
        try {
            const response = await axiosInstance.post('/login', {
                username,
                password,
            });
            const data = response.data;

            // 如果后端返回 data.success === true，并且有 id / authToken
            if (data && data.success && data.id && data.authToken) {
                // 存到 Vuex
                store.commit('setUserId', data.id);
                store.commit('setAuthToken', data.authToken);
                store.commit('setIsLoggedIn', true);

                // 同步 localStorage
                localStorage.setItem('userId', data.id);
                localStorage.setItem('authToken', data.authToken);

                console.log('login success', data.id, data.authToken);
                return data; // 或者返回 { id, authToken }
            } else {
                return null;
            }
        } catch (error) {
            console.error('Login failed:', error);
            return null;
        }
    },

    // 注册
    async register(username, password) {
        try {
            const response = await axiosInstance.post('/register', {
                username,
                password,
            });
            const data = response.data;

            // 例如你后端如果 status=201 表示创建成功
            if (response.status === 201 && data && data.id && data.authToken) {
                // 存到 Vuex
                store.commit('setUserId', data.id);
                store.commit('setAuthToken', data.authToken);
                store.commit('setIsLoggedIn', true);

                // 同步 localStorage
                localStorage.setItem('userId', data.id);
                localStorage.setItem('authToken', data.authToken);

                console.log('register success', data.id, data.authToken);
                return data;
            } else {
                return null;
            }
        } catch (error) {
            console.error('Registration failed:', error);
            return null;
        }
    },
};
