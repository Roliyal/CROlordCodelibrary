<template>
  <div class="login-form">
    <h1>Login</h1>

    <form @submit.prevent="login">
      <label>
        Username:
        <input type="text" v-model="user.username" />
      </label>

      <label>
        Password:
        <input type="password" v-model="user.password" />
      </label>

      <button type="submit">Login</button>
    </form>
  </div>
</template>

<script>
import AuthService from '@/services/auth.js'

export default {
  name: 'LoginForm',

  data() {
    return {
      user: {
        username: '',
        password: ''
      }
    }
  },

  methods: {
    login() {
      AuthService.login(this.user)
        .then(() => {
          this.$router.push('/game')
        })
        .catch(error => {
          console.log(error)
        })
    }
  }
}
</script>
