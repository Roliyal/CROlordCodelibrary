import axios from "axios";



export default {
    isAuthenticated: false,

    async authenticate(username, password) {
        try {
            const response = await axios.post("http://localhost:8083/login", { // 更新这一行
                username,
                password,
            });
            console.log("Response data:", response.data); // 添加这

            if (response.data && response.data.authToken) {
                this.isAuthenticated = true; // 添加这一行

                return {
                    authToken: response.data.authToken,
                    id: response.data.id, // 确保这里获取了 userID
                };
            } else {
                return null;
            }
        } catch (error) {
            console.error("Error authenticating:", error);
            return null;
        }
    },
    // 添加 register 函数使用
    async register(username, password) {
        try {
            const response = await axios.post("http://localhost:8083/register", {
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

