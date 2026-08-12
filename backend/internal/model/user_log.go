package model

import "time"

// LogEvent identifies the kind of user-affecting action a UserLog records.
type LogEvent string

const (
	EventUserCreated     LogEvent = "user.created"
	EventUserUpdated     LogEvent = "user.updated"
	EventUserDeleted     LogEvent = "user.deleted"
	EventUserLogin       LogEvent = "user.login"
	EventUserLoginFailed LogEvent = "user.login_failed"
)

// UserLog is a single audit event, written to the document store.
// Data is event-specific: e.g. a diff for user.updated, empty for user.deleted.
type UserLog struct {
	UserID    string         `bson:"user_id"`
	Event     LogEvent       `bson:"event"`
	Data      map[string]any `bson:"data"`
	CreatedAt time.Time      `bson:"created_at"`
}
