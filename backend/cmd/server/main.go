package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dildogram/backend/internal/config"
	"dildogram/backend/internal/handlers"
	"dildogram/backend/internal/middleware"
	"dildogram/backend/internal/models"
	"dildogram/backend/internal/repository"
	"dildogram/backend/internal/service"
	"dildogram/backend/internal/websocket"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализируем базу данных
	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Создаём репозитории
	userRepo := repository.NewUserRepository(db)
	chatRepo := repository.NewChatRepository(db)
	messageRepo := repository.NewMessageRepository(db)

	// Создаём сервисы
	authService := service.NewAuthService(userRepo, cfg)
	chatService := service.NewChatService(chatRepo, userRepo)
	messageService := service.NewMessageService(messageRepo, chatRepo)

	// Создаём WebSocket хаб
	hub := websocket.NewHub(messageService, chatService, authService, messageRepo, chatRepo, userRepo)
	go hub.Run()

	// Создаём обработчики
	authHandler := handlers.NewAuthHandler(authService)
	chatHandler := handlers.NewChatHandler(chatService, messageService, hub)
	wsHandler := handlers.NewWSHandler(authService, hub)

	// Инициализируем Gin
	r := gin.Default()

	// Middleware
	r.Use(middleware.CORSMiddleware(cfg.FrontendURL))

	// Статические файлы (аватарки)
	r.Static("/uploads", "./uploads")

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		// Аутентификация (публичные эндпоинты)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/sms", authHandler.RequestSMS)
			auth.POST("/verify-sms", authHandler.VerifySMS)
			
			// Защищённые эндпоинты
			protected := auth.Group("")
			protected.Use(middleware.AuthMiddleware(authService))
			{
				protected.GET("/me", authHandler.GetMe)
				protected.PUT("/me", authHandler.UpdateProfile)
				protected.POST("/avatar", authHandler.UploadAvatar)
			}
		}

		// Пользователи
		users := v1.Group("/users")
		users.Use(middleware.AuthMiddleware(authService))
		{
			users.GET("/:id", authHandler.GetUser)
			users.GET("", authHandler.SearchUsers)
		}

		// Чаты
		chats := v1.Group("/chats")
		chats.Use(middleware.AuthMiddleware(authService))
		{
			chats.POST("", chatHandler.CreateChat)
			chats.GET("", chatHandler.GetChats)
			chats.GET("/:id", chatHandler.GetChat)
			chats.PUT("/:id", chatHandler.UpdateChat)
			chats.DELETE("/:id", chatHandler.DeleteChat)
			
			// Участники
			chats.POST("/:id/members", chatHandler.AddMember)
			chats.DELETE("/:id/members/:userId", chatHandler.RemoveMember)
			chats.GET("/:id/members", chatHandler.GetMembers)
			
			// Сообщения
			chats.GET("/:id/messages", chatHandler.GetMessages)
			chats.POST("/:id/messages", chatHandler.SendMessage)
			chats.POST("/:id/read", chatHandler.MarkChatAsRead)
		}

		// WebSocket
		v1.GET("/ws", wsHandler.HandleWebSocket)
	}

	// Создаём директорию для загрузок
	if err := os.MkdirAll("./uploads/avatars", 0755); err != nil {
		log.Printf("Warning: failed to create uploads directory: %v", err)
	}

	// Запускаем сервер
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Server starting on %s:%s", cfg.Server.Host, cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Ожидаем сигнал завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

// initDB инициализирует подключение к базе данных
func initDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DB.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Автоматическая миграция моделей
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// autoMigrate выполняет миграцию моделей
func autoMigrate(db *gorm.DB) error {
	models := []interface{}{
		&models.User{},
		&models.SMSCode{},
		&models.Chat{},
		&models.ChatMembership{},
		&models.Message{},
		&models.MessageRead{},
	}

	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate %T: %w", model, err)
		}
	}

	return nil
}
