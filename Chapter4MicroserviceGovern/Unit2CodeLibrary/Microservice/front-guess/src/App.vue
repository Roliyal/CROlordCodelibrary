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
  },
};
</script>

