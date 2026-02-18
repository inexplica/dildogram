import { create } from 'zustand';
import { Chat, ChatWithLastMessage, Message, User, ChatMembership } from '../types';
import { api } from '../services/api';
import { wsService } from '../services/websocket';

interface TypingUser {
  userId: string;
  userName: string;
  chatId: string;
  timeout: NodeJS.Timeout;
}

interface ChatState {
  chats: ChatWithLastMessage[];
  currentChat: Chat | null;
  messages: Map<string, Message[]>; // Map<chatId, messages>
  members: Map<string, ChatMembership[]>; // Map<chatId, members>
  typingUsers: Map<string, TypingUser[]>; // Map<chatId, typing users>
  isLoading: boolean;
  error: string | null;

  // Actions
  fetchChats: () => Promise<void>;
  selectChat: (chatId: string) => Promise<void>;
  createChat: (type: 'private' | 'group', name?: string, description?: string, memberIds?: string[]) => Promise<Chat>;
  deleteChat: (chatId: string) => Promise<void>;
  addMember: (chatId: string, userId: string) => Promise<void>;
  removeMember: (chatId: string, userId: string) => Promise<void>;
  fetchMembers: (chatId: string) => Promise<void>;
  sendMessage: (chatId: string, content: string, messageType?: string, mediaUrl?: string | null, replyToId?: string | null) => void;
  addMessage: (chatId: string, message: Message) => void;
  updateMessageStatus: (messageId: string, status: string) => void;
  markChatAsRead: (chatId: string) => Promise<void>;
  setTypingUser: (chatId: string, user: User, isTyping: boolean) => void;
  clearTypingUsers: (chatId: string) => void;
  subscribeToChat: (chatId: string) => void;
  unsubscribeFromChat: (chatId: string) => void;
  clearError: () => void;
}

export const useChatStore = create<ChatState>((set, get) => ({
  chats: [],
  currentChat: null,
  messages: new Map(),
  members: new Map(),
  typingUsers: new Map(),
  isLoading: false,
  error: null,

  fetchChats: async () => {
    set({ isLoading: true, error: null });
    try {
      const { chats } = await api.getChats();
      set({ chats, isLoading: false });
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Failed to fetch chats';
      set({ error: message, isLoading: false });
    }
  },

  selectChat: async (chatId: string) => {
    try {
      const { chat } = await api.getChat(chatId);
      const { messages } = await api.getMessages(chatId);
      
      set((state) => ({
        currentChat: chat,
        messages: new Map(state.messages).set(chatId, messages),
      }));

      // Mark as read
      await api.markChatAsRead(chatId);
      wsService.markChatRead(chatId);
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Failed to select chat';
      set({ error: message });
    }
  },

  createChat: async (type, name, description, memberIds) => {
    try {
      const { chat } = await api.createChat({ type, name, description, member_ids: memberIds });
      set((state) => ({
        chats: [chat, ...state.chats],
      }));
      return chat;
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Failed to create chat';
      set({ error: message });
      throw error;
    }
  },

  deleteChat: async (chatId: string) => {
    try {
      await api.deleteChat(chatId);
      set((state) => ({
        chats: state.chats.filter((c) => c.id !== chatId),
        currentChat: state.currentChat?.id === chatId ? null : state.currentChat,
      }));
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Failed to delete chat';
      set({ error: message });
      throw error;
    }
  },

  addMember: async (chatId: string, userId: string) => {
    try {
      await api.addMember(chatId, userId);
      await get().fetchMembers(chatId);
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Failed to add member';
      set({ error: message });
      throw error;
    }
  },

  removeMember: async (chatId: string, userId: string) => {
    try {
      await api.removeMember(chatId, userId);
      await get().fetchMembers(chatId);
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Failed to remove member';
      set({ error: message });
      throw error;
    }
  },

  fetchMembers: async (chatId: string) => {
    try {
      const { members } = await api.getMembers(chatId);
      set((state) => ({
        members: new Map(state.members).set(chatId, members),
      }));
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Failed to fetch members';
      set({ error: message });
    }
  },

  sendMessage: (chatId, content, messageType = 'text', mediaUrl = null, replyToId = null) => {
    wsService.sendMessage(chatId, content, messageType, mediaUrl, replyToId);
  },

  addMessage: (chatId, message) => {
    set((state) => {
      const chatMessages = state.messages.get(chatId) || [];
      const updatedMessages = [...chatMessages, message];
      const newMessages = new Map(state.messages).set(chatId, updatedMessages);

      // Update chat list with last message
      const updatedChats = state.chats.map((chat) =>
        chat.id === chatId
          ? {
              ...chat,
              last_message_content: message.content,
              last_message_created_at: message.created_at,
              last_message_sender_id: message.sender_id,
              last_message_status: message.status,
            }
          : chat
      );

      return {
        messages: newMessages,
        chats: updatedChats,
      };
    });
  },

  updateMessageStatus: (messageId, status) => {
    set((state) => {
      const newMessages = new Map(state.messages);
      newMessages.forEach((messages, chatId) => {
        newMessages.set(
          chatId,
          messages.map((msg) => (msg.id === messageId ? { ...msg, status: status as Message['status'] } : msg))
        );
      });
      return { messages: newMessages };
    });
  },

  markChatAsRead: async (chatId: string) => {
    try {
      await api.markChatAsRead(chatId);
      set((state) => ({
        chats: state.chats.map((chat) =>
          chat.id === chatId ? { ...chat, unread_count: 0 } : chat
        ),
      }));
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Failed to mark as read';
      set({ error: message });
    }
  },

  setTypingUser: (chatId, user, isTyping) => {
    set((state) => {
      const chatTypingUsers = state.typingUsers.get(chatId) || [];
      let updatedTypingUsers;

      if (isTyping) {
        const existing = chatTypingUsers.find((u) => u.userId === user.id);
        if (existing) {
          clearTimeout(existing.timeout);
          updatedTypingUsers = chatTypingUsers;
        } else {
          const timeout = setTimeout(() => {
            get().setTypingUser(chatId, user, false);
          }, 3000);
          updatedTypingUsers = [
            ...chatTypingUsers,
            { userId: user.id, userName: user.username, chatId, timeout },
          ];
        }
      } else {
        const filtered = chatTypingUsers.filter((u) => u.userId !== user.id);
        chatTypingUsers.forEach((u) => clearTimeout(u.timeout));
        updatedTypingUsers = filtered;
      }

      return {
        typingUsers: new Map(state.typingUsers).set(chatId, updatedTypingUsers),
      };
    });
  },

  clearTypingUsers: (chatId) => {
    set((state) => {
      const chatTypingUsers = state.typingUsers.get(chatId) || [];
      chatTypingUsers.forEach((u) => clearTimeout(u.timeout));
      return {
        typingUsers: new Map(state.typingUsers).set(chatId, []),
      };
    });
  },

  subscribeToChat: (chatId: string) => {
    wsService.subscribeChat(chatId);
  },

  unsubscribeFromChat: (chatId: string) => {
    wsService.unsubscribeChat(chatId);
  },

  clearError: () => set({ error: null }),
}));
