// src/router/index.js
import { createRouter, createWebHashHistory } from "vue-router";
import LoginComponent from "../components/LoginComponent.vue";
import GuessNumberComponent from "../components/GuessNumberComponent.vue";
import ScoreboardComponent from "../components/ScoreboardComponent.vue";
import RegisterComponent from "../components/RegisterComponent.vue";
import store from "../store";  // 引入 store.js

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
    if (
        (to.path !== "/login" && to.path !== "/register") && // 排除登录和注册
        !store.state.isLoggedIn // 判断是否已登录
    ) {
        next("/login"); // 未登录重定向到登录页
    } else {
        next(); // 继续访问目标页面
    }
});

export default router;
