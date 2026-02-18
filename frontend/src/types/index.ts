// User types
export interface User {
  id: string;
  phone: string;
  username: string;
  first_name: string;
  last_name: string;
  bio: string;
  avatar_url: string;
  is_active: boolean;
  is_online: boolean;
  last_seen: string;
  created_at: string;
  updated_at: string;
}

export interface UserPreview {
  id: string;
  username: string;
  first_name: string;
  last_name: string;
  avatar_url: string;
  is_online: boolean;
}

// Chat types
export type ChatType = 'private' | 'group';

export interface Chat {
  id: string;
  type: ChatType;
  name: string;
  description: string;
  avatar_url: string;
  created_by: string;
  created_at: string;
  updated_at: string;
  last_message_at: string | null;
}

export interface ChatWithLastMessage extends Chat {
  last_message_id: string | null;
  last_message_content: string | null;
  last_message_sender_id: string | null;
  last_message_created_at: string | null;
  last_message_status: string | null;
  unread_count: number;
}

export interface ChatMembership {
  id: string;
  chat_id: string;
  user_id: string;
  role: 'owner' | 'admin' | 'member';
  joined_at: string;
  left_at: string | null;
  user?: UserPreview;
}

// Message types
export type MessageType = 'text' | 'image' | 'file' | 'voice';
export type MessageStatus = 'pending' | 'sent' | 'delivered' | 'read';

export interface Message {
  id: string;
  chat_id: string;
  sender_id: string;
  content: string;
  message_type: MessageType;
  media_url: string | null;
  reply_to_id: string | null;
  is_edited: boolean;
  is_deleted: boolean;
  status: MessageStatus;
  created_at: string;
  updated_at: string;
  sender?: UserPreview;
}

// Auth types
export interface LoginRequest {
  phone: string;
  password: string;
}

export interface RegisterRequest {
  phone: string;
  username: string;
  password: string;
}

export interface SMSRequest {
  phone: string;
}

export interface VerifySMSRequest {
  phone: string;
  code: string;
}

export interface AuthResponse {
  user: User;
  token: string;
}

// API types
export interface ApiResponse<T> {
  data?: T;
  error?: string;
}

export interface ChatListResponse {
  chats: ChatWithLastMessage[];
}

export interface MessageListResponse {
  messages: Message[];
}

export interface UserSearchResponse {
  users: UserPreview[];
}

// WebSocket types
export type WSMessageType =
  | 'send_message'
  | 'read_message'
  | 'read_chat'
  | 'typing_start'
  | 'typing_stop'
  | 'subscribe_chat'
  | 'unsubscribe_chat'
  | 'message'
  | 'message_status'
  | 'message_read'
  | 'typing'
  | 'user_online'
  | 'user_offline'
  | 'chat_updated'
  | 'new_chat'
  | 'error'
  | 'auth_error';

export interface WSMessage {
  type: WSMessageType;
  payload: unknown;
  request_id?: string;
  timestamp: string;
}

export interface SendMessagePayload {
  chat_id: string;
  content: string;
  message_type?: string;
  media_url?: string | null;
  reply_to_id?: string | null;
}

export interface MessagePayload {
  id: string;
  chat_id: string;
  sender_id: string;
  sender_name: string;
  sender_avatar: string;
  content: string;
  message_type: string;
  media_url: string | null;
  reply_to_id: string | null;
  is_edited: boolean;
  is_deleted: boolean;
  status: string;
  created_at: string;
}

export interface TypingPayload {
  chat_id: string;
  is_typing: boolean;
}

export interface TypingStatusPayload {
  chat_id: string;
  user_id: string;
  user_name: string;
  is_typing: boolean;
}

export interface UserStatusPayload {
  user_id: string;
  username: string;
  is_online: boolean;
  last_seen?: string;
}

export interface ErrorPayload {
  code: string;
  message: string;
}
