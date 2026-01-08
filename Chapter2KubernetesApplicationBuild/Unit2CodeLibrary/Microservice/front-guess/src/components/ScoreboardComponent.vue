<template>
  <div class="container scoreboard-container">
    <h2>排行榜展示</h2>
    <p v-if="!dataFetched">在这里查看您的排行！</p>
    <button @click="fetchScoreboardData" v-if="!dataFetched">获取排行信息</button>
    <table class="scoreboard-table" v-if="dataFetched">
      <thead>
      <tr>
        <th>ID</th>
        <th>Attempts</th>
        <th>Target Number</th>
      </tr>
      </thead>
      <tbody>
      <tr v-for="(game, index) in gameData" :key="index">
        <td>{{ game.id }}</td>
        <td>{{ game.attempts }}</td>
        <td>{{ game.target_number }}</td>
      </tr>
      </tbody>
    </table>
  </div>
</template>


<script>
import config from "../config.js";

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
        const response = await fetch(`${config.scoreboardURL}/scoreboard`);
        if (response.ok) {
          this.gameData = await response.json();
          this.dataFetched = true;
        } else {
          console.error("Error fetching scoreboard data:", response.statusText);
        }
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
