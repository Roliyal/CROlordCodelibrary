import axios from 'axios'

const API_URL = 'http://localhost:8080'

class AuthService {
  login (user) {
    return axios
      .post(`${API_URL}/login`, {
        username: user.username,
        password: user.password
      })
      .then(response => {
        if (response.data.token) {
          localStorage.setItem('token', JSON.stringify(response.data.token))
        }

        return response.data
      })
  }

  logout () {
    localStorage.removeItem('token')
  }

  getToken () {
    return JSON.parse(localStorage.getItem('token'))
  }
}

export default new AuthService()
