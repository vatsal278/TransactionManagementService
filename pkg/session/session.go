// Package session provides a simple mechanism for storing session data in a context.Context object.
package session

import "context"

// session is an empty struct used as a key for retrieving the session data from a context.Context object.
type session struct{}

// SetSession stores the session data in the provided context and returns a new context with the session data.
func SetSession(ctx context.Context, value interface{}) context.Context {
	return context.WithValue(ctx, session{}, value)
}

// GetSession returns the session data stored in the provided context.
func GetSession(ctx context.Context) interface{} {
	return ctx.Value(session{})
}
