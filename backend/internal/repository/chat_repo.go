package repository

import (
	"context"
	"time"

	"dildogram/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ChatRepository определяет интерфейс для работы с чатами
type ChatRepository interface {
	Create(ctx context.Context, chat *models.Chat) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Chat, error)
	GetUserChats(ctx context.Context, userID uuid.UUID) ([]models.ChatWithLastMessage, error)
	Update(ctx context.Context, chat *models.Chat) error
	Delete(ctx context.Context, id uuid.UUID) error
	AddMember(ctx context.Context, membership *models.ChatMembership) error
	RemoveMember(ctx context.Context, chatID, userID uuid.UUID) error
	GetMember(ctx context.Context, chatID, userID uuid.UUID) (*models.ChatMembership, error)
	GetMembers(ctx context.Context, chatID uuid.UUID) ([]models.ChatMembership, error)
	IsMember(ctx context.Context, chatID, userID uuid.UUID) (bool, error)
	FindPrivateChat(ctx context.Context, user1, user2 uuid.UUID) (*models.Chat, error)
}

type chatRepository struct {
	db *gorm.DB
}

// NewChatRepository создаёт новый ChatRepository
func NewChatRepository(db *gorm.DB) ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) Create(ctx context.Context, chat *models.Chat) error {
	return r.db.WithContext(ctx).Create(chat).Error
}

func (r *chatRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Chat, error) {
	var chat models.Chat
	err := r.db.WithContext(ctx).
		Preload("Creator").
		Preload("Members.User").
		First(&chat, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &chat, nil
}

func (r *chatRepository) GetUserChats(ctx context.Context, userID uuid.UUID) ([]models.ChatWithLastMessage, error) {
	var chats []models.ChatWithLastMessage

	query := `
		SELECT 
			c.id,
			c.type,
			c.name,
			c.description,
			c.avatar_url,
			c.created_by,
			c.created_at,
			c.updated_at,
			c.last_message_at,
			c.deleted_at,
			lm.message_id as last_message_id,
			lm.content as last_message_content,
			lm.sender_id as last_message_sender_id,
			lm.created_at as last_message_created_at,
			lm.status as last_message_status,
			COALESCE(ur.unread_count, 0) as unread_count
		FROM chats c
		INNER JOIN chat_members cm ON c.id = cm.chat_id AND cm.left_at IS NULL
		LEFT JOIN LATERAL (
			SELECT id, content, sender_id, created_at, status
			FROM messages
			WHERE chat_id = c.id AND is_deleted = false
			ORDER BY created_at DESC
			LIMIT 1
		) lm ON true
		LEFT JOIN LATERAL (
			SELECT COUNT(*) as unread_count
			FROM messages m
			LEFT JOIN message_reads mr ON m.id = mr.message_id AND mr.user_id = ?
			WHERE m.chat_id = c.id 
				AND m.is_deleted = false 
				AND m.sender_id != ?
				AND mr.read_at IS NULL
		) ur ON true
		WHERE cm.user_id = ?
		ORDER BY COALESCE(c.last_message_at, c.created_at) DESC
	`

	err := r.db.WithContext(ctx).Raw(query, userID, userID, userID).Scan(&chats).Error
	return chats, err
}

func (r *chatRepository) Update(ctx context.Context, chat *models.Chat) error {
	return r.db.WithContext(ctx).Save(chat).Error
}

func (r *chatRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Chat{}).
		Where("id = ?", id).
		Update("deleted_at", now).Error
}

func (r *chatRepository) AddMember(ctx context.Context, membership *models.ChatMembership) error {
	return r.db.WithContext(ctx).Create(membership).Error
}

func (r *chatRepository) RemoveMember(ctx context.Context, chatID, userID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.ChatMembership{}).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Update("left_at", now).Error
}

func (r *chatRepository) GetMember(ctx context.Context, chatID, userID uuid.UUID) (*models.ChatMembership, error) {
	var membership models.ChatMembership
	err := r.db.WithContext(ctx).
		Preload("User").
		First(&membership, "chat_id = ? AND user_id = ?", chatID, userID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &membership, nil
}

func (r *chatRepository) GetMembers(ctx context.Context, chatID uuid.UUID) ([]models.ChatMembership, error) {
	var memberships []models.ChatMembership
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("chat_id = ? AND left_at IS NULL", chatID).
		Find(&memberships).Error
	return memberships, err
}

func (r *chatRepository) IsMember(ctx context.Context, chatID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ChatMembership{}).
		Where("chat_id = ? AND user_id = ? AND left_at IS NULL", chatID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *chatRepository) FindPrivateChat(ctx context.Context, user1, user2 uuid.UUID) (*models.Chat, error) {
	var chat models.Chat

	// Ищем чат где есть оба пользователя
	query := `
		SELECT c.* FROM chats c
		INNER JOIN chat_members cm1 ON c.id = cm1.chat_id AND cm1.user_id = ? AND cm1.left_at IS NULL
		INNER JOIN chat_members cm2 ON c.id = cm2.chat_id AND cm2.user_id = ? AND cm2.left_at IS NULL
		WHERE c.type = 'private' AND c.deleted_at IS NULL
		LIMIT 1
	`

	err := r.db.WithContext(ctx).Raw(query, user1, user2).Scan(&chat).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	if chat.ID == uuid.Nil {
		return nil, nil
	}

	return &chat, nil
}
