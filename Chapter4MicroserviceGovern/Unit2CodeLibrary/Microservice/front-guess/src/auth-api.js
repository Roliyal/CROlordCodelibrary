import axiosInstance from './axiosInstance';
import store from './store';

export default {
    async login(username, password) {
        try {
            const response = await axiosInstance.post('/login', {
                username,
                password,
            });

            console.log('Login response:', response.data);

            if (response.data?.success && response.data.userId && response.data.authToken) {
                // 清除旧数据
                localStorage.removeItem('userId');
                localStorage.removeItem('authToken');
                document.cookie = "X-User-ID=; path=/; domain=.roliyal.com; expires=Thu, 01 Jan 1970 00:00:00 GMT";

                // 存储新数据（HTTP 环境）
                localStorage.setItem('userId', response.data.id);
                localStorage.setItem('authToken', response.data.authToken);
                document.cookie = `X-User-ID=${response.data.id}; path=/; domain=.roliyal.com; SameSite=Lax`;

                // 更新 Vuex
                store.commit('setUserId', response.data.id);
                store.commit('setAuthToken', response.data.authToken);
                store.commit('setIsLoggedIn', true);

                return {
                    userId: response.data.id,
                    authToken: response.data.authToken,
                };
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

            if (response.status === 201 && response.data?.id && response.data?.authToken) {
                // 存储数据（和登录逻辑一致）
                localStorage.setItem('userId', response.data.id);
                localStorage.setItem('authToken', response.data.authToken);
                document.cookie = `X-User-ID=${response.data.id}; path=/; domain=.roliyal.com; SameSite=Lax`;

                // 更新 Vuex
                store.commit('setUserId', response.data.id);
                store.commit('setAuthToken', response.data.authToken);
                store.commit('setIsLoggedIn', true);

                return {
                    userId: response.data.userId,
                    authToken: response.data.authToken,
                };
            }
            return null;
        } catch (error) {
            console.error('Registration failed:', error);
            return null;
        }
    }
}