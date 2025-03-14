// src/store.js
import { reactive } from "vue";

// 使用 reactive 创建响应式的 state
const state = reactive({
    isLoggedIn: false,   // 默认未登录
    userId: null,        // 用户 ID
    authToken: null,     // 用户认证 token
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
    state,             // 导出响应式的 state
    setIsLoggedIn,
    setUserId,
    setAuthToken,
};
