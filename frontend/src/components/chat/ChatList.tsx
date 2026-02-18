import React from 'react';
import { ChatWithLastMessage } from '../../types';
import { Avatar } from '../ui/Avatar';
import { formatChatTime } from '../../utils/formatTime';
import clsx from 'clsx';

interface ChatListProps {
  chats: ChatWithLastMessage[];
  selectedChatId: string | null;
  onSelectChat: (chatId: string) => void;
  isLoading?: boolean;
}

export function ChatList({
  chats,
  selectedChatId,
  onSelectChat,
  isLoading = false,
}: ChatListProps) {
  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600" />
      </div>
    );
  }

  if (chats.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-full text-chat-meta">
        <svg
          className="w-16 h-16 mb-4 opacity-50"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={1.5}
            d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
          />
        </svg>
        <p className="text-sm">Нет чатов</p>
        <p className="text-xs mt-1">Начните новый диалог</p>
      </div>
    );
  }

  return (
    <div className="divide-y divide-chat-hover">
      {chats.map((chat) => (
        <button
          key={chat.id}
          onClick={() => onSelectChat(chat.id)}
          className={clsx(
            'w-full p-3 flex items-center gap-3',
            'hover:bg-chat-hover transition-colors',
            'text-left',
            selectedChatId === chat.id && 'bg-chat-hover'
          )}
        >
          {/* Avatar */}
          <div className="flex-shrink-0">
            {chat.avatar_url ? (
              <Avatar
                src={chat.avatar_url}
                alt={chat.name}
                size="lg"
              />
            ) : (
              <Avatar
                alt={chat.name}
                size="lg"
              />
            )}
          </div>

          {/* Content */}
          <div className="flex-1 min-w-0">
            {/* Header */}
            <div className="flex items-center justify-between mb-1">
              <h3 className="font-medium text-white truncate">{chat.name}</h3>
              <div className="flex items-center gap-2 flex-shrink-0">
                <span className="text-xs text-chat-meta">
                  {formatChatTime(chat.last_message_created_at)}
                </span>
                {chat.unread_count > 0 && (
                  <span className="bg-primary-600 text-white text-xs font-medium px-2 py-0.5 rounded-full min-w-[20px] text-center">
                    {chat.unread_count > 99 ? '99+' : chat.unread_count}
                  </span>
                )}
              </div>
            </div>

            {/* Last message */}
            <p className="text-sm text-chat-meta truncate">
              {chat.last_message_content || (
                <span className="italic">Нет сообщений</span>
              )}
            </p>
          </div>
        </button>
      ))}
    </div>
  );
}
