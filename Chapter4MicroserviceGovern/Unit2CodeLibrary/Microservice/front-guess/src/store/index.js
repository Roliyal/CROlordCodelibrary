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

            // 清除所有存储
            localStorage.removeItem('userId');
            localStorage.removeItem('authToken');
            document.cookie = "X-User-ID=; path=/; domain=.roliyal.com; expires=Thu, 01 Jan 1970 00:00:00 GMT";
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