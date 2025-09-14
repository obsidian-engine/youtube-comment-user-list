package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	ahttp "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/http"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/memory"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/system"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/youtube"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/config"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase"
)

func main() {
	_ = godotenv.Load()

	// 設定の読み込みと検証
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Adapters
	users := memory.NewUserRepo()
	state := memory.NewStateRepo()
	yt := youtube.New(cfg.YouTubeAPIKey)
	clock := system.NewSystemClock()

	// UseCases（未実装のため呼び出し時は 501 を返す想定）
	ucStatus := &usecase.Status{Users: users, State: state}
	ucSwitch := &usecase.SwitchVideo{YT: yt, Users: users, State: state, Clock: clock}
	ucPull := &usecase.Pull{YT: yt, Users: users, State: state}
	ucReset := &usecase.Reset{Users: users, State: state}

	h := &ahttp.Handlers{Status: ucStatus, SwitchVideo: ucSwitch, Pull: ucPull, Reset: ucReset, Users: users}
	srv := &http.Server{Addr: ":" + cfg.Port, Handler: ahttp.NewRouter(h, cfg.FrontendOrigin)}

	log.Printf("Server starting on port %s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
