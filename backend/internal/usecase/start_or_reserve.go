package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
)

// StartOrReserveInput is the input for the StartOrReserve usecase.
type StartOrReserveInput struct {
	VideoID string
}

// StartOrReserveOutput is the output for the StartOrReserve usecase.
type StartOrReserveOutput struct {
	State domain.LiveState
	// Dispatched indicates which sub-usecase was executed: "switch" or "reserve".
	Dispatched string
}

// StartOrReserve dispatches a videoId to SwitchVideo (live started) or Reserve (not yet started).
// Centralises the "live started yet?" predicate so monitor.go and the HTTP handler stay in sync.
type StartOrReserve struct {
	YT          port.YouTubePort
	Clock       port.Clock
	SwitchVideo *SwitchVideo
	Reserve     *Reserve
}

// Execute fetches live details and routes to SwitchVideo or Reserve accordingly.
func (uc *StartOrReserve) Execute(ctx context.Context, in StartOrReserveInput) (StartOrReserveOutput, error) {
	if in.VideoID == "" {
		return StartOrReserveOutput{}, &domain.APIError{Code: domain.ErrCodeInvalidArgument, Message: "videoId is required"}
	}

	details, err := uc.YT.GetVideoLiveDetails(ctx, in.VideoID)
	if err != nil {
		return StartOrReserveOutput{}, fmt.Errorf("get_video_live_details: %w", err)
	}

	if details.IsLiveContent && IsLiveNotStarted(details, uc.Clock.Now()) {
		// pre-fetch 済み details を渡し Reserve 側の GetVideoLiveDetails 再呼び出しを回避する。
		out, rerr := uc.Reserve.Execute(ctx, ReserveInput{VideoID: in.VideoID, Details: &details})
		if rerr != nil {
			return StartOrReserveOutput{}, rerr
		}
		return StartOrReserveOutput{State: out.State, Dispatched: "reserve"}, nil
	}

	out, serr := uc.SwitchVideo.Execute(ctx, SwitchVideoInput{VideoID: in.VideoID})
	if serr != nil {
		return StartOrReserveOutput{}, serr
	}
	return StartOrReserveOutput{State: out.State, Dispatched: "switch"}, nil
}

// IsLiveNotStarted returns true when the live is reserved/scheduled but not yet broadcasting.
// Predicate: ActualStartTime is zero OR ScheduledStartTime is in the future (premiere / test broadcast guard).
// Used by both StartOrReserve (handler-side dispatch) and monitor (tick-side guard) to share the rule.
func IsLiveNotStarted(details port.VideoLiveDetails, now time.Time) bool {
	scheduledFuture := !details.ScheduledStartTime.IsZero() && details.ScheduledStartTime.After(now)
	return details.ActualStartTime.IsZero() || scheduledFuture
}
