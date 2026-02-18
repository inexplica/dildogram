import React from 'react';
import clsx from 'clsx';

interface MessageStatusProps {
  status: 'pending' | 'sent' | 'delivered' | 'read';
  isOwn: boolean;
}

export function MessageStatus({ status, isOwn }: MessageStatusProps) {
  if (!isOwn) return null;

  const statusConfig = {
    pending: {
      icon: (
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      ),
      color: 'text-chat-meta',
      label: 'Отправляется...',
    },
    sent: {
      icon: (
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
        </svg>
      ),
      color: 'text-chat-meta',
      label: 'Отправлено',
    },
    delivered: {
      icon: (
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7M12 7l4 4-4 4" />
        </svg>
      ),
      color: 'text-chat-meta',
      label: 'Доставлено',
    },
    read: {
      icon: (
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7M12 7l4 4-4 4" />
        </svg>
      ),
      color: 'text-blue-400',
      label: 'Прочитано',
    },
  };

  const config = statusConfig[status];

  return (
    <span
      className={clsx('flex items-center gap-0.5', config.color)}
      title={config.label}
    >
      {config.icon}
    </span>
  );
}
