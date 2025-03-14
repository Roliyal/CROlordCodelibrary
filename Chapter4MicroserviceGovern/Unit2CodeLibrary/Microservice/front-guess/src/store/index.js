// src/store/index.js
import { createStore } from 'vuex';

export default createStore({
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
        setIsLoggedIn({ commit }, isLoggedIn) {
            commit('setIsLoggedIn', isLoggedIn);
        },
        setUserId({ commit }, userId) {
            commit('setUserId', userId);
        },
        setAuthToken({ commit }, authToken) {
            commit('setAuthToken', authToken);
        },
        // 添加 logout Action
        logout({ commit }) {
            commit('setIsLoggedIn', false);
            commit('setUserId', null);
            commit('setAuthToken', null);

            // 清除 localStorage 中的用户信息
            localStorage.removeItem('userId');
            localStorage.removeItem('authToken');
        },
    },
    getters: {
        isLoggedIn: state => state.isLoggedIn,
        userId: state => state.userId,
        authToken: state => state.authToken,
    },
});
