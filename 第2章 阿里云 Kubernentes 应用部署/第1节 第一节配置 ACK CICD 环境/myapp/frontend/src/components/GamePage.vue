<!-- src/components/GamePage.vue -->
<template>
  <div class="game-page">
    <h2>Click the Button!</h2>
    <p>Score: {{ score }}</p>
    <button @click="incrementScore">Click Me!</button>
  </div>
</template>

<script>
import GameService from "../services/game";

export default {
  name: "GamePage",
  data() {
    return {
      score: 0,
    };
  },
  methods: {
    async incrementScore() {
      this.score++;
      await GameService.saveScore(this.score);
    },
  },
};
</script>

<style scoped>
.game-page {
  display:
