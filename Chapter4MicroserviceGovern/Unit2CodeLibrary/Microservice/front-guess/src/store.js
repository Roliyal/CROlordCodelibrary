// src/store.js
import { reactive } from "vue";

// 使用 reactive 创建响应式的全局状态
const state = reactive({
    isLoggedIn: false,   // 是否已登录
    userId: null,        // 用户 ID
    authToken: null,     // 用户认证 token
});

// 设置登录状态
const setIsLoggedIn = (isLoggedIn) => {
    state.isLoggedIn = isLoggedIn;
};

// 设置用户 ID
const setUserId = (userId) => {
    state.userId = userId;
};

// 设置 authToken
const setAuthToken = (authToken) => {
    state.authToken = authToken;
};

export default {
    state,
    setIsLoggedIn,
    setUserId,
    setAuthToken,
};
