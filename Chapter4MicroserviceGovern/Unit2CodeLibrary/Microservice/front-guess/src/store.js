// src/store.js
import { reactive } from "vue";

const state = reactive({
    isLoggedIn: false,
    userId: null, // 全局用户 ID
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
    setUserId,
};
