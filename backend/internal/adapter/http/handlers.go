package http

import (
    "encoding/json"
    stdhttp "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/render"
    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase"
)

type Handlers struct {
    Status      *usecase.Status
    SwitchVideo *usecase.SwitchVideo
    Pull        *usecase.Pull
    Reset       *usecase.Reset
}

func NewRouter(h *Handlers, frontendOrigin string) stdhttp.Handler {
    r := chi.NewRouter()

    // シンプル CORS（厳密実装は別途ミドルウェアで）
    r.Use(func(next stdhttp.Handler) stdhttp.Handler {
        return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
            if frontendOrigin != "" {
                w.Header().Set("Access-Control-Allow-Origin", frontendOrigin)
                w.Header().Set("Vary", "Origin")
            }
            if r.Method == stdhttp.MethodOptions {
                w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
                w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
                w.WriteHeader(204)
                return
            }
            next.ServeHTTP(w, r)
        })
    })

    r.Get("/status", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
        render.Status(r, 501)
        render.PlainText(w, r, "not implemented")
    })

    r.Get("/users.json", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
        w.Header().Set("Content-Type", "application/json")
        _ = json.NewEncoder(w).Encode([]string{})
    })

    r.Post("/switch-video", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
        render.Status(r, 501)
        render.PlainText(w, r, "not implemented")
    })
    r.Post("/pull", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
        render.Status(r, 501)
        render.PlainText(w, r, "not implemented")
    })
    r.Post("/reset", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
        render.Status(r, 501)
        render.PlainText(w, r, "not implemented")
    })

    return r
}
