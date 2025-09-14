package http_test

import (
    stdhttp "net/http"
    "net/http/httptest"
    "testing"

    ahttp "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/http"
    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/memory"
    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase"
)

func newTestServer(frontend string) *httptest.Server {
    users := memory.NewUserRepo()
    state := memory.NewStateRepo()
    h := &ahttp.Handlers{
        Status:      &usecase.Status{Users: users, State: state},
        SwitchVideo: &usecase.SwitchVideo{Users: users, State: state},
        Pull:        &usecase.Pull{Users: users, State: state},
        Reset:       &usecase.Reset{Users: users, State: state},
    }
    router := ahttp.NewRouter(h, frontend)
    return httptest.NewServer(router)
}

func TestRoutes_ExistAndReturn501_ForNow(t *testing.T) {
    ts := newTestServer("http://example.com")
    defer ts.Close()

    cases := []struct{
        method string
        path   string
    }{
        {stdhttp.MethodGet, "/status"},
        {stdhttp.MethodGet, "/users.json"},
        {stdhttp.MethodPost, "/switch-video"},
        {stdhttp.MethodPost, "/pull"},
        {stdhttp.MethodPost, "/reset"},
    }

    for _, c := range cases {
        req, _ := stdhttp.NewRequest(c.method, ts.URL+c.path, nil)
        res, err := stdhttp.DefaultClient.Do(req)
        if err != nil { t.Fatalf("request %s %s: %v", c.method, c.path, err) }
        if res.StatusCode != 501 {
            t.Fatalf("%s %s => %d want 501 (placeholder)", c.method, c.path, res.StatusCode)
        }
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

