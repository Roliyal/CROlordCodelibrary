// src/auth-api.js

import axiosInstance from "./axiosInstance";

export default {
    isAuthenticated: false,

    // 新增：检查是否有用户信息
    checkAuth() {
        const userId = localStorage.getItem('userId');
        const authToken = localStorage.getItem('authToken');

        if (userId && authToken) {
            this.isAuthenticated = true;  // 设置已认证状态
            console.log('User is already authenticated');
            return { userId, authToken };
        } else {
            this.isAuthenticated = false;
            console.log('No authentication data found');
            return null;
        }
    },

    async authenticate(username, password) {
        try {
            const response = await axiosInstance.post(`/login`, {
                username,
                password,
            });

            console.log('Login response:', response.data);

            if (response.data && response.data.success && response.data.id !== undefined && response.data.authToken) {
                this.isAuthenticated = true;

                // 存储用户信息
                localStorage.setItem('userId', response.data.id);
                localStorage.setItem('authToken', response.data.authToken);

                return {
                    id: response.data.id,
                    authToken: response.data.authToken, // 返回 authToken
                };
            } else {
                return null;
            }
        } catch (error) {
            console.error("Error authenticating:", error);
            return null;
        }
    },

    async register(username, password) {
        try {
            const response = await axiosInstance.post(`/register`, {
                username,
                password,
            });

            if (response.status === 201) {
                return { status: response.status };
            } else {
                return { status: response.status, error: "注册失败，请重试。" };
            }
        } catch (error) {
            console.error("Error registering:", error);
            return { status: 500, error: "注册失败，请重试。" };
        }
    },
}
