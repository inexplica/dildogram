import React, { useState } from 'react';
import { useChatStore } from '../../store/chatStore';
import { useAuth } from '../../hooks/useAuth';
import { ChatList } from './chat/ChatList';
import { ChatWindow } from './chat/ChatWindow';
import { MessageInput } from './chat/MessageInput';
import { useWebSocket } from '../../hooks/useWebSocket';
import { Modal } from './ui/Modal';
import { Button } from './ui/Button';
import { useState as useReactState } from 'react';

export function Sidebar() {
  const { user, token } = useAuth();
  const {
    chats,
    currentChat,
    messages,
    typingUsers,
    isLoading,
    fetchChats,
    selectChat,
    sendMessage,
    subscribeToChat,
    unsubscribeFromChat,
  } = useChatStore();

  const [isNewChatModalOpen, setIsNewChatModalOpen] = useState(false);
  const [newChatType, setNewChatType] = useState<'private' | 'group'>('private');
  const [newChatName, setNewChatName] = useState('');
  const [newChatMember, setNewChatMember] = useState('');

  // WebSocket хук
  useWebSocket(token);

  // Загрузка чатов при монтировании
  React.useEffect(() => {
    fetchChats();
  }, []);

  // Подписка на чат при выборе
  React.useEffect(() => {
    if (currentChat) {
      subscribeToChat(currentChat.id);
      return () => {
        unsubscribeFromChat(currentChat.id);
      };
    }
  }, [currentChat]);

  const handleSelectChat = (chatId: string) => {
    selectChat(chatId);
  };

  const handleSendMessage = (content: string) => {
    if (currentChat) {
      sendMessage(currentChat.id, content);
    }
  };

  const handleTyping = () => {
    // Можно добавить отправку события набора текста
  };

  const handleCreateChat = async () => {
    try {
      if (newChatType === 'private' && newChatMember) {
        // Для простоты создаём чат с одним участником
        // В реальном приложении нужен поиск пользователей
        alert('Функция поиска пользователей будет добавлена');
      } else if (newChatType === 'group' && newChatName) {
        // Создание группового чата
        alert('Создание группового чата будет добавлено');
      }
      setIsNewChatModalOpen(false);
    } catch (error) {
      console.error('Failed to create chat:', error);
    }
  };

  const currentMessages = currentChat ? messages.get(currentChat.id) || [] : [];
  const chatTypingUsers = currentChat
    ? (typingUsers.get(currentChat.id) || []).map((u) => ({
        userId: u.userId,
        userName: u.userName,
      }))
    : [];

  return (
    <div className="h-screen flex">
      {/* Sidebar */}
      <div className="w-80 md:w-96 bg-chat-secondary border-r border-chat-hover flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-chat-hover">
          <div className="flex items-center gap-3">
            <h1 className="text-xl font-bold text-white">Dildogram</h1>
          </div>
          <button
            onClick={() => setIsNewChatModalOpen(true)}
            className="p-2 text-chat-meta hover:text-white transition-colors"
            title="Новый чат"
          >
            <svg
              className="w-6 h-6"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 4v16m8-8H4"
              />
            </svg>
          </button>
        </div>

        {/* Chat list */}
        <div className="flex-1 overflow-y-auto">
          <ChatList
            chats={chats}
            selectedChatId={currentChat?.id || null}
            onSelectChat={handleSelectChat}
            isLoading={isLoading}
          />
        </div>

        {/* User info */}
        {user && (
          <div className="p-3 border-t border-chat-hover flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-primary-600 flex items-center justify-center text-white font-medium">
              {user.username[0].toUpperCase()}
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium text-white truncate">
                {user.username}
              </p>
              <p className="text-xs text-chat-meta">
                {user.is_online ? 'Онлайн' : 'Офлайн'}
              </p>
            </div>
          </div>
        )}
      </div>

      {/* Main chat area */}
      <div className="flex-1 flex flex-col bg-chat-bg">
        <ChatWindow
          chat={currentChat}
          messages={currentMessages}
          currentUserId={user?.id || null}
          typingUsers={chatTypingUsers}
        />

        {currentChat && (
          <div className="p-4 bg-chat-secondary border-t border-chat-hover">
            <MessageInput
              onSend={handleSendMessage}
              onTyping={handleTyping}
              placeholder="Написать сообщение..."
            />
          </div>
        )}
      </div>

      {/* New Chat Modal */}
      <Modal
        isOpen={isNewChatModalOpen}
        onClose={() => setIsNewChatModalOpen(false)}
        title="Новый чат"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm text-chat-meta mb-2">Тип чата</label>
            <div className="flex gap-2">
              <Button
                variant={newChatType === 'private' ? 'primary' : 'secondary'}
                onClick={() => setNewChatType('private')}
                className="flex-1"
              >
                Личный
              </Button>
              <Button
                variant={newChatType === 'group' ? 'primary' : 'secondary'}
                onClick={() => setNewChatType('group')}
                className="flex-1"
              >
                Групповой
              </Button>
            </div>
          </div>

          {newChatType === 'group' && (
            <div>
              <label className="block text-sm text-chat-meta mb-2">
                Название группы
              </label>
              <input
                type="text"
                value={newChatName}
                onChange={(e) => setNewChatName(e.target.value)}
                className="w-full bg-chat-hover border border-chat-hover rounded-lg px-3 py-2 text-white outline-none focus:border-primary-600"
                placeholder="Введите название"
              />
            </div>
          )}

          {newChatType === 'private' && (
            <div>
              <label className="block text-sm text-chat-meta mb-2">
                ID пользователя
              </label>
              <input
                type="text"
                value={newChatMember}
                onChange={(e) => setNewChatMember(e.target.value)}
                className="w-full bg-chat-hover border border-chat-hover rounded-lg px-3 py-2 text-white outline-none focus:border-primary-600"
                placeholder="Введите ID"
              />
            </div>
          )}

          <div className="flex gap-2 pt-4">
            <Button
              variant="secondary"
              onClick={() => setIsNewChatModalOpen(false)}
              className="flex-1"
            >
              Отмена
            </Button>
            <Button
              variant="primary"
              onClick={handleCreateChat}
              className="flex-1"
            >
              Создать
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
