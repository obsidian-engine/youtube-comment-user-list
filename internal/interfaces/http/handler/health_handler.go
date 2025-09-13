package handler

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/application/usecase"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/constants"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/repository"
)

// HealthHandler ヘルスチェック用のハンドラー
type HealthHandler struct {
	chatMonitoringUC *usecase.ChatMonitoringUseCase
	startTime        time.Time
	logger           repository.Logger
}

// NewHealthHandler 新しいヘルスチェックハンドラーを作成
func NewHealthHandler(chatMonitoringUC *usecase.ChatMonitoringUseCase, logger repository.Logger) *HealthHandler {
	return &HealthHandler{
		chatMonitoringUC: chatMonitoringUC,
		startTime:        time.Now(),
		logger:           logger,
	}
}

// HealthStatus ヘルスチェックのレスポンス
type HealthStatus struct {
	Status      string           `json:"status"`
	Uptime      string           `json:"uptime"`
	MemoryUsage MemoryStats      `json:"memory_usage"`
	Monitoring  MonitoringStatus `json:"monitoring"`
	Timestamp   string           `json:"timestamp"`
}

// MemoryStats メモリ使用状況
type MemoryStats struct {
	AllocMB      float64 `json:"alloc_mb"`
	TotalAllocMB float64 `json:"total_alloc_mb"`
	SysMB        float64 `json:"sys_mb"`
	NumGC        uint32  `json:"num_gc"`
}

// MonitoringStatus 監視状況
type MonitoringStatus struct {
	Active    bool   `json:"active"`
	VideoID   string `json:"video_id,omitempty"`
	UserCount int    `json:"user_count"`
}

// Health ヘルスチェックエンドポイント
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	cid := correlationIDFrom(r, "http")
	h.logger.LogAPI(constants.LogLevelInfo, "Health check request", "", cid, map[string]interface{}{
		"userAgent":  r.Header.Get("User-Agent"),
		"remoteAddr": r.RemoteAddr,
	})

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 監視状況を取得
	videoID, isActive, exists := h.chatMonitoringUC.GetActiveVideoID()
	monitoring := MonitoringStatus{
		Active: isActive,
	}
	if exists {
		monitoring.VideoID = videoID
		if users, err := h.chatMonitoringUC.GetUserList(r.Context(), videoID); err == nil && users != nil {
			monitoring.UserCount = len(users)
		}
	}

	status := HealthStatus{
		Status: "healthy",
		Uptime: time.Since(h.startTime).String(),
		MemoryUsage: MemoryStats{
			AllocMB:      float64(m.Alloc) / 1024 / 1024,
			TotalAllocMB: float64(m.TotalAlloc) / 1024 / 1024,
			SysMB:        float64(m.Sys) / 1024 / 1024,
			NumGC:        m.NumGC,
		},
		Monitoring: monitoring,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(status)
	if err != nil {
		http.Error(w, "Failed to encode health status", http.StatusInternalServerError)
	}
}

// Ready レディネスチェックエンドポイント（Cloud Run用）
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	// アプリケーションが準備完了状態かチェック
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(map[string]string{
		"status": "ready",
	})
	if err != nil {
		http.Error(w, "Failed to encode readiness response", http.StatusInternalServerError)
	}
}
