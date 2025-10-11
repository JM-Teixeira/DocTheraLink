import axios from 'axios';

const apiClient = axios.create({
    baseURL: 'https://localhost:3001', // Replace with your API's base URL
    timeout: 10000, // Set a timeout
    headers: {
        'Content-Type': 'application/json',
    },
});

export default apiClient;
