package domain

import (
	"errors"
)

var (
	ErrNotFound                 = errors.New("Not Found")
	ErrAggregatorAlreadyRunning = errors.New("background process is already running")
)
