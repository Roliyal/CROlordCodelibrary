// src/store/index.js
import { createStore } from 'vuex';

export default createStore({
    state: {
        isLoggedIn: false,
        userId: null,
        authToken: null,
        traceId: null,  // 用于存储 Trace ID
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
        setTraceId(state, traceId) {  // 设置 Trace ID
            state.traceId = traceId;
        },
        logout(state) {
            state.isLoggedIn = false;
            state.userId = null;
            state.authToken = null;
            state.traceId = null;  // 清除 Trace ID
            localStorage.removeItem('userId');
            localStorage.removeItem('authToken');
            document.cookie = 'X-User-ID=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT';
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
        traceId: (state) => state.traceId,  // 获取 Trace ID
    },
});
