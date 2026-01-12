package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	ahttp "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/http"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/memory"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/system"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/youtube"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/config"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase"
)

func main() {
	// .envファイルの読み込み（本番環境では存在しない場合があるため、エラーでも続行）
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Warning: can not read env file (this is normal in production): %v", err)
	}

	// 設定の読み込みと検証
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// ログ設定を適用
	cfg.SetupLogger()

	// Adapters
	users := memory.NewUserRepo()
	comments := memory.NewCommentRepo()
	state := memory.NewStateRepo()
	yt := youtube.New(cfg.YouTubeAPIKey)
	clock := system.NewSystemClock()

	// UseCases
	ucStatus := &usecase.Status{Users: users, State: state}
	ucSwitch := &usecase.SwitchVideo{YT: yt, Users: users, State: state, Clock: clock}
	ucPull := &usecase.Pull{YT: yt, Users: users, Comments: comments, State: state, Clock: clock}
	ucReset := &usecase.Reset{Users: users, State: state}

	h := &ahttp.Handlers{Status: ucStatus, SwitchVideo: ucSwitch, Pull: ucPull, Reset: ucReset, Users: users, Comments: comments}
	srv := &http.Server{Addr: ":" + cfg.Port, Handler: ahttp.NewRouter(h, cfg.FrontendOrigin)}

	// グレースフルシャットダウンのためのコンテキスト設定
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// サーバーを別ゴルーチンで起動
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	// シャットダウンシグナルを待機
	<-ctx.Done()
	log.Println("Shutting down server gracefully...")

	// シャットダウンのタイムアウト設定（30秒）
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("Server shutdown completed")
	}
}
