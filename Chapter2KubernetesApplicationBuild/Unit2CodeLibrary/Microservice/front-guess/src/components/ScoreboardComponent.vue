<template>
  <div class="container scoreboard-container">
    <h2>战绩展示</h2>
    <p>在这里查看您的战绩！</p>
    <!-- 展示战绩的具体实现 -->
    <button @click="fetchScoreboardData">获取战绩</button>
    <ul class="scoreboard-list">
      <li v-for="(score, index) in scores" :key="index">
        {{ score.username }} - {{ score.score }}
      </li>
    </ul>
  </div>
</template>

<script>
export default {
  data() {
    return {
      scores: [],
    };
  },
  methods: {
    async fetchScoreboardData() {
      try {
        const response = await fetch("http://localhost:8085/scoreboard");
        if (response.ok) {
          this.scores = await response.json();
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
}

button:hover {
  background-color: #45a049;
}

.scoreboard-list {
  list-style-type: none;
  padding: 0;
}

.scoreboard-list li {
  padding: 10px;
  background-color: #f1f1f1;
  margin-bottom: 10px;
  border-radius: 5px;
}
</style>
