<template>
  <div class="container">
    <h1 class="title">Login</h1>
    <div class="login-container">
      <form @submit.prevent="login">
        <div class="input-group">
          <label>用户名：</label>
          <input type="text" v-model="username" required />
        </div>
        <div class="input-group">
          <label>密码：</label>
          <input type="password" v-model="password" required />
        </div>
        <button type="submit">登录</button>
        <div class="message-container">
          <div v-if="errorMessage" class="error-message">{{ errorMessage }}</div>
          <div v-if="infoMessage" class="info-message">{{ infoMessage }}</div>
        </div>
      </form>
    </div>
    <footer class="footer">
      <p>&copy; 2023 CROlord. All rights reserved.</p>
    </footer>
  </div>
</template>

<script>
import { useRouter } from "vue-router";
import store from "../store";  // 引入 Vuex store

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
    async login() {
      try {
        const authResult = await store.dispatch("login", {
          username: this.username,
          password: this.password
        });

        if (authResult) {
          // 登录成功，将 userId 和 authToken 存储到 localStorage 和 Vuex
          localStorage.setItem("userId", authResult.id);
          localStorage.setItem("authToken", authResult.authToken);
          store.commit("setUserId", authResult.id);
          store.commit("setAuthToken", authResult.authToken);

          this.infoMessage = "登录成功！正在跳转...";
          setTimeout(() => {
            this.router.push("/game"); // 跳转到游戏页面
          }, 1000);
        } else {
          this.errorMessage = "登录失败，请检查用户名和密码是否正确。";
        }
      } catch (error) {
        console.error("Error during login:", error);
        this.errorMessage = "登录过程中发生错误，请稍后再试。";
      }
    },
  },
};
</script>



<style scoped>
.container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background-color: #f5f5f5;
}

.login-container {
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

.info-message {
  color: green;
  text-align: center;
}
</style>
