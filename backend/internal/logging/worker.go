package logging

import (
	"context"
	"log"

	"github.com/ayechanhan/user-records-app/backend/internal/repository"
)

// Worker drains a Bus and persists each event via a LogRepository.
type Worker struct {
	bus  *Bus
	repo repository.LogRepository
	done chan struct{}
}

func NewWorker(bus *Bus, repo repository.LogRepository) *Worker {
	return &Worker{bus: bus, repo: repo, done: make(chan struct{})}
}

// Start launches the drain loop in a goroutine and returns immediately.
func (w *Worker) Start() {
	go func() {
		defer close(w.done)
		for entry := range w.bus.ch {
			if err := w.repo.Create(context.Background(), entry); err != nil {
				log.Printf("logging: failed to persist %s event for user %q: %v", entry.Event, entry.UserID, err)
			}
		}
	}()
}

// Stop closes the bus so no more events can be enqueued, then blocks until
// the worker has drained everything already buffered.
func (w *Worker) Stop() {
	w.bus.Close()
	<-w.done
}
