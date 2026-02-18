package websocket

import (
	"time"

	"github.com/google/uuid"
)

// MessageType определяет тип WebSocket сообщения
type MessageType string

const (
	// Сообщения от клиента
	MessageTypeSendMessage     MessageType = "send_message"
	MessageTypeReadMessage     MessageType = "read_message"
	MessageTypeReadChat        MessageType = "read_chat"
	MessageTypeTypingStart     MessageType = "typing_start"
	MessageTypeTypingStop      MessageType = "typing_stop"
	MessageTypeSubscribeChat   MessageType = "subscribe_chat"
	MessageTypeUnsubscribeChat MessageType = "unsubscribe_chat"

	// Сообщения от сервера
	MessageTypeMessage       MessageType = "message"
	MessageTypeMessageStatus MessageType = "message_status"
	MessageTypeMessageRead   MessageType = "message_read"
	MessageTypeTyping        MessageType = "typing"
	MessageTypeUserOnline    MessageType = "user_online"
	MessageTypeUserOffline   MessageType = "user_offline"
	MessageTypeChatUpdated   MessageType = "chat_updated"
	MessageTypeNewChat       MessageType = "new_chat"
	MessageTypeError         MessageType = "error"
	MessageTypeAuthError     MessageType = "auth_error"
)

// WSMessage представляет WebSocket сообщение
type WSMessage struct {
	Type      MessageType     `json:"type"`
	Payload   interface{}     `json:"payload,omitempty"`
	RequestID string          `json:"request_id,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// SendMessagePayload payload для отправки сообщения
type SendMessagePayload struct {
	ChatID      string  `json:"chat_id"`
	Content     string  `json:"content"`
	MessageType string  `json:"message_type,omitempty"`
	MediaURL    *string `json:"media_url,omitempty"`
	ReplyToID   *string `json:"reply_to_id,omitempty"`
}

// ReadMessagePayload payload для отметки прочтения сообщения
type ReadMessagePayload struct {
	MessageID string `json:"message_id"`
}

// ReadChatPayload payload для отметки прочтения чата
type ReadChatPayload struct {
	ChatID string `json:"chat_id"`
}

// TypingPayload payload для статуса набора текста
type TypingPayload struct {
	ChatID string `json:"chat_id"`
	IsTyping bool `json:"is_typing"`
}

// SubscribePayload payload для подписки на чат
type SubscribePayload struct {
	ChatID string `json:"chat_id"`
}

// MessagePayload payload с сообщением
type MessagePayload struct {
	ID            string     `json:"id"`
	ChatID        string     `json:"chat_id"`
	SenderID      string     `json:"sender_id"`
	SenderName    string     `json:"sender_name"`
	SenderAvatar  string     `json:"sender_avatar,omitempty"`
	Content       string     `json:"content"`
	MessageType   string     `json:"message_type"`
	MediaURL      *string    `json:"media_url,omitempty"`
	ReplyToID     *string    `json:"reply_to_id,omitempty"`
	IsEdited      bool       `json:"is_edited"`
	IsDeleted     bool       `json:"is_deleted"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
}

// MessageStatusPayload payload со статусом сообщения
type MessageStatusPayload struct {
	MessageID string `json:"message_id"`
	Status    string `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MessageReadPayload payload о прочтении сообщения
type MessageReadPayload struct {
	MessageID string    `json:"message_id"`
	UserID    string    `json:"user_id"`
	ReadAt    time.Time `json:"read_at"`
}

// TypingStatusPayload payload со статусом набора текста
type TypingStatusPayload struct {
	ChatID   string `json:"chat_id"`
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	IsTyping bool   `json:"is_typing"`
}

// UserStatusPayload payload со статусом пользователя
type UserStatusPayload struct {
	UserID   string    `json:"user_id"`
	Username string    `json:"username"`
	IsOnline bool      `json:"is_online"`
	LastSeen time.Time `json:"last_seen,omitempty"`
}

// ChatUpdatedPayload payload об обновлении чата
type ChatUpdatedPayload struct {
	ChatID   string `json:"chat_id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar_url,omitempty"`
	LastMessage *string `json:"last_message,omitempty"`
}

// ErrorPayload payload с ошибкой
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ToMessagePayload конвертирует Message в MessagePayload
func ToMessagePayload(msg interface{}, senderName, senderAvatar string) MessagePayload {
	// Эта функция будет использоваться в handlers
	// Реализация зависит от конкретной модели
	return MessagePayload{}
}

// GenerateRequestID генерирует ID для запроса
func GenerateRequestID() string {
	return uuid.New().String()
}
