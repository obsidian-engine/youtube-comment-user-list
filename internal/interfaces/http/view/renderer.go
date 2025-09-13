package view

import (
    "fmt"
    "html/template"
    "net/http"
    "path/filepath"

    "github.com/obsidian-engine/youtube-comment-user-list/internal/domain/repository"
)

// Renderer は HTML テンプレートの描画を担当します。
// ページごと（home/users/logs）に独立したテンプレートセットを持ち、
// block "content" の衝突で全ページが同一になる問題を回避します。
type Renderer struct {
    pages  map[string]*template.Template
    logger repository.Logger
}

// NewRenderer はテンプレートをパースして Renderer を返します。
func NewRenderer(logger repository.Logger) (*Renderer, error) {
    dir := filepath.Join("internal", "interfaces", "http", "templates")
    base := filepath.Join(dir, "base.gohtml")
    files := map[string]string{
        "home":  filepath.Join(dir, "home.gohtml"),
        "users": filepath.Join(dir, "users.gohtml"),
        "logs":  filepath.Join(dir, "logs.gohtml"),
    }

    pages := make(map[string]*template.Template, len(files))
    for name, page := range files {
        t, err := template.New("base").Funcs(template.FuncMap{
            "eq":  func(a, b any) bool { return a == b },
            "neq": func(a, b any) bool { return a != b },
        }).ParseFiles(base, page)
        if err != nil {
            return nil, fmt.Errorf("parse template %s: %w", name, err)
        }
        pages[name] = t
    }
    return &Renderer{pages: pages, logger: logger}, nil
}

// Render は指定ページ名（"home"|"users"|"logs"）のテンプレートを描画します。
func (r *Renderer) Render(w http.ResponseWriter, name string, data any) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    t, ok := r.pages[name]
    if !ok || t == nil {
        r.logger.LogError("ERROR", "template not found", "", "", fmt.Errorf("template %s not found", name), nil)
        http.Error(w, "template not found", http.StatusInternalServerError)
        return
    }
    if err := t.ExecuteTemplate(w, name, data); err != nil {
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
