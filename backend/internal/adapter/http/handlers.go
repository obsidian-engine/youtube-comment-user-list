package http

import (
    "encoding/json"
    stdhttp "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/render"
    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase"
)

type Handlers struct {
    Status      *usecase.Status
    SwitchVideo *usecase.SwitchVideo
    Pull        *usecase.Pull
    Reset       *usecase.Reset
    Users       port.UserRepo
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
        out, err := h.Status.Execute(r.Context())
        if err != nil {
            render.Status(r, 500)
            render.PlainText(w, r, "internal error")
            return
        }
        
        response := map[string]interface{}{
            "status":    string(out.Status),
            "count":     out.Count,
            "videoId":   out.VideoID,
            "startedAt": out.StartedAt,
            "endedAt":   out.EndedAt,
        }
        render.JSON(w, r, response)
    })

    r.Get("/users.json", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
        users := h.Users.ListDisplayNames()
        render.JSON(w, r, users)
    })

    r.Post("/switch-video", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
        var req struct {
            VideoID string `json:"videoId"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            render.Status(r, 400)
            render.PlainText(w, r, "invalid JSON")
            return
        }
        
        if req.VideoID == "" {
            render.Status(r, 400)
            render.PlainText(w, r, "videoId is required")
            return
        }
        
        out, err := h.SwitchVideo.Execute(r.Context(), usecase.SwitchVideoInput{VideoID: req.VideoID})
        if err != nil {
            render.Status(r, 502)
            render.PlainText(w, r, "backend error")
            return
        }
        
        response := map[string]interface{}{
            "status":      string(out.State.Status),
            "videoId":     out.State.VideoID,
            "liveChatId":  out.State.LiveChatID,
            "startedAt":   out.State.StartedAt,
        }
        render.JSON(w, r, response)
    })
    r.Post("/pull", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
        out, err := h.Pull.Execute(r.Context())
        if err != nil {
            render.Status(r, 500)
            render.PlainText(w, r, "internal error")
            return
        }
        
        response := map[string]interface{}{
            "addedCount": out.AddedCount,
            "autoReset":  out.AutoReset,
        }
        render.JSON(w, r, response)
    })
    r.Post("/reset", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
        out, err := h.Reset.Execute(r.Context())
        if err != nil {
            render.Status(r, 500)
            render.PlainText(w, r, "internal error")
            return
        }
        
        response := map[string]interface{}{
            "status": string(out.State.Status),
        }
        render.JSON(w, r, response)
    })

    return r
}
