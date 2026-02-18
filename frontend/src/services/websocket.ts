import {
  WSMessage,
  SendMessagePayload,
  MessagePayload,
  TypingPayload,
  UserStatusPayload,
  ErrorPayload,
  TypingStatusPayload,
} from '../types';

type WSMessageType = WSMessage['type'];
type MessageHandler = (payload: unknown) => void;

export class WebSocketService {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private handlers: Map<WSMessageType, Set<MessageHandler>> = new Map();
  private messageQueue: WSMessage[] = [];
  private pingInterval: number | null = null;
  private reconnectTimeout: number | null = null;

  constructor() {}

  connect(token: string): Promise<void> {
    return new Promise((resolve, reject) => {
      const wsUrl = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/api/v1/ws?token=${token}`;

      try {
        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
          console.log('WebSocket connected');
          this.reconnectAttempts = 0;
          this.startPing();
          this.flushMessageQueue();
          resolve();
        };

        this.ws.onmessage = (event) => {
          const messages = event.data.split('\n');
          messages.forEach((msgStr: string) => {
            try {
              const message: WSMessage = JSON.parse(msgStr);
              this.handleMessage(message);
            } catch (e) {
              console.error('Failed to parse WebSocket message:', e);
            }
          });
        };

        this.ws.onclose = (event) => {
          console.log('WebSocket closed:', event.code, event.reason);
          this.stopPing();
          this.reconnect(token);
        };

        this.ws.onerror = (error) => {
          console.error('WebSocket error:', error);
          reject(error);
        };
      } catch (error) {
        reject(error);
      }
    });
  }

  disconnect() {
    this.stopPing();
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }
    if (this.ws) {
      this.ws.close(1000, 'Client disconnect');
      this.ws = null;
    }
  }

  private reconnect(token: string) {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached');
      this.notifyHandlers('auth_error', { code: 'max_reconnect', message: 'Connection lost' });
      return;
    }

    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts);
    this.reconnectAttempts++;

    console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);

    this.reconnectTimeout = window.setTimeout(() => {
      this.connect(token).catch(console.error);
    }, delay);
  }

  private startPing() {
    this.pingInterval = window.setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({ type: 'ping', timestamp: new Date().toISOString() }));
      }
    }, 30000);
  }

  private stopPing() {
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
    }
  }

  private handleMessage(message: WSMessage) {
    const handlers = this.handlers.get(message.type);
    if (handlers) {
      handlers.forEach((handler) => handler(message.payload));
    }

    // Handle errors
    if (message.type === 'error' || message.type === 'auth_error') {
      console.error('WebSocket error:', message.payload);
    }
  }

  private notifyHandlers(type: WSMessageType, payload: unknown) {
    const handlers = this.handlers.get(type);
    if (handlers) {
      handlers.forEach((handler) => handler(payload));
    }
  }

  on(type: WSMessageType, handler: MessageHandler) {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set());
    }
    this.handlers.get(type)!.add(handler);

    // Return unsubscribe function
    return () => {
      this.handlers.get(type)?.delete(handler);
    };
  }

  send(type: WSMessageType, payload: unknown) {
    const message: WSMessage = {
      type,
      payload,
      timestamp: new Date().toISOString(),
    };

    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      // Queue message for later
      this.messageQueue.push(message);
      if (this.messageQueue.length > 100) {
        this.messageQueue.shift();
      }
    }
  }

  private flushMessageQueue() {
    while (this.messageQueue.length > 0 && this.ws?.readyState === WebSocket.OPEN) {
      const message = this.messageQueue.shift();
      if (message) {
        this.ws.send(JSON.stringify(message));
      }
    }
  }

  // Convenience methods
  sendMessage(chatId: string, content: string, messageType = 'text', mediaUrl: string | null = null, replyToId: string | null = null) {
    const payload: SendMessagePayload = {
      chat_id: chatId,
      content,
      message_type: messageType,
      media_url: mediaUrl,
      reply_to_id: replyToId,
    };
    this.send('send_message', payload);
  }

  subscribeChat(chatId: string) {
    this.send('subscribe_chat', { chat_id: chatId });
  }

  unsubscribeChat(chatId: string) {
    this.send('unsubscribe_chat', { chat_id: chatId });
  }

  startTyping(chatId: string) {
    const payload: TypingPayload = { chat_id: chatId, is_typing: true };
    this.send('typing_start', payload);
  }

  stopTyping(chatId: string) {
    const payload: TypingPayload = { chat_id: chatId, is_typing: false };
    this.send('typing_stop', payload);
  }

  markMessageRead(messageId: string) {
    this.send('read_message', { message_id: messageId });
  }

  markChatRead(chatId: string) {
    this.send('read_chat', { chat_id: chatId });
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

// Singleton instance
export const wsService = new WebSocketService();
export default wsService;
