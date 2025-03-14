// src/auth-api.js
import axiosInstance from "./axiosInstance";
import store from "./store";  // 引入 store.js

export default {
    isAuthenticated: false,

    // 检查是否已经登录
    checkAuth() {
        const userId = localStorage.getItem('userId');
        const authToken = localStorage.getItem('authToken');

        if (userId && authToken) {
            store.setIsLoggedIn(true); // 设置已登录状态
            store.setUserId(userId);   // 设置用户 ID
            store.setAuthToken(authToken);  // 设置 authToken
            console.log('User is already authenticated');
            return { userId, authToken };
        } else {
            store.setIsLoggedIn(false);
            console.log('No authentication data found');
            return null;
        }
    },

    // 用户登录
    async authenticate(username, password) {
        try {
            const response = await axiosInstance.post(`/login`, {
                username,
                password,
            });

            console.log('Login response:', response.data);

            if (response.data && response.data.success && response.data.id && response.data.authToken) {
                // 更新 store 状态
                store.setIsLoggedIn(true);
                store.setUserId(response.data.id);
                store.setAuthToken(response.data.authToken);

                // 将用户信息存储到 localStorage 中，以便在页面刷新时保持登录状态
                localStorage.setItem('userId', response.data.id);
                localStorage.setItem('authToken', response.data.authToken);

                return {
                    id: response.data.id,
                    authToken: response.data.authToken,
                };
            } else {
                return null;
            }
        } catch (error) {
            console.error("Error authenticating:", error);
            return null;
        }
    },

    // 用户注册
    async register(username, password) {
        try {
            const response = await axiosInstance.post(`/register`, {
                username,
                password,
            });

            if (response.status === 201) {
                return { status: response.status };
            } else {
                return { status: response.status, error: "Registration failed, please try again." };
            }
        } catch (error) {
            console.error("Error registering:", error);
            return { status: 500, error: "Registration failed, please try again." };
        }
    },
};
