package memory

import "sync"

type UserRepo struct {
	mu     sync.RWMutex
	byChan map[string]string
}

func NewUserRepo() *UserRepo {
	return &UserRepo{byChan: make(map[string]string)}
}

func (r *UserRepo) Upsert(channelID string, displayName string) error {
	r.mu.Lock()
	r.byChan[channelID] = displayName
	r.mu.Unlock()
	return nil
}

func (r *UserRepo) ListDisplayNames() []string {
	r.mu.RLock()
	names := make([]string, 0, len(r.byChan))
	for _, n := range r.byChan {
		names = append(names, n)
	}
	r.mu.RUnlock()
	return names
}

func (r *UserRepo) Count() int {
	r.mu.RLock()
	c := len(r.byChan)
	r.mu.RUnlock()
	return c
}

func (r *UserRepo) Clear() {
	r.mu.Lock()
	r.byChan = make(map[string]string)
	r.mu.Unlock()
}
