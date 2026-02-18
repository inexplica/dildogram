import React from 'react';
import clsx from 'clsx';
import { Message } from '../../types';
import { Avatar } from '../ui/Avatar';
import { MessageStatus } from './MessageStatus';
import { formatMessageTime } from '../../utils/formatTime';

interface MessageBubbleProps {
  message: Message;
  isOwn: boolean;
  showAvatar?: boolean;
  senderName?: string;
  senderAvatar?: string;
}

export function MessageBubble({
  message,
  isOwn,
  showAvatar = false,
  senderName,
  senderAvatar,
}: MessageBubbleProps) {
  if (message.is_deleted) {
    return (
      <div
        className={clsx(
          'flex mb-2',
          isOwn ? 'justify-end' : 'justify-start'
        )}
      >
        <div
          className={clsx(
            'px-4 py-2 rounded-lg max-w-xs md:max-w-md lg:max-w-lg',
            'bg-chat-hover text-chat-meta italic'
          )}
        >
          Это сообщение было удалено
        </div>
      </div>
    );
  }

  return (
    <div
      className={clsx(
        'flex mb-2 message-animation group',
        isOwn ? 'justify-end' : 'justify-start',
        showAvatar ? 'items-end' : 'items-end'
      )}
    >
      {/* Avatar для входящих сообщений */}
      {!isOwn && showAvatar && (
        <div className="mr-2 flex-shrink-0">
          <Avatar
            src={senderAvatar || message.sender?.avatar_url}
            alt={senderName || message.sender?.username || ''}
            size="sm"
          />
        </div>
      )}

      {/* Bubble */}
      <div
        className={clsx(
          'px-3 py-2 rounded-lg max-w-xs md:max-w-md lg:max-w-lg shadow-sm',
          isOwn
            ? 'bg-chat-outgoing text-white rounded-br-sm'
            : 'bg-chat-incoming text-white rounded-bl-sm'
        )}
      >
        {/* Имя отправителя в групповом чате */}
        {!isOwn && senderName && (
          <div className="text-xs text-primary-400 font-medium mb-1">
            {senderName}
          </div>
        )}

        {/* Контент сообщения */}
        <div className="text-sm whitespace-pre-wrap break-words">
          {message.content}
        </div>

        {/* Мета информация */}
        <div
          className={clsx(
            'flex items-center justify-end gap-1 mt-1',
            'text-xs text-chat-meta'
          )}
        >
          <span>{formatMessageTime(message.created_at)}</span>
          <MessageStatus status={message.status} isOwn={isOwn} />
        </div>

        {/* Индикатор редактирования */}
        {message.is_edited && (
          <span className="text-xs text-chat-meta ml-1">(ред.)</span>
        )}
      </div>
    </div>
  );
}
