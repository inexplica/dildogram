package repository

import (
	"context"
	"errors"
	"time"

	"dildogram/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository определяет интерфейс для работы с пользователями
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByPhone(ctx context.Context, phone string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdateAvatar(ctx context.Context, id uuid.UUID, avatarURL string) error
	SetOnline(ctx context.Context, id uuid.UUID, isOnline bool) error
	Search(ctx context.Context, query string, limit int) ([]models.User, error)
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository создаёт новый UserRepository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) UpdateAvatar(ctx context.Context, id uuid.UUID, avatarURL string) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", id).
		Update("avatar_url", avatarURL).Error
}

func (r *userRepository) SetOnline(ctx context.Context, id uuid.UUID, isOnline bool) error {
	updates := map[string]interface{}{
		"is_online": isOnline,
	}
	if !isOnline {
		updates["last_seen"] = time.Now()
	}
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *userRepository) Search(ctx context.Context, query string, limit int) ([]models.User, error) {
	var users []models.User
	searchPattern := "%" + query + "%"
	err := r.db.WithContext(ctx).
		Where("username ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?",
			searchPattern, searchPattern, searchPattern).
		Limit(limit).
		Find(&users).Error
	return users, err
}
