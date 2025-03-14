// src/main.js
import { createApp } from 'vue';
import App from './App.vue';
import router from './router';
import store from './store';  // 引入 Vuex store
import './styles.css';  // 引入样式文件

const app = createApp(App);

// 从 localStorage 初始化全局状态
const authToken = localStorage.getItem('authToken');
const storedUserId = localStorage.getItem('userId');

if (authToken && storedUserId) {
    store.commit('setIsLoggedIn', true);  // 只有在获取到有效的 token 和 userId 时，才设置为已登录
    store.commit('setAuthToken', authToken);
    store.commit('setUserId', storedUserId);
} else {
    store.commit('setIsLoggedIn', false);  // 如果没有有效的 token 和 userId，确保用户处于未登录状态
}

app.use(store).use(router).mount('#app');
