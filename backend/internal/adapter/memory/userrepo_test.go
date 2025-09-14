package memory_test

import (
    "testing"

    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/memory"
)

func TestUserRepo_UpsertAndCount(t *testing.T) {
    r := memory.NewUserRepo()

    if err := r.Upsert("ch1", "Alice"); err != nil { t.Fatalf("upsert: %v", err) }
    if err := r.Upsert("ch1", "Alice 2"); err != nil { t.Fatalf("upsert: %v", err) }

    if got, want := r.Count(), 1; got != want {
        t.Fatalf("Count=%d want %d", got, want)
    }

    names := r.ListDisplayNames()
    if len(names) != 1 || names[0] != "Alice 2" {
        t.Fatalf("ListDisplayNames=%v want [Alice 2]", names)
    }
}

