package memory

import (
	"testing"
	"time"
)

func TestUpsertWithMessageResult_ReportsNewOnlyOnFirstComment(t *testing.T) {
	r := NewUserRepo()
	now := time.Now()

	updated, isNew, err := r.UpsertWithMessageResult("UC1", "Alice", now, "m1")
	if err != nil || !updated || !isNew {
		t.Fatalf("first comment: updated=%v isNew=%v err=%v", updated, isNew, err)
	}
	updated, isNew, _ = r.UpsertWithMessageResult("UC1", "Alice", now, "m2")
	if !updated || isNew {
		t.Fatalf("second comment: updated=%v isNew=%v", updated, isNew)
	}
	updated, isNew, _ = r.UpsertWithMessageResult("UC1", "Alice", now, "m2")
	if updated || isNew {
		t.Fatalf("duplicate message: updated=%v isNew=%v", updated, isNew)
	}

	// リセットで一覧を消した後は、同じ人も新しく登録される
	r.Clear()
	_, isNew, _ = r.UpsertWithMessageResult("UC1", "Alice", now, "m3")
	if !isNew {
		t.Fatal("expected new after Clear")
	}
}
