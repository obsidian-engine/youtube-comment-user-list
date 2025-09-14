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

// Validate は設定の妥当性を検証します
func (c *Config) Validate() error {
	// ポート番号の検証
	if port, err := strconv.Atoi(c.Port); err != nil || port < 1 || port > 65535 {
		return errors.New("PORT must be a valid port number (1-65535)")
	}

	// YouTubeAPIKeyは本番環境では必須（開発・テスト環境では任意）
	env := getEnv("GO_ENV", "development")
	if env == "production" && c.YouTubeAPIKey == "" {
		return errors.New("YT_API_KEY is required in production environment")
	}

	// ログレベルの検証
	validLogLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLogLevels, c.LogLevel) {
		return fmt.Errorf("LOG_LEVEL must be one of: %v", validLogLevels)
	}

	log.Printf("Config loaded successfully:")
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