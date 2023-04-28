import axios from "axios";
import config from "./config";

export default {
    isAuthenticated: false,

    async authenticate(username, password) {
        try {
            const response = await axios.post(`${config.loginURL}/login`, {
                username,
                password,
            });

            if (response.data && response.data.authToken) {
                this.isAuthenticated = true;

                return {
                    authToken: response.data.authToken,
                    id: response.data.id,
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
            const response = await axios.post(`${config.registerURL}/register`, {
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
};
