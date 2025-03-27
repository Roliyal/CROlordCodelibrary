import axiosInstance from './axiosInstance';
import store from './store';

export default {
    async login(username, password) {
        try {
            const response = await axiosInstance.post('/login', {
                username,
                password
            });

            if (response.data?.success && response.data.id && response.data.authToken) {
                // 清除旧存储
                localStorage.removeItem('userId');
                localStorage.removeItem('authToken');
                document.cookie = "X-User-ID=; path=/; domain=.roliyal.com; expires=Thu, 01 Jan 1970 00:00:00 GMT";

                // 设置新存储（HTTP专用）
                localStorage.setItem('userId', response.data.id);
                localStorage.setItem('authToken', response.data.authToken);
                document.cookie = `X-User-ID=${response.data.id}; path=/; domain=.roliyal.com; SameSite=Lax`;

                // 更新Vuex
                store.commit('setUserId', response.data.id);
                store.commit('setAuthToken', response.data.authToken);
                store.commit('setIsLoggedIn', true);

                return {
                    id: response.data.id,
                    authToken: response.data.authToken
                };
            }
            return null;
        } catch (error) {
            console.error('Login error:', error);
            return null;
        }
    },

    async register(username, password) {
        try {
            const response = await axiosInstance.post('/register', {
                username,
                password
            });

            if (response.status === 201 && response.data?.id) {
                // 使用和login相同的存储逻辑
                localStorage.setItem('userId', response.data.id);
                localStorage.setItem('authToken', response.data.authToken);
                document.cookie = `X-User-ID=${response.data.id}; path=/; domain=.roliyal.com; SameSite=Lax`;

                store.commit('setUserId', response.data.id);
                store.commit('setAuthToken', response.data.authToken);
                store.commit('setIsLoggedIn', true);

                return {
                    id: response.data.id,
                    authToken: response.data.authToken
                };
            }
            return null;
        } catch (error) {
            console.error('Registration error:', error);
            return null;
        }
    }
}