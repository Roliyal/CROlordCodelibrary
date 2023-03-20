<template>
  <div>
    <h2>Login</h2>
    <form>
      <label for="username">Username:</label>
      <input type="text" id="username" v-model="username">
      <br>
      <label for="password">Password:</label>
      <input type="password" id="password" v-model="password">
      <br>
      <button type="submit" @click.prevent="login()">Login</button>
    </form>
  </div>
</template>

<script>
export default {
  data() {
    return {
      username: '',
      password: ''
    }
  },
  methods: {
    login() {
      // 使用axios向后端发送登录请求
      axios.post('/api/login', {
        username: this.username,
        password: this.password
      }).then(response => {
        // 登录成功，将JWT token保存到本地
        localStorage.setItem('token', response.data.token)
        // 跳转到游戏页面
        this.$router.push('/game')
      }).catch(error => {
        // 登录失败，提示错误信息
        alert('Login failed: ' + error.response.data.error)
      })
    }
  }
}
</script>
