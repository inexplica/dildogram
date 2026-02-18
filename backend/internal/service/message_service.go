package service

import (
	"context"
	"errors"
	"time"

	"dildogram/backend/internal/models"
	"dildogram/backend/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrMessageNotFound = errors.New("message not found")
	ErrEmptyContent    = errors.New("message content cannot be empty")
)

// MessageService предоставляет методы для работы с сообщениями
type MessageService struct {
	messageRepo repository.MessageRepository
	chatRepo    repository.ChatRepository
}

// NewMessageService создаёт новый MessageService
func NewMessageService(messageRepo repository.MessageRepository, chatRepo repository.ChatRepository) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		chatRepo:    chatRepo,
	}
}

// SendMessage отправляет сообщение в чат
func (s *MessageService) SendMessage(ctx context.Context, chatID, senderID uuid.UUID, content string, messageType models.MessageType, mediaURL *string, replyToID *uuid.UUID) (*models.Message, error) {
	if content == "" && messageType == models.MessageTypeText {
		return nil, ErrEmptyContent
	}

	// Проверяем существование чата и доступ
	isMember, err := s.chatRepo.IsMember(ctx, chatID, senderID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotMember
	}

	// Создаём сообщение
	message := &models.Message{
		ChatID:      chatID,
		SenderID:    senderID,
		Content:     content,
		MessageType: messageType,
		MediaURL:    mediaURL,
		ReplyToID:   replyToID,
		Status:      models.MessageStatusSent,
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, err
	}

	// Загружаем отправителя
	message.Sender, _ = s.messageRepo.GetByID(ctx, message.ID)
	if message.Sender != nil {
		sender, _ := s.chatRepo.GetMember(ctx, chatID, senderID)
		if sender != nil {
			message.Sender.Sender = sender.User
		}
	}

	return message, nil
}

// GetMessages получает историю сообщений чата
func (s *MessageService) GetMessages(ctx context.Context, chatID, userID uuid.UUID, limit, offset int) ([]models.Message, error) {
	// Проверяем доступ
	isMember, err := s.chatRepo.IsMember(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotMember
	}

	return s.messageRepo.GetChatMessages(ctx, chatID, limit, offset)
}

// GetMessage получает сообщение по ID
func (s *MessageService) GetMessage(ctx context.Context, messageID, userID uuid.UUID) (*models.Message, error) {
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return nil, err
	}
	if message == nil {
		return nil, ErrMessageNotFound
	}

	// Проверяем доступ к чату
	isMember, err := s.chatRepo.IsMember(ctx, message.ChatID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotMember
	}

	return message, nil
}

// UpdateMessage обновляет сообщение
func (s *MessageService) UpdateMessage(ctx context.Context, messageID, userID uuid.UUID, content string) (*models.Message, error) {
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return nil, err
	}
	if message == nil {
		return nil, ErrMessageNotFound
	}

	// Только отправитель может редактировать
	if message.SenderID != userID {
		return nil, ErrNoPermission
	}

	message.Content = content
	message.IsEdited = true

	if err := s.messageRepo.Update(ctx, message); err != nil {
		return nil, err
	}

	return message, nil
}

// DeleteMessage удаляет сообщение
func (s *MessageService) DeleteMessage(ctx context.Context, messageID, userID uuid.UUID) error {
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return err
	}
	if message == nil {
		return ErrMessageNotFound
	}

	// Только отправитель или админ чата может удалить
	isMember, err := s.chatRepo.GetMember(ctx, message.ChatID, userID)
	if err != nil {
		return err
	}
	if isMember == nil {
		return ErrNotMember
	}

	canDelete := message.SenderID == userID ||
		isMember.Role == models.MemberRoleOwner ||
		isMember.Role == models.MemberRoleAdmin

	if !canDelete {
		return ErrNoPermission
	}

	message.IsDeleted = true
	message.Content = "This message was deleted"

	return s.messageRepo.Update(ctx, message)
}

// MarkAsRead отмечает сообщение как прочитанное
func (s *MessageService) MarkAsRead(ctx context.Context, messageID, userID uuid.UUID) error {
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return err
	}
	if message == nil {
		return ErrMessageNotFound
	}

	// Не можем отметить своё сообщение
	if message.SenderID == userID {
		return nil
	}

	// Создаём запись о прочтении
	read := models.MessageRead{
		MessageID: messageID,
		UserID:    userID,
	}

	// Используем transaction для предотвращения дубликатов
	return s.messageRepo.MarkChatAsRead(ctx, message.ChatID, userID)
}

// MarkChatAsRead отмечает все сообщения в чате как прочитанные
func (s *MessageService) MarkChatAsRead(ctx context.Context, chatID, userID uuid.UUID) error {
	isMember, err := s.chatRepo.IsMember(ctx, chatID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return ErrNotMember
	}

	return s.messageRepo.MarkChatAsRead(ctx, chatID, userID)
}

// GetUnreadCount получает количество непрочитанных сообщений
func (s *MessageService) GetUnreadCount(ctx context.Context, chatID, userID uuid.UUID) (int64, error) {
	isMember, err := s.chatRepo.IsMember(ctx, chatID, userID)
	if err != nil {
		return 0, err
	}
	if !isMember {
		return 0, ErrNotMember
	}

	return s.messageRepo.GetUnreadCount(ctx, chatID, userID)
}

// UpdateMessageStatus обновляет статус сообщения
func (s *MessageService) UpdateMessageStatus(ctx context.Context, messageID uuid.UUID, status models.MessageStatus) error {
	return s.messageRepo.UpdateStatus(ctx, messageID, status)
}

// DeliverMessage обновляет статус сообщения на "delivered"
func (s *MessageService) DeliverMessage(ctx context.Context, messageID uuid.UUID) error {
	return s.messageRepo.UpdateStatus(ctx, messageID, models.MessageStatusDelivered)
}

// ReadMessage обновляет статус сообщения на "read"
func (s *MessageService) ReadMessage(ctx context.Context, messageID uuid.UUID) error {
	return s.messageRepo.UpdateStatus(ctx, messageID, models.MessageStatusRead)
}

// BroadcastReadStatus обновляет статусы всех сообщений от пользователя
func (s *MessageService) BroadcastReadStatus(ctx context.Context, chatID, readerID uuid.UUID) error {
	return s.messageRepo.MarkAsRead(ctx, chatID, readerID)
}
