// src/store/index.js
import { createStore } from 'vuex';

export default createStore({
    state: {
        traceId: null,  // 存储 traceId
        isLoggedIn: false,
        userId: null,
        authToken: null,
    },
    mutations: {
        setTraceId(state, traceId) {
            state.traceId = traceId;
        },
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
            localStorage.removeItem('userId');
            localStorage.removeItem('authToken');
        },
    },
    actions: {
        logout({ commit }) {
            commit('logout');
        },
    },
    getters: {
        isLoggedIn: state => state.isLoggedIn,
        userId: state => state.userId,
        authToken: state => state.authToken,
        traceId: state => state.traceId,  // 获取 traceId
    },
});
