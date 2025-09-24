package http

import (
	"encoding/json"
	stdhttp "net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/logging"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase"
)

// RefactoredHandlers は新しい構造化ログとレスポンスヘルパーを使用するハンドラー
type RefactoredHandlers struct {
	Status      *usecase.Status
	SwitchVideo *usecase.SwitchVideo
	Pull        *usecase.Pull
	Reset       *usecase.Reset
	Users       port.UserRepo
	Logger      logging.ModuleLogger
}

// NewRefactoredRouter は新機能を使用したルーターを作成します
func NewRefactoredRouter(h *RefactoredHandlers, frontendOrigin string) stdhttp.Handler {
	r := chi.NewRouter()

	// ミドルウェアの設定
	r.Use(LoggingMiddleware)
	r.Use(CORSMiddleware(frontendOrigin))

	r.Get("/status", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		h.Logger.Info("Getting current status")
		out, err := h.Status.Execute(r.Context())
		if err != nil {
			h.Logger.Error("Failed to get status", "error", err)
			RenderInternalServerError(w, r, "Failed to get status")
			return
		}

		h.Logger.Info("Current status retrieved", "status", out.Status, "count", out.Count)
		
		// 型安全なレスポンス構造体を使用
		response := TypeSafeStatusResponse{
			Status:       string(out.Status),
			Count:        out.Count,
			VideoID:      out.VideoID,
			LiveChatID:   out.LiveChatID,
			StartedAt:    convertToTimePtr(out.StartedAt),
			EndedAt:      convertToTimePtr(out.EndedAt),
			LastPulledAt: convertToTimePtr(out.LastPulledAt),
		}
		RenderSuccessResponse(w, r, stdhttp.StatusOK, response)
	})

	r.Get("/users.json", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		h.Logger.Info("Getting user list with join time")
		users := h.Users.ListUsersSortedByJoinTime()
		h.Logger.Info("Retrieved user list", "count", len(users))
		RenderSuccessResponse(w, r, stdhttp.StatusOK, users)
	})

	r.Post("/switch-video", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		h.Logger.Info("Processing video switch request")
		var req struct {
			VideoID string `json:"videoId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Logger.Error("Invalid JSON in switch-video request", "error", err)
			RenderBadRequestError(w, r, "Invalid JSON format")
			return
		}

		if req.VideoID == "" {
			h.Logger.Error("Missing videoId in switch-video request")
			RenderBadRequestError(w, r, "videoId is required")
			return
		}

		h.Logger.Info("Processing video switch", "videoId_input", req.VideoID, "length", len(req.VideoID))

		// URL形式の場合はvideo_idを抽出
		videoID, err := ExtractVideoID(req.VideoID)
		if err != nil {
			h.Logger.Error("Invalid video ID or URL", "input", req.VideoID, "error", err)
			RenderBadRequestError(w, r, "Invalid video ID or URL: "+err.Error())
			return
		}

		h.Logger.Info("Extracted video ID", "extracted_id", videoID, "original_input", req.VideoID)
		out, err := h.SwitchVideo.Execute(r.Context(), usecase.SwitchVideoInput{VideoID: videoID})
		if err != nil {
			h.Logger.Error("Failed to switch video", "video_id", videoID, "error", err)
			RenderBadGatewayError(w, r, "Failed to switch video: "+err.Error())
			return
		}

		h.Logger.Info("Successfully switched video", "video_id", out.State.VideoID, "status", out.State.Status)
		response := TypeSafeSwitchVideoResponse{
			Status:     string(out.State.Status),
			VideoID:    out.State.VideoID,
			LiveChatID: out.State.LiveChatID,
			StartedAt:  convertToTimePtr(out.State.StartedAt),
		}
		RenderSuccessResponse(w, r, stdhttp.StatusOK, response)
	})

	r.Post("/pull", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		h.Logger.Info("Processing pull request")
		out, err := h.Pull.Execute(r.Context())
		if err != nil {
			h.Logger.Error("Failed to pull messages", "error", err)
			RenderInternalServerError(w, r, "Failed to pull messages")
			return
		}

		h.Logger.Info("Pull completed", "added_count", out.AddedCount, "auto_reset", out.AutoReset, "polling_ms", out.PollingIntervalMillis)
		response := TypeSafePullResponse{
			AddedCount:            out.AddedCount,
			AutoReset:             out.AutoReset,
			PollingIntervalMillis: out.PollingIntervalMillis,
		}
		RenderSuccessResponse(w, r, stdhttp.StatusOK, response)
	})

	r.Post("/reset", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		h.Logger.Info("Processing reset request")
		out, err := h.Reset.Execute(r.Context())
		if err != nil {
			h.Logger.Error("Failed to reset", "error", err)
			RenderInternalServerError(w, r, "Failed to reset")
			return
		}

		h.Logger.Info("Reset completed", "status", out.State.Status)
		response := TypeSafeResetResponse{
			Status: string(out.State.Status),
		}
		RenderSuccessResponse(w, r, stdhttp.StatusOK, response)
	})

	return r
}

// convertToTimePtr はinterface{}を*time.Timeに安全に変換するヘルパー関数
func convertToTimePtr(value interface{}) *time.Time {
	if value == nil {
		return nil
	}
	if t, ok := value.(time.Time); ok {
		return &t
	}
	return nil
}

// NewRefactoredHandlersFromLegacy は既存のHandlersから新しいRefactoredHandlersを作成
func NewRefactoredHandlersFromLegacy(legacy *Handlers, logger logging.ModuleLogger) *RefactoredHandlers {
	return &RefactoredHandlers{
		Status:      legacy.Status,
		SwitchVideo: legacy.SwitchVideo,
		Pull:        legacy.Pull,
		Reset:       legacy.Reset,
		Users:       legacy.Users,
		Logger:      logger,
	}
}