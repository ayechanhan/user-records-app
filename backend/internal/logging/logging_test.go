package logging

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ayechanhan/user-records-app/backend/internal/model"
)

// mockLogRepo implements repository.LogRepository, recording every event it
// receives. Reads after Worker.Stop() are safe without extra synchronization
// since Stop's channel receive establishes happens-before, but the mutex is
// kept for safety against future concurrent use.
type mockLogRepo struct {
	mu         sync.Mutex
	created    []*model.UserLog
	createFunc func(ctx context.Context, entry *model.UserLog) error
}

func (m *mockLogRepo) Create(ctx context.Context, entry *model.UserLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.created = append(m.created, entry)
	if m.createFunc != nil {
		return m.createFunc(ctx, entry)
	}
	return nil
}

func (m *mockLogRepo) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]model.UserLog, int64, error) {
	return nil, 0, nil
}

func TestBus_Emit_NonBlockingWhenFull(t *testing.T) {
	bus := NewBus(1)
	bus.Emit("u1", model.EventUserCreated, nil)

	done := make(chan struct{})
	go func() {
		bus.Emit("u2", model.EventUserCreated, nil) // buffer already full
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Emit blocked when the buffer was full")
	}

	if len(bus.ch) != 1 {
		t.Fatalf("expected buffer to still hold 1 event, got %d", len(bus.ch))
	}
}

func TestWorker_DrainsQueuedEventsInOrder(t *testing.T) {
	bus := NewBus(10)
	repo := &mockLogRepo{}
	worker := NewWorker(bus, repo)
	worker.Start()

	bus.Emit("user-1", model.EventUserCreated, map[string]any{"name": "Ada"})
	bus.Emit("user-1", model.EventUserUpdated, map[string]any{"name": "Ada L."})

	worker.Stop()

	if len(repo.created) != 2 {
		t.Fatalf("expected 2 events consumed, got %d", len(repo.created))
	}
	if repo.created[0].Event != model.EventUserCreated || repo.created[1].Event != model.EventUserUpdated {
		t.Fatalf("expected events consumed in order, got %v then %v", repo.created[0].Event, repo.created[1].Event)
	}
	if repo.created[0].UserID != "user-1" {
		t.Fatalf("expected user_id to round-trip, got %q", repo.created[0].UserID)
	}
}

func TestWorker_ContinuesAfterRepositoryError(t *testing.T) {
	bus := NewBus(10)
	var calls int
	repo := &mockLogRepo{
		createFunc: func(ctx context.Context, entry *model.UserLog) error {
			calls++
			if calls == 1 {
				return errors.New("mongo: write failed")
			}
			return nil
		},
	}
	worker := NewWorker(bus, repo)
	worker.Start()

	bus.Emit("u1", model.EventUserCreated, nil)
	bus.Emit("u1", model.EventUserUpdated, nil)

	worker.Stop()

	if calls != 2 {
		t.Fatalf("expected worker to keep draining after a write error, got %d calls", calls)
	}
}
