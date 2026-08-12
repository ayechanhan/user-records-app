// Package logging implements the async audit-log path described in
// plan.md: handlers enqueue a UserLog onto an in-process buffered channel,
// and a background worker drains it into the document store. The request
// that triggered an event never waits on the write.
package logging

import (
	"log"
	"time"

	"github.com/ayechanhan/user-records-app/backend/internal/model"
)

// DefaultBufferSize is how many events the channel holds before Emit starts
// dropping — plan.md's documented trade-off for using an in-process channel
// instead of a durable queue: not durable across a crash or a sustained burst.
const DefaultBufferSize = 256

// Bus is the buffered channel producers enqueue UserLog events onto.
type Bus struct {
	ch chan *model.UserLog
}

func NewBus(bufferSize int) *Bus {
	return &Bus{ch: make(chan *model.UserLog, bufferSize)}
}

// Emit builds a UserLog and enqueues it without blocking the caller. If the
// buffer is full the event is dropped and a warning is logged rather than
// blocking the request that triggered it (see spec.md FR5).
func (b *Bus) Emit(userID string, event model.LogEvent, data map[string]any) {
	entry := &model.UserLog{
		UserID:    userID,
		Event:     event,
		Data:      data,
		CreatedAt: time.Now(),
	}
	select {
	case b.ch <- entry:
	default:
		log.Printf("logging: buffer full, dropping %s event for user %q", event, userID)
	}
}

// Close signals no more events will be enqueued. Callers must ensure Emit is
// never called concurrently with or after Close — e.g. call it only after
// the HTTP server has stopped accepting requests.
func (b *Bus) Close() {
	close(b.ch)
}
