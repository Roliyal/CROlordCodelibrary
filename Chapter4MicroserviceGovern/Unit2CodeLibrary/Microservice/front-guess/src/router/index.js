import { createRouter, createWebHashHistory } from "vue-router"; // 修改为 createWebHashHistory
import LoginComponent from "../components/LoginComponent.vue";
import GuessNumberComponent from "../components/GuessNumberComponent.vue";
import ScoreboardComponent from "../components/ScoreboardComponent.vue";
import authApi from "../auth-api";
import RegisterComponent from "../components/RegisterComponent.vue";

const routes = [
    { path: "/login", component: LoginComponent },
    { path: "/register", component: RegisterComponent },
    { path: "/game", component: GuessNumberComponent },
    { path: "/scoreboard", component: ScoreboardComponent },
];

const router = createRouter({
    history: createWebHashHistory(process.env.BASE_URL), // 改为 createWebHashHistory
    routes,
});

router.beforeEach((to, from, next) => {
    if (
        (to.path !== "/login" && to.path !== "/register") && // 排除登录和注册
        !authApi.isAuthenticated // 未认证
    ) {
        next("/login"); // 重定向到登录页
    } else {
        next(); // 继续其他页面
    }
});

export default router;
