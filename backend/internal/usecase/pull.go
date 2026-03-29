package usecase

import (
	"context"
	"strings"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
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
	items, nextToken, pollMs, skippedCount, isEnded, err := uc.YT.ListLiveChatMessages(ctx, state.LiveChatID, state.NextPageToken)
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

	// @ハンドルのチャンネルIDを収集してChannels APIで名前解決
	var handleChannelIDs []string
	for _, msg := range items {
		if strings.HasPrefix(msg.DisplayName, "@") {
			handleChannelIDs = append(handleChannelIDs, msg.ChannelID)
		}
	}
	var channelNames map[string]string
	if len(handleChannelIDs) > 0 {
		channelNames, _ = uc.YT.GetChannelDisplayNames(ctx, handleChannelIDs)
	}

	// 解決した名前で置換、失敗時は@除去でフォールバック
	for i, msg := range items {
		if strings.HasPrefix(msg.DisplayName, "@") {
			if name, ok := channelNames[msg.ChannelID]; ok {
				items[i].DisplayName = name
			} else {
				items[i].DisplayName = strings.TrimPrefix(msg.DisplayName, "@")
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
			return PullOutput{}, err
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
			return PullOutput{}, err
		}
	}

	// 最終取得日時と次ページトークンを更新
	state.LastPulledAt = now
	state.NextPageToken = nextToken
	if err := uc.State.Set(ctx, state); err != nil {
		return PullOutput{}, err
	}

	// minPollingIntervalMillis はYouTube Data API v3の無料枠制限（10,000 units/day）を考慮した最小ポーリング間隔
	// 1回のliveChatMessages.list呼び出しは5 unitsを消費するため、15秒間隔で運用することで1日あたり約5,760回（28,800 units）の呼び出しを制限内に抑える
	const minPollingIntervalMillis = 15000
	if pollMs < minPollingIntervalMillis {
		pollMs = minPollingIntervalMillis
	}

	return PullOutput{AddedCount: addedCount, SkippedCount: skippedCount, AutoReset: false, PollingIntervalMillis: pollMs}, nil
}
