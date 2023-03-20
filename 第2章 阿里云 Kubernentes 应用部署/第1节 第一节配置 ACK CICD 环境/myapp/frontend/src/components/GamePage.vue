<template>
  <div class="game-page">
    <h1>Game</h1>

    <p>Score: {{ score }}</p>

    <button @click="incrementScore">Click me!</button>
  </div>
</template>

<script>
import GameService from '@/services/game.js'

export default {
  name: 'GamePage',

  data () {
    return {
      score: 0
    }
  },

  methods: {
    incrementScore () {
      this.score++

      GameService.saveScore({ score: this.score })
        .catch(error => {
          console.log(error)
        })
    }
  },

  created () {
    GameService.getScoreboard()
      .then(response => {
        console.log(response.data)
      })
      .catch(error => {
        console.log(error)
      })
  }
}
</script>
