// Package main YouTubeコメント監視アプリケーションのエントリーポイントです
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/constants"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/application/usecase"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/service"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/infrastructure/events"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/infrastructure/logging"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/infrastructure/repository/memory"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/infrastructure/youtube"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/interfaces/http/handler"
)

// ApplicationContainer holds all dependencies
// ApplicationContainer はすべての依存関係を保持する
type ApplicationContainer struct {
	// Infrastructure
	// インフラストラクチャ層
	Logger         service.Logger
	YouTubeClient  service.YouTubeClient
	UserRepository service.UserRepository
	ChatRepository service.ChatRepository
	EventPublisher service.EventPublisher

	// Domain Services
	// ドメインサービス層
	PollingService *service.PollingService
	UserService    *service.UserService
	VideoService   *service.VideoService

	// Use Cases
	// ユースケース層
	ChatMonitoringUC *usecase.ChatMonitoringUseCase
	LogManagementUC  *usecase.LogManagementUseCase

	// HTTP Handlers
	// HTTPハンドラー層
	MonitoringHandler *handler.MonitoringHandler
	SSEHandler        *handler.SSEHandler
	LogHandler        *handler.LogHandler
	StaticHandler     *handler.StaticHandler
}

func main() {
	fmt.Println("Starting YouTube Live Chat Monitor...")

	// Load environment variables
	// 環境変数を読み込む
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Get API key from environment
	// 環境変数からAPIキーを取得
	apiKey := os.Getenv("YT_API_KEY")
	if apiKey == "" {
		log.Fatal("YT_API_KEY environment variable is required")
	}

	// Build application container with dependency injection
	// 依存性注入でアプリケーションコンテナを構築
	container := buildContainer(apiKey)

	container.Logger.LogStructured("INFO", "main", "startup", "Application starting", "", "", map[string]interface{}{
		"version": "onion-architecture-refactored",
	})

	// Setup HTTP server
	// HTTPサーバーをセットアップ
	server := setupHTTPServer(container)

	// Start server in background
	// サーバーをバックグラウンドで開始
	go func() {
		container.Logger.LogStructured("INFO", "main", "server_start", fmt.Sprintf("HTTP server starting on %s", server.Addr), "", "", nil)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			container.Logger.LogError("FATAL", "HTTP server failed", "", "", err, nil)
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Wait for shutdown signal
	// シャットダウンシグナルを待機
	waitForShutdown(container, server)
}

// buildContainer builds the application container with all dependencies
// buildContainer は全ての依存関係を持つアプリケーションコンテナを構築する
func buildContainer(apiKey string) *ApplicationContainer {
	container := &ApplicationContainer{}

	// Infrastructure layer
	// インフラストラクチャ層

	// Infrastructure
	// インフラストラクチャ層 layer
	container.Logger = logging.NewStructuredLogger()
	container.YouTubeClient = youtube.NewClient(apiKey)
	container.UserRepository = memory.NewUserRepository()
	container.ChatRepository = memory.NewChatRepository(constants.DefaultMaxChatMessages)
	// 最大10,000メッセージ/動画
	container.EventPublisher = events.NewSimplePublisher(container.Logger)

	// Domain services
	// ドメインサービス
	container.PollingService = service.NewPollingService(
		container.YouTubeClient,
		container.ChatRepository,
		container.Logger,
		container.EventPublisher,
	)

	container.UserService = service.NewUserService(
		container.UserRepository,
		container.Logger,
		container.EventPublisher,
	)

	container.VideoService = service.NewVideoService(
		container.YouTubeClient,
		container.Logger,
	)

	// Use cases
	// ユースケース
	container.ChatMonitoringUC = usecase.NewChatMonitoringUseCase(
		container.PollingService,
		container.UserService,
		container.VideoService,
		container.Logger,
	)

	container.LogManagementUC = usecase.NewLogManagementUseCase(
		container.Logger,
		constants.DefaultMaxLogEntries,
		// 最大1,000ログエントリ
	)

	// HTTP handlers
	// HTTPハンドラー
	container.MonitoringHandler = handler.NewMonitoringHandler(
		container.ChatMonitoringUC,
		container.Logger,
	)

	container.SSEHandler = handler.NewSSEHandler(
		container.ChatMonitoringUC,
		container.Logger,
	)

	container.LogHandler = handler.NewLogHandler(
		container.LogManagementUC,
		container.Logger,
	)

	container.StaticHandler = handler.NewStaticHandler(
		container.Logger,
	)

	return container
}

// setupHTTPServer configures and returns the HTTP server
// setupHTTPServer はHTTPサーバーを設定して返す
func setupHTTPServer(container *ApplicationContainer) *http.Server {
	mux := http.NewServeMux()

	// Static pages
	// 静的ページ
	mux.HandleFunc("/", container.StaticHandler.ServeHome)
	mux.HandleFunc("/users", container.StaticHandler.ServeUserListPage)
	mux.HandleFunc("/logs", container.StaticHandler.ServeLogsPage)

	// API endpoints
	// APIエンドポイント
	mux.HandleFunc("/api/monitoring/start", container.MonitoringHandler.StartMonitoring)
	mux.HandleFunc("/api/monitoring/stop/", container.MonitoringHandler.StopMonitoring)
	mux.HandleFunc("/api/monitoring/active", container.MonitoringHandler.GetActiveVideos)

	// Dynamic video ID endpoints (simple pattern matching)
	// 動的動画IDエンドポイント（シンプルパターンマッチング）
	mux.HandleFunc("/api/monitoring/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > len("/api/monitoring/") {
			// Extract operation from path
			// パスから操作を抽出
			remaining := path[len("/api/monitoring/"):]
			if len(remaining) >= constants.YouTubeVideoIDLength {
				// 動画IDの最小長
				switch remaining[constants.YouTubeVideoIDLength:] {
				case "/users":
					container.MonitoringHandler.GetUserList(w, r)
				case "/status":
					container.MonitoringHandler.GetVideoStatus(w, r)
				default:
					http.NotFound(w, r)
				}
			} else {
				http.NotFound(w, r)
			}
		} else {
			http.NotFound(w, r)
		}
	})

	// SSE endpoints
	// SSEエンドポイント
	mux.HandleFunc("/api/sse/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) <= len("/api/sse/") {
			http.NotFound(w, r)
			return
		}

		remaining := path[len("/api/sse/"):]
		if len(remaining) < constants.YouTubeVideoIDLength {
			http.NotFound(w, r)
			return
		}

		// 動画IDの最小長
		if len(remaining) == constants.YouTubeVideoIDLength {
			// /api/sse/{videoId}
			// /api/sse/{動画ID}
			container.SSEHandler.StreamMessages(w, r)
			return
		}

		if remaining[constants.YouTubeVideoIDLength:] == "/users" {
			// /api/sse/{videoId}/users
			// /api/sse/{動画ID}/users
			container.SSEHandler.StreamUserList(w, r)
			return
		}

		http.NotFound(w, r)
	})

	// Log endpoints
	// ログエンドポイント
	mux.HandleFunc("/api/logs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			container.LogHandler.GetLogs(w, r)
		case http.MethodDelete:
			container.LogHandler.ClearLogs(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/logs/stats", container.LogHandler.GetLogStats)
	mux.HandleFunc("/api/logs/export", container.LogHandler.ExportLogs)

	// Create server with proper timeouts
	// 適切なタイムアウト設定でサーバーを作成
	server := &http.Server{
		Addr:         getEnv("PORT", ":8080"),
		Handler:      loggingMiddleware(container.Logger, mux),
		ReadTimeout:  constants.HTTPReadTimeout,
		WriteTimeout: constants.HTTPWriteTimeout,
		IdleTimeout:  constants.HTTPIdleTimeout,
	}

	return server
}

// loggingMiddleware はリクエストログ機能を追加する
func loggingMiddleware(logger service.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		correlationID := fmt.Sprintf("req-%d", start.Unix())

		// Add correlation ID to context
		// コンテキストに相関IDを追加
		ctx := context.WithValue(r.Context(), "requestId", start.Unix())
		r = r.WithContext(ctx)

		logger.LogAPI("INFO", "Request received", "", correlationID, map[string]interface{}{
			"method":     r.Method,
			"path":       r.URL.Path,
			"userAgent":  r.Header.Get("User-Agent"),
			"remoteAddr": r.RemoteAddr,
		})

		// Wrap ResponseWriter to capture status code
		// ステータスコードをキャプチャするためResponseWriterをラップ
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		logger.LogAPI("INFO", "Request completed", "", correlationID, map[string]interface{}{
			"method":     r.Method,
			"path":       r.URL.Path,
			"statusCode": wrapped.statusCode,
			"duration":   duration.String(),
		})
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
// responseWriter はステータスコードをキャプチャするためにhttp.ResponseWriterをラップする
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// waitForShutdown waits for a shutdown signal and gracefully shuts down the server
// waitForShutdown はシャットダウンシグナルを待機し、サーバーを優雅にシャットダウンする
func waitForShutdown(container *ApplicationContainer, server *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	container.Logger.LogStructured("INFO", "main", "shutdown_start", "Shutdown signal received", "", "", nil)

	ctx, cancel := context.WithTimeout(context.Background(), constants.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		container.Logger.LogError("ERROR", "Server shutdown error", "", "", err, nil)
	}

	container.Logger.LogStructured("INFO", "main", "shutdown_complete", "Application shutdown complete", "", "", nil)
}

// getEnv gets an environment variable with a default value
// getEnv はデフォルト値付きで環境変数を取得する
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
