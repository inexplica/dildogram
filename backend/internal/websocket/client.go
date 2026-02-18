package websocket

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Время ожидания записи
	writeWait = 10 * time.Second

	// Время ожидания pong от клиента
	pongWait = 60 * time.Second

	// Период отправки ping (должен быть меньше pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Максимальный размер сообщения
	maxMessageSize = 512 * 1024 // 512KB
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Разрешаем все origin'ы (в продакшене нужно ограничить)
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client представляет WebSocket клиента
type Client struct {
	hub        *Hub
	conn       *websocket.Conn
	userID     uuid.UUID
	username   string
	send       chan []byte
	mu         sync.RWMutex
	subscribed map[uuid.UUID]bool // Подписки на чаты
	typing     map[uuid.UUID]bool // Статус набора текста по чатам
	lastSeen   time.Time
}

// NewClient создаёт нового клиента
func NewClient(hub *Hub, conn *websocket.Conn, userID uuid.UUID, username string) *Client {
	return &Client{
		hub:        hub,
		conn:       conn,
		userID:     userID,
		username:   username,
		send:       make(chan []byte, 256),
		subscribed: make(map[uuid.UUID]bool),
		typing:     make(map[uuid.UUID]bool),
		lastSeen:   time.Now(),
	}
}

// Read reads messages from the wire.
func (c *Client) Read() {
	defer func() {
		c.hub.Unregister <- c
		c.hub.broadcastUserOffline(c.userID)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket error: %v", err)
			}
			break
		}

		// Парсим сообщение
		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			c.sendError("invalid_json", "Failed to parse message")
			continue
		}

		// Обрабатываем сообщение
		c.hub.handleMessage(c, &wsMsg)
	}
}

// Write writes messages to the wire.
func (c *Client) Write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub закрыл канал
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Отправляем все ожидающие сообщения в том же пакете
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Send отправляет сообщение клиенту
func (c *Client) Send(msg *WSMessage) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("failed to marshal message: %v", err)
		return
	}

	select {
	case c.send <- data:
	default:
		// Канал переполнен, закрываем соединение
		log.Printf("send buffer full for user %s", c.userID)
		close(c.send)
	}
}

// SendError отправляет ошибку клиенту
func (c *Client) SendError(code, message string) {
	c.Send(&WSMessage{
		Type:      MessageTypeError,
		Timestamp: time.Now(),
		Payload: ErrorPayload{
			Code:    code,
			Message: message,
		},
	})
}

// Subscribe подписывает клиента на чат
func (c *Client) Subscribe(chatID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subscribed[chatID] = true
}

// Unsubscribe отписывает клиента от чата
func (c *Client) Unsubscribe(chatID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.subscribed, chatID)
	delete(c.typing, chatID)
}

// IsSubscribed проверяет подписку на чат
func (c *Client) IsSubscribed(chatID uuid.UUID) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.subscribed[chatID]
}

// SetTyping устанавливает статус набора текста
func (c *Client) SetTyping(chatID uuid.UUID, isTyping bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if isTyping {
		c.typing[chatID] = true
		c.lastSeen = time.Now()
	} else {
		delete(c.typing, chatID)
	}
}

// IsTyping проверяет статус набора текста
func (c *Client) IsTyping(chatID uuid.UUID) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.typing[chatID]
}

// GetUserID возвращает ID пользователя
func (c *Client) GetUserID() uuid.UUID {
	return c.userID
}

// GetUsername возвращает имя пользователя
func (c *Client) GetUsername() string {
	return c.username
}

// Broadcast отправляет сообщение всем подключенным клиентам
func (c *Client) Broadcast(msg *WSMessage, excludeSelf bool) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("failed to marshal broadcast message: %v", err)
		return
	}

	c.hub.broadcast <- broadcastMessage{
		message:   data,
		excludeID: c.userID,
		skipCheck: !excludeSelf,
	}
}

// BroadcastToChat отправляет сообщение всем подписчикам чата
func (c *Client) BroadcastToChat(chatID uuid.UUID, msg *WSMessage, excludeSelf bool) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("failed to marshal chat broadcast message: %v", err)
		return
	}

	c.hub.broadcastToChat <- chatBroadcastMessage{
		chatID:    chatID,
		message:   data,
		excludeID: c.userID,
		skipCheck: !excludeSelf,
	}
}

// broadcastMessage сообщение для широковещательной рассылки
type broadcastMessage struct {
	message   []byte
	excludeID uuid.UUID
	skipCheck bool
}

// chatBroadcastMessage сообщение для рассылки по чату
type chatBroadcastMessage struct {
	chatID    uuid.UUID
	message   []byte
	excludeID uuid.UUID
	skipCheck bool
}

// wsResponse записывает WebSocket ответ
func wsResponse(conn *websocket.Conn, status int, message interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(message); err != nil {
		return err
	}
	return conn.WriteMessage(status, buf.Bytes())
}
