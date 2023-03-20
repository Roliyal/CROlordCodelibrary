import axios from 'axios'

const API_URL = 'http://localhost:8080'

class GameService {
  getScoreboard () {
    return axios.get(`${API_URL}/scoreboard`)
  }

  saveScore (score) {
    const token = JSON.parse(localStorage.getItem('token'))

    return axios
      .post(`${API_URL}/score`, score, {
        headers: {
          Authorization: `Bearer ${token}`
        }
      })
  }
}

export default new GameService()
