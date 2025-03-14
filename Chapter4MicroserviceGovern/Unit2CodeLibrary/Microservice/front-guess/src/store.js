// src/store.js
import { reactive } from "vue";

// 使用 reactive 创建响应式状态
const state = reactive({
    isLoggedIn: false,
    userId: null,
    authToken: null,
});

// 设置状态的更新方法
const setIsLoggedIn = (isLoggedIn) => {
    state.isLoggedIn = isLoggedIn;
};

const setUserId = (userId) => {
    state.userId = userId;
};

const setAuthToken = (authToken) => {
    state.authToken = authToken;
};

export default {
    state,
    setIsLoggedIn,
    setUserId,
    setAuthToken,
};
