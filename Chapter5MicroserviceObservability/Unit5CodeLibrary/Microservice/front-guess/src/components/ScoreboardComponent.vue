<!-- src/components/ScoreboardComponent.vue -->
<template>
  <div class="container scoreboard-container">
    <h2>猜测次数最少排行榜</h2>
    <p v-if="!dataFetched">在这里查看您的排行！</p>
    <button @click="fetchScoreboardData" v-if="!dataFetched">获取排行信息</button>
    <table class="scoreboard-table" v-if="dataFetched">
      <thead>
      <tr>
        <th>ID</th>
        <th>Username</th>
        <th>Attempts</th>
        <th>Target Number</th>
      </tr>
      </thead>
      <tbody>
      <tr v-for="(game, index) in gameData" :key="index">
        <td>{{ game.id }}</td>
        <td>{{ game.username }}</td>
        <td>{{ game.attempts }}</td>
        <td>{{ game.target_number }}</td>
      </tr>
      </tbody>
    </table>
  </div>
</template>

<script>
import axiosInstance from "../axiosInstance"; // 使用 axiosInstance
import config from "../config.js";            // 你定义的配置文件

export default {
  data() {
    return {
      scores: [],
      gameData: [],
      dataFetched: false,
    };
  },
  methods: {
    async fetchScoreboardData() {
      try {
        const response = await axiosInstance.get(`${config.scoreboardURL}/scoreboard`);
        this.gameData = response.data;
        this.dataFetched = true;
      } catch (error) {
        console.error("Error fetching scoreboard data:", error);
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
}

.scoreboard-container {
  max-width: 600px;
  padding: 30px;
  box-shadow: 0 0 8px rgba(0, 0, 0, 0.1);
  border-radius: 10px;
  margin-top: 20px;
}

h2 {
  margin-bottom: 20px;
}

button {
  padding: 10px 20px;
  background-color: #4caf50;
  border: none;
  border-radius: 5px;
  color: white;
  font-weight: bold;
  cursor: pointer;
  margin-bottom: 10px;
}

button:hover {
  background-color: #45a049;
}

.scoreboard-table {
  border-collapse: collapse;
  width: 100%;
}

.scoreboard-table th,
.scoreboard-table td {
  border: 1px solid #ddd;
  padding: 8px;
  text-align: center;
}

.scoreboard-table th {
  padding-top: 12px;
  padding-bottom: 12px;
  background-color: #4caf50;
  color: white;
}

.scoreboard-table tr:nth-child(even) {
  background-color: #f2f2f2;
}

.scoreboard-table tr:hover {
  background-color: #ddd;
}
</style>
