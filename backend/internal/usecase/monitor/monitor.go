// Package monitor は配信開始を待ち受ける background goroutine を提供する。
// RESERVED → SwitchVideo 呼出、ACTIVE+AutonomousMonitoring=true → Pull 呼出 を 60s tick で実行する。
package monitor

import (
	"context"
	"errors"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/logging"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase"
)

const (
	DefaultInterval = 60 * time.Second
	DefaultBuffer   = 5 * time.Minute
)

// switcher は SwitchVideo usecase の依存を抽象化する (test fake 注入用)。
type switcher interface {
	Execute(ctx context.Context, in usecase.SwitchVideoInput) (usecase.SwitchVideoOutput, error)
}

// puller は Pull usecase の依存を抽象化する (test fake 注入用)。
type puller interface {
	Execute(ctx context.Context) (usecase.PullOutput, error)
}

// detailsFetcher は YouTubePort.GetVideoLiveDetails のサブセット (test fake 注入用)。
type detailsFetcher interface {
	GetVideoLiveDetails(ctx context.Context, videoID string) (port.VideoLiveDetails, error)
}

// Monitor は配信開始を待ち受けて SwitchVideo / Pull を自動呼出しする。
type Monitor struct {
	SwitchVideo switcher
	Pull        puller
	YT          detailsFetcher
	State       port.StateRepo
	Clock       port.Clock
	Interval    time.Duration // 0 なら DefaultInterval
	Buffer      time.Duration // 0 なら DefaultBuffer

	// TickC を inject すると Interval 無視で外部 channel を使う (test 用)。
	TickC <-chan time.Time
}

// New は *usecase.SwitchVideo / *usecase.Pull を受け取って Monitor を返す。
// production での convenience constructor。
func New(
	sv *usecase.SwitchVideo,
	pl *usecase.Pull,
	yt port.YouTubePort,
	state port.StateRepo,
	clock port.Clock,
) *Monitor {
	return &Monitor{
		SwitchVideo: sv,
		Pull:        pl,
		YT:          yt,
		State:       state,
		Clock:       clock,
	}
}

// Run は ctx.Done まで tick loop を回す。
// RESERVED + ScheduledStartTime - Buffer が未来なら最初に sleep してから tick 開始する。
func (m *Monitor) Run(ctx context.Context) {
	interval := m.Interval
	if interval == 0 {
		interval = DefaultInterval
	}
	buffer := m.Buffer
	if buffer == 0 {
		buffer = DefaultBuffer
	}

	// 初期 sleep: RESERVED + scheduledStartTime - buffer が未来なら待つ (API quota 節約)
	m.waitUntilScheduled(ctx, buffer)
	if ctx.Err() != nil {
		return
	}

	tickC := m.TickC
	var ticker *time.Ticker
	if tickC == nil {
		ticker = time.NewTicker(interval)
		defer ticker.Stop()
		tickC = ticker.C
	}

	// sleep 完了直後に即実行 (体感応答を良くする)
	m.process(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-tickC:
			m.process(ctx)
		}
	}
}

// waitUntilScheduled は RESERVED かつ scheduledStartTime - buffer が未来なら sleep する。
func (m *Monitor) waitUntilScheduled(ctx context.Context, buffer time.Duration) {
	st, err := m.State.Get(ctx)
	if err != nil {
		logging.Log(ctx, "warn", "MONITOR", "state_get on init failed: %v", err)
		return
	}
	if st.Status != domain.StatusReserved || st.ScheduledStartTime.IsZero() {
		return
	}
	wakeAt := st.ScheduledStartTime.Add(-buffer)
	wait := wakeAt.Sub(m.Clock.Now())
	if wait <= 0 {
		return
	}
	logging.Log(ctx, "info", "MONITOR", "sleeping until %s (buffer=%s) before tick start", wakeAt.Format(time.RFC3339), buffer)
	select {
	case <-ctx.Done():
	case <-time.After(wait):
	}
}

// process は State を再評価して RESERVED / ACTIVE+AM の場合に適切な usecase を呼ぶ。
// tick ごとに毎回 State.Get することで Reserve / Cancel / Reset 等の動的変化に追従する。
func (m *Monitor) process(ctx context.Context) {
	st, err := m.State.Get(ctx)
	if err != nil {
		logging.Log(ctx, "warn", "MONITOR", "state_get failed: %v", err)
		return
	}

	switch {
	case st.Status == domain.StatusReserved:
		// 配信開始済みかを判定: ActualStartTime が立っていて、かつ ScheduledStartTime が過去 (or zero) であること
		// premiere / test broadcast 中は ActualStartTime が先行して立つことがあるため、ScheduledStartTime も併用する
		if m.YT != nil {
			details, derr := m.YT.GetVideoLiveDetails(ctx, st.VideoID)
			if derr != nil {
				logging.Log(ctx, "warn", "MONITOR", "get_video_live_details failed: %v", derr)
				return
			}
			now := m.Clock.Now()
			scheduledFuture := !details.ScheduledStartTime.IsZero() && details.ScheduledStartTime.After(now)
			if details.ActualStartTime.IsZero() || scheduledFuture {
				logging.Log(ctx, "info", "MONITOR", "tick: RESERVED, waiting (videoId=%s actualStart=%s scheduled=%s)",
					st.VideoID, details.ActualStartTime.Format(time.RFC3339), details.ScheduledStartTime.Format(time.RFC3339))
				return
			}
		}
		logging.Log(ctx, "info", "MONITOR", "tick: RESERVED → SwitchVideo (videoId=%s)", st.VideoID)
		_, err := m.SwitchVideo.Execute(ctx, usecase.SwitchVideoInput{VideoID: st.VideoID})
		if err != nil {
			// 配信未開放は期待される失敗 (info)、それ以外は warn
			var apiErr *domain.APIError
			if errors.As(err, &apiErr) && apiErr.Code == domain.ErrCodeLiveChatEnded {
				logging.Log(ctx, "info", "MONITOR", "switch_video: chat not yet open: %v", err)
			} else {
				logging.Log(ctx, "warn", "MONITOR", "switch_video failed: %v", err)
			}
		}
	case st.Status == domain.StatusActive && st.AutonomousMonitoring:
		if _, err := m.Pull.Execute(ctx); err != nil {
			logging.Log(ctx, "warn", "MONITOR", "pull failed: %v", err)
		}
	default:
		// WAITING / ACTIVE+AM=false → no-op
	}
}
