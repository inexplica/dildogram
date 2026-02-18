package models

import (
	"time"

	"github.com/google/uuid"
)

// ChatType определяет тип чата
type ChatType string

const (
	ChatTypePrivate ChatType = "private"
	ChatTypeGroup   ChatType = "group"
)

// MemberRole определяет роль участника
type MemberRole string

const (
	MemberRoleOwner  MemberRole = "owner"
	MemberRoleAdmin  MemberRole = "admin"
	MemberRoleMember MemberRole = "member"
)

// Chat представляет чат (личный или групповой)
type Chat struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Type          ChatType   `gorm:"size:20;not null" json:"type"`
	Name          string     `gorm:"size:100;not null;default:''" json:"name"`
	Description   string     `gorm:"type:text;not null;default:''" json:"description"`
	AvatarURL     string     `gorm:"size:500;not null;default:''" json:"avatar_url"`
	CreatedBy     uuid.UUID  `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt     time.Time  `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"not null;default:now()" json:"updated_at"`
	LastMessageAt *time.Time `gorm:"index" json:"last_message_at"`
	DeletedAt     *time.Time `gorm:"index" json:"-"`

	// Связи
	Creator   *User            `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Members   []ChatMembership `gorm:"foreignKey:ChatID" json:"members,omitempty"`
	Messages  []Message        `gorm:"foreignKey:ChatID" json:"-"`
}

// TableName возвращает имя таблицы
func (Chat) TableName() string {
	return "chats"
}

// ChatMembership представляет участника чата
type ChatMembership struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	ChatID    uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_chat_user" json:"chat_id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_chat_user" json:"user_id"`
	Role      MemberRole `gorm:"size:20;not null;default:'member'" json:"role"`
	JoinedAt  time.Time  `gorm:"not null;default:now()" json:"joined_at"`
	LeftAt    *time.Time `gorm:"index" json:"left_at"`

	// Связи
	Chat *Chat `gorm:"foreignKey:ChatID" json:"-"`
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName возвращает имя таблицы
func (ChatMembership) TableName() string {
	return "chat_members"
}

// IsActive проверяет, активен ли участник
func (m *ChatMembership) IsActive() bool {
	return m.LeftAt == nil
}

// ChatWithLastMessage представляет чат с последним сообщением
type ChatWithLastMessage struct {
	Chat
	LastMessageID     *uuid.UUID `json:"last_message_id"`
	LastMessageContent *string   `json:"last_message_content"`
	LastMessageSenderID *uuid.UUID `json:"last_message_sender_id"`
	LastMessageCreatedAt *time.Time `json:"last_message_created_at"`
	LastMessageStatus *string    `json:"last_message_status"`
	UnreadCount       int64      `json:"unread_count"`
}
