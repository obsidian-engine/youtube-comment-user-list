package http_test

import (
    stdhttp "net/http"
    "net/http/httptest"
    "testing"

    ahttp "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/http"
    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/memory"
    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/system"
    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/youtube"
    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase"
)

func newTestServer(frontend string) *httptest.Server {
    users := memory.NewUserRepo()
    state := memory.NewStateRepo()
    yt := youtube.New("") // テスト用の空キー
    clock := system.NewSystemClock()
    
    h := &ahttp.Handlers{
        Status:      &usecase.Status{Users: users, State: state},
        SwitchVideo: &usecase.SwitchVideo{YT: yt, Users: users, State: state, Clock: clock},
        Pull:        &usecase.Pull{YT: yt, Users: users, State: state},
        Reset:       &usecase.Reset{Users: users, State: state},
        Users:       users,
    }
    router := ahttp.NewRouter(h, frontend)
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
    if res.StatusCode != 200 {
        t.Errorf("GET /status => %d want 200", res.StatusCode)
    }

    // GET /users.json should return 200
    req, _ = stdhttp.NewRequest(stdhttp.MethodGet, ts.URL+"/users.json", nil)
    res, err = stdhttp.DefaultClient.Do(req)
    if err != nil { 
        t.Fatalf("request GET /users.json: %v", err) 
    }
    if res.StatusCode != 200 {
        t.Errorf("GET /users.json => %d want 200", res.StatusCode)
    }

    // POST /reset should return 200
    req, _ = stdhttp.NewRequest(stdhttp.MethodPost, ts.URL+"/reset", nil)
    res, err = stdhttp.DefaultClient.Do(req)
    if err != nil { 
        t.Fatalf("request POST /reset: %v", err) 
    }
    if res.StatusCode != 200 {
        t.Errorf("POST /reset => %d want 200", res.StatusCode)
    }

    // POST /pull should return 200
    req, _ = stdhttp.NewRequest(stdhttp.MethodPost, ts.URL+"/pull", nil)
    res, err = stdhttp.DefaultClient.Do(req)
    if err != nil { 
        t.Fatalf("request POST /pull: %v", err) 
    }
    if res.StatusCode != 200 {
        t.Errorf("POST /pull => %d want 200", res.StatusCode)
    }
}

func TestCORS_AllowsFrontendOrigin(t *testing.T) {
    origin := "https://frontend.example"
    ts := newTestServer(origin)
    defer ts.Close()

    req, _ := stdhttp.NewRequest(stdhttp.MethodOptions, ts.URL+"/status", nil)
    req.Header.Set("Origin", origin)
    res, err := stdhttp.DefaultClient.Do(req)
    if err != nil { t.Fatalf("preflight: %v", err) }
    if got := res.Header.Get("Access-Control-Allow-Origin"); got != origin {
        t.Fatalf("Allow-Origin=%q want %q", got, origin)
    }
    if res.StatusCode != 204 {
        t.Fatalf("preflight status=%d want 204", res.StatusCode)
    }
}

