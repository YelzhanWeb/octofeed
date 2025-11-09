package services

import (
	"context"
	"fmt"
	"log"
	"time"
)

func (s *AggregatorService) keepAliveLoop() {
	defer s.wg.Done()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			if err := s.ipcLock.KeepAlive(context.Background()); err != nil {
				log.Printf("Warning: failed to keep lock alive: %v\n", err)
			}
		}
	}
}

func (s *AggregatorService) commandListener() {
	defer s.wg.Done()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.checkCommands()
		}
	}
}

func (s *AggregatorService) checkCommands() {
	ctx := context.Background()

	if intervalStr, err := s.ipcLock.GetCommand(ctx, "set_interval"); err == nil && intervalStr != "" {
		if d, err := time.ParseDuration(intervalStr); err == nil {
			log.Printf("DEBUG: Found 'set_interval' command: %s", intervalStr)
			s.applySetInterval(d)
			s.ipcLock.SetCommand(ctx, "set_interval", "")
		}
	}

	if workersStr, err := s.ipcLock.GetCommand(ctx, "set_workers"); err == nil && workersStr != "" {
		var workers int
		if _, err := fmt.Sscanf(workersStr, "%d", &workers); err == nil {
			log.Printf("DEBUG: Found 'set_interval' command: %s", workersStr)
			s.applyResize(workers)
			s.ipcLock.SetCommand(ctx, "set_workers", "")
		}
	}
}
