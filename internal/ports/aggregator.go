package ports

import (
	"context"
	"time"
)

type AggregatorPort interface {
	Start(ctx context.Context) error
	Stop() error
	SetInterval(d time.Duration) error
	Resize(workers int) error
	IsRunning() bool
	GetInterval() time.Duration
	GetWorkersCount() int
}
