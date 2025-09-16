package http

import (
	"encoding/json"
	"log"
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

	// ミドルウェアの設定
	r.Use(LoggingMiddleware)
	r.Use(CORSMiddleware(frontendOrigin))

	r.Get("/status", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		log.Printf("[STATUS] Getting current status")
		out, err := h.Status.Execute(r.Context())
		if err != nil {
			log.Printf("[STATUS] Error: %v", err)
			renderInternalError(w, r, "Failed to get status")
			return
		}

		log.Printf("[STATUS] Current status: %s, Users: %d", out.Status, out.Count)
		response := map[string]interface{}{
			"status":       string(out.Status),
			"count":        out.Count,
			"videoId":      out.VideoID,
			"startedAt":    out.StartedAt,
			"endedAt":      out.EndedAt,
			"lastPulledAt": out.LastPulledAt,
		}
		render.JSON(w, r, response)
	})

	r.Get("/users.json", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		log.Printf("[USERS] Getting user list with join time")
		users := h.Users.ListUsersSortedByJoinTime()
		log.Printf("[USERS] Returning %d users sorted by join time", len(users))
		render.JSON(w, r, users)
	})

	r.Post("/switch-video", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		log.Printf("[SWITCH_VIDEO] Processing video switch request")
		var req struct {
			VideoID string `json:"videoId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[SWITCH_VIDEO] Invalid JSON: %v", err)
			renderBadRequest(w, r, "Invalid JSON format")
			return
		}

		if req.VideoID == "" {
			log.Printf("[SWITCH_VIDEO] Missing videoId")
			renderBadRequest(w, r, "videoId is required")
			return
		}

		log.Printf("[SWITCH_VIDEO] Switching to video: %s", req.VideoID)
		out, err := h.SwitchVideo.Execute(r.Context(), usecase.SwitchVideoInput{VideoID: req.VideoID})
		if err != nil {
			log.Printf("[SWITCH_VIDEO] Execute error: %v", err)
			renderBadGateway(w, r, "Failed to switch video: "+err.Error())
			return
		}

		log.Printf("[SWITCH_VIDEO] Successfully switched to video %s, status: %s", out.State.VideoID, out.State.Status)
		response := map[string]interface{}{
			"status":     string(out.State.Status),
			"videoId":    out.State.VideoID,
			"liveChatId": out.State.LiveChatID,
			"startedAt":  out.State.StartedAt,
		}
		render.JSON(w, r, response)
	})

	r.Post("/pull", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		log.Printf("[PULL] Processing pull request")
		out, err := h.Pull.Execute(r.Context())
		if err != nil {
			log.Printf("[PULL] Error: %v", err)
			renderInternalError(w, r, "Failed to pull messages")
			return
		}

		log.Printf("[PULL] Added %d users, AutoReset: %v", out.AddedCount, out.AutoReset)
		response := map[string]interface{}{
			"addedCount": out.AddedCount,
			"autoReset":  out.AutoReset,
		}
		render.JSON(w, r, response)
	})

	r.Post("/reset", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		log.Printf("[RESET] Processing reset request")
		out, err := h.Reset.Execute(r.Context())
		if err != nil {
			log.Printf("[RESET] Error: %v", err)
			renderInternalError(w, r, "Failed to reset")
			return
		}

		log.Printf("[RESET] Reset complete, status: %s", out.State.Status)
		response := map[string]interface{}{
			"status": string(out.State.Status),
		}
		render.JSON(w, r, response)
	})

	return r
}
