import { createApp } from "vue";
import App from "./App.vue";
import router from "./router";
import './styles.css';
import store from "./store";
import axios from 'axios';

axios.defaults.baseURL = 'http://localhost:8080';


createApp(App)
    .use(store)
    .use(router)
    .mount("#app");

const authToken = localStorage.getItem("authToken");// 获取存储的authToken

axios.defaults.headers.common["Authorization"] = `Bearer ${authToken}`; // 设置全局默认请求头
