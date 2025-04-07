// src/main.js
import { createApp } from 'vue';
import App from './App.vue';
import store from './store';
import router from './router';

// 1. 在应用启动时，把 localStorage 的值恢复到 Vuex
const storedUserId = localStorage.getItem('userId');
const storedAuthToken = localStorage.getItem('authToken');

if (storedUserId && storedAuthToken) {
    store.commit('setUserId', storedUserId);
    store.commit('setAuthToken', storedAuthToken);
    store.commit('setIsLoggedIn', true);
} else {
    store.commit('setIsLoggedIn', false);
}

// 2. 创建应用并挂载
createApp(App).use(store).use(router).mount('#app');
