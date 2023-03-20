import axios from 'axios';

const API_URL = 'http://localhost:8000';

export default {
  getGame: function(token) {
    return axios.get(`${API_URL}/game`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
  },
  saveScore: function(token, score) {
    return axios.post(`${API_URL}/game`, {
      score: score
    }, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
  }
};
