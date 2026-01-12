package usecase

import (
	"context"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
)

type PullOutput struct {
	AddedCount            int
	AutoReset             bool
	PollingIntervalMillis int64
}

type Pull struct {
	YT       port.YouTubePort
	Users    port.UserRepo
	Comments port.CommentRepo
	State    port.StateRepo
	Clock    port.Clock
}

// Execute: コメント取得・ユーザー追加、終了検知→WAITING へ（autoReset）。
func (uc *Pull) Execute(ctx context.Context) (PullOutput, error) {
	// 現在の状態を取得
	state, err := uc.State.Get(ctx)
	if err != nil {
		return PullOutput{}, err
	}

	// WAITING状態の場合は何もしない
	if state.Status != domain.StatusActive {
		return PullOutput{AddedCount: 0, AutoReset: false}, nil
	}

	// YouTube APIからメッセージを取得（ページトークン対応）
	items, nextToken, pollMs, isEnded, err := uc.YT.ListLiveChatMessages(ctx, state.LiveChatID, state.NextPageToken)
	if err != nil {
		return PullOutput{}, err
	}

	// 配信終了検知
	if isEnded {
		// ユーザークリア
		uc.Users.Clear()

		// コメントクリア
		uc.Comments.Clear()

		// WAITINGに戻す（現在時刻を終了時刻として設定）
		state.Status = domain.StatusWaiting
		state.EndedAt = uc.Clock.Now()
		state.NextPageToken = ""
		if err := uc.State.Set(ctx, state); err != nil {
			return PullOutput{}, err
		}

		return PullOutput{AddedCount: 0, AutoReset: true, PollingIntervalMillis: 0}, nil
	}

	// ユーザー追加 - メッセージIDによる重複チェックを使用
	addedCount := 0
	now := uc.Clock.Now()
	for _, msg := range items {
		// UpsertWithMessageUpdatedを使用してメッセージIDによる重複チェックを実行し、実際に更新された場合のみカウント
		updated, err := uc.Users.UpsertWithMessageUpdated(msg.ChannelID, msg.DisplayName, msg.PublishedAt, msg.ID)
		if err != nil {
			return PullOutput{}, err
		}
		if updated {
			addedCount++
		}

		// コメント保存
		uc.Comments.Add(domain.Comment{
			ID:          msg.ID,
			ChannelID:   msg.ChannelID,
			DisplayName: msg.DisplayName,
			Message:     msg.Message,
			PublishedAt: msg.PublishedAt,
		})
	}

	// 最終取得日時と次ページトークンを更新
	state.LastPulledAt = now
	state.NextPageToken = nextToken
	if err := uc.State.Set(ctx, state); err != nil {
		return PullOutput{}, err
	}

	// 最小ポーリング間隔を15秒に設定（無料枠での運用を考慮）
	const minPollingIntervalMillis = 15000
	if pollMs < minPollingIntervalMillis {
		pollMs = minPollingIntervalMillis
	}

	return PullOutput{AddedCount: addedCount, AutoReset: false, PollingIntervalMillis: pollMs}, nil
}
