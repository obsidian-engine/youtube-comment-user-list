package usecase

import (
	"context"
	"fmt"

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
	YT       port.YouTubePort
	Users    port.UserRepo    // 別 videoId 予約時のクリア対象。nil 許容 (test 用)。
	Comments port.CommentRepo // 別 videoId 予約時のクリア対象。nil 許容 (test 用)。
	State    port.StateRepo
	Clock    port.Clock
	Snap     snapshot.Coordinator
}

// Execute: videoId を RESERVED 状態に遷移する。
// state 遷移成功後に旧配信の in-memory data / snapshot pointer を新 videoId に切替える (transitionSnapshot)。
func (uc *Reserve) Execute(ctx context.Context, in ReserveInput) (ReserveOutput, error) {
	if in.VideoID == "" {
		return ReserveOutput{}, &domain.APIError{Code: domain.ErrCodeInvalidArgument, Message: "videoId is required"}
	}

	cur, err := uc.State.Get(ctx)
	if err != nil {
		return ReserveOutput{}, fmt.Errorf("state_get: %w", err)
	}
	if err := cur.CanReserve(); err != nil {
		return ReserveOutput{}, err
	}

	details, err := uc.resolveDetails(ctx, in)
	if err != nil {
		return ReserveOutput{}, err
	}
	if !details.IsLiveContent {
		return ReserveOutput{}, &domain.APIError{Code: domain.ErrCodeInvalidArgument, Message: "video is not a live stream"}
	}

	newState := domain.NewReservedState(in.VideoID, details.LiveChatID, details.ScheduledStartTime, uc.Clock.Now())
	if err := uc.State.Set(ctx, newState); err != nil {
		return ReserveOutput{}, fmt.Errorf("state_set: %w", err)
	}

	clearForNewVideo(uc.Users, uc.Comments, cur.VideoID, in.VideoID)
	syncSnapshotVideo(ctx, uc.Snap, in.VideoID, details.LiveChatID, "", "", "reserve")

	return ReserveOutput{State: newState}, nil
}

// resolveDetails は Details が pre-fetched (StartOrReserve 経由) ならそれを、なければ YouTube API から取得する。
func (uc *Reserve) resolveDetails(ctx context.Context, in ReserveInput) (port.VideoLiveDetails, error) {
	if in.Details != nil {
		return *in.Details, nil
	}
	d, err := uc.YT.GetVideoLiveDetails(ctx, in.VideoID)
	if err != nil {
		return port.VideoLiveDetails{}, fmt.Errorf("get_video_live_details: %w", err)
	}
	return d, nil
}
