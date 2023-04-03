export default {
    get isAuthenticated() {
        return !!localStorage.getItem("isLoggedIn");
    },

    async authenticate(username, password) {
        // 模拟验证，如果用户名和密码都不为空，则允许登录
        if (username && password) {
            this.isAuthenticated = true;
            localStorage.setItem("isLoggedIn", "true"); // 添加这一行
            return true;
        } else {
            this.isAuthenticated = false;
            localStorage.removeItem("isLoggedIn"); // 添加这一行
            return false;
        }
    },
};
