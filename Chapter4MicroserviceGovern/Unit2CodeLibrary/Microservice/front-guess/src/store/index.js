// src/store/index.js
import { createStore } from 'vuex';

export default createStore({
    state: {
        isLoggedIn: false,
        userId: null,
        authToken: null,
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
        logout(state) {
            state.isLoggedIn = false;
            state.userId = null;
            state.authToken = null;

            // 清除 localStorage 中的用户信息
            localStorage.removeItem('userId');
            localStorage.removeItem('authToken');

            // 删除 cookie 中的 X-User-ID
            document.cookie = "X-User-ID=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT";
        },
    },
    actions: {
        logout({ commit }) {
            commit('logout');
        },
    },
    getters: {
        isLoggedIn: (state) => state.isLoggedIn,
        userId: (state) => state.userId,
        authToken: (state) => state.authToken,
    },
});
