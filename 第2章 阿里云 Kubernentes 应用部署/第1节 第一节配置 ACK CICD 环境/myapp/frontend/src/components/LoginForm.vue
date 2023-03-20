<template>
  <div class="login-form">
    <h2>Login Form</h2>
    <form @submit.prevent="login">
      <div>
        <label for="username">Username:</label>
        <input type="text" id="username" v-model="username" required>
      </div>
      <div>
        <label for="password">Password:</label>
        <input type="password" id="password" v-model="password" required>
      </div>
      <button type="submit">Login</button>
    </form>
  </div>
</template>

<script>
import AuthService from "../services/auth";

export default {
  name: "LoginForm",
  data() {
    return {
      username: "",
      password: "",
    };
  },
  methods: {
    async login() {
      try {
        await AuthService.login(this.username, this.password);
        this.$router.push("/game");
      } catch (error) {
        console.log(error);
      }
    },
  },
};
</script>

<style scoped>
.login-form {
  display: flex;
  flex-direction: column;
  align-items: center;
}

form {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-top: 20px;
}

div {
  margin: 10px 0;
}

input {
  padding: 5px;
  margin-left: 5px;
}

button {
  padding: 10px;
  margin-top: 20px;
  border: none;
  border-radius: 5px;
  background-color: #3f51b5;
  color: white;
  cursor: pointer;
}

button:hover {
  background-color: #2c3e50;
}
</style>
