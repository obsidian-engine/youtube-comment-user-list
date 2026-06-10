package http

import (
	"encoding/json"
	"errors"
	"log"
	stdhttp "net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/logging"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase/snapshot"
)

type Handlers struct {
	Status      *usecase.Status
	SwitchVideo *usecase.SwitchVideo
	Pull        *usecase.Pull
	Reset       *usecase.Reset
	Users       port.UserRepo
	Comments    port.CommentRepo
	Coord       snapshot.Coordinator
	ListHistory *usecase.ListHistorySnapshots
	GetHistory  *usecase.GetHistorySnapshot
}

// StatusResponse represents the response for /status endpoint
type StatusResponse struct {
	Status          string      `json:"status"`
	Count           int         `json:"count"`
	VideoID         string      `json:"videoId"`
	LiveChatID      string      `json:"liveChatId"`
	StartedAt       any         `json:"startedAt"`
	EndedAt         any         `json:"endedAt"`
	LastPulledAt    any         `json:"lastPulledAt"`
	SnapshotSavedAt *time.Time  `json:"snapshotSavedAt,omitempty"`
	Logs            []LogDetail `json:"logs,omitempty"`
}

// SwitchVideoResponse represents the response for /switch-video endpoint
type SwitchVideoResponse struct {
	Status     string      `json:"status"`
	VideoID    string      `json:"videoId"`
	LiveChatID string      `json:"liveChatId"`
	StartedAt  any         `json:"startedAt"`
	Logs       []LogDetail `json:"logs,omitempty"`
}

// LogDetail represents a single log entry in the response
type LogDetail struct {
	Level   string `json:"level"`
	Source  string `json:"source"`
	Message string `json:"message"`
}

// PullResponse represents the response for /pull endpoint
type PullResponse struct {
	AddedCount            int         `json:"addedCount"`
	SkippedCount          int         `json:"skippedCount"`
	AutoReset             bool        `json:"autoReset"`
	PollingIntervalMillis int64       `json:"pollingIntervalMillis"`
	Logs                  []LogDetail `json:"logs,omitempty"`
}

// ResetResponse represents the response for /reset endpoint
type ResetResponse struct {
	Status string      `json:"status"`
	Logs   []LogDetail `json:"logs,omitempty"`
}

func collectLogs(collector *logging.Collector) []LogDetail {
	if collector == nil {
		return nil
	}
	entries := collector.Entries()
	logs := make([]LogDetail, len(entries))
	for i, e := range entries {
		logs[i] = LogDetail{Level: e.Level, Source: e.Source, Message: e.Message}
	}
	return logs
}

func NewRouter(h *Handlers, frontendOrigin string) stdhttp.Handler {
	r := chi.NewRouter()

	// ミドルウェアの設定
	// chi の r.Use 登録順 = 外側から内側
	// RecoverMiddleware を最外周にし、内側全 handler の panic を catch して Cloud Run server log に stack trace を残す。
	// CollectorMiddleware は usecase / handler 層で context から取り出して frontend に logs を返すためのもの。
	// 設計判断: panic 時の stack trace は frontend に流さず (security 観点と非対称性回避)、server log のみ。
	r.Use(RecoverMiddleware)
	r.Use(LoggingMiddleware)
	r.Use(CORSMiddleware(frontendOrigin))
	r.Use(CollectorMiddleware)

	r.Get("/status", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		log.Printf("[STATUS] Getting current status")
		collector := collectorFromRequest(r)
		out, err := h.Status.Execute(r.Context())
		if err != nil {
			log.Printf("[STATUS] Error: %v", err)
			renderInternalErrorWithCollector(w, r, "Failed to get status", collector)
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
			Logs:         collectLogs(collector),
		}
		if h.Coord != nil {
			if savedAt := h.Coord.LastSavedAt(); !savedAt.IsZero() {
				response.SnapshotSavedAt = &savedAt
			}
		}
		render.JSON(w, r, response)
	})

	// [logs-non-conformant] /users.json は root=array endpoint のため logs を同梱しない。
	// 他全 endpoint (/status /switch-video /pull /reset /history/*) は
	// ErrorResponse / SuccessResponse の logs フィールド ([]LogDetail) を持つが、
	// /users.json は domain.User スライスをそのまま root に返す設計のため
	// response wrapper を挿入できず logs 非対応のまま維持する。
	//
	// frontend 実装上の注意: root array endpoint は logs を期待しないこと。
	// 将来的に {users: [...], logs: [...]} でラップする re-design 案があるが現時点では着手しない。
	r.Get("/users.json", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		log.Printf("[USERS] Getting user list with join time")
		users := h.Users.ListUsersSortedByJoinTime()
		log.Printf("[USERS] Returning %d users sorted by join time", len(users))
		render.JSON(w, r, users)
	})

	r.Post("/switch-video", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		log.Printf("[SWITCH_VIDEO] Processing video switch request")
		collector := collectorFromRequest(r)
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
			renderUsecaseError(w, r, err, "Failed to switch video: "+err.Error(), collector, StatusBadGateway, "bad_gateway")
			return
		}

		log.Printf("[SWITCH_VIDEO] Successfully switched to video %s, status: %s", out.State.VideoID, out.State.Status)
		response := SwitchVideoResponse{
			Status:     string(out.State.Status),
			VideoID:    out.State.VideoID,
			LiveChatID: out.State.LiveChatID,
			StartedAt:  out.State.StartedAt,
			Logs:       collectLogs(collector),
		}
		render.JSON(w, r, response)
	})

	r.Post("/pull", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		log.Printf("[PULL] Processing pull request")
		// CollectorMiddleware が既に context に inject 済みのため、middleware 経由で取得
		collector := collectorFromRequest(r)
		out, err := h.Pull.Execute(r.Context())
		if err != nil {
			log.Printf("[PULL] Error: %v", err)
			renderUsecaseError(w, r, err, err.Error(), collector, StatusInternalServerError, "internal_error")
			return
		}

		log.Printf("[PULL] Added %d messages, AutoReset: %v, Polling(ms): %d", out.AddedCount, out.AutoReset, out.PollingIntervalMillis)
		response := PullResponse{
			AddedCount:            out.AddedCount,
			SkippedCount:          out.SkippedCount,
			AutoReset:             out.AutoReset,
			PollingIntervalMillis: out.PollingIntervalMillis,
			Logs:                  collectLogs(collector),
		}
		render.JSON(w, r, response)
	})

	r.Post("/reset", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		log.Printf("[RESET] Processing reset request")
		collector := collectorFromRequest(r)
		out, err := h.Reset.Execute(r.Context())
		if err != nil {
			log.Printf("[RESET] Error: %v", err)
			renderUsecaseError(w, r, err, "Failed to reset: "+err.Error(), collector, StatusInternalServerError, "internal_error")
			return
		}

		log.Printf("[RESET] Reset complete, status: %s", out.State.Status)
		response := ResetResponse{
			Status: string(out.State.Status),
			Logs:   collectLogs(collector),
		}
		render.JSON(w, r, response)
	})

	r.Get("/history/snapshots", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		collector := collectorFromRequest(r)
		if h.ListHistory == nil {
			renderInternalErrorWithCollector(w, r, "history listing is not available", collector)
			return
		}
		out, err := h.ListHistory.Execute(r.Context())
		if err != nil {
			log.Printf("[HISTORY_LIST] Error: %v", err)
			renderInternalErrorWithCollector(w, r, "Failed to list history snapshots", collector)
			return
		}
		items := make([]HistorySummaryResponse, len(out.Items))
		for i, s := range out.Items {
			items[i] = newHistorySummaryResponse(s)
		}
		render.JSON(w, r, HistoryListResponse{Items: items, Logs: collectLogs(collector)})
	})

	r.Get("/history/snapshots/{videoID}", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		collector := collectorFromRequest(r)
		if h.GetHistory == nil {
			renderInternalErrorWithCollector(w, r, "history detail is not available", collector)
			return
		}
		videoID := chi.URLParam(r, "videoID")
		out, err := h.GetHistory.Execute(r.Context(), videoID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				RenderNotFoundError(w, r, "snapshot not found")
				return
			}
			log.Printf("[HISTORY_GET] Error: %v", err)
			renderInternalErrorWithCollector(w, r, "Failed to get history snapshot", collector)
			return
		}
		resp := newHistorySnapshotResponse(out.Snapshot)
		resp.Logs = collectLogs(collector)
		render.JSON(w, r, resp)
	})

	r.Get("/comments", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		collector := collectorFromRequest(r)
		keywordsParam := r.URL.Query().Get("keywords")
		if keywordsParam == "" {
			renderBadRequest(w, r, "keywords parameter is required")
			return
		}

		const maxKeywordLength = 100
		const maxKeywords = 20

		keywords := []string{}
		for keyword := range strings.SplitSeq(keywordsParam, ",") {
			trimmed := strings.TrimSpace(keyword)
			if trimmed != "" {
				if len(trimmed) > maxKeywordLength {
					renderBadRequest(w, r, "keyword too long (max 100 characters)")
					return
				}
				keywords = append(keywords, trimmed)
				if len(keywords) > maxKeywords {
					renderBadRequest(w, r, "too many keywords (max 20)")
					return
				}
			}
		}

		if len(keywords) == 0 {
			renderBadRequest(w, r, "at least one keyword is required")
			return
		}

		// CommentRepo未初期化チェック
		if h.Comments == nil {
			log.Printf("[COMMENTS] CommentRepo not initialized")
			renderInternalErrorWithCollector(w, r, "comment search is not available", collector)
			return
		}

		log.Printf("[COMMENTS] Searching comments with keywords: %v", keywords)
		comments := h.Comments.SearchByKeywords(keywords)
		log.Printf("[COMMENTS] Found %d comments", len(comments))

		render.JSON(w, r, comments)
	})

	return r
}
