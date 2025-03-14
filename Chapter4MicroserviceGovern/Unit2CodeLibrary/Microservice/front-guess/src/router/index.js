// src/router/index.js
import { createRouter, createWebHashHistory } from "vue-router";
import LoginComponent from "../components/LoginComponent.vue";
import GuessNumberComponent from "../components/GuessNumberComponent.vue";
import ScoreboardComponent from "../components/ScoreboardComponent.vue";
import RegisterComponent from "../components/RegisterComponent.vue";
import { mapState } from "vuex";

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
    const store = mapState(["isLoggedIn"]);
    if (
        (to.path !== "/login" && to.path !== "/register") && // 排除登录和注册
        !store.isLoggedIn  // 使用 Vuex 判断登录状态
    ) {
        next("/login"); // 如果未登录，重定向到登录页
    } else {
        next(); // 继续其他页面
    }
});

export default router;
