<template>
  <div class="container">
    <h1 class="title">Register</h1>
    <div class="register-container">
      <form @submit.prevent="register">
        <div class="input-group">
          <label>用户名：</label>
          <input type="text" v-model="username" />
        </div>
        <div class="input-group">
          <label>密码：</label>
          <input type="password" v-model="password" />
        </div>
        <button type="submit">注册</button>
        <div class="message-container">
          <div v-show="errorMessage" class="error-message">{{ errorMessage }}</div>
          <div v-show="infoMessage" class="info-message">{{ infoMessage }}</div>
        </div>
      </form>
    </div>
    <footer class="footer">
      <p>&copy; 2023 CROlord. All rights reserved.</p>
    </footer>
  </div>
</template>


<script>
import authApi from "../auth-api";
import { useRouter } from "vue-router";

export default {
  data() {
    return {
      username: "",
      password: "",
      errorMessage: "",
      infoMessage: "",
    };
  },
  setup() {
    const router = useRouter();
    return { router };
  },
  methods: {
    async register() {
      const registerResult = await authApi.register(this.username, this.password);

      if (registerResult && registerResult.status === 201) {
        this.errorMessage = "";
        this.infoMessage = "注册成功！正在跳转到登录页面...";
        setTimeout(() => {
          this.router.push("/login");
        }, 2000);
      } else {
        this.errorMessage = "注册失败，请重试。";
      }
    },
  },
};
</script>


<style scoped>
body {
  font-family: Arial, sans-serif;
}

.container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background-color: #f5f5f5;
}

.register-container {
  width: 370px;
  padding: 30px;
  box-shadow: 0 0 8px rgba(0, 0, 0, 0.1);
  border-radius: 10px;
}

.input-group {
  margin-bottom: 15px;
}

label {
  display: block;
  margin-bottom: 5px;
}

input {
  width: 100%;
  padding: 5px;
  border: 1px solid #ccc;
  border-radius: 5px;
}

button {
  width: 100%;
  padding: 8px;
  background-color: #4caf50;
  border: none;
  border-radius: 5px;
  color: white;
  font-weight: bold;
  cursor: pointer;
}

button:hover {
  background-color: #45a049;
}

.message-container {
  height: 20px;
  margin-top: 10px;
  width: 100%;
}

.error-message {
  color: red;
  text-align: center;
}
</style>
