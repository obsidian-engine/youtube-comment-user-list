package usecase

import (
	"context"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/logging"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase/snapshot"
)

// clearForNewVideo は videoId 遷移で旧配信の in-memory data を破棄する。
// prevVideoID != newVideoID かつ prevVideoID が非空のときのみクリアする (同 videoId 再切替は維持)。
// users は必須、comments は nil 許容 (test 用途)。
//
// 呼び出し規約: state.Set 成功後に呼ぶこと。state.Set 失敗時に in-memory だけ消える不整合を防ぐ。
func clearForNewVideo(users port.UserRepo, comments port.CommentRepo, prevVideoID, newVideoID string) {
	if prevVideoID == "" || prevVideoID == newVideoID {
		return
	}
	if users != nil {
		users.Clear()
	}
	if comments != nil {
		comments.Clear()
	}
}

// syncSnapshotVideo は Coordinator の video pointer を新 videoId に切替え、即時 Flush する。
// SetVideo を呼ばずに Flush すると coord.videoID が旧値のままで save が旧 key に書く事故を防ぐ。
// Flush 失敗は warn のみで遷移は継続する (in-memory は既に更新済み、次 tick で retry される)。
func syncSnapshotVideo(
	ctx context.Context,
	snap snapshot.Coordinator,
	videoID, liveChatID, videoTitle, channelTitle, logTag string,
) {
	snap.SetVideo(videoID, liveChatID, videoTitle, channelTitle)
	snap.MarkDirty()
	if err := snap.Flush(ctx); err != nil {
		logging.Log(ctx, "warn", "SNAPSHOT", "%s: snapshot flush failed: %v", logTag, err)
	}
}
