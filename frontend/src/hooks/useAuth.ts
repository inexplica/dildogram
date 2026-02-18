import { useEffect } from 'react';
import { useAuthStore } from '../store/authStore';
import { wsService } from '../services/websocket';

export function useAuth() {
  const {
    user,
    token,
    isAuthenticated,
    isLoading,
    error,
    login,
    register,
    requestSMS,
    verifySMS,
    logout,
    updateProfile,
    uploadAvatar,
    fetchUser,
    clearError,
  } = useAuthStore();

  // Подключение WebSocket при аутентификации
  useEffect(() => {
    if (token && isAuthenticated) {
      wsService.connect(token).catch(console.error);
    }

    return () => {
      wsService.disconnect();
    };
  }, [token, isAuthenticated]);

  // Проверка токена при загрузке
  useEffect(() => {
    if (token && !user) {
      fetchUser().catch(console.error);
    }
  }, [token, user, fetchUser]);

  return {
    user,
    token,
    isAuthenticated,
    isLoading,
    error,
    login,
    register,
    requestSMS,
    verifySMS,
    logout,
    updateProfile,
    uploadAvatar,
    clearError,
  };
}
