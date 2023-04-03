import axios from "axios";

const apiClient = axios.create({
    baseURL: "/",
    withCredentials: false,
    headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
    },
});

export default {
    isAuthenticated: false,

    async authenticate(username, password) {
        try {
            const response = await apiClient.post("/login", { username, password });

            if (response.data && response.data.success) {
                this.isAuthenticated = true;
                return true;
            } else {
                this.isAuthenticated = false;
                return false;
            }
        } catch (error) {
            console.error("Error during authentication:", error);
            this.isAuthenticated = false;
            return false;
        }
    },
};
