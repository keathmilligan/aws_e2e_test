import axios from 'axios';

// Get the API URL from environment variables or use a default
const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

// Create an axios instance with the base URL
const api = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add a request interceptor to add the auth token to requests
api.interceptors.request.use(
  (config) => {
    // Get the token from local storage
    const idToken = localStorage.getItem('idToken');
    
    // If token exists, add it to the headers
    if (idToken) {
      config.headers.Authorization = `Bearer ${idToken}`;
    }
    
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Add a response interceptor to handle authentication errors
api.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    // Handle 401 Unauthorized errors
    if (error.response && error.response.status === 401) {
      // Clear local storage and redirect to login
      localStorage.removeItem('idToken');
      localStorage.removeItem('accessToken');
      localStorage.removeItem('refreshToken');
      
      // Redirect to login page
      window.location.href = '/login';
    }
    
    return Promise.reject(error);
  }
);

// Define the Message interface
export interface Message {
  id: string;
  text: string;
  timestamp: string;
}

// API functions
export const getMessages = async (): Promise<Message[]> => {
  const response = await api.get<Message[]>('/messages');
  return response.data;
};

export const createMessage = async (text: string): Promise<Message> => {
  const response = await api.post<Message>('/messages', { text });
  return response.data;
};

// User API functions
export interface User {
  email: string;
  firstName: string;
  lastName: string;
  status: string;
  createdAt: string;
  updatedAt: string;
}

export const getUser = async (email: string): Promise<User> => {
  const response = await api.get<User>(`/users/${email}`);
  return response.data;
};

export const updateUser = async (email: string, userData: Partial<User>): Promise<User> => {
  const response = await api.put<User>(`/users/${email}`, userData);
  return response.data;
};

export default api;