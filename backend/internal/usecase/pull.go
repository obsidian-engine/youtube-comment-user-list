package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/logging"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase/snapshot"
)

type PullOutput struct {
	AddedCount            int
	SkippedCount          int
	AutoReset             bool
	PollingIntervalMillis int64
}

type Pull struct {
	YT       port.YouTubePort
	Users    port.UserRepo
	Comments port.CommentRepo
	State    port.StateRepo
	Clock    port.Clock
	Snap     snapshot.Coordinator // 必須 (GCS 不要な場合は NopCoordinator を渡す)
}

// Execute: コメント取得・ユーザー追加、終了検知→WAITING へ（autoReset）。
func (uc *Pull) Execute(ctx context.Context) (PullOutput, error) {
	// 現在の状態を取得
	state, err := uc.State.Get(ctx)
	if err != nil {
		return PullOutput{}, fmt.Errorf("state_get: %w", err)
	}

	// WAITING状態の場合は何もしない
	if state.Status != domain.StatusActive {
		return PullOutput{AddedCount: 0, AutoReset: false}, nil
	}

	// YouTube APIからメッセージを取得（ページトークン対応）
	items, nextToken, pollMs, skippedCount, isEnded, err := uc.YT.ListLiveChatMessages(ctx, state.LiveChatID, state.NextPageToken)
	if err != nil {
		return PullOutput{}, fmt.Errorf("list_messages: %w", err)
	}

	// 配信終了検知
	if isEnded {
		// 配信中の users/comments はメモリに保持したまま snapshot へ永続化する
		// （同じ videoId で再度「切替」を押したら復元できるようにするため）
		uc.Snap.MarkDirty()
		if err := uc.Snap.Flush(ctx); err != nil {
			logging.Log(ctx, "warn", "SNAPSHOT", "pull: snapshot flush on stream end failed: %v", err)
			// Flush 失敗は警告のみ、終了処理は継続
		}

		// WAITINGに戻す（現在時刻を終了時刻として設定）
		// Users / Comments は意図的にクリアしない
		state.Status = domain.StatusWaiting
		state.EndedAt = uc.Clock.Now()
		state.NextPageToken = ""
		if err := uc.State.Set(ctx, state); err != nil {
			return PullOutput{}, fmt.Errorf("state_set: %w", err)
		}

		return PullOutput{AddedCount: 0, AutoReset: true, PollingIntervalMillis: 0}, nil
	}

	// @ハンドルのチャンネルIDを収集してChannels APIで名前解決（重複排除）
	seen := make(map[string]bool)
	var handleChannelIDs []string
	for _, msg := range items {
		if strings.HasPrefix(msg.DisplayName, "@") && !seen[msg.ChannelID] {
			seen[msg.ChannelID] = true
			handleChannelIDs = append(handleChannelIDs, msg.ChannelID)
		}
	}
	var channelNames map[string]string
	if len(handleChannelIDs) > 0 {
		var err error
		channelNames, err = uc.YT.GetChannelDisplayNames(ctx, handleChannelIDs)
		if err != nil {
			logging.Log(ctx, "warn", "PULL", "Failed to resolve channel display names: %v", err)
		}
	}

	// 解決した名前で置換、失敗時は@除去でフォールバック
	for i, msg := range items {
		if after, found := strings.CutPrefix(msg.DisplayName, "@"); found {
			if name, ok := channelNames[msg.ChannelID]; ok {
				items[i].DisplayName = name
			} else {
				items[i].DisplayName = after
			}
		}
	}

	// ユーザー追加 - メッセージIDによる重複チェックを使用
	addedCount := 0
	now := uc.Clock.Now()
	for _, msg := range items {
		// UpsertWithMessageUpdatedを使用してメッセージIDによる重複チェックを実行し、実際に更新された場合のみカウント
		updated, err := uc.Users.UpsertWithMessageUpdated(msg.ChannelID, msg.DisplayName, msg.PublishedAt, msg.ID)
		if err != nil {
			return PullOutput{}, fmt.Errorf("user_upsert: %w", err)
		}
		if updated {
			addedCount++
		}

		// コメント保存
		if err := uc.Comments.Add(domain.Comment{
			ID:          msg.ID,
			ChannelID:   msg.ChannelID,
			DisplayName: msg.DisplayName,
			Message:     msg.Message,
			PublishedAt: msg.PublishedAt,
		}); err != nil {
			return PullOutput{}, fmt.Errorf("comment_add: %w", err)
		}
	}

	// 最終取得日時と次ページトークンを更新
	state.LastPulledAt = now
	state.NextPageToken = nextToken
	if err := uc.State.Set(ctx, state); err != nil {
		return PullOutput{}, fmt.Errorf("state_set: %w", err)
	}

	// minPollingIntervalMillis はYouTube Data API v3の無料枠制限（10,000 units/day）を考慮した最小ポーリング間隔
	// 1回のliveChatMessages.list呼び出しは5 unitsを消費するため、60秒間隔で運用することで24時間連続稼働でも約1,440回（7,200 units）に収まり、他 API 呼び出し分の余裕も確保する
	const minPollingIntervalMillis = 60000
	if pollMs < minPollingIntervalMillis {
		pollMs = minPollingIntervalMillis
	}

	// 差分あり（新規ユーザー追加 or コメント追加）の場合にスナップショット dirty フラグを立てる
	if addedCount > 0 || len(items) > 0 {
		uc.Snap.MarkDirty()
	}

	return PullOutput{AddedCount: addedCount, SkippedCount: skippedCount, AutoReset: false, PollingIntervalMillis: pollMs}, nil
}
