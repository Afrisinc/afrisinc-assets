package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server  ServerConfig
	DB      DBConfig
	Storage StorageConfig
	Auth    AuthConfig
	Upload  UploadConfig
}

type ServerConfig struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type DBConfig struct {
	DSN         string
	MaxConns    int
	MinConns    int
	MaxConnLife time.Duration
}

type StorageConfig struct {
	// "local" or "s3"
	Driver    string
	LocalRoot string

	// S3 / S3-compatible (MinIO, Backblaze, etc.)
	S3Endpoint  string
	S3Bucket    string
	S3Region    string
	S3AccessKey string
	S3SecretKey string
	S3UseSSL    bool
	BaseURL     string // public base URL for served assets
}

type AuthConfig struct {
	// Static API key for now — swap for JWT/OAuth as needed
	APIKey string
}

type UploadConfig struct {
	MaxFileSizeBytes int64
	AllowedMIMETypes []string
}

// Load reads .env (if present) then falls back to environment variables.
func Load() (*Config, error) {
	// Non-fatal if .env is missing (production may rely on real env vars)
	_ = godotenv.Load()

	cfg := &Config{}

	cfg.Server = ServerConfig{
		Addr:            getEnv("SERVER_ADDR", ":8080"),
		ReadTimeout:     getDuration("SERVER_READ_TIMEOUT", 30*time.Second),
		WriteTimeout:    getDuration("SERVER_WRITE_TIMEOUT", 60*time.Second),
		IdleTimeout:     getDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
		ShutdownTimeout: getDuration("SERVER_SHUTDOWN_TIMEOUT", 15*time.Second),
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	cfg.DB = DBConfig{
		DSN:         dsn,
		MaxConns:    getInt("DB_MAX_CONNS", 20),
		MinConns:    getInt("DB_MIN_CONNS", 2),
		MaxConnLife: getDuration("DB_MAX_CONN_LIFE", 30*time.Minute),
	}

	cfg.Storage = StorageConfig{
		Driver:      getEnv("STORAGE_DRIVER", "local"),
		LocalRoot:   getEnv("STORAGE_LOCAL_ROOT", "./data/uploads"),
		S3Endpoint:  os.Getenv("S3_ENDPOINT"),
		S3Bucket:    os.Getenv("S3_BUCKET"),
		S3Region:    getEnv("S3_REGION", "auto"),
		S3AccessKey: os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey: os.Getenv("S3_SECRET_KEY"),
		S3UseSSL:    getBool("S3_USE_SSL", true),
		BaseURL:     getEnv("ASSET_BASE_URL", "http://localhost:8080"),
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("API_KEY is required")
	}
	cfg.Auth = AuthConfig{APIKey: apiKey}

	cfg.Upload = UploadConfig{
		MaxFileSizeBytes: int64(getInt("UPLOAD_MAX_MB", 50)) * 1024 * 1024,
		AllowedMIMETypes: []string{
			"image/jpeg", "image/png", "image/webp", "image/gif", "image/svg+xml",
			"video/mp4", "video/webm",
			"application/pdf",
			"font/ttf", "font/otf", "font/woff", "font/woff2",
		},
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
