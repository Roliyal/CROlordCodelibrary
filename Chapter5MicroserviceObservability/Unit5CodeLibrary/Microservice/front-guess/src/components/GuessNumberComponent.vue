// src/components/GameComponent.vue
<template>
  <div class="container game-container">
    <h2>猜数字游戏</h2>
    <div v-if="gameStatus === 'idle'">
      <button @click="startGame">开始游戏</button>
    </div>
    <div v-else-if="gameStatus === 'loading'">
      <p>加载中...</p>
    </div>
    <div v-else-if="gameStatus === 'playing'">
      <p>猜一个 1 到 100 之间的数字：</p>
      <input v-model.number="number" type="number" min="1" max="100" />
      <button @click="submitGuess(number)">提交</button>
      <p v-if="message">{{ message }}</p>
      <p v-if="attempts">尝试次数：{{ attempts }}</p>
    </div>
    <div v-else-if="gameStatus === 'error'">
      <p>发生错误，请重试。</p>
      <button @click="startGame">重试</button>
    </div>
  </div>
</template>

<script>
import axiosInstance from '../axiosInstance';  // 确保导入 axiosInstance

import { mapState } from "vuex";  // 引入 mapState 用于映射 state

export default {
  data() {
    return {
      number: null,
      message: null,
      attempts: null,
      gameStatus: "idle",
    };
  },
  computed: {
    ...mapState(["traceId"]),  // 从 Vuex 获取 traceId
  },
  methods: {
    async startGame() {
      this.gameStatus = "playing";
      this.message = null;
      this.attempts = null;
    },

    async submitGuess(guess) {
      try {
        // 提交猜测
        const response = await axiosInstance.post('/game', { number: guess });
        const resData = response.data;
        if (resData.success) {
          alert(resData.message);
          this.message = resData.message;
          this.attempts = resData.attempts;
          this.gameStatus = "idle";
        } else {
          this.message = resData.message;
          this.attempts = resData.attempts;
        }
      } catch (error) {
        console.error("Error submitting guess:", error);
        this.gameStatus = "error";
      }
    },
  },
};
</script>



<style scoped>
.container {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-top: 50px;
}

.game-container {
  max-width: 600px;
  padding: 40px;
  box-shadow: 0 0 10px rgba(0, 0, 0, 0.2);
  border-radius: 12px;
}

h2 {
  margin-bottom: 30px;
  font-size: 32px;
  color: #4a4a4a;
}

p {
  font-size: 18px;
  color: #4a4a4a;
  margin-bottom: 15px;
}

input[type="number"] {
  width: 100%;
  padding: 8px 12px;
  font-size: 18px;
  border: 1px solid #ccc;
  border-radius: 4px;
  box-sizing: border-box;
  margin-bottom: 20px;
}

button {
  padding: 12px 24px;
  background-color: #4caf50;
  border: none;
  border-radius: 6px;
  color: white;
  font-weight: bold;
  font-size: 18px;
  cursor: pointer;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.15);
  transition: background-color 0.2s ease;
}

button:hover {
  background-color: #45a049;
}

button:active {
  box-shadow: none;
  transform: translateY(1px);
}
</style>
