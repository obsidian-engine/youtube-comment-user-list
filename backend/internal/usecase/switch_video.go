package usecase

import (
	"context"
	"log"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase/snapshot"
)

type SwitchVideoInput struct {
	VideoID string
}

type SwitchVideoOutput struct {
	State domain.LiveState
}

type SwitchVideo struct {
	YT    port.YouTubePort
	Users port.UserRepo
	State port.StateRepo
	Clock port.Clock
	Snap  snapshot.Coordinator // 必須 (GCS 不要な場合は NopCoordinator を渡す)
}

// Execute: videoId 切替、ユーザー初期化、State=ACTIVE に遷移。
// 同じ videoId に対する切替の場合は users / state.StartedAt を維持する（配信再開ユースケース）。
func (uc *SwitchVideo) Execute(ctx context.Context, in SwitchVideoInput) (SwitchVideoOutput, error) {
	// 1. YouTube APIでliveChatIDを取得（失敗時はここで返るので snapshot 操作はしない）
	liveChatID, err := uc.YT.GetActiveLiveChatID(ctx, in.VideoID)
	if err != nil {
		return SwitchVideoOutput{}, err
	}

	// 2. 切替前の状態を snapshot に保存（旧 video の最終状態を確実に残す）
	if err := uc.Snap.Flush(ctx); err != nil {
		log.Printf("[WARN] switch_video: snapshot flush (pre-switch) failed: %v", err)
		// Flush 失敗は警告のみ、切替処理は継続
	}

	// 3. 同じ videoId への再切替は既存 users を維持する（別 videoId の場合のみクリア）
	prevState, _ := uc.State.Get(ctx)
	sameVideo := prevState.VideoID == in.VideoID
	if !sameVideo {
		uc.Users.Clear()
	}

	// 4. StateをACTIVEに更新
	now := uc.Clock.Now()
	startedAt := now
	if sameVideo && !prevState.StartedAt.IsZero() {
		startedAt = prevState.StartedAt
	}
	newState := domain.LiveState{
		Status:        domain.StatusActive,
		VideoID:       in.VideoID,
		LiveChatID:    liveChatID,
		StartedAt:     startedAt,
		NextPageToken: "",
	}

	if err := uc.State.Set(ctx, newState); err != nil {
		return SwitchVideoOutput{}, err
	}

	// 5. 新 videoId を Coordinator に設定し、current.json を即時更新
	uc.Snap.SetVideo(in.VideoID, liveChatID)
	uc.Snap.MarkDirty()
	if err := uc.Snap.Flush(ctx); err != nil {
		log.Printf("[WARN] switch_video: snapshot flush (post-switch) failed: %v", err)
	}

	return SwitchVideoOutput{State: newState}, nil
}
