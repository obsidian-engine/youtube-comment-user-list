package view

import (
    "html/template"
    "net/http"
    "path/filepath"

    "github.com/obsidian-engine/youtube-comment-user-list/internal/domain/repository"
)

// Renderer は HTML テンプレートの描画を担当します。
type Renderer struct {
    tmpl   *template.Template
    logger repository.Logger
}

// NewRenderer はテンプレートをパースして Renderer を返します。
// テンプレートは `internal/interfaces/http/templates/*.tmpl` に配置します。
func NewRenderer(logger repository.Logger) (*Renderer, error) {
    // ルートからの相対パスで読み込む（Cloud Run/ローカル双方で動作）
    pattern := filepath.Join("internal", "interfaces", "http", "templates", "*.tmpl")
    // html/template を使用
    tmpl, err := template.New("base").Funcs(template.FuncMap{
        // eq/neq 等のユーティリティ（必要最低限）。主に string 用で使用。
        "eq": func(a, b any) bool { return a == b },
        "neq": func(a, b any) bool { return a != b },
    }).ParseGlob(pattern)
    if err != nil {
        return nil, err
    }
    return &Renderer{tmpl: tmpl, logger: logger}, nil
}

// Render は指定ページ名（例: "home", "users", "logs"）でテンプレートを描画します。
// 各ページのテンプレートは自分で base を呼び出す薄いラッパーにします。
func (r *Renderer) Render(w http.ResponseWriter, name string, data any) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    if err := r.tmpl.ExecuteTemplate(w, name, data); err != nil {
        r.logger.LogError("ERROR", "template render failed", "", "", err, map[string]interface{}{"template": name})
        http.Error(w, "template render error", http.StatusInternalServerError)
        return
    }
}

// PageData は全ページ共通の最低限のデータです。
type PageData struct {
    Title  string
    Active string // "home" | "users" | "logs"
}
