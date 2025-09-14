package main

import (
    "log"
    "net/http"
    "os"

    "github.com/joho/godotenv"
    ahttp "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/http"
    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/memory"
    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/youtube"
    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase"
)

func main() {
    _ = godotenv.Load()

    port := getenv("PORT", "8080")
    frontend := os.Getenv("FRONTEND_ORIGIN")
    ytKey := os.Getenv("YT_API_KEY")

    // Adapters
    users := memory.NewUserRepo()
    state := memory.NewStateRepo()
    yt := youtube.New(ytKey)

    // UseCases（未実装のため呼び出し時は 501 を返す想定）
    ucStatus := &usecase.Status{Users: users, State: state}
    ucSwitch := &usecase.SwitchVideo{YT: yt, Users: users, State: state}
    ucPull := &usecase.Pull{YT: yt, Users: users, State: state}
    ucReset := &usecase.Reset{Users: users, State: state}

    h := &ahttp.Handlers{Status: ucStatus, SwitchVideo: ucSwitch, Pull: ucPull, Reset: ucReset}
    srv := &http.Server{Addr: ":" + port, Handler: ahttp.NewRouter(h, frontend)}

    log.Printf("listening on :%s", port)
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatal(err)
    }
}

func getenv(k, def string) string {
    if v := os.Getenv(k); v != "" {
        return v
    }
    return def
}
