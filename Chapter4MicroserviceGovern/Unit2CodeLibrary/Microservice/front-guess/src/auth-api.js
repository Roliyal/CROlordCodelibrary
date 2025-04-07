// src/auth-api.js
import axiosInstance from './axiosInstance';
import store from './store';

export default {
    async login(username, password) {
        try {
            const response = await axiosInstance.post('/login', {
                username,
                password,
            });
            const data = response.data;
            if (data && data.success && data.id && data.authToken) {
                // 记录到 Vuex
                store.commit('setUserId', data.id);
                store.commit('setAuthToken', data.authToken);
                store.commit('setIsLoggedIn', true);

                // 同步 localStorage
                localStorage.setItem('userId', data.id);
                localStorage.setItem('authToken', data.authToken);

                console.log('login success', data.id, data.authToken);
                return data;
            }
            return null;
        } catch (error) {
            console.error('Login failed:', error);
            return null;
        }
    },

    async register(username, password) {
        try {
            const response = await axiosInstance.post('/register', {
                username,
                password,
            });
            const data = response.data;
            // 后端若返回 201 + 正常payload
            if (response.status === 201 && data && data.id && data.authToken) {
                // 记录到 Vuex
                store.commit('setUserId', data.id);
                store.commit('setAuthToken', data.authToken);
                store.commit('setIsLoggedIn', true);

                // 同步 localStorage
                localStorage.setItem('userId', data.id);
                localStorage.setItem('authToken', data.authToken);

                console.log('register success', data.id, data.authToken);
                return data;
            }
            return null;
        } catch (error) {
            console.error('Registration failed:', error);
            return null;
        }
    },
};
