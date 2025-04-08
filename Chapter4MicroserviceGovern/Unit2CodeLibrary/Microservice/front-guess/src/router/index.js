// src/router/index.js
import { createRouter, createWebHashHistory } from 'vue-router';
import LoginComponent from '../components/LoginComponent.vue';
import GuessNumberComponent from '../components/GuessNumberComponent.vue';
import ScoreboardComponent from '../components/ScoreboardComponent.vue';
import RegisterComponent from '../components/RegisterComponent.vue';
import store from '../store';

const routes = [
    { path: '/login', component: LoginComponent },
    { path: '/register', component: RegisterComponent },
    { path: '/game', component: GuessNumberComponent },
    { path: '/scoreboard', component: ScoreboardComponent },
];

const router = createRouter({
    history: createWebHashHistory(process.env.BASE_URL),
    routes,
});

// 导航守卫：如果未登录则跳转到 /login
router.beforeEach((to, from, next) => {
    if (
        to.path !== '/login' &&
        to.path !== '/register' &&
        !store.state.isLoggedIn
    ) {
        next('/login');
    } else {
        next();
    }
});

export default router;
