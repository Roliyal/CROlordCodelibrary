import { createRouter, createWebHashHistory } from "vue-router"; // 使用 createWebHashHistory
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
    history: createWebHashHistory(process.env.BASE_URL), // 使用 createWebHashHistory
    routes,
});

router.beforeEach((to, from, next) => {
    // 判断是否已认证，且是否访问非登录/注册页面
    if (
        (to.path !== "/login" && to.path !== "/register") && // 排除登录和注册
        !store.state.isLoggedIn // 使用 store 中的 isLoggedIn 判断是否已登录
    ) {
        next("/login"); // 如果未登录，重定向到登录页
    } else {
        next(); // 否则继续访问目标页面
    }
});

export default router;
