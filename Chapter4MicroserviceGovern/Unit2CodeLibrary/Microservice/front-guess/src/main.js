// src/main.js
import { createApp } from "vue";
import App from "./App.vue";
import router from "./router";
import store from "./store";  // 确保正确导入 store.js

const app = createApp(App);

// 从 localStorage 初始化全局状态
const authToken = localStorage.getItem("authToken");
const storedUserId = localStorage.getItem("userId");

if (authToken) {
    store.setIsLoggedIn(true);   // 设置为已登录
    store.setAuthToken(authToken);  // 设置 authToken
}

if (storedUserId) {
    store.setUserId(storedUserId);  // 设置 userId
}

app.use(store).use(router).mount("#app");
