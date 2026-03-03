package logging

import (
	"context"
	"fmt"
	"log"
	"sync"
)

type LogEntry struct {
	Level   string
	Source  string
	Message string
}

type Collector struct {
	mu      sync.RWMutex
	entries []LogEntry
}

func NewCollector() *Collector {
	return &Collector{}
}

func (c *Collector) Add(level, source, message string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = append(c.entries, LogEntry{Level: level, Source: source, Message: message})
}

func (c *Collector) Entries() []LogEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]LogEntry, len(c.entries))
	copy(out, c.entries)
	return out
}

type contextKey struct{}

func WithCollector(ctx context.Context, c *Collector) context.Context {
	return context.WithValue(ctx, contextKey{}, c)
}

func FromContext(ctx context.Context) *Collector {
	c, _ := ctx.Value(contextKey{}).(*Collector)
	return c
}

// Log はログ出力とコレクターへの記録を同時に行います。
func Log(ctx context.Context, level, source, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	log.Printf("[%s] %s", source, msg)
	if c := FromContext(ctx); c != nil {
		c.Add(level, source, msg)
	}
}
