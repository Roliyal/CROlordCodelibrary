<template>
  <div id="app">
    <div v-if="!loggedIn">
      <Login @login="onLogin" />
    </div>
    <div v-else>
      <GuessGame />
    </div>
  </div>
</template>

<script>
import Login from "./LogInGame.vue";
import GuessGame from "./GuessGame.vue";

export default {
  components: {
    Login,
    GuessGame,
  },
  data() {
    return {
      loggedIn: false,
    };
  },
  methods: {
    onLogin(credentials) {
      // 这里可以添加实际的登录验证逻辑
      // 暂时只检查用户名和密码是否相同
      if (credentials.username === credentials.password) {
        this.loggedIn = true;
      } else {
        alert("登录失败，请检查用户名和密码");
      }
    },
  },
};
</script>
