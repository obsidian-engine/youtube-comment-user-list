package logging

import (
	"context"
	"sync"
	"testing"
)

func TestCollector_AddAndEntries(t *testing.T) {
	c := NewCollector()
	c.Add("info", "SRC", "msg1")
	c.Add("warn", "SRC", "msg2")

	entries := c.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Level != "info" || entries[0].Message != "msg1" {
		t.Errorf("unexpected first entry: %+v", entries[0])
	}
	if entries[1].Level != "warn" || entries[1].Message != "msg2" {
		t.Errorf("unexpected second entry: %+v", entries[1])
	}
}

func TestCollector_EntriesReturnsCopy(t *testing.T) {
	c := NewCollector()
	c.Add("info", "SRC", "original")

	entries := c.Entries()
	entries[0].Message = "modified"

	got := c.Entries()
	if got[0].Message != "original" {
		t.Errorf("Entries should return a copy, but original was modified")
	}
}

func TestCollector_ConcurrentAccess(t *testing.T) {
	c := NewCollector()
	var wg sync.WaitGroup

	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Add("info", "SRC", "concurrent")
		}()
	}
	wg.Wait()

	entries := c.Entries()
	if len(entries) != 100 {
		t.Fatalf("expected 100 entries, got %d", len(entries))
	}
}

func TestWithCollectorAndFromContext(t *testing.T) {
	c := NewCollector()
	ctx := WithCollector(context.Background(), c)

	got := FromContext(ctx)
	if got != c {
		t.Errorf("FromContext should return the same collector")
	}
}

func TestFromContext_NoCollector(t *testing.T) {
	got := FromContext(context.Background())
	if got != nil {
		t.Errorf("FromContext should return nil when no collector is set, got %v", got)
	}
}

func TestLog_WithCollector(t *testing.T) {
	c := NewCollector()
	ctx := WithCollector(context.Background(), c)

	Log(ctx, "warn", "TEST", "hello %s", "world")

	entries := c.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Level != "warn" {
		t.Errorf("expected level 'warn', got %q", entries[0].Level)
	}
	if entries[0].Source != "TEST" {
		t.Errorf("expected source 'TEST', got %q", entries[0].Source)
	}
	if entries[0].Message != "hello world" {
		t.Errorf("expected message 'hello world', got %q", entries[0].Message)
	}
}

func TestLog_WithoutCollector(t *testing.T) {
	// Collectorなしでもpanicしないことを確認
	Log(context.Background(), "info", "TEST", "no collector")
}
