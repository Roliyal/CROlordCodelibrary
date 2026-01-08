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
            // Log the error and traceId if available
            const traceId = error?.response?.data?.traceId || 'No traceId available';
            console.error('Login failed:', error.message || error);
            console.log('Trace ID:', traceId);

            // Customize the error message for users
            this.errorMessage = error?.response?.data?.message || '登录失败，请检查用户名和密码是否正确。';

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
            // Log the error and traceId if available
            const traceId = error?.response?.data?.traceId || 'No traceId available';
            console.error('Registration failed:', error.message || error);
            console.log('Trace ID:', traceId);

            // Customize the error message for users
            this.errorMessage = error?.response?.data?.message || '注册失败，请重试。';

            return null;
        }
    },
};

