<template>
  <div id="app">
    <nav class="navbar">
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
