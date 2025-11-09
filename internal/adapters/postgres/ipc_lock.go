package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const (
	lockTimeout = 5 * time.Minute
)

type IPCLock struct {
	db     *DB
	lockID string
}

func NewIPCLock(db *DB) *IPCLock {
	return &IPCLock{
		db:     db,
		lockID: fmt.Sprintf("%d", time.Now().UnixNano()),
	}
}

func (l *IPCLock) TryAcquire(ctx context.Context) (bool, error) {
	var existingLockID string
	var updatedAt time.Time
	query := `SELECT lock_id, updated_at FROM aggregator_lock WHERE id = 1`

	err := l.db.conn.QueryRowContext(ctx, query).Scan(&existingLockID, &updatedAt)
	if err == nil {
		if time.Since(updatedAt) < lockTimeout {
			return false, nil
		}
	} else if err != sql.ErrNoRows {
		return false, fmt.Errorf("failed to check existing lock: %w", err)
	}

	insertQuery := `
		INSERT INTO aggregator_lock (id, lock_id, locked_at, updated_at)
		VALUES (1, $1, NOW(), NOW())
		ON CONFLICT (id) DO UPDATE 
		SET lock_id = $1, locked_at = NOW(), updated_at = NOW()
		WHERE aggregator_lock.updated_at < NOW() - INTERVAL '5 minutes' 
		OR aggregator_lock.lock_id = $1
		RETURNING lock_id
	`

	var returnedLockID string
	err = l.db.conn.QueryRowContext(ctx, insertQuery, l.lockID).Scan(&returnedLockID)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}

	return returnedLockID == l.lockID, nil
}

func (l *IPCLock) Release(ctx context.Context) error {
	query := `DELETE FROM aggregator_lock WHERE id = 1 AND lock_id = $1`
	result, err := l.db.conn.ExecContext(ctx, query, l.lockID)
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("lock was not held by this process")
	}

	return nil
}

func (l *IPCLock) KeepAlive(ctx context.Context) error {
	query := `UPDATE aggregator_lock SET updated_at = NOW() WHERE id = 1 AND lock_id = $1`
	result, err := l.db.conn.ExecContext(ctx, query, l.lockID)
	if err != nil {
		return fmt.Errorf("failed to keep lock alive: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("lock was lost")
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
