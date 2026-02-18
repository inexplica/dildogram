package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"dildogram/backend/internal/middleware"
	"dildogram/backend/internal/models"
	"dildogram/backend/internal/service"
	"dildogram/backend/internal/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WSHandler обрабатывает WebSocket подключения
type WSHandler struct {
	authService *service.AuthService
	hub         *websocket.Hub
	upgrader    websocket.Upgrader
}

// NewWSHandler создаёт новый WSHandler
func NewWSHandler(authService *service.AuthService, hub *websocket.Hub) *WSHandler {
	return &WSHandler{
		authService: authService,
		hub:         hub,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // В продакшене нужно ограничить
			},
		},
	}
}

// HandleWebSocket обрабатывает WebSocket подключения
func (h *WSHandler) HandleWebSocket(c *gin.Context) {
	// Проверяем токен
	tokenString := c.Query("token")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Token required",
		})
		return
	}

	// Проверяем токен
	claims, err := h.authService.ValidateToken(c.Request.Context(), tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid or expired token",
		})
		return
	}

	// Получаем пользователя
	user, err := h.authService.GetUserByID(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user",
		})
		return
	}

	// Upgrader'им соединение
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// Создаём клиента
	client := websocket.NewClient(h.hub, conn, claims.UserID, user.Username)

	// Регистрируем клиента
	h.hub.Register <- client

	// Запускаем обработчики
	go client.Write()
	go client.Read()
}

// AuthHandler обрабатывает запросы аутентификации
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler создаёт новый AuthHandler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterRequest запрос на регистрацию
type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest запрос на вход
type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// SMSRequest запрос на SMS код
type SMSRequest struct {
	Phone string `json:"phone" binding:"required"`
}

// VerifySMSRequest запрос на проверку SMS
type VerifySMSRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required,len=6"`
}

// Register регистрирует пользователя
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, token, err := h.authService.Register(c.Request.Context(), req.Phone, req.Username, req.Password)
	if err != nil {
		if err == service.ErrUserExists {
			c.JSON(http.StatusConflict, gin.H{
				"error": "User already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":  user,
		"token": token,
	})
}

// Login выполняет вход
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, token, err := h.authService.Login(c.Request.Context(), req.Phone, req.Password)
	if err != nil {
		if err == service.ErrUserNotFound || err == service.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid phone or password",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

// RequestSMS запрашивает SMS код
func (h *AuthHandler) RequestSMS(c *gin.Context) {
	var req SMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	code, err := h.authService.RequestSMSCode(c.Request.Context(), req.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// В реальном приложении код отправляется по SMS
	// Для разработки возвращаем код в ответе (удалить в продакшене!)
	c.JSON(http.StatusOK, gin.H{
		"message": "SMS code sent",
		"code":    code, // Удалить в продакшене!
	})
}

// VerifySMS проверяет SMS код
func (h *AuthHandler) VerifySMS(c *gin.Context) {
	var req VerifySMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, token, err := h.authService.VerifySMSCode(c.Request.Context(), req.Phone, req.Code)
	if err != nil {
		if err == service.ErrInvalidCode {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired code",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

// GetMe возвращает текущего пользователя
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// UpdateProfileRequest запрос на обновление профиля
type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Bio       string `json:"bio"`
}

// UpdateProfile обновляет профиль
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := h.authService.UpdateProfile(c.Request.Context(), userID, req.FirstName, req.LastName, req.Bio)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// UploadAvatar загружает аватар
func (h *AuthHandler) UploadAvatar(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Avatar file required",
		})
		return
	}

	// Проверяем расширение
	ext := filepath.Ext(file.Filename)
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid file type. Allowed: jpg, jpeg, png, gif, webp",
		})
		return
	}

	// Создаём уникальное имя файла
	filename := uuid.New().String() + ext
	uploadDir := "./uploads/avatars"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create upload directory",
		})
		return
	}

	filePath := filepath.Join(uploadDir, filename)

	// Сохраняем файл
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save file",
		})
		return
	}

	// Формируем URL
	avatarURL := "/uploads/avatars/" + filename

	// Обновляем аватар в БД
	user, err := h.authService.UpdateAvatar(c.Request.Context(), userID, avatarURL)
	if err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update avatar",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":       user,
		"avatar_url": avatarURL,
	})
}

// GetUser получает пользователя по ID
func (h *AuthHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// SearchUsers ищет пользователей
func (h *AuthHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Query parameter 'q' required",
		})
		return
	}

	limit := 20
	users, err := h.authService.SearchUsers(c.Request.Context(), query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
	})
}

// ChatHandler обрабатывает запросы чатов
type ChatHandler struct {
	chatService    *service.ChatService
	messageService *service.MessageService
	hub            *websocket.Hub
}

// NewChatHandler создаёт новый ChatHandler
func NewChatHandler(chatService *service.ChatService, messageService *service.MessageService, hub *websocket.Hub) *ChatHandler {
	return &ChatHandler{
		chatService:    chatService,
		messageService: messageService,
		hub:            hub,
	}
}

// CreateChatRequest запрос на создание чата
type CreateChatRequest struct {
	Type        string    `json:"type" binding:"required,oneof=private group"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	MemberIDs   []string  `json:"member_ids"`
}

// CreateChat создаёт чат
func (h *ChatHandler) CreateChat(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req CreateChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if req.Type == "private" {
		if len(req.MemberIDs) != 1 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Private chat requires exactly one other member",
			})
			return
		}

		otherUserID, err := uuid.Parse(req.MemberIDs[0])
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid member ID",
			})
			return
		}

		chat, err := h.chatService.CreatePrivateChat(c.Request.Context(), userID, otherUserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"chat": chat,
		})
		return
	}

	// Групповой чат
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Group chat requires a name",
		})
		return
	}

	memberIDs := make([]uuid.UUID, 0, len(req.MemberIDs))
	for _, idStr := range req.MemberIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid member ID: " + idStr,
			})
			return
		}
		memberIDs = append(memberIDs, id)
	}

	chat, err := h.chatService.CreateGroupChat(c.Request.Context(), userID, req.Name, req.Description, memberIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"chat": chat,
	})
}

// GetChats получает список чатов пользователя
func (h *ChatHandler) GetChats(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	chats, err := h.chatService.GetUserChats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chats": chats,
	})
}

// GetChat получает чат по ID
func (h *ChatHandler) GetChat(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid chat ID",
		})
		return
	}

	chat, err := h.chatService.GetChat(c.Request.Context(), chatID, userID)
	if err != nil {
		if err == service.ErrChatNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Chat not found",
			})
			return
		}
		if err == service.ErrNotMember {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chat": chat,
	})
}

// UpdateChatRequest запрос на обновление чата
type UpdateChatRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateChat обновляет чат
func (h *ChatHandler) UpdateChat(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid chat ID",
		})
		return
	}

	var req UpdateChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	chat, err := h.chatService.UpdateChat(c.Request.Context(), chatID, userID, req.Name, req.Description, "")
	if err != nil {
		if err == service.ErrChatNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Chat not found",
			})
			return
		}
		if err == service.ErrNoPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chat": chat,
	})
}

// DeleteChat удаляет чат
func (h *ChatHandler) DeleteChat(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid chat ID",
		})
		return
	}

	if err := h.chatService.DeleteChat(c.Request.Context(), chatID, userID); err != nil {
		if err == service.ErrChatNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Chat not found",
			})
			return
		}
		if err == service.ErrNoPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Chat deleted",
	})
}

// AddMemberRequest запрос на добавление участника
type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// AddMember добавляет участника в чат
func (h *ChatHandler) AddMember(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid chat ID",
		})
		return
	}

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	newMemberID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	if err := h.chatService.AddMember(c.Request.Context(), chatID, userID, newMemberID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Member added",
	})
}

// RemoveMember удаляет участника из чата
func (h *ChatHandler) RemoveMember(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid chat ID",
		})
		return
	}

	memberID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	if err := h.chatService.RemoveMember(c.Request.Context(), chatID, userID, memberID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Member removed",
	})
}

// GetMembers получает участников чата
func (h *ChatHandler) GetMembers(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid chat ID",
		})
		return
	}

	members, err := h.chatService.GetMembers(c.Request.Context(), chatID, userID)
	if err != nil {
		if err == service.ErrNotMember {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"members": members,
	})
}

// GetMessages получает сообщения чата
func (h *ChatHandler) GetMessages(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid chat ID",
		})
		return
	}

	limit := 50
	offset := 0

	if l := c.Query("limit"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &limit); err != nil {
			limit = 50
		}
	}
	if o := c.Query("offset"); o != "" {
		if _, err := fmt.Sscanf(o, "%d", &offset); err != nil {
			offset = 0
		}
	}

	messages, err := h.messageService.GetMessages(c.Request.Context(), chatID, userID, limit, offset)
	if err != nil {
		if err == service.ErrNotMember {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
	})
}

// SendMessageRequest запрос на отправку сообщения
type SendMessageRequest struct {
	Content   string  `json:"content" binding:"required"`
	MessageType string  `json:"message_type"`
	MediaURL  *string `json:"media_url"`
	ReplyToID *string `json:"reply_to_id"`
}

// SendMessage отправляет сообщение
func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid chat ID",
		})
		return
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	messageType := models.MessageTypeText
	if req.MessageType != "" {
		messageType = models.MessageType(req.MessageType)
	}

	var replyToID *uuid.UUID
	if req.ReplyToID != nil {
		id, err := uuid.Parse(*req.ReplyToID)
		if err == nil {
			replyToID = &id
		}
	}

	message, err := h.messageService.SendMessage(
		c.Request.Context(),
		chatID,
		userID,
		req.Content,
		messageType,
		req.MediaURL,
		replyToID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": message,
	})
}

// MarkChatAsRead отмечает чат как прочитанный
func (h *ChatHandler) MarkChatAsRead(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid chat ID",
		})
		return
	}

	if err := h.messageService.MarkChatAsRead(c.Request.Context(), chatID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Chat marked as read",
	})
}
