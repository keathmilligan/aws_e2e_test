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

export default api;