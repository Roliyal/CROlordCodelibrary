<!-- src/App.vue -->
<template>
  <div id="app">
    <nav class="navbar">
      <div v-if="!isLoggedIn">
        <router-link to="/login">登录</router-link>
        <span>/</span>
        <router-link to="/register">注册</router-link>
      </div>
      <a v-else href="#" @click="logout">退出</a>
      <router-link to="/game">猜数字游戏-this-gary</router-link>
      <router-link to="/scoreboard">猜测次数最少排行榜-this-gary</router-link>
    </nav>
    <router-view></router-view>
  </div>
</template>

<script>
import { mapState, mapActions } from "vuex";

export default {
  name: "App",
  computed: {
    ...mapState(["isLoggedIn"]),
  },
  methods: {
    ...mapActions(["logout"]),
    logout() {
      this.$store.dispatch('logout');
      localStorage.removeItem('userId');
      localStorage.removeItem('authToken');
      // 删除旧 X-User-ID Cookie
      document.cookie = "X-User-ID=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT";
    },
  },
  mounted() {
    const storedUserId = localStorage.getItem('userId');
    const storedAuthToken = localStorage.getItem('authToken');
    if (storedUserId && storedAuthToken) {
      this.$store.commit('setIsLoggedIn', true);
      this.$store.commit('setUserId', storedUserId);
      this.$store.commit('setAuthToken', storedAuthToken);
    } else {
      this.$store.commit('setIsLoggedIn', false);
    }
  },
};
</script>
