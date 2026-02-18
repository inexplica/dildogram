package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"dildogram/backend/internal/models"
	"dildogram/backend/internal/repository"
	"dildogram/backend/internal/service"
	"github.com/google/uuid"
)

// Hub управляет WebSocket соединениями
type Hub struct {
	clients        map[uuid.UUID]*Client // Клиенты по ID пользователя
	clientsByChat  map[uuid.UUID]map[uuid.UUID]*Client // Клиенты по ID чата
	Register       chan *Client
	Unregister     chan *Client
	broadcast      chan broadcastMessage
	broadcastToChat chan chatBroadcastMessage
	mu             sync.RWMutex

	// Сервисы
	messageService *service.MessageService
	chatService    *service.ChatService
	authService    *service.AuthService
	messageRepo    repository.MessageRepository
	chatRepo       repository.ChatRepository
	userRepo       repository.UserRepository
}

// chatSubscriber хранит информацию о подписчике чата
type chatSubscriber struct {
	userID   uuid.UUID
	client   *Client
	chatID   uuid.UUID
}

// NewHub создаёт новый Hub
func NewHub(
	messageService *service.MessageService,
	chatService *service.ChatService,
	authService *service.AuthService,
	messageRepo repository.MessageRepository,
	chatRepo repository.ChatRepository,
	userRepo repository.UserRepository,
) *Hub {
	return &Hub{
		clients:        make(map[uuid.UUID]*Client),
		clientsByChat:  make(map[uuid.UUID]map[uuid.UUID]*Client),
		Register:       make(chan *Client),
		Unregister:     make(chan *Client),
		broadcast:      make(chan broadcastMessage, 256),
		broadcastToChat: make(chan chatBroadcastMessage, 256),
		messageService: messageService,
		chatService:    chatService,
		authService:    authService,
		messageRepo:    messageRepo,
		chatRepo:       chatRepo,
		userRepo:       userRepo,
	}
}

// Run запускает Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case msg := <-h.broadcast:
			h.handleBroadcast(msg)

		case msg := <-h.broadcastToChat:
			h.handleBroadcastToChat(msg)
		}
	}
}

// registerClient регистрирует клиента
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Если уже есть соединение для этого пользователя, закрываем старое
	if existing, ok := h.clients[client.userID]; ok {
		close(existing.send)
	}

	h.clients[client.userID] = client

	// Устанавливаем статус онлайн
	_ = h.authService.SetOnline(context.Background(), client.userID, true)

	// Отправляем уведомление о статусе онлайн
	h.broadcastUserOnline(client.userID, client.username)

	log.Printf("client connected: %s (%s)", client.username, client.userID)
}

// unregisterClient отключает клиента
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client.userID]; ok {
		delete(h.clients, client.userID)
		close(client.send)

		// Отписываем от всех чатов
		for chatID := range client.subscribed {
			h.unsubscribeFromChat(client, chatID)
		}

		// Устанавливаем статус офлайн
		_ = h.authService.SetOnline(context.Background(), client.userID, false)

		// Отправляем уведомление о статусе офлайн
		h.broadcastUserOffline(client.userID)

		log.Printf("client disconnected: %s (%s)", client.username, client.userID)
	}
}

// handleBroadcast обрабатывает широковещательную рассылку
func (h *Hub) handleBroadcast(msg broadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		if msg.skipCheck || client.userID != msg.excludeID {
			select {
			case client.send <- msg.message:
			default:
				close(client.send)
				delete(h.clients, client.userID)
			}
		}
	}
}

// handleBroadcastToChat отправляет сообщение подписчикам чата
func (h *Hub) handleBroadcastToChat(msg chatBroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.clientsByChat[msg.chatID]
	if !ok {
		return
	}

	for _, client := range clients {
		if msg.skipCheck || client.userID != msg.excludeID {
			select {
			case client.send <- msg.message:
			default:
				close(client.send)
				delete(h.clients, client.userID)
			}
		}
	}
}

// SubscribeToChat подписывает клиента на чат
func (h *Hub) SubscribeToChat(client *Client, chatID uuid.UUID) error {
	// Проверяем доступ к чату
	_, err := h.chatService.GetChat(context.Background(), chatID, client.userID)
	if err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Добавляем в список подписчиков чата
	if _, ok := h.clientsByChat[chatID]; !ok {
		h.clientsByChat[chatID] = make(map[uuid.UUID]*Client)
	}
	h.clientsByChat[chatID][client.userID] = client
	client.Subscribe(chatID)

	// Отправляем непрочитанные сообщения
	h.sendUnreadMessages(client, chatID)

	return nil
}

// UnsubscribeFromChat отписывает клиента от чата
func (h *Hub) UnsubscribeFromChat(client *Client, chatID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.unsubscribeFromChat(client, chatID)
}

// unsubscribeFromChat внутренняя функция отписки
func (h *Hub) unsubscribeFromChat(client *Client, chatID uuid.UUID) {
	if clients, ok := h.clientsByChat[chatID]; ok {
		delete(clients, client.userID)
		if len(clients) == 0 {
			delete(h.clientsByChat, chatID)
		}
	}
	client.Unsubscribe(chatID)
}

// handleMessage обрабатывает входящее WebSocket сообщение
func (h *Hub) handleMessage(client *Client, msg *WSMessage) {
	switch msg.Type {
	case MessageTypeSendMessage:
		h.handleSendMessage(client, msg)
	case MessageTypeReadMessage:
		h.handleReadMessage(client, msg)
	case MessageTypeReadChat:
		h.handleReadChat(client, msg)
	case MessageTypeTypingStart:
		h.handleTyping(client, msg, true)
	case MessageTypeTypingStop:
		h.handleTyping(client, msg, false)
	case MessageTypeSubscribeChat:
		h.handleSubscribeChat(client, msg)
	case MessageTypeUnsubscribeChat:
		h.handleUnsubscribeChat(client, msg)
	default:
		client.SendError("unknown_type", "Unknown message type")
	}
}

// handleSendMessage обрабатывает отправку сообщения
func (h *Hub) handleSendMessage(client *Client, msg *WSMessage) {
	var payload SendMessagePayload
	if err := json.Unmarshal(msg.Payload.(json.RawMessage), &payload); err != nil {
		client.SendError("invalid_payload", "Failed to parse payload")
		return
	}

	chatID, err := uuid.Parse(payload.ChatID)
	if err != nil {
		client.SendError("invalid_chat_id", "Invalid chat ID")
		return
	}

	var replyToID *uuid.UUID
	if payload.ReplyToID != nil {
		id, err := uuid.Parse(*payload.ReplyToID)
		if err == nil {
			replyToID = &id
		}
	}

	messageType := models.MessageTypeText
	if payload.MessageType != "" {
		messageType = models.MessageType(payload.MessageType)
	}

	// Отправляем сообщение через сервис
	sentMsg, err := h.messageService.SendMessage(
		context.Background(),
		chatID,
		client.userID,
		payload.Content,
		messageType,
		payload.MediaURL,
		replyToID,
	)
	if err != nil {
		client.SendError("send_failed", err.Error())
		return
	}

	// Получаем данные отправителя
	senderName := client.username
	senderAvatar := ""
	if user, _ := h.userRepo.GetByID(context.Background(), client.userID); user != nil {
		senderName = user.GetFullName()
		senderAvatar = user.AvatarURL
	}

	// Формируем ответ
	response := &WSMessage{
		Type:      MessageTypeMessage,
		Timestamp: time.Now(),
		Payload: MessagePayload{
			ID:          sentMsg.ID.String(),
			ChatID:      sentMsg.ChatID.String(),
			SenderID:    sentMsg.SenderID.String(),
			SenderName:  senderName,
			SenderAvatar: senderAvatar,
			Content:     sentMsg.Content,
			MessageType: string(sentMsg.MessageType),
			MediaURL:    sentMsg.MediaURL,
			IsEdited:    sentMsg.IsEdited,
			IsDeleted:   sentMsg.IsDeleted,
			Status:      string(sentMsg.Status),
			CreatedAt:   sentMsg.CreatedAt,
		},
	}

	// Отправляем отправителю
	client.Send(response)

	// Рассылаем другим подписчикам чата
	h.BroadcastToChat(chatID, response, true)
}

// handleReadMessage обрабатывает отметку прочтения сообщения
func (h *Hub) handleReadMessage(client *Client, msg *WSMessage) {
	var payload ReadMessagePayload
	if err := json.Unmarshal(msg.Payload.(json.RawMessage), &payload); err != nil {
		client.SendError("invalid_payload", "Failed to parse payload")
		return
	}

	messageID, err := uuid.Parse(payload.MessageID)
	if err != nil {
		client.SendError("invalid_message_id", "Invalid message ID")
		return
	}

	// Получаем сообщение
	message, err := h.messageRepo.GetByID(context.Background(), messageID)
	if err != nil || message == nil {
		client.SendError("message_not_found", "Message not found")
		return
	}

	// Отмечаем как прочитанное
	_ = h.messageService.MarkAsRead(context.Background(), messageID, client.userID)

	// Отправляем уведомление
	response := &WSMessage{
		Type:      MessageTypeMessageRead,
		Timestamp: time.Now(),
		Payload: MessageReadPayload{
			MessageID: messageID.String(),
			UserID:    client.userID.String(),
			ReadAt:    time.Now(),
		},
	}

	// Рассылаем подписчикам чата
	h.BroadcastToChat(message.ChatID, response, false)
}

// handleReadChat обрабатывает отметку прочтения чата
func (h *Hub) handleReadChat(client *Client, msg *WSMessage) {
	var payload ReadChatPayload
	if err := json.Unmarshal(msg.Payload.(json.RawMessage), &payload); err != nil {
		client.SendError("invalid_payload", "Failed to parse payload")
		return
	}

	chatID, err := uuid.Parse(payload.ChatID)
	if err != nil {
		client.SendError("invalid_chat_id", "Invalid chat ID")
		return
	}

	// Отмечаем все сообщения как прочитанные
	_ = h.messageService.MarkChatAsRead(context.Background(), chatID, client.userID)
}

// handleTyping обрабатывает статус набора текста
func (h *Hub) handleTyping(client *Client, msg *WSMessage, isTyping bool) {
	var payload TypingPayload
	if err := json.Unmarshal(msg.Payload.(json.RawMessage), &payload); err != nil {
		client.SendError("invalid_payload", "Failed to parse payload")
		return
	}

	chatID, err := uuid.Parse(payload.ChatID)
	if err != nil {
		client.SendError("invalid_chat_id", "Invalid chat ID")
		return
	}

	// Обновляем статус
	client.SetTyping(chatID, isTyping)

	// Отправляем уведомление другим участникам
	response := &WSMessage{
		Type:      MessageTypeTyping,
		Timestamp: time.Now(),
		Payload: TypingStatusPayload{
			ChatID:   chatID.String(),
			UserID:   client.userID.String(),
			UserName: client.username,
			IsTyping: isTyping,
		},
	}

	h.BroadcastToChat(chatID, response, true)
}

// handleSubscribeChat обрабатывает подписку на чат
func (h *Hub) handleSubscribeChat(client *Client, msg *WSMessage) {
	var payload SubscribePayload
	if err := json.Unmarshal(msg.Payload.(json.RawMessage), &payload); err != nil {
		client.SendError("invalid_payload", "Failed to parse payload")
		return
	}

	chatID, err := uuid.Parse(payload.ChatID)
	if err != nil {
		client.SendError("invalid_chat_id", "Invalid chat ID")
		return
	}

	if err := h.SubscribeToChat(client, chatID); err != nil {
		client.SendError("subscribe_failed", err.Error())
		return
	}
}

// handleUnsubscribeChat обрабатывает отписку от чата
func (h *Hub) handleUnsubscribeChat(client *Client, msg *WSMessage) {
	var payload SubscribePayload
	if err := json.Unmarshal(msg.Payload.(json.RawMessage), &payload); err != nil {
		client.SendError("invalid_payload", "Failed to parse payload")
		return
	}

	chatID, err := uuid.Parse(payload.ChatID)
	if err != nil {
		client.SendError("invalid_chat_id", "Invalid chat ID")
		return
	}

	h.UnsubscribeFromChat(client, chatID)
}

// sendUnreadMessages отправляет непрочитанные сообщения
func (h *Hub) sendUnreadMessages(client *Client, chatID uuid.UUID) {
	messages, err := h.messageRepo.GetChatMessages(context.Background(), chatID, 50, 0)
	if err != nil {
		return
	}

	for _, msg := range messages {
		senderName := ""
		senderAvatar := ""
		if msg.Sender != nil {
			senderName = msg.Sender.GetFullName()
			senderAvatar = msg.Sender.AvatarURL
		}

		client.Send(&WSMessage{
			Type:      MessageTypeMessage,
			Timestamp: time.Now(),
			Payload: MessagePayload{
				ID:          msg.ID.String(),
				ChatID:      msg.ChatID.String(),
				SenderID:    msg.SenderID.String(),
				SenderName:  senderName,
				SenderAvatar: senderAvatar,
				Content:     msg.Content,
				MessageType: string(msg.MessageType),
				MediaURL:    msg.MediaURL,
				IsEdited:    msg.IsEdited,
				IsDeleted:   msg.IsDeleted,
				Status:      string(msg.Status),
				CreatedAt:   msg.CreatedAt,
			},
		})
	}
}

// BroadcastToChat отправляет сообщение всем подписчикам чата
func (h *Hub) BroadcastToChat(chatID uuid.UUID, msg *WSMessage, excludeSelf bool) {
	h.broadcastToChat <- chatBroadcastMessage{
		chatID:    chatID,
		message:   mustMarshal(msg),
		excludeID: uuid.Nil,
		skipCheck: !excludeSelf,
	}
}

// broadcastUserOnline отправляет уведомление о статусе онлайн
func (h *Hub) broadcastUserOnline(userID uuid.UUID, username string) {
	msg := &WSMessage{
		Type:      MessageTypeUserOnline,
		Timestamp: time.Now(),
		Payload: UserStatusPayload{
			UserID:   userID.String(),
			Username: username,
			IsOnline: true,
		},
	}

	h.broadcast <- broadcastMessage{
		message:   mustMarshal(msg),
		excludeID: userID,
		skipCheck: false,
	}
}

// broadcastUserOffline отправляет уведомление о статусе офлайн
func (h *Hub) broadcastUserOffline(userID uuid.UUID) {
	msg := &WSMessage{
		Type:      MessageTypeUserOffline,
		Timestamp: time.Now(),
		Payload: UserStatusPayload{
			UserID:   userID.String(),
			IsOnline: false,
			LastSeen: time.Now(),
		},
	}

	h.broadcast <- broadcastMessage{
		message:   mustMarshal(msg),
		excludeID: userID,
		skipCheck: false,
	}
}

// GetClient возвращает клиента по ID пользователя
func (h *Hub) GetClient(userID uuid.UUID) *Client {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.clients[userID]
}

// GetOnlineUsers возвращает список онлайн пользователей
func (h *Hub) GetOnlineUsers() []uuid.UUID {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]uuid.UUID, 0, len(h.clients))
	for userID := range h.clients {
		users = append(users, userID)
	}
	return users
}

// IsUserOnline проверяет, онлайн ли пользователь
func (h *Hub) IsUserOnline(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}

func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("failed to marshal: %v", err)
		return []byte("{}")
	}
	return data
}
