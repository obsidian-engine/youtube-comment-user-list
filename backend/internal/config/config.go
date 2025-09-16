package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
)

// Config はアプリケーション設定を保持します
type Config struct {
	Port           string
	FrontendOrigin string
	YouTubeAPIKey  string
	LogLevel       string
}

// Load は環境変数から設定を読み込み、検証します
func Load() (*Config, error) {
	config := &Config{
		Port:           getEnv("PORT", "8080"),
		FrontendOrigin: os.Getenv("FRONTEND_ORIGIN"),
		YouTubeAPIKey:  os.Getenv("YT_API_KEY"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// SetupLogger はログレベルに応じてログ設定を行います
func (c *Config) SetupLogger() {
	switch c.LogLevel {
	case "debug":
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	case "info":
		log.SetFlags(log.LstdFlags)
	case "warn", "error":
		log.SetFlags(log.LstdFlags)
	default:
		log.SetFlags(log.LstdFlags)
	}
	
	log.Printf("Logger configured with level: %s", c.LogLevel)
}

// Validate は設定の妥当性を検証します
func (c *Config) Validate() error {
	// ポート番号の検証
	if port, err := strconv.Atoi(c.Port); err != nil || port < 1 || port > 65535 {
		return errors.New("PORT must be a valid port number (1-65535)")
	}

	// 環境変数の取得
	env := getEnv("GO_ENV", "development")

	// 本番環境での必須項目チェック
	if env == "production" {
		// YouTubeAPIKeyは本番環境では必須
		if c.YouTubeAPIKey == "" {
			return errors.New("YT_API_KEY is required in production environment")
		}
		
		// FrontendOriginは本番環境では必須（CORS設定のため）
		if c.FrontendOrigin == "" {
			return errors.New("FRONTEND_ORIGIN is required in production environment")
		}
	}

	// ログレベルの検証
	validLogLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLogLevels, c.LogLevel) {
		return fmt.Errorf("LOG_LEVEL must be one of: %v", validLogLevels)
	}

	log.Printf("Config loaded successfully:")
	log.Printf("  Environment: %s", env)
	log.Printf("  Port: %s", c.Port)
	log.Printf("  Frontend Origin: %s", maskString(c.FrontendOrigin))
	log.Printf("  YouTube API Key: %s", maskString(c.YouTubeAPIKey))
	log.Printf("  Log Level: %s", c.LogLevel)

	return nil
}

// getEnv は環境変数を取得し、存在しない場合はデフォルト値を返します
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// contains はスライスに値が含まれているかチェックします
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// maskString は機密情報をマスクして表示用に変換します
func maskString(s string) string {
	if s == "" {
		return "(not set)"
	}
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}
