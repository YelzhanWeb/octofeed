package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

const aggregatorLockID = 123456789

type IPCLock struct {
	db *DB
}

func NewIPCLock(db *DB) *IPCLock {
	return &IPCLock{db: db}
}

func (l *IPCLock) TryAcquire(ctx context.Context) (bool, error) {
	var acquired bool
	query := `SELECT pg_try_advisory_lock($1)`
	err := l.db.conn.QueryRowContext(ctx, query, aggregatorLockID).Scan(&acquired)
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}
	return acquired, nil
}

func (l *IPCLock) Release(ctx context.Context) error {
	var released bool
	query := `SELECT pg_advisory_unlock($1)`
	err := l.db.conn.QueryRowContext(ctx, query, aggregatorLockID).Scan(&released)
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}
	if !released {
		return fmt.Errorf("lock was not held")
	}
	return nil
}

func (l *IPCLock) SetCommand(ctx context.Context, command string, value string) error {
	query := `
		INSERT INTO ipc_commands (command, value, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (command) DO UPDATE SET value = $2, created_at = NOW()
	`
	_, err := l.db.conn.ExecContext(ctx, query, command, value)
	return err
}

func (l *IPCLock) GetCommand(ctx context.Context, command string) (string, error) {
	var value string
	query := `SELECT value FROM ipc_commands WHERE command = $1`
	err := l.db.conn.QueryRowContext(ctx, query, command).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}
