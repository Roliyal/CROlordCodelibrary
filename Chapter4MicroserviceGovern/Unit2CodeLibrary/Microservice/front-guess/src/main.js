// src/main.js
import { createApp } from "vue";
import App from "./App.vue";
import router from "./router";
import './styles.css';
import store from "./store";

const app = createApp(App);

// 从 localStorage 初始化全局状态
const authToken = localStorage.getItem("authToken");
const storedUserId = localStorage.getItem("id");

if (authToken) {
    store.setIsLoggedIn(true);
}

if (storedUserId) {
    store.setUserId(storedUserId);
}

app.use(store).use(router).mount("#app");
