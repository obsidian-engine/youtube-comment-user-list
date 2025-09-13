package handler

import (
    "net/http"

    "github.com/obsidian-engine/youtube-comment-user-list/internal/constants"
    "github.com/obsidian-engine/youtube-comment-user-list/internal/domain/repository"
    "github.com/obsidian-engine/youtube-comment-user-list/internal/interfaces/http/view"
)

// StaticHandler 静的ファイル配信とHTMLページを処理します
type StaticHandler struct {
    logger   repository.Logger
    renderer *view.Renderer
}

// NewStaticHandler 新しい静的ハンドラーを作成します
func NewStaticHandler(logger repository.Logger) *StaticHandler {
    r, err := view.NewRenderer(logger)
    if err != nil {
        logger.LogError("ERROR", "failed to init template renderer", "", "", err, nil)
    }
    return &StaticHandler{logger: logger, renderer: r}
}

// ServeHome GET / を処理します
func (h *StaticHandler) ServeHome(w http.ResponseWriter, r *http.Request) {
    h.logger.LogAPI(constants.LogLevelInfo, "Home page request", "", "", map[string]interface{}{
        "userAgent":  r.Header.Get("User-Agent"),
        "remoteAddr": r.RemoteAddr,
    })
    // ホームはユーザー一覧と統合: 同一テンプレートを表示
    h.renderer.Render(w, "users", &view.PageData{Title: "YouTube Live Chat Monitor", Active: "users"})
}

// ServeUserListPage GET /users を処理します
func (h *StaticHandler) ServeUserListPage(w http.ResponseWriter, r *http.Request) {
    h.logger.LogAPI(constants.LogLevelInfo, "User list page request", "", "", map[string]interface{}{
        "userAgent":  r.Header.Get("User-Agent"),
        "remoteAddr": r.RemoteAddr,
    })
    h.renderer.Render(w, "users", &view.PageData{Title: "User List - YouTube Live Chat Monitor", Active: "users"})
}

// ServeLogsPage GET /logs を処理します
func (h *StaticHandler) ServeLogsPage(w http.ResponseWriter, r *http.Request) {
    h.logger.LogAPI(constants.LogLevelInfo, "Logs page request", "", "", map[string]interface{}{
        "userAgent": r.Header.Get("User-Agent"),
        "remoteAddr": r.RemoteAddr,
    })
    h.renderer.Render(w, "logs", &view.PageData{Title: "System Logs", Active: "logs"})
}
