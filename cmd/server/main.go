// Package main YouTubeコメント監視アプリケーションのエントリーポイントです
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
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

// IdleTimeoutManager アイドルタイムアウト管理
type IdleTimeoutManager struct {
	lastRequest int64 // Unix timestamp of last request
	stopChan    chan struct{}
}

// NewIdleTimeoutManager 新しいアイドルタイムアウトマネージャーを作成
func NewIdleTimeoutManager() *IdleTimeoutManager {
	return &IdleTimeoutManager{
		lastRequest: time.Now().Unix(),
		stopChan:    make(chan struct{}),
	}
}

// UpdateLastRequest 最後のリクエスト時刻を更新
func (itm *IdleTimeoutManager) UpdateLastRequest() {
	atomic.StoreInt64(&itm.lastRequest, time.Now().Unix())
}

// StartIdleMonitor アイドル監視を開始
func (itm *IdleTimeoutManager) StartIdleMonitor(logger service.Logger) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute) // 1分ごとにチェック
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				lastRequest := atomic.LoadInt64(&itm.lastRequest)
				if time.Since(time.Unix(lastRequest, 0)) > constants.IdleTimeout {
					logger.LogStructured("INFO", "main", "idle_timeout",
						fmt.Sprintf("Server has been idle for %v, shutting down", constants.IdleTimeout),
						"", "", nil)
					close(itm.stopChan)
					return
				}
			case <-itm.stopChan:
				return
			}
		}
	}()
}

// GetStopChannel 停止チャンネルを取得
func (itm *IdleTimeoutManager) GetStopChannel() <-chan struct{} {
	return itm.stopChan
}

// ApplicationContainer はすべての依存関係を保持する
type ApplicationContainer struct {
	// Infrastructure
	IdleManager    *IdleTimeoutManager
	Logger         service.Logger
	YouTubeClient  service.YouTubeClient
	UserRepository service.UserRepository
	ChatRepository service.ChatRepository
	EventPublisher service.EventPublisher

	// ドメインサービス層
	PollingService *service.PollingService
	UserService    *service.UserService
	VideoService   *service.VideoService

	// ユースケース層
	ChatMonitoringUC *usecase.ChatMonitoringUseCase
	LogManagementUC  *usecase.LogManagementUseCase

	// HTTPハンドラー層
	MonitoringHandler *handler.MonitoringHandler
	SSEHandler        *handler.SSEHandler
	LogHandler        *handler.LogHandler
	StaticHandler     *handler.StaticHandler
}

func main() {
	fmt.Println("Starting YouTube Live Chat Monitor...")

	// 環境変数を読み込む
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// 環境変数からAPIキーを取得
	apiKey := os.Getenv("YT_API_KEY")
	if apiKey == "" {
		log.Fatal("YT_API_KEY environment variable is required")
	}

	// 依存性注入でアプリケーションコンテナを構築
	container := buildContainer(apiKey)

	container.Logger.LogStructured("INFO", "main", "startup", "Application starting", "", "", map[string]interface{}{
		"version": "onion-architecture-refactored",
	})

	// アイドルタイムアウト監視を開始
	container.IdleManager.StartIdleMonitor(container.Logger)

	server := setupHTTPServer(container)

	// サーバーをバックグラウンドで開始
	go func() {
		container.Logger.LogStructured("INFO", "main", "server_start", fmt.Sprintf("HTTP server starting on %s", server.Addr), "", "", nil)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			container.Logger.LogError("FATAL", "HTTP server failed", "", "", err, nil)
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// シャットダウンシグナルを待機
	waitForShutdown(container, server)
}

// buildContainer は全ての依存関係を持つアプリケーションコンテナを構築する
func buildContainer(apiKey string) *ApplicationContainer {
	container := &ApplicationContainer{}

	container.IdleManager = NewIdleTimeoutManager()
	container.Logger = logging.NewStructuredLogger()
	container.YouTubeClient = youtube.NewClient(apiKey)
	container.UserRepository = memory.NewUserRepository()
	container.ChatRepository = memory.NewChatRepository(constants.DefaultMaxChatMessages)
	// 最大10,000メッセージ/動画
	container.EventPublisher = events.NewSimplePublisher(container.Logger)

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

// setupHTTPServer はHTTPサーバーを設定して返す
func setupHTTPServer(container *ApplicationContainer) *http.Server {
	// Ginルーターを初期化
	r := gin.New()

	// ミドルウェアを追加
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// アイドルタイムアウト更新ミドルウェア
	r.Use(func(c *gin.Context) {
		container.IdleManager.UpdateLastRequest()
		c.Next()
	})

	// CORS設定を追加
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 静的ページ
	r.GET("/", container.StaticHandler.ServeHome)
	r.GET("/users", container.StaticHandler.ServeUserListPage)
	r.GET("/logs", container.StaticHandler.ServeLogsPage)

	// APIエンドポイント
	api := r.Group("/api")
	{
		// 監視エンドポイント
		monitoring := api.Group("/monitoring")
		{
			monitoring.POST("/start", container.MonitoringHandler.StartMonitoring)
			monitoring.DELETE("/stop", container.MonitoringHandler.StopMonitoring)
			monitoring.POST("/stop", container.MonitoringHandler.StopMonitoring)
			monitoring.GET("/active", container.MonitoringHandler.GetActiveVideoID)
			monitoring.GET("/:videoId/users", container.MonitoringHandler.GetUserList)
			monitoring.GET("/:videoId/status", container.MonitoringHandler.GetVideoStatus)
		}

		// SSEエンドポイント
		sse := api.Group("/sse")
		{
			sse.GET("/:videoId", container.SSEHandler.StreamMessages)
			sse.GET("/:videoId/users", container.SSEHandler.StreamUserList)
		}

		// ログエンドポイント
		logs := api.Group("/logs")
		{
			logs.GET("", container.LogHandler.GetLogs)
			logs.DELETE("", container.LogHandler.ClearLogs)
			logs.GET("/stats", container.LogHandler.GetLogStats)
			logs.GET("/export", container.LogHandler.ExportLogs)
		}
	}

	// 適切なタイムアウト設定でサーバーを作成
	server := &http.Server{
		Addr:         getEnv("PORT", ":8080"),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server
}

// waitForShutdown はシャットダウンシグナルを待機し、サーバーを優雅にシャットダウンする
func waitForShutdown(container *ApplicationContainer, server *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// シグナルまたはアイドルタイムアウトを待機
	select {
	case <-quit:
		container.Logger.LogStructured("INFO", "main", "shutdown_start", "Shutdown signal received", "", "", nil)
	case <-container.IdleManager.GetStopChannel():
		container.Logger.LogStructured("INFO", "main", "shutdown_start", "Idle timeout reached, shutting down", "", "", nil)
	}
	ctx, cancel := context.WithTimeout(context.Background(), constants.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		container.Logger.LogError("ERROR", "Server shutdown error", "", "", err, nil)
	}

	container.Logger.LogStructured("INFO", "main", "shutdown_complete", "Application shutdown complete", "", "", nil)
}

// getEnv はデフォルト値付きで環境変数を取得する
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
