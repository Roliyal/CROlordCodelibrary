<template>
  <div id="app">
    <nav class="navbar">
      <div v-if="!store.state.isLoggedIn">
        <router-link to="/login">登录</router-link>
        <span>/</span>
        <router-link to="/register">注册</router-link>
      </div>
      <a v-else href="#" @click="logout">退出</a>
      <router-link to="/game">猜数字游戏</router-link>
      <router-link to="/scoreboard">排行榜展示</router-link>
    </nav>
    <router-view></router-view>
  </div>
</template>
<script>
import store from "./store";

export default {
  name: "App",
  computed: {
    store() {
      return store;
    },
  },
  methods: {
    logout() {
      store.setIsLoggedIn(false); // Update this line
      this.$router.push("/login");
    },
  },
};
</script>
