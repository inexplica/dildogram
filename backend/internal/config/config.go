package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DB        DBConfig
	JWT       JWTConfig
	Server    ServerConfig
	Upload    UploadConfig
	Redis     RedisConfig
	SMS       SMSConfig
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	DSN      string
}

type JWTConfig struct {
	Secret     string
	ExpireHours int
	ExpireDur  time.Duration
}

type ServerConfig struct {
	Host string
	Port string
}

type UploadConfig struct {
	Dir         string
	MaxFileSize int64
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	Addr     string
}

type SMSConfig struct {
	CodeExpireMinutes int
	CodeExpireDur     time.Duration
}

func Load() (*Config, error) {
	// Загружаем .env файл (игнорируем ошибку если нет)
	_ = godotenv.Load()

	cfg := &Config{}

	// Database
	cfg.DB.Host = getEnv("DB_HOST", "localhost")
	cfg.DB.Port = getEnv("DB_PORT", "5432")
	cfg.DB.User = getEnv("DB_USER", "dildogram")
	cfg.DB.Password = getEnv("DB_PASSWORD", "dildogram_secret")
	cfg.DB.DBName = getEnv("DB_NAME", "dildogram")
	cfg.DB.SSLMode = getEnv("DB_SSLMODE", "disable")
	cfg.DB.DSN = fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.DBName, cfg.DB.SSLMode,
	)

	// JWT
	cfg.JWT.Secret = getEnv("JWT_SECRET", "change-this-secret-key")
	cfg.JWT.ExpireHours = getEnvInt("JWT_EXPIRE_HOURS", 72)
	cfg.JWT.ExpireDur = time.Duration(cfg.JWT.ExpireHours) * time.Hour

	// Server
	cfg.Server.Host = getEnv("SERVER_HOST", "0.0.0.0")
	cfg.Server.Port = getEnv("SERVER_PORT", "8080")

	// Upload
	cfg.Upload.Dir = getEnv("UPLOAD_DIR", "./uploads")
	cfg.Upload.MaxFileSize = getEnvInt64("MAX_UPLOAD_SIZE", 10*1024*1024)

	// Redis
	cfg.Redis.Host = getEnv("REDIS_HOST", "localhost")
	cfg.Redis.Port = getEnv("REDIS_PORT", "6379")
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.Redis.Addr = fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)

	// SMS
	cfg.SMS.CodeExpireMinutes = getEnvInt("SMS_CODE_EXPIRE_MINUTES", 5)
	cfg.SMS.CodeExpireDur = time.Duration(cfg.SMS.CodeExpireMinutes) * time.Minute

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}
