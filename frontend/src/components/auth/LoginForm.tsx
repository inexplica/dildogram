import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import { Button } from '../ui/Button';

export function LoginForm() {
  const navigate = useNavigate();
  const { login, requestSMS, verifySMS, isLoading, error, clearError } = useAuth();
  
  const [loginMethod, setLoginMethod] = useState<'password' | 'sms'>('password');
  const [phone, setPhone] = useState('');
  const [password, setPassword] = useState('');
  const [code, setCode] = useState('');
  const [smsSent, setSmsSent] = useState(false);

  const handlePasswordLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await login(phone, password);
      navigate('/');
    } catch (err) {
      // Error handled by store
    }
  };

  const handleRequestSMS = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await requestSMS(phone);
      setSmsSent(true);
    } catch (err) {
      // Error handled by store
    }
  };

  const handleVerifySMS = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await verifySMS(phone, code);
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
          <h2 className="text-xl font-semibold text-white mb-6">Вход</h2>

          {/* Выбор метода входа */}
          <div className="flex gap-2 mb-6">
            <button
              onClick={() => setLoginMethod('password')}
              className={`flex-1 py-2 rounded-lg text-sm font-medium transition-colors ${
                loginMethod === 'password'
                  ? 'bg-primary-600 text-white'
                  : 'bg-chat-hover text-chat-meta hover:text-white'
              }`}
            >
              Пароль
            </button>
            <button
              onClick={() => setLoginMethod('sms')}
              className={`flex-1 py-2 rounded-lg text-sm font-medium transition-colors ${
                loginMethod === 'sms'
                  ? 'bg-primary-600 text-white'
                  : 'bg-chat-hover text-chat-meta hover:text-white'
              }`}
            >
              SMS
            </button>
          </div>

          {error && (
            <div className="mb-4 p-3 bg-red-500/10 border border-red-500/50 rounded-lg text-red-400 text-sm">
              {error}
            </div>
          )}

          {loginMethod === 'password' ? (
            <form onSubmit={handlePasswordLogin} className="space-y-4">
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

              <Button type="submit" className="w-full" isLoading={isLoading}>
                Войти
              </Button>
            </form>
          ) : (
            <>
              {!smsSent ? (
                <form onSubmit={handleRequestSMS} className="space-y-4">
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

                  <Button type="submit" className="w-full" isLoading={isLoading}>
                    Получить код
                  </Button>
                </form>
              ) : (
                <form onSubmit={handleVerifySMS} className="space-y-4">
                  <div>
                    <label className="block text-sm text-chat-meta mb-2">
                      Код из SMS
                    </label>
                    <input
                      type="text"
                      value={code}
                      onChange={(e) => {
                        setCode(e.target.value);
                        clearError();
                      }}
                      className="w-full bg-chat-hover border border-chat-hover rounded-lg px-4 py-2.5 text-white outline-none focus:border-primary-600 transition-colors text-center text-2xl tracking-widest"
                      placeholder="000000"
                      maxLength={6}
                      required
                    />
                  </div>

                  <Button type="submit" className="w-full" isLoading={isLoading}>
                    Войти
                  </Button>

                  <button
                    type="button"
                    onClick={() => setSmsSent(false)}
                    className="w-full text-sm text-chat-meta hover:text-white transition-colors"
                  >
                    Изменить номер
                  </button>
                </form>
              )}
            </>
          )}

          <p className="mt-6 text-center text-sm text-chat-meta">
            Нет аккаунта?{' '}
            <Link to="/register" className="text-primary-400 hover:text-primary-300">
              Зарегистрироваться
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
