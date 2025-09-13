// Package main YouTubeコメント監視アプリケーションのエントリーポイントです
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/constants"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/application/service"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/application/usecase"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/repository"
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
func (itm *IdleTimeoutManager) StartIdleMonitor(logger repository.Logger) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute) // 1分ごとにチェック
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				lastRequest := atomic.LoadInt64(&itm.lastRequest)
				if time.Since(time.Unix(lastRequest, 0)) > constants.IdleTimeout {
                logger.LogStructured(constants.LogLevelInfo, "main", "idle_timeout",
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
	IdleManager    *IdleTimeoutManager
	Logger         repository.Logger
	YouTubeClient  repository.YouTubeClient
	UserRepository repository.UserRepository
	ChatRepository repository.ChatRepository
	EventPublisher repository.EventPublisher

	PollingService *service.PollingService
	UserService    *service.UserService
	VideoService   *service.VideoService

	ChatMonitoringUC *usecase.ChatMonitoringUseCase
	LogManagementUC  *usecase.LogManagementUseCase

	MonitoringHandler *handler.MonitoringHandler
	SSEHandler        *handler.SSEHandler
	LogHandler        *handler.LogHandler
	StaticHandler     *handler.StaticHandler
	HealthHandler     *handler.HealthHandler
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

    container.Logger.LogStructured(constants.LogLevelInfo, "main", "startup", "Application starting", "", "", map[string]interface{}{
		"version": "onion-architecture-refactored",
	})

	// アイドルタイムアウト監視を開始
	container.IdleManager.StartIdleMonitor(container.Logger)

	server := setupHTTPServer(container)

	// サーバーをバックグラウンドで開始
	go func() {
        container.Logger.LogStructured(constants.LogLevelInfo, "main", "server_start", fmt.Sprintf("HTTP server starting on %s", server.Addr), "", "", nil)
        if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            container.Logger.LogError(constants.LogLevelFatal, "HTTP server failed", "", "", err, nil)
            log.Fatalf("HTTP server failed: %v", err)
        }
	}()

	// シャットダウンシグナルを待機
	gracefulShutdown(container, server)
}

// buildContainer は全ての依存関係を持つアプリケーションコンテナを構築する
func buildContainer(apiKey string) *ApplicationContainer {
    container := &ApplicationContainer{}

    container.IdleManager = NewIdleTimeoutManager()
    // LOG_LEVEL を反映（未指定は INFO）
    logLevel := getEnv("LOG_LEVEL", constants.LogLevelInfo)
    container.Logger = logging.NewSlogLoggerWithLevel(logLevel) // slog対応ロガー（レベル指定）
	container.YouTubeClient = youtube.NewClient(apiKey)
	container.UserRepository = memory.NewUserRepository()
	// Cloud Run用のメモリ制限設定を環境変数から取得
	maxChatMessages := getEnvAsInt("MAX_CHAT_MESSAGES", 500) // デフォルト500（無料枠向け）

	container.ChatRepository = memory.NewChatRepository(maxChatMessages)
	// 最大メッセージ数/動画（環境変数で調整可能）
	container.EventPublisher = events.NewSimplePublisher(container.Logger)

	// ドメインサービス
	// 先に VideoService を作成（PollingService から利用）
	container.VideoService = service.NewVideoService(
		container.YouTubeClient,
		container.Logger,
	)

	container.PollingService = service.NewPollingService(
		container.YouTubeClient,
		container.ChatRepository,
		container.Logger,
		container.EventPublisher,
		container.VideoService,
	)

	container.UserService = service.NewUserService(
		container.UserRepository,
		container.Logger,
		container.EventPublisher,
	)

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

    // ログをバッファにも集約（SlogLogger のシンクへ接続）
    if sinkable, ok := container.Logger.(interface{ SetSink(func(string, string, string, string, string, string, map[string]interface{})) }); ok {
        sinkable.SetSink(func(level, component, event, message, videoID, correlationID string, ctx map[string]interface{}) {
            container.LogManagementUC.AddLogEntry(level, component, event, message, videoID, correlationID, ctx)
        })
    }

	// StructuredLoggerにLogManagementUseCaseを設定（循環依存回避）
	container.MonitoringHandler = handler.NewMonitoringHandler(
		container.ChatMonitoringUC,
		container.Logger,
	)

    container.SSEHandler = handler.NewSSEHandler(
        container.ChatMonitoringUC,
        container.Logger,
    )
    // SSE 送信時にもアイドル更新
    container.SSEHandler.SetIdleTouch(container.IdleManager.UpdateLastRequest)

	container.LogHandler = handler.NewLogHandler(
		container.LogManagementUC,
		container.Logger,
	)

	container.StaticHandler = handler.NewStaticHandler(
		container.Logger,
	)

	// ヘルスチェックハンドラー
	container.HealthHandler = handler.NewHealthHandler(container.ChatMonitoringUC, container.Logger)

	return container
}

// getEnvAsInt 環境変数を整数として取得（デフォルト値付き）
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// setupHTTPServer はHTTPサーバーを設定して返す
func setupHTTPServer(container *ApplicationContainer) *http.Server {
	// Chiルーターを初期化
	r := chi.NewRouter()

	// ミドルウェアを追加
	r.Use(middleware.RequestID) // リクエストID生成（slog対応）
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// アイドルタイムアウト更新ミドルウェア
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			container.IdleManager.UpdateLastRequest()
			next.ServeHTTP(w, r)
		})
	})

	// CORS設定を追加
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(204)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// ヘルスチェックエンドポイント（Cloud Run用）
	r.Get("/health", container.HealthHandler.Health)
	r.Get("/ready", container.HealthHandler.Ready)

	// favicon(空) で404ログ抑止
	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.WriteHeader(http.StatusNoContent) // 204
	})

	// 静的ページ
	r.Get("/", container.StaticHandler.ServeHome)
	r.Get("/users", container.StaticHandler.ServeUserListPage)
	r.Get("/logs", container.StaticHandler.ServeLogsPage)

	// 静的アセット
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// APIエンドポイント
	r.Route("/api", func(r chi.Router) {
		// 監視エンドポイント
		r.Route("/monitoring", func(r chi.Router) {
			r.Post("/start", container.MonitoringHandler.StartMonitoring)
			r.Post("/resume", container.MonitoringHandler.ResumeMonitoring)
			r.Delete("/stop", container.MonitoringHandler.StopMonitoring)
			r.Get("/active", container.MonitoringHandler.GetActiveVideoID)
			r.Get("/{videoId}/users", container.MonitoringHandler.GetUserList)
			r.Get("/{videoId}/status", container.MonitoringHandler.GetVideoStatus)
			// 自動終了検知の切替/取得
			r.Get("/auto-end", container.MonitoringHandler.GetAutoEndSetting)
			r.Post("/auto-end", container.MonitoringHandler.SetAutoEndSetting)
		})

		// SSEエンドポイント
		r.Route("/sse", func(r chi.Router) {
			r.Get("/{videoId}/users", container.SSEHandler.StreamUserList)
		})

		// ログエンドポイント（統合）
		r.Route("/logs", func(r chi.Router) {
			r.Get("/", container.LogHandler.GetLogs)
			r.Delete("/", container.LogHandler.ClearLogs)
		})
	})

	// 適切なタイムアウト設定でサーバーを作成（YouTube APIポーリング対応）
	server := &http.Server{
		Addr:              ":" + getEnv("PORT", "8080"),
		Handler:           r,
		ReadTimeout:       30 * time.Second,  // 延長: YouTube APIリクエスト対応
		WriteTimeout:      60 * time.Second,  // 大幅延長: ポーリング処理対応
		IdleTimeout:       120 * time.Second, // 延長: 長時間接続対応
		ReadHeaderTimeout: 10 * time.Second,  // 延長: セキュリティとバランス
	}

	return server
}

// gracefulShutdown はシャットダウンシグナルを待機し、サーバーを優雅にシャットダウンする
func gracefulShutdown(container *ApplicationContainer, server *http.Server) {
	// Context-based signal handling (Go 1.16+)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// SSE接続のクリーンアップ処理を登録
	server.RegisterOnShutdown(func() {
        container.Logger.LogStructured(constants.LogLevelInfo, "main", "shutdown_cleanup", "Cleaning up active connections", "", "", nil)
		// TODO: SSEハンドラーに接続クリーンアップメソッドを追加する予定
	})

	// シグナルまたはアイドルタイムアウトを待機
	select {
    case <-ctx.Done():
        container.Logger.LogStructured(constants.LogLevelInfo, "main", "shutdown_start", "Shutdown signal received", "", "", nil)
    case <-container.IdleManager.GetStopChannel():
        container.Logger.LogStructured(constants.LogLevelInfo, "main", "shutdown_start", "Idle timeout reached, shutting down", "", "", nil)
	}

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), constants.ShutdownTimeout)
	defer cancel()

    container.Logger.LogStructured(constants.LogLevelInfo, "main", "shutdown_progress", "Starting graceful shutdown", "", "", map[string]interface{}{
        "timeout": constants.ShutdownTimeout.String(),
    })

    if err := server.Shutdown(shutdownCtx); err != nil {
        container.Logger.LogError(constants.LogLevelError, "Server shutdown error", "", "", err, map[string]interface{}{
            "timeout": constants.ShutdownTimeout.String(),
        })
        // Force shutdown if graceful shutdown fails
        container.Logger.LogStructured(constants.LogLevelWarning, "main", "shutdown_force", "Forcing server shutdown", "", "", nil)
        err := server.Close()
        if err != nil {
            container.Logger.LogError(constants.LogLevelError, "Server force close error", "", "", err, nil)
        }
    }

    container.Logger.LogStructured(constants.LogLevelInfo, "main", "shutdown_complete", "Application shutdown complete", "", "", nil)
}

// getEnv はデフォルト値付きで環境変数を取得する
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
