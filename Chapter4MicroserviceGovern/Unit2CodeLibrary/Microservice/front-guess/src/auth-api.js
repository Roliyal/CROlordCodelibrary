// src/auth-api.js
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
                // 登录成功后更新 Vuex 和 localStorage
                store.commit('setUserId', response.data.id);
                store.commit('setAuthToken', response.data.authToken);
                store.commit('setIsLoggedIn', true);

                // 存储在 localStorage 中
                localStorage.setItem('userId', response.data.id);
                localStorage.setItem('authToken', response.data.authToken);

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
