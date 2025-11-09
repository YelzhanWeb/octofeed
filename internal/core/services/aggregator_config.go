package services

import (
	"context"
	"fmt"
	"log"
	"time"
)

func (s *AggregatorService) GetInterval() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.interval
}

func (s *AggregatorService) GetWorkersCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.workersCount
}

func (s *AggregatorService) SetInterval(d time.Duration) error {
	ctx := context.Background()

	err := s.ipcLock.SetCommand(ctx, "set_interval", d.String())
	if err != nil {
		return fmt.Errorf("failed to set interval command: %w", err)
	}

	s.mu.RLock()
	isRunning := s.running
	s.mu.RUnlock()

	if isRunning {
		log.Printf("Interval change scheduled: %v\n", d)
	} else {
		log.Printf("Interval change command sent (will be applied by running aggregator)\n")
	}

	return nil
}

func (s *AggregatorService) Resize(workers int) error {
	ctx := context.Background()

	err := s.ipcLock.SetCommand(ctx, "set_workers", fmt.Sprintf("%d", workers))
	if err != nil {
		return fmt.Errorf("failed to set workers command: %w", err)
	}

	s.mu.RLock()
	isRunning := s.running
	s.mu.RUnlock()

	if isRunning {
		log.Printf("Workers count change scheduled: %d\n", workers)
	} else {
		log.Printf("Workers count change command sent (will be applied by running aggregator)\n")
	}

	return nil
}

func (s *AggregatorService) applySetInterval(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldInterval := s.interval
	s.interval = d

	if s.ticker != nil {
		s.ticker.Reset(d)
	}

	log.Printf("Interval of fetching feeds changed from %v to %v\n", oldInterval, d)
}

func (s *AggregatorService) applyResize(newCount int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldCount := s.workersCount
	if newCount == oldCount {
		return
	}

	if newCount > oldCount {
		s.startWorkers(newCount - oldCount)
	} else {
		diff := oldCount - newCount
		for i := 0; i < diff && len(s.workerContexts) > 0; i++ {
			lastIdx := len(s.workerContexts) - 1
			s.workerContexts[lastIdx]()
			s.workerContexts = s.workerContexts[:lastIdx]
		}
	}

	s.workersCount = newCount
	log.Printf("Number of workers changed from %d to %d\n", oldCount, newCount)
}
