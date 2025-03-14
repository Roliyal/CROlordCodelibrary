// src/auth-api.js
import axiosInstance from "./axiosInstance";

export default {
    // 用户登录
    async authenticate(username, password) {
        try {
            const response = await axiosInstance.post(`/login`, {
                username,
                password,
            });

            console.log('Login response:', response.data);

            if (response.data && response.data.success && response.data.id && response.data.authToken) {
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
