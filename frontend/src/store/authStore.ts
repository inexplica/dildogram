import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { User } from '../types';
import { api } from '../services/api';

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;

  // Actions
  login: (phone: string, password: string) => Promise<void>;
  register: (phone: string, username: string, password: string) => Promise<void>;
  requestSMS: (phone: string) => Promise<string>;
  verifySMS: (phone: string, code: string) => Promise<void>;
  logout: () => void;
  updateProfile: (data: { first_name?: string; last_name?: string; bio?: string }) => Promise<void>;
  uploadAvatar: (file: File) => Promise<void>;
  fetchUser: () => Promise<void>;
  clearError: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,

      login: async (phone: string, password: string) => {
        set({ isLoading: true, error: null });
        try {
          const { user, token } = await api.login(phone, password);
          set({ user, token, isAuthenticated: true, isLoading: false });
        } catch (error: unknown) {
          const message = error instanceof Error ? error.message : 'Login failed';
          set({ error: message, isLoading: false });
          throw error;
        }
      },

      register: async (phone: string, username: string, password: string) => {
        set({ isLoading: true, error: null });
        try {
          const { user, token } = await api.register(phone, username, password);
          set({ user, token, isAuthenticated: true, isLoading: false });
        } catch (error: unknown) {
          const message = error instanceof Error ? error.message : 'Registration failed';
          set({ error: message, isLoading: false });
          throw error;
        }
      },

      requestSMS: async (phone: string) => {
        set({ isLoading: true, error: null });
        try {
          const { code } = await api.requestSMS(phone);
          set({ isLoading: false });
          return code;
        } catch (error: unknown) {
          const message = error instanceof Error ? error.message : 'SMS request failed';
          set({ error: message, isLoading: false });
          throw error;
        }
      },

      verifySMS: async (phone: string, code: string) => {
        set({ isLoading: true, error: null });
        try {
          const { user, token } = await api.verifySMS(phone, code);
          set({ user, token, isAuthenticated: true, isLoading: false });
        } catch (error: unknown) {
          const message = error instanceof Error ? error.message : 'SMS verification failed';
          set({ error: message, isLoading: false });
          throw error;
        }
      },

      logout: () => {
        localStorage.removeItem('token');
        set({ user: null, token: null, isAuthenticated: false });
      },

      updateProfile: async (data) => {
        try {
          const { user } = await api.updateProfile(data);
          set({ user });
        } catch (error: unknown) {
          const message = error instanceof Error ? error.message : 'Update failed';
          set({ error: message });
          throw error;
        }
      },

      uploadAvatar: async (file: File) => {
        try {
          const { user } = await api.uploadAvatar(file);
          set({ user });
        } catch (error: unknown) {
          const message = error instanceof Error ? error.message : 'Upload failed';
          set({ error: message });
          throw error;
        }
      },

      fetchUser: async () => {
        const { token } = get();
        if (!token) return;

        try {
          const { user } = await api.getMe();
          set({ user, isAuthenticated: true });
        } catch (error) {
          set({ user: null, token: null, isAuthenticated: false });
        }
      },

      clearError: () => set({ error: null }),
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({ token: state.token }),
    }
  )
);
