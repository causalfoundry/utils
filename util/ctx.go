package util

import (
	"context"
	"time"
)

func WithTimeout(dur time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), dur)
}
