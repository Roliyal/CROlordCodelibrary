<template>
  <div>
    <h2>登录</h2>
    <input v-model="username" placeholder="用户名" />
    <button @click="login">登录</button>
  </div>
</template>

<script>
import axios from "axios";

export default {
  data() {
    return {
      username: ""
    };
  },
  methods: {
    async login() {
      try {
        const response = await axios.post("https://your-api-gateway/login", {
          username: this.username
        });
        this.$emit("loggedIn", response.data);
      } catch (error) {
        console.error("登录失败", error);
      }
    }
  }
};
</script>
