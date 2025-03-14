<template>
  <div id="app">
    <nav class="navbar">
      <!-- 登录和注册链接显示，仅在用户未登录时显示 -->
      <div v-if="!store.state.isLoggedIn">
        <router-link to="/login">登录-this is gra</router-link>
        <span>/</span>
        <router-link to="/register">注册-this is gray</router-link>
      </div>
      <!-- 退出链接显示，仅在用户已登录时显示 -->
      <a v-else href="#" @click="logout">退出</a>
      <!-- 其他路由链接 -->
      <router-link to="/game">猜数字游戏-this is gra</router-link>
      <router-link to="/scoreboard">猜测次数最少排行榜-this is gra</router-link>
    </nav>
    <router-view></router-view>
  </div>
</template>

<script>
import store from "./store";

export default {
  name: "App",
  methods: {
    logout() {
      // 更新登录状态
      store.setIsLoggedIn(false);
      store.setUserId(null); // 清除 userId
      store.setAuthToken(null); // 清除 authToken

      // 清除 localStorage 中存储的用户信息
      localStorage.removeItem('userId');
      localStorage.removeItem('authToken');

      // 跳转到登录页面
      this.$router.push("/login");
    },
  },
};
</script>
