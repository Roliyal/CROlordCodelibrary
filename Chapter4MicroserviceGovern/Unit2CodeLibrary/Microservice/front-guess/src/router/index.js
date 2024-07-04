import { createRouter, createWebHistory } from "vue-router";
import LoginComponent from "../components/LoginComponent.vue";
import GuessNumberComponent from "../components/GuessNumberComponent.vue";
import ScoreboardComponent from "../components/ScoreboardComponent.vue";
import authApi from "../auth-api"; // 更新这里
import RegisterComponent from "../components/RegisterComponent.vue"; // 导入 Register 组件


const routes = [
    { path: "/login", component: LoginComponent },
    { path: "/register", component: RegisterComponent }, // 添加新的路由
    { path: "/game", component: GuessNumberComponent },
    { path: "/scoreboard", component: ScoreboardComponent },

];

const router = createRouter({
    history: createWebHistory(process.env.BASE_URL),
    routes,
});

router.beforeEach((to, from, next) => {
    if (to.path !== "/login" && to.path !== "/register" && !authApi.isAuthenticated) {
        next("/login");
    } else {
        next();
    }
});

export default router;
