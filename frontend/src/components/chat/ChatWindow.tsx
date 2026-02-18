import React, { useRef, useEffect } from 'react';
import { Message, User, Chat } from '../../types';
import { MessageBubble } from './MessageBubble';
import { Avatar } from '../ui/Avatar';
import clsx from 'clsx';

interface ChatWindowProps {
  chat: Chat | null;
  messages: Message[];
  currentUserId: string | null;
  typingUsers: Array<{ userId: string; userName: string }>;
  isLoading?: boolean;
}

export function ChatWindow({
  chat,
  messages,
  currentUserId,
  typingUsers,
  isLoading = false,
}: ChatWindowProps) {
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const messagesContainerRef = useRef<HTMLDivElement>(null);

  // Автопрокрутка к последнему сообщению
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  // Группировка сообщений по отправителю для отображения аватара
  const shouldShowAvatar = (message: Message, index: number): boolean => {
    const nextMessage = messages[index + 1];
    if (!nextMessage) return true;
    
    // Показываем аватар, если следующий сообщение от другого пользователя
    // или если это первое сообщение в группе от одного отправителя
    return message.sender_id !== nextMessage.sender_id;
  };

  if (!chat) {
    return (
      <div className="flex-1 flex items-center justify-center bg-chat-bg">
        <div className="text-center text-chat-meta">
          <svg
            className="w-20 h-20 mx-auto mb-4 opacity-30"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1}
              d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
            />
          </svg>
          <p className="text-lg">Выберите чат</p>
          <p className="text-sm mt-2">для начала общения</p>
        </div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="flex-1 flex items-center justify-center bg-chat-bg">
        <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-primary-600" />
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col bg-chat-bg overflow-hidden">
      {/* Header */}
      <div className="flex items-center gap-3 p-3 bg-chat-secondary border-b border-chat-hover">
        {chat.avatar_url ? (
          <Avatar src={chat.avatar_url} alt={chat.name} size="md" />
        ) : (
          <Avatar alt={chat.name} size="md" />
        )}
        <div className="flex-1 min-w-0">
          <h2 className="font-semibold text-white truncate">{chat.name}</h2>
          <p className="text-xs text-chat-meta">
            {chat.type === 'group' ? chat.description : 'Личный чат'}
          </p>
        </div>
      </div>

      {/* Messages */}
      <div
        ref={messagesContainerRef}
        className="flex-1 overflow-y-auto p-4"
      >
        {messages.length === 0 ? (
          <div className="flex items-center justify-center h-full text-chat-meta">
            <p className="text-sm">Сообщений пока нет</p>
          </div>
        ) : (
          messages.map((message, index) => {
            const isOwn = message.sender_id === currentUserId;
            const showAvatar = shouldShowAvatar(message, index);

            return (
              <MessageBubble
                key={message.id}
                message={message}
                isOwn={isOwn}
                showAvatar={showAvatar && chat.type === 'group'}
                senderName={message.sender?.username}
                senderAvatar={message.sender?.avatar_url}
              />
            );
          })
        )}

        {/* Typing indicator */}
        {typingUsers.length > 0 && (
          <div className="flex items-center gap-2 mt-2 ml-1">
            <div className="bg-chat-incoming rounded-lg rounded-bl-sm px-4 py-3">
              <div className="flex gap-1">
                <span className="w-2 h-2 bg-chat-meta rounded-full typing-dot" />
                <span className="w-2 h-2 bg-chat-meta rounded-full typing-dot" />
                <span className="w-2 h-2 bg-chat-meta rounded-full typing-dot" />
              </div>
            </div>
            <span className="text-xs text-chat-meta">
              {typingUsers.map((u) => u.userName).join(', ')}{' '}
              {typingUsers.length === 1 ? 'печатает' : 'печатают'}...
            </span>
          </div>
        )}

        <div ref={messagesEndRef} />
      </div>
    </div>
  );
}
