// src/store/index.js
import { createStore } from "vuex";

const store = createStore({
    state: {
        isLoggedIn: false,  // 默认未登录
        userId: null,       // 用户 ID
        authToken: null,    // 用户认证 token
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
        login({ commit }, { userId, authToken }) {
            commit("setIsLoggedIn", true);
            commit("setUserId", userId);
            commit("setAuthToken", authToken);
        },
        logout({ commit }) {
            commit("setIsLoggedIn", false);
            commit("setUserId", null);
            commit("setAuthToken", null);
        },
    },
    getters: {
        isLoggedIn: (state) => state.isLoggedIn,
        userId: (state) => state.userId,
        authToken: (state) => state.authToken,
    },
});

export default store;
