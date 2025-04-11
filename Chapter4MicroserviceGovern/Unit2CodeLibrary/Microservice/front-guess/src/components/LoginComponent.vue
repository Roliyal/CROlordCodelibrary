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
  </div>
</template>

<script>
import { useRouter } from 'vue-router';
import store from '../store';       // 这里引入了 store
import authApi from '../auth-api';  // 这里引入了登录接口封装

export default {
  data() {
    return {
      username: '',
      password: '',
      errorMessage: '',
      infoMessage: '',
    };
  },
  setup() {
    const router = useRouter();
    return { router };
  },
  methods: {
    async login() {
      try {
        const authResult = await authApi.login(this.username, this.password);

        if (authResult) {
          // 更新 Vuex 和 localStorage
          store.commit('setUserId', authResult.id);
          store.commit('setAuthToken', authResult.authToken);
          store.commit('setIsLoggedIn', true);
          localStorage.setItem('userId', authResult.id);
          localStorage.setItem('authToken', authResult.authToken);

          document.cookie = `X-User-ID=${authResult.id}; path=/;`;
          document.cookie = `x-pre-higress-tag=base; path=/;`;

          this.infoMessage = '登录成功！正在刷新页面以应用配置...';
          setTimeout(() => {
            window.location.reload(); // 刷新后用户状态、cookie、灰度策略全生效
          }, 800);
        } else {
          this.errorMessage = '登录失败，请检查用户名和密码是否正确。';
        }
      } catch (error) {
        console.error('Error during login:', error);
        this.errorMessage = '登录过程中发生错误，请稍后再试。';
      }
    }
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
