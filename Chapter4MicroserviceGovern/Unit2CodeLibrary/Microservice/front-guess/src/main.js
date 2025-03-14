// src/main.js
import { createApp } from 'vue';
import App from './App.vue';
import router from './router';
import store from './store';  // 确保引入 Vuex store
import './styles.css';  // 引入样式文件

const app = createApp(App);

// 从 localStorage 初始化全局状态
const authToken = localStorage.getItem('authToken');
const storedUserId = localStorage.getItem('userId');

if (authToken) {
    store.commit('setIsLoggedIn', true);   // 通过 mutation 设置已登录状态
    store.commit('setAuthToken', authToken);  // 设置 authToken
}

if (storedUserId) {
    store.commit('setUserId', storedUserId);  // 设置 userId
}

app.use(store).use(router).mount('#app');
