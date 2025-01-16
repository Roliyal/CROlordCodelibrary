// src/store.js
import { reactive } from "vue";

const state = reactive({
    isLoggedIn: false,
    userId: null, // 全局用户 ID
    authToken: null, // 全局 authToken
});

const setIsLoggedIn = (isLoggedIn) => {
    state.isLoggedIn = isLoggedIn;
};

const setUserId = (userId) => {
    state.userId = userId;
};

const setAuthToken = (authToken) => { // 新增方法
    state.authToken = authToken;
};

export default {
    state,
    setIsLoggedIn,
    setUserId,
    setAuthToken, // 导出新增方法
};
