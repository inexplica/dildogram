package models

import (
	"time"

	"github.com/google/uuid"
)

// User представляет пользователя в системе
type User struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Phone        string     `gorm:"size:20;uniqueIndex;not null" json:"phone"`
	Username     string     `gorm:"size:50;uniqueIndex;not null" json:"username"`
	PasswordHash *string    `gorm:"size:255" json:"-"` // Pointer - может быть NULL для SMS
	FirstName    string     `gorm:"size:50;not null;default:''" json:"first_name"`
	LastName     string     `gorm:"size:50;not null;default:''" json:"last_name"`
	Bio          string     `gorm:"type:text;not null;default:''" json:"bio"`
	AvatarURL    string     `gorm:"size:500;not null;default:''" json:"avatar_url"`
	IsActive     bool       `gorm:"not null;default:true" json:"is_active"`
	IsOnline     bool       `gorm:"not null;default:false" json:"is_online"`
	LastSeen     time.Time  `gorm:"not null;default:now()" json:"last_seen"`
	CreatedAt    time.Time  `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"not null;default:now()" json:"updated_at"`

	// Связи
	OwnedChats   []Chat          `gorm:"foreignKey:CreatedBy" json:"-"`
	Memberships  []ChatMembership `gorm:"foreignKey:UserID" json:"-"`
	SentMessages []Message       `gorm:"foreignKey:SenderID" json:"-"`
}

// TableName возвращает имя таблицы
func (User) TableName() string {
	return "users"
}

// GetFullName возвращает полное имя пользователя
func (u *User) GetFullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return u.Username
	}
	return u.FirstName + " " + u.LastName
}

// SMSCode представляет код для SMS авторизации
type SMSCode struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Phone     string    `gorm:"size:20;not null;index" json:"phone"`
	Code      string    `gorm:"size:6;not null" json:"-"`
	IsUsed    bool      `gorm:"not null;default:false" json:"is_used"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
}

func (SMSCode) TableName() string {
	return "sms_codes"
}

// IsExpired проверяет истёк ли код
func (s *SMSCode) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
