import React, { useState, useRef, useEffect } from 'react';
import clsx from 'clsx';

interface MessageInputProps {
  onSend: (content: string) => void;
  onTyping?: () => void;
  disabled?: boolean;
  placeholder?: string;
}

export function MessageInput({
  onSend,
  onTyping,
  disabled = false,
  placeholder = 'Написать сообщение...',
}: MessageInputProps) {
  const [content, setContent] = useState('');
  const [isFocused, setIsFocused] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const typingTimeoutRef = useRef<NodeJS.Timeout>();

  const adjustHeight = () => {
    const textarea = textareaRef.current;
    if (!textarea) return;

    textarea.style.height = 'auto';
    const newHeight = Math.min(textarea.scrollHeight, 200);
    textarea.style.height = `${newHeight}px`;
  };

  useEffect(() => {
    adjustHeight();
  }, [content]);

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setContent(e.target.value);
    adjustHeight();

    // Отправляем событие набора текста
    if (onTyping) {
      onTyping();
      
      // Debounce для остановки набора
      if (typingTimeoutRef.current) {
        clearTimeout(typingTimeoutRef.current);
      }
      typingTimeoutRef.current = setTimeout(() => {
        // Можно отправить событие остановки набора
      }, 1000);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit();
    }
  };

  const handleSubmit = () => {
    const trimmed = content.trim();
    if (trimmed && !disabled) {
      onSend(trimmed);
      setContent('');
      
      if (textareaRef.current) {
        textareaRef.current.style.height = 'auto';
      }
    }
  };

  return (
    <div
      className={clsx(
        'flex items-end gap-2 p-3 bg-chat-secondary rounded-xl',
        'border border-transparent transition-colors',
        isFocused && 'border-primary-600/50'
      )}
    >
      <textarea
        ref={textareaRef}
        value={content}
        onChange={handleChange}
        onKeyDown={handleKeyDown}
        onFocus={() => setIsFocused(true)}
        onBlur={() => setIsFocused(false)}
        disabled={disabled}
        placeholder={placeholder}
        rows={1}
        className={clsx(
          'flex-1 bg-transparent text-white placeholder-chat-meta',
          'resize-none outline-none text-sm',
          'max-h-[200px] min-h-[40px]',
          'scrollbar-thin scrollbar-thumb-chat-hover'
        )}
      />

      <button
        onClick={handleSubmit}
        disabled={!content.trim() || disabled}
        className={clsx(
          'flex-shrink-0 w-10 h-10 rounded-full',
          'flex items-center justify-center',
          'transition-all duration-200',
          content.trim() && !disabled
            ? 'bg-primary-600 hover:bg-primary-700 text-white'
            : 'bg-chat-hover text-chat-meta cursor-not-allowed'
        )}
      >
        <svg
          className="w-5 h-5 transform rotate-90"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
          />
        </svg>
      </button>
    </div>
  );
}
