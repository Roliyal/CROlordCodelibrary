<template>
  <div class="container scoreboard-container">
    <h2>战绩展示</h2>
    <p>在这里查看您的战绩！</p>
    <!-- 展示战绩的具体实现 -->
    <button @click="fetchScoreboardData">获取战绩</button>
  </div>
</template>

<script>
export default {
  methods: {
    async fetchScoreboardData() {
      // 在此处添加API调用以获取战绩数据
      console.log("获取战绩数据");
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
</style>
