<template>
  <div>
    <h1>登录</h1>
    <label for="username">用户名:</label>
    <input id="username" type="text" v-model="username" />
    <label for="password">密码:</label>
    <input id="password" type="password" v-model="password" />
    <button @click="login">登录</button>
  </div>
</template>

<script>
export default {
  data() {
    return {
      username: "",
      password: "",
    };
  },
  methods: {
    login() {
      if (this.username && this.password) {
        this.$emit("login", { username: this.username, password: this.password });
      } else {
        alert("请填写用户名和密码");
      }
    },
  },
};
</script>
