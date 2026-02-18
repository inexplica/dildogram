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
	ErrChatNotFound     = errors.New("chat not found")
	ErrChatExists       = errors.New("chat already exists with these users")
	ErrNotMember        = errors.New("user is not a member of this chat")
	ErrNoPermission     = errors.New("no permission to perform this action")
	ErrCannotAddSelf    = errors.New("cannot add yourself to chat")
	ErrCannotRemoveOwner = errors.New("cannot remove chat owner")
)

// ChatService предоставляет методы для управления чатами
type ChatService struct {
	chatRepo repository.ChatRepository
	userRepo repository.UserRepository
}

// NewChatService создаёт новый ChatService
func NewChatService(chatRepo repository.ChatRepository, userRepo repository.UserRepository) *ChatService {
	return &ChatService{
		chatRepo: chatRepo,
		userRepo: userRepo,
	}
}

// CreatePrivateChat создаёт личный чат между двумя пользователями
func (s *ChatService) CreatePrivateChat(ctx context.Context, userID, otherUserID uuid.UUID) (*models.Chat, error) {
	// Проверяем существование чата
	existingChat, err := s.chatRepo.FindPrivateChat(ctx, userID, otherUserID)
	if err != nil {
		return nil, err
	}
	if existingChat != nil {
		return existingChat, nil
	}

	// Проверяем существование другого пользователя
	otherUser, err := s.userRepo.GetByID(ctx, otherUserID)
	if err != nil {
		return nil, err
	}
	if otherUser == nil {
		return nil, ErrUserNotFound
	}

	// Создаём чат
	chat := &models.Chat{
		Type:      models.ChatTypePrivate,
		CreatedBy: userID,
		Name:      otherUser.GetFullName(),
	}

	if err := s.chatRepo.Create(ctx, chat); err != nil {
		return nil, err
	}

	// Добавляем создателя
	ownerMembership := &models.ChatMembership{
		ChatID: chat.ID,
		UserID: userID,
		Role:   models.MemberRoleOwner,
	}
	if err := s.chatRepo.AddMember(ctx, ownerMembership); err != nil {
		return nil, err
	}

	// Добавляем второго участника
	memberMembership := &models.ChatMembership{
		ChatID: chat.ID,
		UserID: otherUserID,
		Role:   models.MemberRoleMember,
	}
	if err := s.chatRepo.AddMember(ctx, memberMembership); err != nil {
		return nil, err
	}

	return chat, nil
}

// CreateGroupChat создаёт групповой чат
func (s *ChatService) CreateGroupChat(ctx context.Context, userID uuid.UUID, name, description string, memberIDs []uuid.UUID) (*models.Chat, error) {
	// Создаём чат
	chat := &models.Chat{
		Type:        models.ChatTypeGroup,
		Name:        name,
		Description: description,
		CreatedBy:   userID,
	}

	if err := s.chatRepo.Create(ctx, chat); err != nil {
		return nil, err
	}

	// Добавляем создателя как владельца
	ownerMembership := &models.ChatMembership{
		ChatID: chat.ID,
		UserID: userID,
		Role:   models.MemberRoleOwner,
	}
	if err := s.chatRepo.AddMember(ctx, ownerMembership); err != nil {
		return nil, err
	}

	// Добавляем остальных участников
	for _, memberID := range memberIDs {
		if memberID == userID {
			continue // Пропускаем создателя
		}

		membership := &models.ChatMembership{
			ChatID: chat.ID,
			UserID: memberID,
			Role:   models.MemberRoleMember,
		}
		if err := s.chatRepo.AddMember(ctx, membership); err != nil {
			return nil, err
		}
	}

	return chat, nil
}

// GetChat получает чат по ID
func (s *ChatService) GetChat(ctx context.Context, chatID, userID uuid.UUID) (*models.Chat, error) {
	chat, err := s.chatRepo.GetByID(ctx, chatID)
	if err != nil {
		return nil, err
	}
	if chat == nil {
		return nil, ErrChatNotFound
	}

	// Проверяем доступ
	isMember, err := s.chatRepo.IsMember(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotMember
	}

	return chat, nil
}

// GetUserChats получает список чатов пользователя
func (s *ChatService) GetUserChats(ctx context.Context, userID uuid.UUID) ([]models.ChatWithLastMessage, error) {
	return s.chatRepo.GetUserChats(ctx, userID)
}

// UpdateChat обновляет чат
func (s *ChatService) UpdateChat(ctx context.Context, chatID, userID uuid.UUID, name, description, avatarURL string) (*models.Chat, error) {
	chat, err := s.chatRepo.GetByID(ctx, chatID)
	if err != nil {
		return nil, err
	}
	if chat == nil {
		return nil, ErrChatNotFound
	}

	// Проверяем права (только админ и владелец)
	membership, err := s.chatRepo.GetMember(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}
	if membership == nil || (membership.Role != models.MemberRoleOwner && membership.Role != models.MemberRoleAdmin) {
		return nil, ErrNoPermission
	}

	// Обновляем поля
	if name != "" {
		chat.Name = name
	}
	if description != "" {
		chat.Description = description
	}
	if avatarURL != "" {
		chat.AvatarURL = avatarURL
	}

	if err := s.chatRepo.Update(ctx, chat); err != nil {
		return nil, err
	}

	return chat, nil
}

// DeleteChat удаляет чат
func (s *ChatService) DeleteChat(ctx context.Context, chatID, userID uuid.UUID) error {
	chat, err := s.chatRepo.GetByID(ctx, chatID)
	if err != nil {
		return err
	}
	if chat == nil {
		return ErrChatNotFound
	}

	// Только владелец может удалить чат
	membership, err := s.chatRepo.GetMember(ctx, chatID, userID)
	if err != nil {
		return err
	}
	if membership == nil || membership.Role != models.MemberRoleOwner {
		return ErrNoPermission
	}

	return s.chatRepo.Delete(ctx, chatID)
}

// AddMember добавляет участника в чат
func (s *ChatService) AddMember(ctx context.Context, chatID, userID, newMemberID uuid.UUID) error {
	if userID == newMemberID {
		return ErrCannotAddSelf
	}

	chat, err := s.chatRepo.GetByID(ctx, chatID)
	if err != nil {
		return err
	}
	if chat == nil {
		return ErrChatNotFound
	}

	// Только владелец и админ могут добавлять
	membership, err := s.chatRepo.GetMember(ctx, chatID, userID)
	if err != nil {
		return err
	}
	if membership == nil || (membership.Role != models.MemberRoleOwner && membership.Role != models.MemberRoleAdmin) {
		return ErrNoPermission
	}

	// Проверяем, не состоит ли уже
	existing, err := s.chatRepo.GetMember(ctx, chatID, newMemberID)
	if err != nil {
		return err
	}
	if existing != nil && existing.LeftAt == nil {
		return ErrChatExists
	}

	// Добавляем участника
	newMembership := &models.ChatMembership{
		ChatID: chatID,
		UserID: newMemberID,
		Role:   models.MemberRoleMember,
	}
	return s.chatRepo.AddMember(ctx, newMembership)
}

// RemoveMember удаляет участника из чата
func (s *ChatService) RemoveMember(ctx context.Context, chatID, userID, removeMemberID uuid.UUID) error {
	chat, err := s.chatRepo.GetByID(ctx, chatID)
	if err != nil {
		return err
	}
	if chat == nil {
		return ErrChatNotFound
	}

	// Проверяем права
	membership, err := s.chatRepo.GetMember(ctx, chatID, userID)
	if err != nil {
		return err
	}
	if membership == nil {
		return ErrNotMember
	}

	// Проверяем удаляемого
	removeMembership, err := s.chatRepo.GetMember(ctx, chatID, removeMemberID)
	if err != nil {
		return err
	}
	if removeMembership == nil {
		return ErrNotMember
	}

	// Нельзя удалить владельца
	if removeMembership.Role == models.MemberRoleOwner {
		return ErrCannotRemoveOwner
	}

	// Владелец и админ могут удалять участников
	if membership.Role == models.MemberRoleOwner || membership.Role == models.MemberRoleAdmin {
		return s.chatRepo.RemoveMember(ctx, chatID, removeMemberID)
	}

	// Пользователь может удалить сам себя
	if userID == removeMemberID {
		return s.chatRepo.RemoveMember(ctx, chatID, userID)
	}

	return ErrNoPermission
}

// GetMembers получает список участников чата
func (s *ChatService) GetMembers(ctx context.Context, chatID, userID uuid.UUID) ([]models.ChatMembership, error) {
	// Проверяем доступ
	isMember, err := s.chatRepo.IsMember(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotMember
	}

	return s.chatRepo.GetMembers(ctx, chatID)
}

// LeaveChat покидает чат
func (s *ChatService) LeaveChat(ctx context.Context, chatID, userID uuid.UUID) error {
	membership, err := s.chatRepo.GetMember(ctx, chatID, userID)
	if err != nil {
		return err
	}
	if membership == nil {
		return ErrNotMember
	}

	// Владелец не может покинуть чат, должен передать права
	if membership.Role == models.MemberRoleOwner {
		return ErrCannotRemoveOwner
	}

	return s.chatRepo.RemoveMember(ctx, chatID, userID)
}

// MarkChatRead отмечает все сообщения в чате как прочитанные
func (s *ChatService) MarkChatRead(ctx context.Context, chatID, userID uuid.UUID) error {
	isMember, err := s.chatRepo.IsMember(ctx, chatID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return ErrNotMember
	}

	// Обновляем время last_seen пользователя
	return s.userRepo.SetOnline(ctx, userID, true)
}
