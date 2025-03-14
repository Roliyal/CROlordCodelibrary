// src/store/index.js
import { createStore } from "vuex";
import axiosInstance from "../axiosInstance";  // 引入 axios 实例

const store = createStore({
    state: {
        isLoggedIn: false,   // 默认未登录
        userId: null,        // 用户 ID
        authToken: null,     // 用户认证 token
    },
    mutations: {
        setIsLoggedIn(state, isLoggedIn) {
            state.isLoggedIn = isLoggedIn;
        },
        setUserId(state, userId) {
            state.userId = userId;
        },
        setAuthToken(state, authToken) {
            state.authToken = authToken;
        },
    },
    actions: {
        // 登录 action
        async login({ commit }, { username, password }) {
            try {
                const response = await axiosInstance.post("/login", { username, password });

                if (response.data && response.data.success && response.data.id && response.data.authToken) {
                    // 更新 Vuex 状态
                    commit("setIsLoggedIn", true);
                    commit("setUserId", response.data.id);
                    commit("setAuthToken", response.data.authToken);

                    // 将用户信息存储到 localStorage 中，以便在页面刷新时保持登录状态
                    localStorage.setItem("userId", response.data.id);
                    localStorage.setItem("authToken", response.data.authToken);

                    return response.data;  // 返回用户数据
                }
                return null;
            } catch (error) {
                console.error("Login failed:", error);
                return null;
            }
        },

        // 注册 action
        async register({ commit }, { username, password }) {
            try {
                const response = await axiosInstance.post("/register", { username, password });

                if (response.status === 201) {
                    // 注册成功，设置用户状态
                    commit("setIsLoggedIn", true);
                    commit("setUserId", response.data.id);
                    commit("setAuthToken", response.data.authToken);

                    // 将用户信息存储到 localStorage
                    localStorage.setItem("userId", response.data.id);
                    localStorage.setItem("authToken", response.data.authToken);

                    return response.data;  // 返回注册后的用户数据
                }
                return null;
            } catch (error) {
                console.error("Registration failed:", error);
                return null;
            }
        },

        // 登出 action
        async logout({ commit }) {
            commit("setIsLoggedIn", false);
            commit("setUserId", null);
            commit("setAuthToken", null);

            localStorage.removeItem("userId");
            localStorage.removeItem("authToken");

            // 跳转到登录页
            return "Logged out";
        },
    },
    getters: {
        isLoggedIn: (state) => state.isLoggedIn,
        userId: (state) => state.userId,
        authToken: (state) => state.authToken,
    },
});

export default store;
