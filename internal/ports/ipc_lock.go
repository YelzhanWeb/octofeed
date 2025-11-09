package ports

import "context"

type IPCLock interface {
	TryAcquire(ctx context.Context) (bool, error)
	Release(ctx context.Context) error
	SetCommand(ctx context.Context, command string, value string) error
	GetCommand(ctx context.Context, command string) (string, error)
}
