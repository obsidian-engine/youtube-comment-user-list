package usecase

import (
	"context"
	"fmt"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/logging"
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

// Execute: videoId を ACTIVE に切替える。API error 時は 2 段階の fallback で終了配信を復元する。
// 同 videoId 切替は既存 users/StartedAt を維持、別 videoId は新規開始。
func (uc *SwitchVideo) Execute(ctx context.Context, in SwitchVideoInput) (SwitchVideoOutput, error) {
	meta, err := uc.YT.GetActiveLiveChatID(ctx, in.VideoID)
	if err != nil {
		return uc.fallbackOnAPIError(ctx, in.VideoID, err)
	}

	// 切替前の状態を snapshot に保存 (旧 video の最終状態を確実に残す)
	if flushErr := uc.Snap.Flush(ctx); flushErr != nil {
		logging.Log(ctx, "warn", "SNAPSHOT", "switch_video: snapshot flush (pre-switch) failed: %v", flushErr)
	}

	prevState, _ := uc.State.Get(ctx)
	sameVideo := prevState.VideoID == in.VideoID

	// 別 videoId 切替時のみ users/comments をクリアし GCS snapshot から復元を試みる
	if !sameVideo {
		uc.Users.Clear()
		if uc.Comments != nil {
			uc.Comments.Clear()
		}
		if _, rerr := uc.Snap.RestoreFor(ctx, in.VideoID); rerr != nil {
			logging.Log(ctx, "warn", "SNAPSHOT", "switch_video: restoreFor failed: %v", rerr)
		}
	}

	// RestoreFor で state.StartedAt が上書きされている可能性があるので再取得
	basisState, _ := uc.State.Get(ctx)
	newState := basisState.Activate(in.VideoID, meta.LiveChatID, uc.Clock.Now())
	if err := uc.State.Set(ctx, newState); err != nil {
		return SwitchVideoOutput{}, fmt.Errorf("state_set: %w", err)
	}

	syncSnapshotVideo(ctx, uc.Snap, in.VideoID, meta.LiveChatID, meta.Title, meta.ChannelTitle, "switch_video")

	return SwitchVideoOutput{State: newState}, nil
}

// fallbackOnAPIError は GetActiveLiveChatID が error を返した場合の 2 段階 fallback を実行する。
//  1. 同 videoId + in-memory に users あり → in-memory snapshot を WAITING で表示
//  2. GCS snapshot に該当 videoId あり → GCS から復元して WAITING で表示
//
// いずれも該当しなければ元の API error を wrap して返す。
func (uc *SwitchVideo) fallbackOnAPIError(ctx context.Context, videoID string, apiErr error) (SwitchVideoOutput, error) {
	prevState, _ := uc.State.Get(ctx)
	now := uc.Clock.Now()

	// Fallback 1: 同 videoId + in-memory users あり → 既存 in-memory を WAITING で再表示
	if prevState.VideoID == videoID && uc.Users.Count() > 0 {
		logging.Log(ctx, "info", "SNAPSHOT",
			"switch_video: API error on same videoId, restoring from in-memory snapshot (videoId=%s, users=%d): %v",
			videoID, uc.Users.Count(), apiErr)
		restored := prevState.MarkEnded(videoID, prevState.LiveChatID, now)
		if err := uc.State.Set(ctx, restored); err != nil {
			return SwitchVideoOutput{}, fmt.Errorf("state_set: %w", err)
		}
		return SwitchVideoOutput{State: restored}, nil
	}

	// Fallback 2: GCS snapshot から復元 (M シナリオ: cold start / 別 video 経由)
	if prevState.VideoID != videoID {
		uc.Users.Clear()
		if uc.Comments != nil {
			uc.Comments.Clear()
		}
	}
	gcsRestored, rerr := uc.Snap.RestoreFor(ctx, videoID)
	if rerr != nil {
		logging.Log(ctx, "warn", "SNAPSHOT", "switch_video: GCS fallback restoreFor failed: %v", rerr)
		return SwitchVideoOutput{}, fmt.Errorf("get_live_chat_id: %w", apiErr)
	}
	if !gcsRestored {
		return SwitchVideoOutput{}, fmt.Errorf("get_live_chat_id: %w", apiErr)
	}

	// GCS 復元後の state から MarkEnded を導出
	basisState, _ := uc.State.Get(ctx)
	finalState := basisState.MarkEnded(videoID, basisState.LiveChatID, now)
	if err := uc.State.Set(ctx, finalState); err != nil {
		return SwitchVideoOutput{}, fmt.Errorf("state_set: %w", err)
	}
	logging.Log(ctx, "info", "SNAPSHOT", "switch_video: restored from GCS (videoId=%s, users=%d)", videoID, uc.Users.Count())
	return SwitchVideoOutput{State: finalState}, nil
}
