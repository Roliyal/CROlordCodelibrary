import { createApp } from 'vue';
import App from './App.vue';
import store from './store';
import router from './router';
import './styles.css';
// 引入 ARMS SDK 和配置文件
import ArmsRum from '@arms/rum-browser';
import { armsConfig } from './config/armsConfig'; // 引入配置文件

ArmsRum.init(armsConfig);
ArmsRum.setConfig('debug', true);

// 同步用户状态
const storedUserId = localStorage.getItem('userId');
const storedAuthToken = localStorage.getItem('authToken');
const justLoggedIn = localStorage.getItem('justLoggedIn');

if (storedUserId && storedAuthToken) {
    store.commit('setUserId', storedUserId);
    store.commit('setAuthToken', storedAuthToken);
    store.commit('setIsLoggedIn', true);
} else {
    store.commit('setIsLoggedIn', false);
}

if (justLoggedIn === 'true') {
    localStorage.removeItem('justLoggedIn');
    window.location.href = '#/game';
}

// 挂载应用/
createApp(App).use(store).use(router).mount('#app');
