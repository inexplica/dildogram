package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims представляет JWT claims с пользовательскими данными
type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	jwt.RegisteredClaims
}

// TokenManager управляет созданием и валидацией JWT токенов
type TokenManager struct {
	secretKey string
	expireDur time.Duration
}

// NewTokenManager создаёт новый TokenManager
func NewTokenManager(secretKey string, expireHours int) *TokenManager {
	return &TokenManager{
		secretKey: secretKey,
		expireDur: time.Duration(expireHours) * time.Hour,
	}
}

// Generate создаёт новый JWT токен для пользователя
func (tm *TokenManager) Generate(userID uuid.UUID, username string) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(tm.expireDur)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(tm.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// Verify проверяет и парсит JWT токен
func (tm *TokenManager) Verify(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(tm.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}

// GetExpiration возвращает время истечения токена
func (tm *TokenManager) GetExpiration() time.Duration {
	return tm.expireDur
}
