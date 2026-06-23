package usecase

import (
	"context"
	"fmt"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/logging"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase/snapshot"
)

type ReserveInput struct {
	VideoID string
	// Details が非 nil のとき GetVideoLiveDetails をスキップ (StartOrReserve からの再利用)。
	Details *port.VideoLiveDetails
}

type ReserveOutput struct {
	State domain.LiveState
}

type Reserve struct {
	YT    port.YouTubePort
	State port.StateRepo
	Clock port.Clock
	Snap  snapshot.Coordinator
}

// Execute: videoId を予約状態に遷移。ACTIVE 中なら conflict、非 live なら invalid argument。
// users/comments は触らない (実際の切替は monitor → SwitchVideo に任せる)。
func (uc *Reserve) Execute(ctx context.Context, in ReserveInput) (ReserveOutput, error) {
	if in.VideoID == "" {
		return ReserveOutput{}, &domain.APIError{Code: domain.ErrCodeInvalidArgument, Message: "videoId is required"}
	}

	cur, err := uc.State.Get(ctx)
	if err != nil {
		return ReserveOutput{}, fmt.Errorf("state_get: %w", err)
	}
	if cur.Status == domain.StatusActive {
		return ReserveOutput{}, &domain.APIError{Code: domain.ErrCodeConflict, Message: "stream is currently active, reset first"}
	}

	var details port.VideoLiveDetails
	if in.Details != nil {
		details = *in.Details
	} else {
		var err error
		details, err = uc.YT.GetVideoLiveDetails(ctx, in.VideoID)
		if err != nil {
			return ReserveOutput{}, fmt.Errorf("get_video_live_details: %w", err)
		}
	}
	if !details.IsLiveContent {
		return ReserveOutput{}, &domain.APIError{Code: domain.ErrCodeInvalidArgument, Message: "video is not a live stream"}
	}

	now := uc.Clock.Now()
	newState := domain.LiveState{
		Status:               domain.StatusReserved,
		VideoID:              in.VideoID,
		LiveChatID:           details.LiveChatID, // チャット未開なら空
		AutonomousMonitoring: true,
		ReservedAt:           now,
		ScheduledStartTime:   details.ScheduledStartTime,
	}
	if err := uc.State.Set(ctx, newState); err != nil {
		return ReserveOutput{}, fmt.Errorf("state_set: %w", err)
	}

	uc.Snap.MarkDirty()
	if err := uc.Snap.Flush(ctx); err != nil {
		logging.Log(ctx, "warn", "SNAPSHOT", "reserve: snapshot flush failed: %v", err)
	}

	return ReserveOutput{State: newState}, nil
}
