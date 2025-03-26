<template>
  <div id="app">
    <nav class="navbar">
      <div v-if="!isLoggedIn">
        <router-link to="/login">登录</router-link>
        <span>/</span>
        <router-link to="/register">注册</router-link>
      </div>
      <a v-else href="#" @click="logout">退出</a>
      <router-link to="/game">猜数字游戏</router-link>
      <router-link to="/scoreboard">猜测次数最少排行榜</router-link>
    </nav>
    <router-view></router-view>
  </div>
</template>

<script>
import { mapState, mapActions } from "vuex";

export default {
  name: "App",
  computed: {
    ...mapState(["isLoggedIn"]),  // 映射 Vuex 状态到组件的计算属性
  },
  methods: {
    ...mapActions(["logout"]),  // 映射 Vuex actions 到组件方法

    // 退出时清除 Vuex 状态和 localStorage 中的用户信息
    logout() {
      this.$store.dispatch('logout');  // 清除 Vuex 状态
      localStorage.removeItem('userId'); // 清除 localStorage 中的 userId
      localStorage.removeItem('authToken'); // 清除 localStorage 中的 authToken
      document.cookie = "X-User-ID=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT";  // 删除 cookie 中的 X-User-ID
    },
  },
  mounted() {
    // 在组件加载时初始化登录状态
    const storedUserId = localStorage.getItem('userId');
    const storedAuthToken = localStorage.getItem('authToken');

    if (storedUserId && storedAuthToken) {
      this.$store.commit('setIsLoggedIn', true);  // 更新 Vuex 中的登录状态
      this.$store.commit('setUserId', storedUserId);  // 设置 userId
      this.$store.commit('setAuthToken', storedAuthToken);  // 设置 authToken
    } else {
      this.$store.commit('setIsLoggedIn', false);  // 如果没有 userId 或 authToken，设置为未登录
    }
  },
};
</script>
