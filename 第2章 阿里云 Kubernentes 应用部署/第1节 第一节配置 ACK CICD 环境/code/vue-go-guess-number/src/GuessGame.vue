<template>
  <div>
    <h1>猜数字游戏</h1>
    <p>请输入一个数字（1-100）：</p>
    <input type="number" v-model="userGuess" @input="checkGuess" />
    <p>{{ message }}</p>
  </div>
</template>

<script>
import axios from "axios";

export default {
  data() {
    return {
      userGuess: null,
      message: "",
    };
  },
  methods: {
    async checkGuess() {
      if (this.userGuess < 1 || this.userGuess > 100) {
        this.message = "请输入一个 1-100 之间的数字";
        return;
      }
      const response = await axios.post("http://localhost:8081/check-guess", {
        guess: this.userGuess,
      });
      this.message = response.data.message;
    },
  },
};
</script>
