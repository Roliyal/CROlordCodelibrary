import { reactive } from "vue";

const state = reactive({
    isLoggedIn: false,
});

const setIsLoggedIn = (isLoggedIn) => {
    state.isLoggedIn = isLoggedIn;
};

export default {
    state,
    setIsLoggedIn, // Add this line
};
