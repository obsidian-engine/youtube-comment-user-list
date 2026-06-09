package gcs_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/gcs"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"google.golang.org/api/option"
)

const testBucket = "test-bucket"

// gcsObject は GCS JSON API のオブジェクトメタデータ表現です。
type gcsObject struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

// gcsObjectList は GCS JSON API のオブジェクト一覧レスポンスです。
type gcsObjectList struct {
	Kind  string      `json:"kind"`
	Items []gcsObject `json:"items"`
}

// fakeGCSServer は GCS JSON API の一部 (list / get) を in-memory で実装します。
type fakeGCSServer struct {
	objects map[string][]byte // name -> body
}

func newFakeGCSServer() *fakeGCSServer {
	return &fakeGCSServer{objects: make(map[string][]byte)}
}

func (f *fakeGCSServer) put(name string, body []byte) {
	f.objects[name] = body
}

// ServeHTTP は GCS SDK が使う 2 種類のエンドポイントをサポートします。
//
//	JSON API (list):    GET /b/{bucket}/o?prefix=...
//	XML API (download): GET /{bucket}/{object}
func (f *fakeGCSServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// JSON API: /b/{bucket}/o[/{objname}]
	if strings.HasPrefix(path, "/b/") {
		rest := strings.TrimPrefix(path, "/b/")
		slashIdx := strings.Index(rest, "/o")
		if slashIdx < 0 {
			http.NotFound(w, r)
			return
		}
		afterO := rest[slashIdx+2:] // "" for list, "/{objname}" for get

		if afterO == "" || afterO == "/" {
			// list
			prefix := r.URL.Query().Get("prefix")
			resp := gcsObjectList{Kind: "storage#objects"}
			for name := range f.objects {
				if strings.HasPrefix(name, prefix) {
					resp.Items = append(resp.Items, gcsObject{Kind: "storage#object", Name: name})
				}
			}
			sort.Slice(resp.Items, func(i, j int) bool {
				return resp.Items[i].Name < resp.Items[j].Name
			})
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		// metadata: /b/{bucket}/o/{objname}
		objName := strings.TrimPrefix(afterO, "/")
		_, ok := f.objects[objName]
		if !ok {
			http.Error(w, `{"error":{"code":404,"message":"not found"}}`, http.StatusNotFound)
			return
		}
		meta := gcsObject{Kind: "storage#object", Name: objName}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(meta)
		return
	}

	// XML API (download): /{bucket}/{object}
	// GCS SDK の NewReader は storage.googleapis.com/{bucket}/{object} を使うため
	// option.WithEndpoint 指定時は /{bucket}/{object} になる
	parts := strings.SplitN(strings.TrimPrefix(path, "/"), "/", 2)
	if len(parts) == 2 {
		objName := parts[1]
		body, ok := f.objects[objName]
		if !ok {
			http.Error(w, `<?xml version='1.0' encoding='UTF-8'?><Error><Code>NoSuchKey</Code></Error>`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
		return
	}

	http.NotFound(w, r)
}

// newTestStore は fake GCS server に向いた SnapshotStore を返します。
func newTestStore(t *testing.T, srv *httptest.Server) *gcs.SnapshotStore {
	t.Helper()
	client, err := storage.NewClient(
		context.Background(),
		option.WithEndpoint(srv.URL),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatalf("storage.NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return gcs.NewSnapshotStore(client, testBucket)
}

func marshalSnapshot(t *testing.T, snap port.Snapshot) []byte {
	t.Helper()
	b, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	return b
}

// TestList_returnsAllSnapshots は bucket に 3 件入れると 3 件返ることを確認します。
func TestList_returnsAllSnapshots(t *testing.T) {
	fake := newFakeGCSServer()
	srv := httptest.NewServer(fake)
	defer srv.Close()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	snaps := []port.Snapshot{
		{VideoID: "vid1", SavedAt: now.Add(0), Users: []domain.User{{ChannelID: "u1"}, {ChannelID: "u2"}}, Comments: []domain.Comment{{ID: "c1"}}},
		{VideoID: "vid2", SavedAt: now.Add(time.Hour), Users: []domain.User{{ChannelID: "u3"}}, Comments: []domain.Comment{{ID: "c2"}, {ID: "c3"}}},
		{VideoID: "vid3", SavedAt: now.Add(2 * time.Hour), Users: []domain.User{}, Comments: []domain.Comment{}},
	}
	for _, snap := range snaps {
		fake.put(fmt.Sprintf("snapshots/%s.json", snap.VideoID), marshalSnapshot(t, snap))
	}

	store := newTestStore(t, srv)
	summaries, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(summaries) != 3 {
		t.Fatalf("want 3 summaries, got %d", len(summaries))
	}

	// videoID をキーにしてアサート
	byID := make(map[string]port.SnapshotSummary)
	for _, s := range summaries {
		byID[s.VideoID] = s
	}

	cases := []struct {
		videoID  string
		users    int
		comments int
	}{
		{"vid1", 2, 1},
		{"vid2", 1, 2},
		{"vid3", 0, 0},
	}
	for _, tc := range cases {
		s, ok := byID[tc.videoID]
		if !ok {
			t.Errorf("summary for %s not found", tc.videoID)
			continue
		}
		if s.UserCount != tc.users {
			t.Errorf("%s: UserCount = %d, want %d", tc.videoID, s.UserCount, tc.users)
		}
		if s.CommentCount != tc.comments {
			t.Errorf("%s: CommentCount = %d, want %d", tc.videoID, s.CommentCount, tc.comments)
		}
		if s.SavedAt.IsZero() {
			t.Errorf("%s: SavedAt is zero", tc.videoID)
		}
	}
}

// TestList_empty は bucket が空の場合に空 slice + err nil を返すことを確認します。
func TestList_empty(t *testing.T) {
	fake := newFakeGCSServer()
	srv := httptest.NewServer(fake)
	defer srv.Close()

	store := newTestStore(t, srv)
	summaries, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if summaries == nil {
		t.Fatal("want empty slice, got nil")
	}
	if len(summaries) != 0 {
		t.Fatalf("want 0 summaries, got %d", len(summaries))
	}
}

// TestList_skipsCurrentJson は current.json が結果に含まれないことを確認します。
func TestList_skipsCurrentJson(t *testing.T) {
	fake := newFakeGCSServer()
	srv := httptest.NewServer(fake)
	defer srv.Close()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	snap := port.Snapshot{VideoID: "vid1", SavedAt: now, Users: []domain.User{}, Comments: []domain.Comment{}}
	fake.put("snapshots/vid1.json", marshalSnapshot(t, snap))

	// current.json も入れる
	currentBody, _ := json.Marshal(port.CurrentPointer{VideoID: "vid1", SavedAt: now})
	fake.put("snapshots/current.json", currentBody)

	store := newTestStore(t, srv)
	summaries, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("want 1 summary (current.json excluded), got %d", len(summaries))
	}
	if summaries[0].VideoID != "vid1" {
		t.Errorf("want vid1, got %s", summaries[0].VideoID)
	}
}

// TestList_skipsMalformed は不正 JSON の file を skip してエラーを返さないことを確認します。
func TestList_skipsMalformed(t *testing.T) {
	fake := newFakeGCSServer()
	srv := httptest.NewServer(fake)
	defer srv.Close()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	snap := port.Snapshot{VideoID: "valid", SavedAt: now, Users: []domain.User{{ChannelID: "u1"}}, Comments: []domain.Comment{}}
	fake.put("snapshots/valid.json", marshalSnapshot(t, snap))
	fake.put("snapshots/broken.json", []byte("not-json{{{"))

	store := newTestStore(t, srv)
	summaries, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("List should not return error on malformed file, got: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("want 1 summary (broken.json skipped), got %d", len(summaries))
	}
	if summaries[0].VideoID != "valid" {
		t.Errorf("want valid, got %s", summaries[0].VideoID)
	}
}
