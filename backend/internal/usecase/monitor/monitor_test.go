package monitor_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/memory"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase/monitor"
)

// fakeYT は monitor.detailsFetcher を満たす最小 fake。
type fakeYT struct {
	details port.VideoLiveDetails
	err     error
	calls   atomic.Int32
}

func (f *fakeYT) GetVideoLiveDetails(_ context.Context, _ string) (port.VideoLiveDetails, error) {
	f.calls.Add(1)
	return f.details, f.err
}

// --- fake implementations ---

type fakeSwitcher struct {
	calls atomic.Int32
	err   error
}

func (f *fakeSwitcher) Execute(_ context.Context, _ usecase.SwitchVideoInput) (usecase.SwitchVideoOutput, error) {
	f.calls.Add(1)
	return usecase.SwitchVideoOutput{}, f.err
}

type fakePuller struct {
	calls atomic.Int32
	err   error
}

func (f *fakePuller) Execute(_ context.Context) (usecase.PullOutput, error) {
	f.calls.Add(1)
	return usecase.PullOutput{}, f.err
}

// fixedClock は Clock interface の固定時刻実装。
type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

// buildMonitor は tick channel inject 済みの Monitor を生成するヘルパー。
// yt が nil の場合は actualStartTime 判定をスキップする (= 古い挙動)。
func buildMonitor(sw *fakeSwitcher, pl *fakePuller, yt *fakeYT, state *memory.StateRepo, tickC <-chan time.Time) *monitor.Monitor {
	m := &monitor.Monitor{
		SwitchVideo: sw,
		Pull:        pl,
		State:       state,
		Clock:       fixedClock{t: time.Now()},
		TickC:       tickC,
	}
	if yt != nil {
		m.YT = yt
	}
	return m
}

// --- tests ---

// TestMonitor_Reserved_CallsSwitchVideo: RESERVED 状態で tick 1 回 → SwitchVideo が 1 回呼ばれる。
func TestMonitor_Reserved_CallsSwitchVideo(t *testing.T) {
	sw := &fakeSwitcher{}
	pl := &fakePuller{}
	yt := &fakeYT{details: port.VideoLiveDetails{ActualStartTime: time.Now().Add(-1 * time.Minute)}}
	state := memory.NewStateRepo()
	_ = state.Set(context.Background(), domain.LiveState{
		Status:  domain.StatusReserved,
		VideoID: "vid001",
	})

	tickC := make(chan time.Time, 1)
	m := buildMonitor(sw, pl, yt, state, tickC)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		m.Run(ctx)
	}()

	// Run は起動直後に process を 1 回実行する (initial call)
	// tickC からの tick を送って 2 回目を駆動してからキャンセル
	tickC <- time.Now()
	// キャンセル前に goroutine が process を処理できるよう少し待つ
	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after ctx cancel")
	}

	// initial process (1) + tick driven process (1) = 2 回
	if got := sw.calls.Load(); got < 1 {
		t.Errorf("SwitchVideo.Execute call count = %d, want >= 1", got)
	}
	if pl.calls.Load() != 0 {
		t.Errorf("Pull.Execute should not be called, got %d", pl.calls.Load())
	}
}

// TestMonitor_ActiveAM_CallsPull: ACTIVE+AutonomousMonitoring=true で tick → Pull が呼ばれる。
func TestMonitor_ActiveAM_CallsPull(t *testing.T) {
	sw := &fakeSwitcher{}
	pl := &fakePuller{}
	state := memory.NewStateRepo()
	_ = state.Set(context.Background(), domain.LiveState{
		Status:               domain.StatusActive,
		VideoID:              "vid002",
		AutonomousMonitoring: true,
	})

	tickC := make(chan time.Time, 1)
	m := buildMonitor(sw, pl, nil, state, tickC)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		m.Run(ctx)
	}()

	tickC <- time.Now()
	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after ctx cancel")
	}

	if got := pl.calls.Load(); got < 1 {
		t.Errorf("Pull.Execute call count = %d, want >= 1", got)
	}
	if sw.calls.Load() != 0 {
		t.Errorf("SwitchVideo.Execute should not be called, got %d", sw.calls.Load())
	}
}

// TestMonitor_Waiting_NoOp: WAITING 状態では SwitchVideo も Pull も呼ばれない。
func TestMonitor_Waiting_NoOp(t *testing.T) {
	sw := &fakeSwitcher{}
	pl := &fakePuller{}
	state := memory.NewStateRepo()
	_ = state.Set(context.Background(), domain.LiveState{
		Status: domain.StatusWaiting,
	})

	tickC := make(chan time.Time, 1)
	m := buildMonitor(sw, pl, nil, state, tickC)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		m.Run(ctx)
	}()

	tickC <- time.Now()
	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after ctx cancel")
	}

	if sw.calls.Load() != 0 {
		t.Errorf("SwitchVideo.Execute should not be called, got %d", sw.calls.Load())
	}
	if pl.calls.Load() != 0 {
		t.Errorf("Pull.Execute should not be called, got %d", pl.calls.Load())
	}
}

// TestMonitor_CtxCancel_Exits: ctx cancel で Run が return する (goroutine leak なし)。
func TestMonitor_CtxCancel_Exits(t *testing.T) {
	sw := &fakeSwitcher{}
	pl := &fakePuller{}
	state := memory.NewStateRepo()
	_ = state.Set(context.Background(), domain.LiveState{Status: domain.StatusWaiting})

	// tickC に何も送らない状態でキャンセルしても終わる
	tickC := make(chan time.Time)
	m := buildMonitor(sw, pl, nil, state, tickC)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		m.Run(ctx)
	}()

	// 少し待ってから cancel
	time.Sleep(10 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// OK
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after ctx cancel — goroutine leak suspected")
	}
}

// TestMonitor_ScheduledStartTime_SleepSkippedIfPast: scheduledStartTime が過去なら sleep せず即 process する。
func TestMonitor_ScheduledStartTime_SleepSkippedIfPast(t *testing.T) {
	sw := &fakeSwitcher{}
	pl := &fakePuller{}
	state := memory.NewStateRepo()
	pastScheduled := time.Now().Add(-10 * time.Minute) // 過去 → buffer 後も過去
	_ = state.Set(context.Background(), domain.LiveState{
		Status:             domain.StatusReserved,
		VideoID:            "vid_past",
		ScheduledStartTime: pastScheduled,
	})

	yt := &fakeYT{details: port.VideoLiveDetails{ActualStartTime: time.Now().Add(-1 * time.Minute)}}
	tickC := make(chan time.Time, 1)
	m := &monitor.Monitor{
		SwitchVideo: sw,
		Pull:        pl,
		YT:          yt,
		State:       state,
		Clock:       fixedClock{t: time.Now()},
		Buffer:      5 * time.Minute,
		TickC:       tickC,
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	start := time.Now()
	go func() {
		defer close(done)
		m.Run(ctx)
	}()

	// sleep なしなら即 process (initial call) が走るはず
	time.Sleep(30 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after ctx cancel")
	}

	elapsed := time.Since(start)
	// sleep せず即実行 → 100ms 以内に process が呼ばれているはず
	if elapsed > 200*time.Millisecond {
		t.Errorf("expected no sleep (past scheduledStartTime), but elapsed=%s", elapsed)
	}
	if sw.calls.Load() == 0 {
		t.Error("SwitchVideo.Execute should have been called at least once")
	}
}

// TestMonitor_ScheduledStartTime_SleepsUntilBuffer: scheduledStartTime が未来なら buffer 後まで sleep する。
func TestMonitor_ScheduledStartTime_SleepsUntilBuffer(t *testing.T) {
	sw := &fakeSwitcher{}
	pl := &fakePuller{}
	state := memory.NewStateRepo()

	now := time.Now()
	// scheduledStartTime = now + 150ms, buffer = 100ms → wakeAt = now + 50ms
	scheduled := now.Add(150 * time.Millisecond)
	buffer := 100 * time.Millisecond
	_ = state.Set(context.Background(), domain.LiveState{
		Status:             domain.StatusReserved,
		VideoID:            "vid_future",
		ScheduledStartTime: scheduled,
	})

	tickC := make(chan time.Time) // tick を送らない (sleep 中に cancel する)
	m := &monitor.Monitor{
		SwitchVideo: sw,
		Pull:        pl,
		State:       state,
		Clock:       fixedClock{t: now},
		Buffer:      buffer,
		TickC:       tickC,
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	start := time.Now()
	go func() {
		defer close(done)
		m.Run(ctx)
	}()

	// wakeAt (50ms) 前にキャンセル → process が呼ばれないことを確認
	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after ctx cancel")
	}

	elapsed := time.Since(start)
	_ = elapsed

	// 20ms でキャンセルしたので process は呼ばれていないはず
	if sw.calls.Load() != 0 {
		t.Errorf("SwitchVideo.Execute should not be called during sleep, got %d calls", sw.calls.Load())
	}
}

// TestMonitor_Reserved_ActualStartTimeZero_NoSwitch: actualStartTime が zero (未開始) なら SwitchVideo を呼ばない。
// 修正前は ActiveLiveChatId が立っただけで SwitchVideo が成功し ACTIVE に遷移していた (premature activation)。
func TestMonitor_Reserved_ActualStartTimeZero_NoSwitch(t *testing.T) {
	sw := &fakeSwitcher{}
	pl := &fakePuller{}
	yt := &fakeYT{details: port.VideoLiveDetails{}} // ActualStartTime zero
	state := memory.NewStateRepo()
	_ = state.Set(context.Background(), domain.LiveState{
		Status:  domain.StatusReserved,
		VideoID: "vid_not_started",
	})

	tickC := make(chan time.Time, 1)
	m := buildMonitor(sw, pl, yt, state, tickC)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		m.Run(ctx)
	}()

	tickC <- time.Now()
	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after ctx cancel")
	}

	if sw.calls.Load() != 0 {
		t.Errorf("SwitchVideo.Execute should NOT be called when actualStartTime is zero, got %d", sw.calls.Load())
	}
	if yt.calls.Load() < 1 {
		t.Errorf("GetVideoLiveDetails should be called at least once, got %d", yt.calls.Load())
	}
}

// TestMonitor_Reserved_ActualStartTimeSet_Switches: actualStartTime が立っていれば SwitchVideo を呼ぶ。
func TestMonitor_Reserved_ActualStartTimeSet_Switches(t *testing.T) {
	sw := &fakeSwitcher{}
	pl := &fakePuller{}
	yt := &fakeYT{details: port.VideoLiveDetails{ActualStartTime: time.Now().Add(-30 * time.Second)}}
	state := memory.NewStateRepo()
	_ = state.Set(context.Background(), domain.LiveState{
		Status:  domain.StatusReserved,
		VideoID: "vid_started",
	})

	tickC := make(chan time.Time, 1)
	m := buildMonitor(sw, pl, yt, state, tickC)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		m.Run(ctx)
	}()

	tickC <- time.Now()
	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after ctx cancel")
	}

	if sw.calls.Load() < 1 {
		t.Errorf("SwitchVideo.Execute should be called when actualStartTime is set, got %d", sw.calls.Load())
	}
}
