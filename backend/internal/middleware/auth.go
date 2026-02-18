package middleware

import (
	"net/http"
	"strings"

	"dildogram/backend/internal/service"
	"dildogram/backend/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	UserIDKey = "userID"
	UsernameKey = "username"
)

// AuthMiddleware создаёт middleware для JWT аутентификации
func AuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			return
		}

		// Извлекаем токен из "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization format",
			})
			return
		}

		tokenString := parts[1]

		// Проверяем токен
		claims, err := authService.ValidateToken(c.Request.Context(), tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		// Сохраняем данные пользователя в контексте
		c.Set(UserIDKey, claims.UserID.String())
		c.Set(UsernameKey, claims.Username)

		c.Next()
	}
}

// OptionalAuthMiddleware создаёт middleware для опциональной аутентификации
func OptionalAuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString := parts[1]
			claims, err := authService.ValidateToken(c.Request.Context(), tokenString)
			if err == nil {
				c.Set(UserIDKey, claims.UserID.String())
				c.Set(UsernameKey, claims.Username)
			}
		}

		c.Next()
	}
}

// GetUserID извлекает ID пользователя из контекста
func GetUserID(c *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := c.Get(UserIDKey)
	if !exists {
		return uuid.Nil, nil
	}

	if userIDStr, ok := userIDStr.(string); ok {
		return uuid.Parse(userIDStr)
	}

	if userID, ok := userIDStr.(uuid.UUID); ok {
		return userID, nil
	}

	return uuid.Nil, nil
}

// GetUsername извлекает имя пользователя из контекста
func GetUsername(c *gin.Context) string {
	username, _ := c.Get(UsernameKey)
	return username.(string)
}
