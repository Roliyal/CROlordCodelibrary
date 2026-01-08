import { createApp } from 'vue';
import App from './App.vue';
import store from './store';
import router from './router';
import './styles.css';

// 引入 ARMS SDK 和配置文件工厂函数
import ArmsRum from '@arms/rum-browser';
import { createArmsConfig } from './config/armsConfig';  // 引入工厂函数

// 获取 userId
const userId = store.state.userId || localStorage.getItem('userId');

// 创建 ARMS SDK 配置
const armsConfig = createArmsConfig(userId);

// 初始化 ARMS SDK
ArmsRum.init(armsConfig);

// 启用调试模式
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

// 挂载应用
createApp(App).use(store).use(router).mount('#app');
