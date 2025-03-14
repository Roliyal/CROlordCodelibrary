<template>
  <div id="app">
    <nav class="navbar">
      <!-- 判断是否已登录 -->
      <div v-if="!store.state.isLoggedIn">
        <router-link to="/login">登录-this is gra</router-link>
        <span>/</span>
        <router-link to="/register">注册-this is gray</router-link>
      </div>
      <a v-else href="#" @click="logout">退出</a>
      <router-link to="/game">猜数字游戏-this is gra</router-link>
      <router-link to="/scoreboard">猜测次数最少排行榜-this is gra</router-link>
    </nav>
    <router-view></router-view>
  </div>
</template>

<script>
import store from "./store";  // 确保正确导入 store.js

export default {
  name: "App",
  methods: {
    logout() {
      store.setIsLoggedIn(false); // 更新登录状态
      store.setUserId(null); // 清除用户 ID
      store.setAuthToken(null); // 清除认证 token

      // 清除 localStorage 中的数据
      localStorage.removeItem('userId');
      localStorage.removeItem('authToken');

      // 跳转到登录页面
      this.$router.push("/login");
    },
  },
};
</script>
