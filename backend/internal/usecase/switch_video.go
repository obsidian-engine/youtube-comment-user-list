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
	Snap  snapshot.Coordinator
}

// Execute: videoId 切替、ユーザー初期化、State=ACTIVE に遷移。
func (uc *SwitchVideo) Execute(ctx context.Context, in SwitchVideoInput) (SwitchVideoOutput, error) {
	// 1. YouTube APIでliveChatIDを取得（失敗時はここで返るので snapshot 操作はしない）
	liveChatID, err := uc.YT.GetActiveLiveChatID(ctx, in.VideoID)
	if err != nil {
		return SwitchVideoOutput{}, err
	}

	// 2. 切替前の状態を snapshot に保存（旧 video の最終状態を確実に残す）
	if uc.Snap != nil {
		if err := uc.Snap.Flush(ctx); err != nil {
			log.Printf("[WARN] switch_video: snapshot flush (pre-switch) failed: %v", err)
			// Flush 失敗は警告のみ、切替処理は継続
		}
	}

	// 3. ユーザーをクリア
	uc.Users.Clear()

	// 4. StateをACTIVEに更新
	now := uc.Clock.Now()
	newState := domain.LiveState{
		Status:        domain.StatusActive,
		VideoID:       in.VideoID,
		LiveChatID:    liveChatID,
		StartedAt:     now,
		NextPageToken: "",
	}

	if err := uc.State.Set(ctx, newState); err != nil {
		return SwitchVideoOutput{}, err
	}

	// 5. 新 videoId を Coordinator に設定し、current.json を即時更新
	if uc.Snap != nil {
		uc.Snap.SetVideo(in.VideoID, liveChatID)
		uc.Snap.MarkDirty()
		if err := uc.Snap.Flush(ctx); err != nil {
			log.Printf("[WARN] switch_video: snapshot flush (post-switch) failed: %v", err)
		}
	}

	return SwitchVideoOutput{State: newState}, nil
}
