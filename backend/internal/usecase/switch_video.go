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
	YT       port.YouTubePort
	Users    port.UserRepo
	Comments port.CommentRepo
	State    port.StateRepo
	Clock    port.Clock
	Snap     snapshot.Coordinator // 必須 (GCS 不要な場合は NopCoordinator を渡す)
}

// Execute: videoId 切替、ユーザー初期化、State=ACTIVE に遷移。
// 同じ videoId に対する切替の場合は users / state.StartedAt を維持する（配信再開ユースケース）。
func (uc *SwitchVideo) Execute(ctx context.Context, in SwitchVideoInput) (SwitchVideoOutput, error) {
	// 1. YouTube APIでliveChatIDを取得（失敗時はここで返るので snapshot 操作はしない）
	liveChatID, err := uc.YT.GetActiveLiveChatID(ctx, in.VideoID)
	if err != nil {
		// 配信終了済 video への再切替で API error が発生した場合、同一 videoId かつ
		// in-memory に users が残っていれば snapshot 復元データとして WAITING 状態で表示する。
		prevState, _ := uc.State.Get(ctx)
		if prevState.VideoID == in.VideoID && uc.Users.Count() > 0 {
			log.Printf("[INFO] switch_video: API error on same videoId, restoring from in-memory snapshot (videoId=%s, users=%d): %v",
				in.VideoID, uc.Users.Count(), err)
			now := uc.Clock.Now()
			startedAt := prevState.StartedAt
			if startedAt.IsZero() {
				startedAt = now
			}
			restoredState := domain.LiveState{
				Status:        domain.StatusWaiting,
				VideoID:       in.VideoID,
				LiveChatID:    prevState.LiveChatID,
				StartedAt:     startedAt,
				EndedAt:       now,
				NextPageToken: "",
			}
			if setErr := uc.State.Set(ctx, restoredState); setErr != nil {
				return SwitchVideoOutput{}, setErr
			}
			return SwitchVideoOutput{State: restoredState}, nil
		}
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
	restored := false
	if !sameVideo {
		uc.Users.Clear()
		if uc.Comments != nil {
			uc.Comments.Clear()
		}
		r, rerr := uc.Snap.RestoreFor(ctx, in.VideoID)
		if rerr != nil {
			log.Printf("[WARN] switch_video: restoreFor failed: %v", rerr)
		} else {
			restored = r
		}
	}

	// 4. StateをACTIVEに更新
	now := uc.Clock.Now()
	startedAt := now
	if sameVideo && !prevState.StartedAt.IsZero() {
		startedAt = prevState.StartedAt
	} else if restored {
		// RestoreFor が State を復元している場合、復元された StartedAt を引き継ぐ
		if restoredState, err := uc.State.Get(ctx); err == nil && !restoredState.StartedAt.IsZero() {
			startedAt = restoredState.StartedAt
		}
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
