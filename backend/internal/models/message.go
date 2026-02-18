package models

import (
	"time"

	"github.com/google/uuid"
)

// MessageType определяет тип сообщения
type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeImage MessageType = "image"
	MessageTypeFile  MessageType = "file"
	MessageTypeVoice MessageType = "voice"
)

// MessageStatus определяет статус сообщения
type MessageStatus string

const (
	MessageStatusPending  MessageStatus = "pending"
	MessageStatusSent     MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead     MessageStatus = "read"
)

// Message представляет сообщение в чате
type Message struct {
	ID          uuid.UUID    `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	ChatID      uuid.UUID    `gorm:"type:uuid;not null;index:idx_chat_created" json:"chat_id"`
	SenderID    uuid.UUID    `gorm:"type:uuid;not null" json:"sender_id"`
	Content     string       `gorm:"type:text;not null" json:"content"`
	MessageType MessageType  `gorm:"size:20;not null;default:'text'" json:"message_type"`
	MediaURL    *string      `gorm:"size:500" json:"media_url,omitempty"`
	ReplyToID   *uuid.UUID   `gorm:"type:uuid" json:"reply_to_id,omitempty"`
	IsEdited    bool         `gorm:"not null;default:false" json:"is_edited"`
	IsDeleted   bool         `gorm:"not null;default:false;index" json:"is_deleted"`
	Status      MessageStatus `gorm:"size:20;not null;default:'sent';index" json:"status"`
	CreatedAt   time.Time    `gorm:"not null;default:now();index:idx_chat_created" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt   *time.Time   `gorm:"index" json:"-"`

	// Связи
	Chat      *Chat       `gorm:"foreignKey:ChatID" json:"-"`
	Sender    *User       `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	ReplyTo   *Message    `gorm:"foreignKey:ReplyToID" json:"reply_to,omitempty"`
	Reads     []MessageRead `gorm:"foreignKey:MessageID" json:"reads,omitempty"`
}

// TableName возвращает имя таблицы
func (Message) TableName() string {
	return "messages"
}

// MessageRead представляет факт прочтения сообщения
type MessageRead struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	MessageID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_message_user" json:"message_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_message_user" json:"user_id"`
	ReadAt    time.Time `gorm:"not null;default:now()" json:"read_at"`

	// Связи
	Message *Message `gorm:"foreignKey:MessageID" json:"-"`
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName возвращает имя таблицы
func (MessageRead) TableName() string {
	return "message_reads"
}

// MessageWithSender представляет сообщение с данными отправителя
type MessageWithSender struct {
	Message
	SenderUsername  *string `json:"sender_username"`
	SenderFirstName *string `json:"sender_first_name"`
	SenderLastName  *string `json:"sender_last_name"`
	SenderAvatarURL *string `json:"sender_avatar_url"`
}
