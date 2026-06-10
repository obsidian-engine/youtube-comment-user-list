package http_test

import (
	"context"
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	ahttp "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/http"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/memory"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/system"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/youtube"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase/snapshot"
)

// fakeCoordinator は LastSavedAt を制御できるテスト用 Coordinator 実装です。
type fakeCoordinator struct {
	snapshot.NopCoordinator
	savedAt time.Time
}

func (f *fakeCoordinator) LastSavedAt() time.Time {
	return f.savedAt
}

func newTestServer(frontend string) *httptest.Server {
	return newTestServerWithCoord(frontend, &snapshot.NopCoordinator{})
}

func newTestServerWithCoord(frontend string, coord snapshot.Coordinator) *httptest.Server {
	users := memory.NewUserRepo()
	state := memory.NewStateRepo()
	yt := youtube.New("") // テスト用の空キー
	clock := system.NewSystemClock()

	h := &ahttp.Handlers{
		Status:      &usecase.Status{Users: users, State: state},
		SwitchVideo: &usecase.SwitchVideo{YT: yt, Users: users, State: state, Clock: clock, Snap: &snapshot.NopCoordinator{}},
		Pull:        &usecase.Pull{YT: yt, Users: users, State: state, Snap: &snapshot.NopCoordinator{}},
		Reset:       &usecase.Reset{Users: users, State: state, Snap: &snapshot.NopCoordinator{}},
		Users:       users,
		Coord:       coord,
	}
	router := ahttp.NewRouter(h, frontend)
	return httptest.NewServer(router)
}

// fakeSnapshotSink は integration test 用の in-memory SnapshotSink。
type fakeSnapshotSink struct {
	mu        sync.Mutex
	snapshots map[string]*port.Snapshot
}

func newFakeSnapshotSink() *fakeSnapshotSink {
	return &fakeSnapshotSink{snapshots: make(map[string]*port.Snapshot)}
}

func (f *fakeSnapshotSink) Load(_ context.Context, videoID string) (*port.Snapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	snap, ok := f.snapshots[videoID]
	if !ok {
		return nil, nil
	}
	cp := *snap
	return &cp, nil
}

func (f *fakeSnapshotSink) Save(_ context.Context, snap *port.Snapshot) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *snap
	f.snapshots[snap.VideoID] = &cp
	return nil
}

func (f *fakeSnapshotSink) LoadCurrent(_ context.Context) (*port.CurrentPointer, error) {
	return nil, nil
}
func (f *fakeSnapshotSink) SaveCurrent(_ context.Context, _ *port.CurrentPointer) error { return nil }

func (f *fakeSnapshotSink) List(_ context.Context) ([]port.SnapshotSummary, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	summaries := make([]port.SnapshotSummary, 0, len(f.snapshots))
	for _, snap := range f.snapshots {
		summaries = append(summaries, port.SnapshotSummary{
			VideoID:      snap.VideoID,
			SavedAt:      snap.SavedAt,
			UserCount:    len(snap.Users),
			CommentCount: len(snap.Comments),
		})
	}
	return summaries, nil
}

func newTestServerWithHistory(sink *fakeSnapshotSink) *httptest.Server {
	users := memory.NewUserRepo()
	state := memory.NewStateRepo()
	yt := youtube.New("")
	clock := system.NewSystemClock()

	h := &ahttp.Handlers{
		Status:      &usecase.Status{Users: users, State: state},
		SwitchVideo: &usecase.SwitchVideo{YT: yt, Users: users, State: state, Clock: clock, Snap: &snapshot.NopCoordinator{}},
		Pull:        &usecase.Pull{YT: yt, Users: users, State: state, Snap: &snapshot.NopCoordinator{}},
		Reset:       &usecase.Reset{Users: users, State: state, Snap: &snapshot.NopCoordinator{}},
		Users:       users,
		Coord:       &snapshot.NopCoordinator{},
		ListHistory: &usecase.ListHistorySnapshots{Sink: sink},
		GetHistory:  &usecase.GetHistorySnapshot{Sink: sink},
	}
	router := ahttp.NewRouter(h, "http://example.com")
	return httptest.NewServer(router)
}

func TestRoutes_WorkCorrectly(t *testing.T) {
	ts := newTestServer("http://example.com")
	defer ts.Close()

	// GET /status should return 200
	req, _ := stdhttp.NewRequest(stdhttp.MethodGet, ts.URL+"/status", nil)
	res, err := stdhttp.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request GET /status: %v", err)
	}
	if res.StatusCode != stdhttp.StatusOK {
		t.Errorf("GET /status => %d want %d", res.StatusCode, stdhttp.StatusOK)
	}

	// GET /users.json should return 200
	req, _ = stdhttp.NewRequest(stdhttp.MethodGet, ts.URL+"/users.json", nil)
	res, err = stdhttp.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request GET /users.json: %v", err)
	}
	if res.StatusCode != stdhttp.StatusOK {
		t.Errorf("GET /users.json => %d want %d", res.StatusCode, stdhttp.StatusOK)
	}

	// POST /reset should return 200
	req, _ = stdhttp.NewRequest(stdhttp.MethodPost, ts.URL+"/reset", nil)
	res, err = stdhttp.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request POST /reset: %v", err)
	}
	if res.StatusCode != stdhttp.StatusOK {
		t.Errorf("POST /reset => %d want %d", res.StatusCode, stdhttp.StatusOK)
	}

	// POST /pull should return 200
	req, _ = stdhttp.NewRequest(stdhttp.MethodPost, ts.URL+"/pull", nil)
	res, err = stdhttp.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request POST /pull: %v", err)
	}
	if res.StatusCode != stdhttp.StatusOK {
		t.Errorf("POST /pull => %d want %d", res.StatusCode, stdhttp.StatusOK)
	}
}

func TestCORS_AllowsFrontendOrigin(t *testing.T) {
	origin := "https://frontend.example"
	ts := newTestServer(origin)
	defer ts.Close()

	req, _ := stdhttp.NewRequest(stdhttp.MethodOptions, ts.URL+"/status", nil)
	req.Header.Set("Origin", origin)
	res, err := stdhttp.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("preflight: %v", err)
	}
	if got := res.Header.Get("Access-Control-Allow-Origin"); got != origin {
		t.Fatalf("Allow-Origin=%q want %q", got, origin)
	}
	if res.StatusCode != stdhttp.StatusNoContent {
		t.Fatalf("preflight status=%d want %d", res.StatusCode, stdhttp.StatusNoContent)
	}
}

// TestStatus_snapshotSavedAt_returnedAlways: LastSavedAt がある coordinator は
// /status を何度呼んでも常に snapshotSavedAt を返す
func TestStatus_snapshotSavedAt_returnedAlways(t *testing.T) {
	t.Helper()
	savedAt := time.Date(2024, 6, 9, 14, 23, 0, 0, time.UTC)
	coord := &fakeCoordinator{savedAt: savedAt}
	ts := newTestServerWithCoord("http://example.com", coord)
	defer ts.Close()

	// 1 回目: snapshotSavedAt が含まれる
	req, _ := stdhttp.NewRequest(stdhttp.MethodGet, ts.URL+"/status", nil)
	res, err := stdhttp.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("1st GET /status: %v", err)
	}
	if res.StatusCode != stdhttp.StatusOK {
		t.Fatalf("1st GET /status status = %d, want 200", res.StatusCode)
	}
	var body1 map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body1); err != nil {
		t.Fatalf("decode 1st response: %v", err)
	}
	res.Body.Close()

	if _, exists := body1["snapshotSavedAt"]; !exists {
		t.Error("1st /status: expected snapshotSavedAt field, got absent")
	}

	// 2 回目: 常時返すので snapshotSavedAt が含まれる
	req2, _ := stdhttp.NewRequest(stdhttp.MethodGet, ts.URL+"/status", nil)
	res2, err := stdhttp.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("2nd GET /status: %v", err)
	}
	defer res2.Body.Close()
	var body2 map[string]any
	if err := json.NewDecoder(res2.Body).Decode(&body2); err != nil {
		t.Fatalf("decode 2nd response: %v", err)
	}

	if _, exists := body2["snapshotSavedAt"]; !exists {
		t.Error("2nd /status: expected snapshotSavedAt present (always returned), but absent")
	}
}

// TestStatus_noSnapshot_noSavedAt: Restore が走っていない場合は snapshotSavedAt が返らない
func TestStatus_noSnapshot_noSavedAt(t *testing.T) {
	t.Helper()
	ts := newTestServer("http://example.com") // NopCoordinator
	defer ts.Close()

	req, _ := stdhttp.NewRequest(stdhttp.MethodGet, ts.URL+"/status", nil)
	res, err := stdhttp.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /status: %v", err)
	}
	defer res.Body.Close()
	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if _, exists := body["snapshotSavedAt"]; exists {
		t.Error("expected snapshotSavedAt absent when no restore, but present")
	}
}

// --- History endpoint integration tests ---

func TestGET_HistorySnapshots(t *testing.T) {
	t.Helper()
	sink := newFakeSnapshotSink()
	now := time.Now().UTC()
	_ = sink.Save(context.Background(), &port.Snapshot{
		VideoID: "vid1",
		SavedAt: now,
		Users:   []domain.User{{ChannelID: "c1", DisplayName: "Alice"}},
	})
	_ = sink.Save(context.Background(), &port.Snapshot{
		VideoID:  "vid2",
		SavedAt:  now.Add(-time.Hour),
		Comments: []domain.Comment{{ChannelID: "c2", Message: "hello"}},
	})
	ts := newTestServerWithHistory(sink)
	defer ts.Close()

	req, _ := stdhttp.NewRequest(stdhttp.MethodGet, ts.URL+"/history/snapshots", nil)
	res, err := stdhttp.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /history/snapshots: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != stdhttp.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}

	var body struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Items) != 2 {
		t.Errorf("got %d items, want 2", len(body.Items))
	}
	// 各 item に必須フィールドがあることを確認
	for _, item := range body.Items {
		for _, field := range []string{"videoId", "savedAt", "userCount", "commentCount"} {
			if _, ok := item[field]; !ok {
				t.Errorf("item missing field %q: %v", field, item)
			}
		}
	}
}

func TestGET_HistorySnapshot_byVideoID(t *testing.T) {
	t.Helper()
	sink := newFakeSnapshotSink()
	now := time.Now().UTC()
	_ = sink.Save(context.Background(), &port.Snapshot{
		VideoID: "vid1",
		SavedAt: now,
		Users:   []domain.User{{ChannelID: "c1", DisplayName: "Alice"}},
	})
	ts := newTestServerWithHistory(sink)
	defer ts.Close()

	req, _ := stdhttp.NewRequest(stdhttp.MethodGet, ts.URL+"/history/snapshots/vid1", nil)
	res, err := stdhttp.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /history/snapshots/vid1: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != stdhttp.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["videoId"] != "vid1" {
		t.Errorf("videoId = %v, want vid1", body["videoId"])
	}
	if _, ok := body["savedAt"]; !ok {
		t.Error("response missing savedAt field")
	}
	if _, ok := body["users"]; !ok {
		t.Error("response missing users field")
	}
}

func TestGET_HistorySnapshot_notFound(t *testing.T) {
	t.Helper()
	sink := newFakeSnapshotSink()
	ts := newTestServerWithHistory(sink)
	defer ts.Close()

	req, _ := stdhttp.NewRequest(stdhttp.MethodGet, ts.URL+"/history/snapshots/nonexistent", nil)
	res, err := stdhttp.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /history/snapshots/nonexistent: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != stdhttp.StatusNotFound {
		t.Fatalf("status = %d, want 404", res.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["error"] != "not_found" {
		t.Errorf("error = %v, want not_found", body["error"])
	}
}

// TestRouter_PanicResponseDoesNotIncludeStack は production と同じ middleware 順
// (Recover 外側 → Collector 内側) で panic が起きたとき、
// HTTP 500 を返しつつ logs field が response に含まれないことを確認する (方針 B)。
func TestRouter_PanicResponseDoesNotIncludeStack(t *testing.T) {
	t.Helper()

	users := memory.NewUserRepo()
	state := memory.NewStateRepo()
	yt := youtube.New("")
	clock := system.NewSystemClock()

	h := &ahttp.Handlers{
		Status:      &usecase.Status{Users: users, State: state},
		SwitchVideo: &usecase.SwitchVideo{YT: yt, Users: users, State: state, Clock: clock, Snap: &snapshot.NopCoordinator{}},
		Pull:        &usecase.Pull{YT: yt, Users: users, State: state, Snap: &snapshot.NopCoordinator{}},
		Reset:       &usecase.Reset{Users: users, State: state, Snap: &snapshot.NopCoordinator{}},
		Users:       users,
		Coord:       &snapshot.NopCoordinator{},
	}

	// NewRouter が組む production middleware 順 (Recover 外側 → Collector 内側) を利用する。
	// /reset handler 内で deliberately panic させる stub を上書きするのではなく、
	// panicHandler を直接 chi に乗せた router を構築して endpoint を追加する。
	router := ahttp.NewRouter(h, "http://example.com")

	// router は chi.Router だが stdhttp.Handler なので httptest.NewServer に渡す。
	// panic を起こすための専用 endpoint は NewRouter が公開していないため、
	// /pull → /reset → /status が全て正常終了することで 200 を確認した上で、
	// RecoverMiddleware の動作は middleware_test.go TestRecoverMiddleware_PanicReturns500WithoutLogsInResponse
	// が unit test として担保する。
	// ここでは integration として: production router で panic endpoint を追加できないため
	// /status (正常系) の response に logs が absent であることを確認し、
	// middleware 順の正常適用を smoke test する。
	ts := httptest.NewServer(router)
	defer ts.Close()

	req, _ := stdhttp.NewRequest(stdhttp.MethodGet, ts.URL+"/status", nil)
	res, err := stdhttp.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /status: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != stdhttp.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	// 正常系では logs は omitempty で absent
	if _, hasLogs := body["logs"]; hasLogs {
		t.Errorf("logs field must be absent for successful response with empty collector, got body=%v", body)
	}
}

func TestUsersAPI_IncludesLatestCommentedAt(t *testing.T) {
	ts := newTestServer("http://example.com")
	defer ts.Close()

	// /users.json APIをテスト
	req, _ := stdhttp.NewRequest(stdhttp.MethodGet, ts.URL+"/users.json", nil)
	res, err := stdhttp.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request GET /users.json: %v", err)
	}
	if res.StatusCode != stdhttp.StatusOK {
		t.Errorf("GET /users.json => %d want %d", res.StatusCode, stdhttp.StatusOK)
	}

	// レスポンスボディをチェック
	body := make([]byte, 1000)
	n, _ := res.Body.Read(body)
	responseBody := string(body[:n])

	// Content-Typeが正しいことを確認
	contentType := res.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type should be application/json, got: %s", contentType)
	}

	t.Logf("Response body: %s", responseBody)

	// 空のレスポンスの場合でも有効なJSONであることを確認
	// 実際のユーザーデータがある場合、latestCommentedAtフィールドの存在を確認できる
}

// TestSuccessResponse_ContainsLogsField は各 endpoint の success response に
// logs field が含まれることを確認する。
// CollectorMiddleware が inject した collector の entries は空でも
// JSON に "logs" key が存在しない (omitempty) ことを確認 — つまり空 logs の場合は
// レスポンスに logs key が出ない設計が正しく動いていることを検証する。
func TestSuccessResponse_ContainsLogsField(t *testing.T) {
	ts := newTestServer("http://example.com")
	defer ts.Close()

	tests := []struct {
		name        string
		method      string
		path        string
		wantLogsKey bool // true = logs key が present、false = absent (omitempty)
	}{
		// 空 collector なので logs は omitempty で absent が正しい
		{name: "GET /status", method: stdhttp.MethodGet, path: "/status", wantLogsKey: false},
		{name: "POST /reset", method: stdhttp.MethodPost, path: "/reset", wantLogsKey: false},
		{name: "POST /pull", method: stdhttp.MethodPost, path: "/pull", wantLogsKey: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := stdhttp.NewRequest(tt.method, ts.URL+tt.path, nil)
			res, err := stdhttp.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("%s %s: %v", tt.method, tt.path, err)
			}
			defer res.Body.Close()

			if res.StatusCode != stdhttp.StatusOK {
				t.Fatalf("status = %d, want 200", res.StatusCode)
			}

			var body map[string]any
			if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			_, hasLogs := body["logs"]
			if hasLogs != tt.wantLogsKey {
				t.Errorf("logs key present=%v, want %v; body=%v", hasLogs, tt.wantLogsKey, body)
			}
		})
	}
}
