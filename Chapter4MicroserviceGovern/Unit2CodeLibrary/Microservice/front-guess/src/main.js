// src/main.js
import { createApp } from 'vue';
import App from './App.vue';
import store from './store';
import router from './router';
import './styles.css';



// 继续从 localStorage 中读取
const storedUserId = localStorage.getItem('userId');
const storedAuthToken = localStorage.getItem('authToken');

if (storedUserId && storedAuthToken) {
    store.commit('setUserId', storedUserId);
    store.commit('setAuthToken', storedAuthToken);
    store.commit('setIsLoggedIn', true);
} else {
    store.commit('setIsLoggedIn', false);
}

createApp(App).use(store).use(router).mount('#app');
