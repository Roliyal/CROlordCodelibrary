// src/router/index.js
import { createRouter, createWebHashHistory } from "vue-router";
import LoginComponent from "../components/LoginComponent.vue";
import GuessNumberComponent from "../components/GuessNumberComponent.vue";
import ScoreboardComponent from "../components/ScoreboardComponent.vue";
import RegisterComponent from "../components/RegisterComponent.vue";
import store from "../store";  // 引入 Vuex store

const routes = [
    { path: "/login", component: LoginComponent },
    { path: "/register", component: RegisterComponent },
    { path: "/game", component: GuessNumberComponent },
    { path: "/scoreboard", component: ScoreboardComponent },
];

const router = createRouter({
    history: createWebHashHistory(process.env.BASE_URL),
    routes,
});

router.beforeEach((to, from, next) => {
    // 判断是否是登录或注册页面，或者是否已经登录
    if (
        (to.path !== "/login" && to.path !== "/register") &&  // 排除登录和注册页面
        !store.state.isLoggedIn  // 未登录时
    ) {
        next("/login");  // 强制跳转到登录页
    } else {
        next();  // 继续访问目标页面
    }
});

export default router;import { createStore } from 'vuex';

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