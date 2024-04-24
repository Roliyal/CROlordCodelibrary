<template>
  <div class="container">
    <div class="row mt-5">
      <div class="col-md-6 offset-md-3">
        <div class="card">
          <div class="card-body">
            <h1 class="card-title text-center">猜数字游戏</h1>
            <p class="text-center">请输入一个数字（1-100）：</p>
            <div class="input-group mb-3">
              <input type="number" class="form-control" v-model="userGuess" />
              <button class="btn btn-primary" @click="checkGuess">提交</button>
            </div>
            <p class="text-center">{{ message }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from "axios";
import config from "@/config";


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
      try {
        const response = await axios.post(config.apiUrl, {
          guess: this.userGuess,
        });
        this.message = response.data.message;
      } catch (error) {
        console.error(error);
        this.message = "连接后端服务时发生错误，请稍后重试。";

    },
  },
};
</script>
