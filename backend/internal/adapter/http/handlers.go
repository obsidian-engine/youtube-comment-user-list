package http

import (
	"encoding/json"
	"log"
	stdhttp "net/http"
	"strings"

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
	Comments    port.CommentRepo
}

// StatusResponse represents the response for /status endpoint
type StatusResponse struct {
	Status       string      `json:"status"`
	Count        int         `json:"count"`
	VideoID      string      `json:"videoId"`
	LiveChatID   string      `json:"liveChatId"`
	StartedAt    interface{} `json:"startedAt"`
	EndedAt      interface{} `json:"endedAt"`
	LastPulledAt interface{} `json:"lastPulledAt"`
}

// SwitchVideoResponse represents the response for /switch-video endpoint
type SwitchVideoResponse struct {
	Status     string      `json:"status"`
	VideoID    string      `json:"videoId"`
	LiveChatID string      `json:"liveChatId"`
	StartedAt  interface{} `json:"startedAt"`
}

// PullResponse represents the response for /pull endpoint
type PullResponse struct {
	AddedCount            int   `json:"addedCount"`
	AutoReset             bool  `json:"autoReset"`
	PollingIntervalMillis int64 `json:"pollingIntervalMillis"`
}

// ResetResponse represents the response for /reset endpoint
type ResetResponse struct {
	Status string `json:"status"`
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
		response := StatusResponse{
			Status:       string(out.Status),
			Count:        out.Count,
			VideoID:      out.VideoID,
			LiveChatID:   out.LiveChatID,
			StartedAt:    out.StartedAt,
			EndedAt:      out.EndedAt,
			LastPulledAt: out.LastPulledAt,
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

		log.Printf("[SWITCH_VIDEO] Received videoId: '%s' (length: %d)", req.VideoID, len(req.VideoID))

		// URL形式の場合はvideo_idを抽出
		log.Printf("[SWITCH_VIDEO] Calling ExtractVideoID with: '%s'", req.VideoID)
		videoID, err := ExtractVideoID(req.VideoID)
		if err != nil {
			log.Printf("[SWITCH_VIDEO] Invalid video ID or URL: %v", err)
			renderBadRequest(w, r, "Invalid video ID or URL: "+err.Error())
			return
		}

		log.Printf("[SWITCH_VIDEO] Successfully extracted videoID: '%s' (length: %d) from: '%s'", videoID, len(videoID), req.VideoID)
		log.Printf("[SWITCH_VIDEO] Calling SwitchVideo.Execute with videoID: '%s'", videoID)
		out, err := h.SwitchVideo.Execute(r.Context(), usecase.SwitchVideoInput{VideoID: videoID})
		if err != nil {
			log.Printf("[SWITCH_VIDEO] Execute error: %v", err)
			renderBadGateway(w, r, "Failed to switch video: "+err.Error())
			return
		}

		log.Printf("[SWITCH_VIDEO] Successfully switched to video %s, status: %s", out.State.VideoID, out.State.Status)
		response := SwitchVideoResponse{
			Status:     string(out.State.Status),
			VideoID:    out.State.VideoID,
			LiveChatID: out.State.LiveChatID,
			StartedAt:  out.State.StartedAt,
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

		log.Printf("[PULL] Added %d messages, AutoReset: %v, Polling(ms): %d", out.AddedCount, out.AutoReset, out.PollingIntervalMillis)
		response := PullResponse{
			AddedCount:            out.AddedCount,
			AutoReset:             out.AutoReset,
			PollingIntervalMillis: out.PollingIntervalMillis,
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
		response := ResetResponse{
			Status: string(out.State.Status),
		}
		render.JSON(w, r, response)
	})

	r.Get("/comments", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		keywordsParam := r.URL.Query().Get("keywords")
		if keywordsParam == "" {
			renderBadRequest(w, r, "keywords parameter is required")
			return
		}

		keywords := []string{}
		for _, keyword := range strings.Split(keywordsParam, ",") {
			trimmed := strings.TrimSpace(keyword)
			if trimmed != "" {
				keywords = append(keywords, trimmed)
			}
		}

		if len(keywords) == 0 {
			renderBadRequest(w, r, "at least one keyword is required")
			return
		}

		// CommentRepo未初期化チェック
		if h.Comments == nil {
			log.Printf("[COMMENTS] CommentRepo not initialized")
			renderInternalError(w, r, "comment search is not available")
			return
		}

		log.Printf("[COMMENTS] Searching comments with keywords: %v", keywords)
		comments := h.Comments.SearchByKeywords(keywords)
		log.Printf("[COMMENTS] Found %d comments", len(comments))

		render.JSON(w, r, comments)
	})

	return r
}
