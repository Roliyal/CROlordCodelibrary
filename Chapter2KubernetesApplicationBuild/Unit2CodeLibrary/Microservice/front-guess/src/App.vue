<template>
  <div id="app">
    <nav class="navbar">
      <router-link v-if="!store.state.isLoggedIn" to="/login">登录</router-link>
      <a v-else href="#" @click="logout">退出</a>
      <router-link to="/game">猜数字游戏</router-link>
      <router-link to="/scoreboard">战绩展示</router-link>
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
