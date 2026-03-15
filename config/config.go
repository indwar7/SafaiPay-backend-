package config

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	DB       DBConfig
	Redis    RedisConfig
	JWT      JWTConfig
	MSG91    MSG91Config
	R2       R2Config
	Razorpay RazorpayConfig
	FCM      FCMConfig
}

type ServerConfig struct {
	Port string
	Mode string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret      string
	ExpiryHours int
}

type MSG91Config struct {
	AuthKey    string
	TemplateID string
	SenderID   string
}

type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey  string
	BucketName      string
	PublicURL       string
}

type RazorpayConfig struct {
	KeyID     string
	KeySecret string
}

type FCMConfig struct {
	ProjectID          string
	ServiceAccountJSON string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Warn("no .env file found, using environment variables")
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnvMulti([]string{"PORT", "SERVER_PORT"}, "8080"),
			Mode: getEnv("GIN_MODE", "release"),
		},
		DB: DBConfig{
			Host:     getEnvMulti([]string{"DB_HOST", "PGHOST"}, "localhost"),
			Port:     getEnvMulti([]string{"DB_PORT", "PGPORT"}, "5432"),
			User:     getEnvMulti([]string{"DB_USER", "PGUSER"}, "safaipay"),
			Password: getEnvMulti([]string{"DB_PASSWORD", "PGPASSWORD"}, ""),
			Name:     getEnvMulti([]string{"DB_NAME", "PGDATABASE"}, "safaipay"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Addr:     getEnvMulti([]string{"REDIS_ADDR", "REDIS_URL"}, "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", ""),
			ExpiryHours: getEnvInt("JWT_EXPIRY_HOURS", 720),
		},
		MSG91: MSG91Config{
			AuthKey:    getEnv("MSG91_AUTH_KEY", ""),
			TemplateID: getEnv("MSG91_TEMPLATE_ID", ""),
			SenderID:   getEnv("MSG91_SENDER_ID", "SFPAY"),
		},
		R2: R2Config{
			AccountID:      getEnv("R2_ACCOUNT_ID", ""),
			AccessKeyID:    getEnv("R2_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
			BucketName:     getEnv("R2_BUCKET_NAME", "safaipay"),
			PublicURL:      getEnv("R2_PUBLIC_URL", ""),
		},
		Razorpay: RazorpayConfig{
			KeyID:     getEnv("RAZORPAY_KEY_ID", ""),
			KeySecret: getEnv("RAZORPAY_KEY_SECRET", ""),
		},
		FCM: FCMConfig{
			ProjectID:          getEnv("FCM_PROJECT_ID", ""),
			ServiceAccountJSON: getEnv("FCM_SERVICE_ACCOUNT_JSON", ""),
		},
	}
}

func (c *DBConfig) DSN() string {
	return "host=" + c.Host +
		" port=" + c.Port +
		" user=" + c.User +
		" password=" + c.Password +
		" dbname=" + c.Name +
		" sslmode=" + c.SSLMode +
		" TimeZone=Asia/Kolkata"
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvMulti(keys []string, fallback string) string {
	for _, key := range keys {
		if val := os.Getenv(key); val != "" {
			return val
		}
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}
