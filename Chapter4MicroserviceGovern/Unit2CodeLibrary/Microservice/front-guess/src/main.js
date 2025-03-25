// src/main.js
import { createApp } from 'vue';
import App from './App.vue';
import store from './store';  // 引入 Vuex store
import router from './router'; // 引入路由
import './styles.css';  // 引入样式文件

// 从 localStorage 初始化全局状态
const authToken = localStorage.getItem('authToken');
const storedUserId = localStorage.getItem('userId');

// 确保 Vuex 状态初始化
if (authToken && storedUserId) {
    store.commit('setIsLoggedIn', true);  // 登录状态
    store.commit('setAuthToken', authToken);  // 更新 authToken
    store.commit('setUserId', storedUserId);  // 更新 userId
} else {
    store.commit('setIsLoggedIn', false);  // 如果没有登录信息，设置为未登录
}

// 创建应用实例
const app = createApp(App);

// 使用 store 和 router
app.use(store).use(router).mount('#app');
