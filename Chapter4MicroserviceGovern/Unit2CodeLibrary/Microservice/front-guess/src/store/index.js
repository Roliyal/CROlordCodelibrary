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
        // 退出时清除状态
        logout(state) {
            state.isLoggedIn = false;
            state.userId = null;
            state.authToken = null;

            // 清除 localStorage 中的用户信息
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
    },
});
