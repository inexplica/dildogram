package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"dildogram/backend/internal/config"
	"dildogram/backend/internal/models"
	"dildogram/backend/internal/repository"
	"dildogram/backend/pkg/hasher"
	"dildogram/backend/pkg/jwt"
	"github.com/google/uuid"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists        = errors.New("user already exists")
	ErrInvalidCode       = errors.New("invalid or expired code")
)

// AuthService предоставляет методы для аутентификации
type AuthService struct {
	userRepo   repository.UserRepository
	smsRepo    *smsCodeStorage
	tokenMgr   *jwt.TokenManager
	config     *config.Config
}

// smsCodeStorage хранит SMS коды в памяти (для имитации)
type smsCodeStorage struct {
	codes map[string]*models.SMSCode
}

func newSMSCodeStorage() *smsCodeStorage {
	return &smsCodeStorage{
		codes: make(map[string]*models.SMSCode),
	}
}

func (s *smsCodeStorage) Save(code *models.SMSCode) {
	s.codes[code.Phone] = code
}

func (s *smsCodeStorage) Get(phone string) *models.SMSCode {
	return s.codes[phone]
}

func (s *smsCodeStorage) Delete(phone string) {
	delete(s.codes, phone)
}

// NewAuthService создаёт новый AuthService
func NewAuthService(userRepo repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		smsRepo:  newSMSCodeStorage(),
		tokenMgr: jwt.NewTokenManager(cfg.JWT.Secret, cfg.JWT.ExpireHours),
		config:   cfg,
	}
}

// Register регистрирует нового пользователя с паролем
func (s *AuthService) Register(ctx context.Context, phone, username, password string) (*models.User, string, error) {
	// Проверяем существование пользователя
	existing, _ := s.userRepo.GetByPhone(ctx, phone)
	if existing != nil {
		return nil, "", ErrUserExists
	}

	existing, _ = s.userRepo.GetByUsername(ctx, username)
	if existing != nil {
		return nil, "", ErrUserExists
	}

	// Хешируем пароль
	hash, err := hasher.HashPassword(password)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Создаём пользователя
	user := &models.User{
		Phone:        phone,
		Username:     username,
		PasswordHash: &hash,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	// Генерируем токен
	token, err := s.tokenMgr.Generate(user.ID, user.Username)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}

// Login выполняет вход по паролю
func (s *AuthService) Login(ctx context.Context, phone, password string) (*models.User, string, error) {
	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", ErrUserNotFound
	}

	// Проверяем пароль
	if user.PasswordHash == nil || !hasher.VerifyPassword(password, *user.PasswordHash) {
		return nil, "", ErrInvalidCredentials
	}

	// Генерируем токен
	token, err := s.tokenMgr.Generate(user.ID, user.Username)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}

// RequestSMSCode запрашивает SMS код (имитация)
func (s *AuthService) RequestSMSCode(ctx context.Context, phone string) (string, error) {
	// Генерируем 6-значный код
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate code: %w", err)
	}
	code := hex.EncodeToString(bytes)[:6]

	// Сохраняем код
	smsCode := &models.SMSCode{
		Phone:     phone,
		Code:      code,
		ExpiresAt: time.Now().Add(s.config.SMS.CodeExpireDur),
	}
	s.smsRepo.Save(smsCode)

	// В реальном приложении здесь была бы отправка SMS
	// Для разработки выводим код в лог
	fmt.Printf("[SMS CODE] Phone: %s, Code: %s\n", phone, code)

	return code, nil
}

// VerifySMSCode проверяет SMS код и выполняет вход
func (s *AuthService) VerifySMSCode(ctx context.Context, phone, code string) (*models.User, string, error) {
	smsCode := s.smsRepo.Get(phone)
	if smsCode == nil {
		return nil, "", ErrInvalidCode
	}

	if smsCode.IsUsed || smsCode.IsExpired() {
		s.smsRepo.Delete(phone)
		return nil, "", ErrInvalidCode
	}

	if smsCode.Code != code {
		return nil, "", ErrInvalidCode
	}

	// Помечаем код как использованный
	smsCode.IsUsed = true
	s.smsRepo.Delete(phone)

	// Ищем или создаём пользователя
	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return nil, "", err
	}

	if user == nil {
		// Создаём нового пользователя с phone как username
		user = &models.User{
			Phone:    phone,
			Username: phone,
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, "", fmt.Errorf("failed to create user: %w", err)
		}
	}

	// Генерируем токен
	token, err := s.tokenMgr.Generate(user.ID, user.Username)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}

// ValidateToken проверяет JWT токен
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*jwt.Claims, error) {
	return s.tokenMgr.Verify(tokenString)
}

// GetUserByID получает пользователя по ID
func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// UpdateProfile обновляет профиль пользователя
func (s *AuthService) UpdateProfile(ctx context.Context, userID uuid.UUID, firstName, lastName, bio string) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	user.FirstName = firstName
	user.LastName = lastName
	user.Bio = bio

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateAvatar обновляет аватар пользователя
func (s *AuthService) UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) (*models.User, error) {
	if err := s.userRepo.UpdateAvatar(ctx, userID, avatarURL); err != nil {
		return nil, err
	}

	return s.userRepo.GetByID(ctx, userID)
}

// SetOnline устанавливает статус онлайн
func (s *AuthService) SetOnline(ctx context.Context, userID uuid.UUID, isOnline bool) error {
	return s.userRepo.SetOnline(ctx, userID, isOnline)
}

// SearchUsers ищет пользователей
func (s *AuthService) SearchUsers(ctx context.Context, query string, limit int) ([]models.User, error) {
	return s.userRepo.Search(ctx, query, limit)
}
