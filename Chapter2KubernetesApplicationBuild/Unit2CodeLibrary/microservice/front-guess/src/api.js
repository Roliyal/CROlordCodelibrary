import axios from 'axios';

const apiClient = axios.create({
    baseURL: '/', // 修改为 '/'
    withCredentials: false,
    headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json',
    },
});

// ...

export default {
    async login(username, password) {
        try {
            const response = await apiClient.post('/login', { username, password }); // 修改为 '/login'
            return response.data;
        } catch (error) {
            console.error('Error during login:', error);
            return null;
        }
    },
};
