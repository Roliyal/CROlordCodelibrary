<template>
  <div class="game">
    <div class="game__score">Score: {{ score }}</div>
    <button class="game__button" @click="incrementScore">Click me!</button>
  </div>
</template>

<script>
export default {
  name: 'Game',
  data() {
    return {
      score: 0
    }
  },
  methods: {
    incrementScore() {
      this.score++
    }
  }
}
</script>

<style>
.game {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
}

.game__score {
  font-size: 36px;
  font-weight: bold;
  margin-bottom: 20px;
}

.game__button {
  font-size: 24px;
  padding: 10px 20px;
  border: none;
  border-radius: 5px;
  background-color: #3f51b5;
  color: #fff;
  cursor: pointer;
  transition: all 0.3s ease;
}

.game__button:hover {
  background-color: #2c3e50;
}
</style>
