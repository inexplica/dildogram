import { useEffect, useCallback, useRef } from 'react';
import { wsService } from '../services/websocket';
import {
  MessagePayload,
  TypingStatusPayload,
  UserStatusPayload,
  ErrorPayload,
} from '../types';
import { useChatStore } from '../store/chatStore';

export function useWebSocket(token: string | null) {
  const addMessage = useChatStore((state) => state.addMessage);
  const updateMessageStatus = useChatStore((state) => state.updateMessageStatus);
  const setTypingUser = useChatStore((state) => state.setTypingUser);
  const clearTypingUsers = useChatStore((state) => state.clearTypingUsers);

  const typingTimeoutRef = useRef<Map<string, NodeJS.Timeout>>(new Map());

  useEffect(() => {
    if (!token) return;

    let isMounted = true;

    const connect = async () => {
      try {
        await wsService.connect(token);
      } catch (error) {
        console.error('Failed to connect WebSocket:', error);
      }
    };

    connect();

    // Подписка на события
    const unsubscribers: (() => void)[] = [];

    // Новое сообщение
    unsubscribers.push(
      wsService.on('message', (payload) => {
        if (!isMounted) return;
        const msgPayload = payload as MessagePayload;
        const message = {
          id: msgPayload.id,
          chat_id: msgPayload.chat_id,
          sender_id: msgPayload.sender_id,
          content: msgPayload.content,
          message_type: msgPayload.message_type as 'text' | 'image' | 'file' | 'voice',
          media_url: msgPayload.media_url,
          reply_to_id: msgPayload.reply_to_id,
          is_edited: msgPayload.is_edited,
          is_deleted: msgPayload.is_deleted,
          status: msgPayload.status as 'pending' | 'sent' | 'delivered' | 'read',
          created_at: msgPayload.created_at,
          updated_at: msgPayload.created_at,
          sender: {
            id: msgPayload.sender_id,
            username: msgPayload.sender_name,
            first_name: '',
            last_name: '',
            avatar_url: msgPayload.sender_avatar,
            is_online: false,
          },
        };
        addMessage(msgPayload.chat_id, message);
      })
    );

    // Статус сообщения
    unsubscribers.push(
      wsService.on('message_status', (payload) => {
        if (!isMounted) return;
        const { message_id, status } = payload as { message_id: string; status: string };
        updateMessageStatus(message_id, status);
      })
    );

    // Набор текста
    unsubscribers.push(
      wsService.on('typing', (payload) => {
        if (!isMounted) return;
        const typingPayload = payload as TypingStatusPayload;
        
        // Очистка предыдущего таймаута
        const existingTimeout = typingTimeoutRef.current.get(typingPayload.chat_id);
        if (existingTimeout) {
          clearTimeout(existingTimeout);
        }

        if (typingPayload.is_typing) {
          setTypingUser(typingPayload.chat_id, {
            id: typingPayload.user_id,
            username: typingPayload.user_name,
            first_name: '',
            last_name: '',
            bio: '',
            avatar_url: '',
            is_active: true,
            is_online: true,
            last_seen: new Date().toISOString(),
            created_at: '',
            updated_at: '',
            phone: '',
          }, true);

          // Автоматическая очистка через 3 секунды
          const timeout = setTimeout(() => {
            setTypingUser(typingPayload.chat_id, {
              id: typingPayload.user_id,
              username: typingPayload.user_name,
              first_name: '',
              last_name: '',
              bio: '',
              avatar_url: '',
              is_active: true,
              is_online: true,
              last_seen: new Date().toISOString(),
              created_at: '',
              updated_at: '',
              phone: '',
            }, false);
          }, 3000);

          typingTimeoutRef.current.set(typingPayload.chat_id, timeout);
        } else {
          setTypingUser(typingPayload.chat_id, {
            id: typingPayload.user_id,
            username: typingPayload.user_name,
            first_name: '',
            last_name: '',
            bio: '',
            avatar_url: '',
            is_active: true,
            is_online: true,
            last_seen: new Date().toISOString(),
            created_at: '',
            updated_at: '',
            phone: '',
          }, false);
        }
      })
    );

    // Пользователь онлайн
    unsubscribers.push(
      wsService.on('user_online', (payload) => {
        if (!isMounted) return;
        const { user_id } = payload as UserStatusPayload;
        console.log('User online:', user_id);
        // Здесь можно обновить статус пользователя в store
      })
    );

    // Пользователь офлайн
    unsubscribers.push(
      wsService.on('user_offline', (payload) => {
        if (!isMounted) return;
        const { user_id } = payload as UserStatusPayload;
        console.log('User offline:', user_id);
      })
    );

    // Ошибки
    unsubscribers.push(
      wsService.on('error', (payload) => {
        if (!isMounted) return;
        const { code, message } = payload as ErrorPayload;
        console.error('WebSocket error:', code, message);
      })
    );

    return () => {
      isMounted = false;
      unsubscribers.forEach((unsub) => unsub());
      
      // Очистка всех таймаутов
      typingTimeoutRef.current.forEach((timeout) => clearTimeout(timeout));
      typingTimeoutRef.current.clear();
    };
  }, [token, addMessage, updateMessageStatus, setTypingUser, clearTypingUsers]);

  const subscribeToChat = useCallback((chatId: string) => {
    wsService.subscribeChat(chatId);
  }, []);

  const unsubscribeFromChat = useCallback((chatId: string) => {
    wsService.unsubscribeChat(chatId);
  }, []);

  const startTyping = useCallback((chatId: string) => {
    wsService.startTyping(chatId);
  }, []);

  const stopTyping = useCallback((chatId: string) => {
    wsService.stopTyping(chatId);
  }, []);

  const markMessageRead = useCallback((messageId: string) => {
    wsService.markMessageRead(messageId);
  }, []);

  const markChatRead = useCallback((chatId: string) => {
    wsService.markChatRead(chatId);
  }, []);

  return {
    isConnected: wsService.isConnected(),
    subscribeToChat,
    unsubscribeFromChat,
    startTyping,
    stopTyping,
    markMessageRead,
    markChatRead,
  };
}
