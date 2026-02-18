import axios, { AxiosInstance, AxiosError } from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

class ApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Interceptor для добавления токена
    this.client.interceptors.request.use((config) => {
      const token = localStorage.getItem('token');
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      return config;
    });

    // Interceptor для обработки ошибок
    this.client.interceptors.response.use(
      (response) => response,
      (error: AxiosError<{ error?: string }>) => {
        if (error.response?.status === 401) {
          localStorage.removeItem('token');
          localStorage.removeItem('user');
          window.location.href = '/login';
        }
        return Promise.reject(error);
      }
    );
  }

  // Auth endpoints
  async login(phone: string, password: string) {
    const response = await this.client.post('/auth/login', { phone, password });
    return response.data;
  }

  async register(phone: string, username: string, password: string) {
    const response = await this.client.post('/auth/register', {
      phone,
      username,
      password,
    });
    return response.data;
  }

  async requestSMS(phone: string) {
    const response = await this.client.post('/auth/sms', { phone });
    return response.data;
  }

  async verifySMS(phone: string, code: string) {
    const response = await this.client.post('/auth/verify-sms', { phone, code });
    return response.data;
  }

  async getMe() {
    const response = await this.client.get('/auth/me');
    return response.data;
  }

  async updateProfile(data: { first_name?: string; last_name?: string; bio?: string }) {
    const response = await this.client.put('/auth/me', data);
    return response.data;
  }

  async uploadAvatar(file: File) {
    const formData = new FormData();
    formData.append('avatar', file);
    const response = await this.client.post('/auth/avatar', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  }

  // User endpoints
  async getUser(id: string) {
    const response = await this.client.get(`/users/${id}`);
    return response.data;
  }

  async searchUsers(query: string) {
    const response = await this.client.get('/users', { params: { q: query } });
    return response.data;
  }

  // Chat endpoints
  async getChats() {
    const response = await this.client.get('/chats');
    return response.data;
  }

  async createChat(data: {
    type: 'private' | 'group';
    name?: string;
    description?: string;
    member_ids?: string[];
  }) {
    const response = await this.client.post('/chats', data);
    return response.data;
  }

  async getChat(id: string) {
    const response = await this.client.get(`/chats/${id}`);
    return response.data;
  }

  async updateChat(id: string, data: { name?: string; description?: string }) {
    const response = await this.client.put(`/chats/${id}`, data);
    return response.data;
  }

  async deleteChat(id: string) {
    const response = await this.client.delete(`/chats/${id}`);
    return response.data;
  }

  async addMember(chatId: string, userId: string) {
    const response = await this.client.post(`/chats/${chatId}/members`, {
      user_id: userId,
    });
    return response.data;
  }

  async removeMember(chatId: string, userId: string) {
    const response = await this.client.delete(`/chats/${chatId}/members/${userId}`);
    return response.data;
  }

  async getMembers(chatId: string) {
    const response = await this.client.get(`/chats/${chatId}/members`);
    return response.data;
  }

  // Message endpoints
  async getMessages(chatId: string, limit = 50, offset = 0) {
    const response = await this.client.get(`/chats/${chatId}/messages`, {
      params: { limit, offset },
    });
    return response.data;
  }

  async sendMessage(chatId: string, data: {
    content: string;
    message_type?: string;
    media_url?: string | null;
    reply_to_id?: string | null;
  }) {
    const response = await this.client.post(`/chats/${chatId}/messages`, data);
    return response.data;
  }

  async markChatAsRead(chatId: string) {
    const response = await this.client.post(`/chats/${chatId}/read`);
    return response.data;
  }
}

export const api = new ApiClient();
export default api;
