// src/auth-api.js
import axiosInstance from './axiosInstance';  // 使用我们配置好的 axiosInstance
import store from './store';                  // 引入 Vuex store

export default {
    // 用户登录
    async login(username, password) {
        try {
            const response = await axiosInstance.post('/login', {
                username,
                password,
            });

            console.log('Login response:', response.data);

            if (response.data && response.data.success && response.data.id && response.data.authToken) {
                // 清除之前的缓存数据（如果有）
                localStorage.removeItem('userId');
                localStorage.removeItem('authToken');
                document.cookie = "X-User-ID=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT"; // 删除旧 Cookie

                // 登录成功后更新 Vuex
                store.commit('setUserId', response.data.id);
                store.commit('setAuthToken', response.data.authToken);
                store.commit('setIsLoggedIn', true);

                // 同步到 localStorage
                localStorage.setItem('userId', response.data.id);
                localStorage.setItem('authToken', response.data.authToken);

                // 设置 Cookie：X-User-ID
                document.cookie = `X-User-ID=${response.data.id}; path=/;`;

                console.log('Stored userId and authToken in localStorage:', response.data.id, response.data.authToken);

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
            const response = await axiosInstance.post('/register', {
                username,
                password,
            });

            console.log('Register response:', response.data);

            // 假设后端返回的字段为： { id, authToken }，并且 status === 201 表示注册成功
            if (response.status === 201 && response.data.id && response.data.authToken) {
                // 注册成功后更新 Vuex
                store.commit('setUserId', response.data.id);
                store.commit('setAuthToken', response.data.authToken);
                store.commit('setIsLoggedIn', true);

                // 写入 localStorage
                localStorage.setItem('userId', response.data.id);
                localStorage.setItem('authToken', response.data.authToken);

                // 设置 Cookie：X-User-ID
                document.cookie = `X-User-ID=${response.data.id}; path=/;`;

                console.log('Stored userId and authToken in localStorage after registration:', response.data.id, response.data.authToken);

                return {
                    id: response.data.id,
                    authToken: response.data.authToken,
                };
            }

            return null;
        } catch (error) {
            console.error('Registration failed:', error);
            return null;
        }
    },
};
