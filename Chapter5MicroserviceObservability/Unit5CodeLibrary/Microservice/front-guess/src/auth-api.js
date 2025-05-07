// src/auth-api.js
import axiosInstance from './axiosInstance';
import store from './store';

export default {
    // 登录
    async login(username, password) {
        try {
            const response = await axiosInstance.post('/login', { username, password });
            const data = response.data;
            console.log('Login response:', data);

            if (data && data.success && data.id && data.authToken) {
                // 清除之前的
                localStorage.removeItem('userId');
                localStorage.removeItem('authToken');

                // Vuex
                store.commit('setUserId', data.id);
                store.commit('setAuthToken', data.authToken);
                store.commit('setIsLoggedIn', true);

                // localStorage
                localStorage.setItem('userId', data.id);
                localStorage.setItem('authToken', data.authToken);

                console.log('Stored userId and authToken in localStorage:', data.id, data.authToken);
                return { id: data.id, authToken: data.authToken };
            }
            return null;
        } catch (error) {
            console.error('Login failed:', error);

            return null;
        }
    },

    // 注册
    async register(username, password) {
        try {
            const response = await axiosInstance.post('/register', { username, password });
            const data = response.data;
            console.log('Register response:', data);

            if (response.status === 201 && data && data.id && data.authToken) {
                // Vuex
                store.commit('setUserId', data.id);
                store.commit('setAuthToken', data.authToken);
                store.commit('setIsLoggedIn', true);

                localStorage.setItem('userId', data.id);
                localStorage.setItem('authToken', data.authToken);

                console.log('Stored userId and authToken in localStorage after registration:', data.id, data.authToken);
                return { id: data.id, authToken: data.authToken };
            }
            return null;
        } catch (error) {
            console.error('Registration failed:', error);
            return null;
        }
    },
};
