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
                // 登录成功
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
};
