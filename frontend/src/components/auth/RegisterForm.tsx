import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import { Button } from '../ui/Button';

export function RegisterForm() {
  const navigate = useNavigate();
  const { register, isLoading, error, clearError } = useAuth();

  const [phone, setPhone] = useState('');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (password !== confirmPassword) {
      alert('Пароли не совпадают');
      return;
    }

    if (password.length < 6) {
      alert('Пароль должен быть не менее 6 символов');
      return;
    }

    try {
      await register(phone, username, password);
      navigate('/');
    } catch (err) {
      // Error handled by store
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-chat-bg p-4">
      <div className="w-full max-w-md">
        {/* Logo */}
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-white mb-2">Dildogram</h1>
          <p className="text-chat-meta">Мессенджер нового поколения</p>
        </div>

        {/* Form */}
        <div className="bg-chat-secondary rounded-xl p-6 shadow-xl">
          <h2 className="text-xl font-semibold text-white mb-6">Регистрация</h2>

          {error && (
            <div className="mb-4 p-3 bg-red-500/10 border border-red-500/50 rounded-lg text-red-400 text-sm">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm text-chat-meta mb-2">
                Телефон
              </label>
              <input
                type="tel"
                value={phone}
                onChange={(e) => {
                  setPhone(e.target.value);
                  clearError();
                }}
                className="w-full bg-chat-hover border border-chat-hover rounded-lg px-4 py-2.5 text-white outline-none focus:border-primary-600 transition-colors"
                placeholder="+7 999 123 45 67"
                required
              />
            </div>

            <div>
              <label className="block text-sm text-chat-meta mb-2">
                Имя пользователя
              </label>
              <input
                type="text"
                value={username}
                onChange={(e) => {
                  setUsername(e.target.value);
                  clearError();
                }}
                className="w-full bg-chat-hover border border-chat-hover rounded-lg px-4 py-2.5 text-white outline-none focus:border-primary-600 transition-colors"
                placeholder="@username"
                required
                minLength={3}
                maxLength={50}
              />
            </div>

            <div>
              <label className="block text-sm text-chat-meta mb-2">
                Пароль
              </label>
              <input
                type="password"
                value={password}
                onChange={(e) => {
                  setPassword(e.target.value);
                  clearError();
                }}
                className="w-full bg-chat-hover border border-chat-hover rounded-lg px-4 py-2.5 text-white outline-none focus:border-primary-600 transition-colors"
                placeholder="••••••••"
                required
                minLength={6}
              />
            </div>

            <div>
              <label className="block text-sm text-chat-meta mb-2">
                Подтвердите пароль
              </label>
              <input
                type="password"
                value={confirmPassword}
                onChange={(e) => {
                  setConfirmPassword(e.target.value);
                  clearError();
                }}
                className="w-full bg-chat-hover border border-chat-hover rounded-lg px-4 py-2.5 text-white outline-none focus:border-primary-600 transition-colors"
                placeholder="••••••••"
                required
              />
            </div>

            <Button type="submit" className="w-full" isLoading={isLoading}>
              Зарегистрироваться
            </Button>
          </form>

          <p className="mt-6 text-center text-sm text-chat-meta">
            Уже есть аккаунт?{' '}
            <Link to="/login" className="text-primary-400 hover:text-primary-300">
              Войти
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
