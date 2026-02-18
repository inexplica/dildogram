package repository

import (
	"context"

	"dildogram/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessageRepository определяет интерфейс для работы с сообщениями
type MessageRepository interface {
	Create(ctx context.Context, message *models.Message) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Message, error)
	GetChatMessages(ctx context.Context, chatID uuid.UUID, limit, offset int) ([]models.Message, error)
	Update(ctx context.Context, message *models.Message) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.MessageStatus) error
	MarkAsRead(ctx context.Context, chatID, userID uuid.UUID) error
	GetUnreadCount(ctx context.Context, chatID, userID uuid.UUID) (int64, error)
	MarkChatAsRead(ctx context.Context, chatID, userID uuid.UUID) error
}

type messageRepository struct {
	db *gorm.DB
}

// NewMessageRepository создаёт новый MessageRepository
func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, message *models.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *messageRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Message, error) {
	var message models.Message
	err := r.db.WithContext(ctx).
		Preload("Sender").
		Preload("Reads").
		First(&message, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &message, nil
}

func (r *messageRepository) GetChatMessages(ctx context.Context, chatID uuid.UUID, limit, offset int) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.WithContext(ctx).
		Preload("Sender").
		Preload("Reads").
		Where("chat_id = ? AND is_deleted = false", chatID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error

	// Реверсируем порядок для хронологического отображения
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, err
}

func (r *messageRepository) Update(ctx context.Context, message *models.Message) error {
	return r.db.WithContext(ctx).Save(message).Error
}

func (r *messageRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.MessageStatus) error {
	return r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *messageRepository) MarkAsRead(ctx context.Context, chatID, userID uuid.UUID) error {
	tx := r.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Получаем все непрочитанные сообщения в чате от других пользователей
	var messages []models.Message
	err := tx.Where("chat_id = ? AND sender_id != ? AND is_deleted = false", chatID, userID).
		Find(&messages).Error
	if err != nil {
		return err
	}

	// Для каждого сообщения создаём запись о прочтении
	for _, msg := range messages {
		// Проверяем, не прочитано ли уже
		var existing models.MessageRead
		result := tx.Where("message_id = ? AND user_id = ?", msg.ID, userID).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			read := models.MessageRead{
				MessageID: msg.ID,
				UserID:    userID,
			}
			if err := tx.Create(&read).Error; err != nil {
				return err
			}
		}
	}

	// Обновляем статус последних сообщений на "read"
	if len(messages) > 0 {
		var messageIDs []uuid.UUID
		for _, msg := range messages {
			messageIDs = append(messageIDs, msg.ID)
		}
		tx.Model(&models.Message{}).
			Where("id IN ? AND status != ?", messageIDs, models.MessageStatusRead).
			Update("status", models.MessageStatusRead)
	}

	return tx.Commit().Error
}

func (r *messageRepository) GetUnreadCount(ctx context.Context, chatID, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Joins("LEFT JOIN message_reads mr ON messages.id = mr.message_id AND mr.user_id = ?", userID).
		Where("messages.chat_id = ? AND messages.sender_id != ? AND messages.is_deleted = false AND mr.read_at IS NULL",
			chatID, userID).
		Count(&count).Error
	return count, err
}

func (r *messageRepository) MarkChatAsRead(ctx context.Context, chatID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Получаем все сообщения в чате от других пользователей
		var messages []models.Message
		err := tx.Where("chat_id = ? AND sender_id != ? AND is_deleted = false", chatID, userID).
			Find(&messages).Error
		if err != nil {
			return err
		}

		// Создаём записи о прочтении для каждого сообщения
		for _, msg := range messages {
			read := models.MessageRead{
				MessageID: msg.ID,
				UserID:    userID,
			}
			// Используем OnConflict для предотвращения дубликатов
			tx.Clauses(gorm.OnConflict{
				Columns:   []gorm.Column{{Name: "message_id"}, {Name: "user_id"}},
				DoNothing: true,
			}).Create(&read)
		}

		return nil
	})
}
