// src/auth-api.js
import axiosInstance from './axiosInstance'; // Import axiosInstance from axiosInstance.js
import store from './store';  // Import Vuex store

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
                // 清除之前的缓存数据
                localStorage.removeItem('userId');
                localStorage.removeItem('authToken');
                document.cookie = "X-User-ID=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT";  // 删除旧的 cookie

                // 登录成功后更新 Vuex 和 localStorage
                store.commit('setUserId', response.data.id);
                store.commit('setAuthToken', response.data.authToken);
                store.commit('setIsLoggedIn', true);

                // 存储新数据到 localStorage 和 cookie 中
                localStorage.setItem('userId', response.data.id);
                localStorage.setItem('authToken', response.data.authToken);
                localStorage.removeItem('id');  // 删除 id 字段


                // 更新 cookie 中的 X-User-ID
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
}