// src/main.js
import { createApp } from 'vue';
import App from './App.vue';
import store from './store';
import router from './router';

// ---- 1) 如果发现本地没有存过 userId，就强行写入一个 FAKE 值
let storedUserId = localStorage.getItem('userId');
let storedAuthToken = localStorage.getItem('authToken');

if (!storedUserId || !storedAuthToken) {
    // 强行写一个假的 userId/token
    localStorage.setItem('userId', 'CROLORD_USER_001');
    localStorage.setItem('authToken', 'CROLORD_TOKEN_issac');
}

// 再读一次
storedUserId = localStorage.getItem('userId');
storedAuthToken = localStorage.getItem('authToken');

// ---- 2) 把它们写入 Vuex
if (storedUserId && storedAuthToken) {
    store.commit('setUserId', storedUserId);
    store.commit('setAuthToken', storedAuthToken);
    store.commit('setIsLoggedIn', true);
} else {
    store.commit('setIsLoggedIn', false);
}

// ---- 3) 创建应用并挂载
createApp(App).use(store).use(router).mount('#app');
