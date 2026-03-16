package database

import (
	"context"
	"time"
)

// ContextWithTimeout crea un contexto con timeout derivado del contexto padre.
// Si parent es nil se usa context.Background().
func ContextWithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithTimeout(parent, timeout)
}

// ContextWithDeadline crea un contexto con deadline derivado del contexto padre.
func ContextWithDeadline(parent context.Context, deadline time.Time) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithDeadline(parent, deadline)
}
