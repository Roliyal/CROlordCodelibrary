// src/store.js
import { reactive } from "vue";

const state = reactive({
    isLoggedIn: false,
    userId: null, // 添加 userId
});

const setIsLoggedIn = (isLoggedIn) => {
    state.isLoggedIn = isLoggedIn;
};

const setUserId = (userId) => {
    state.userId = userId;
};

export default {
    state,
    setIsLoggedIn,
    setUserId, // 导出 setUserId 方法
};
