package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"rsshub/internal/domain"
	"rsshub/internal/ports"
)

type AggregatorService struct {
	feedRepo    ports.FeedRepository
	articleRepo ports.ArticleRepository
	rssFetcher  ports.RSSFetcher
	ipcLock     ports.IPCLock

	mu             sync.RWMutex
	interval       time.Duration
	workersCount   int
	running        bool
	ticker         *time.Ticker
	jobs           chan *domain.Feed
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	stopOnce       sync.Once
	workerContexts []context.CancelFunc
}

func NewAggregatorService(
	feedRepo ports.FeedRepository,
	articleRepo ports.ArticleRepository,
	rssFetcher ports.RSSFetcher,
	ipcLock ports.IPCLock,
	defaultInterval time.Duration,
	defaultWorkers int,
) ports.AggregatorPort {
	return &AggregatorService{
		feedRepo:     feedRepo,
		articleRepo:  articleRepo,
		rssFetcher:   rssFetcher,
		ipcLock:      ipcLock,
		interval:     defaultInterval,
		workersCount: defaultWorkers,
		jobs:         make(chan *domain.Feed, 100),
	}
}

func (s *AggregatorService) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("aggregator is already running in this process")
	}

	acquired, err := s.ipcLock.TryAcquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to check lock: %w", err)
	}
	if !acquired {
		return domain.ErrAggregatorAlreadyRunning
	}

	s.running = true
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.ticker = time.NewTicker(s.interval)
	s.workerContexts = make([]context.CancelFunc, 0)

	s.startWorkers(s.workersCount)

	s.wg.Add(3)
	go s.fetchLoop()
	go s.commandListener()
	go s.keepAliveLoop()

	log.Printf("The background process for fetching feeds has started (interval = %v, workers = %d)\n",
		s.interval, s.workersCount)

	return nil
}

func (s *AggregatorService) Stop() error {
	var stopErr error
	s.stopOnce.Do(func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		if !s.running {
			stopErr = fmt.Errorf("aggregator is not running")
			return
		}

		s.running = false
		if s.cancel != nil {
			s.cancel()
		}
		if s.ticker != nil {
			s.ticker.Stop()
		}

		for _, cancelFunc := range s.workerContexts {
			cancelFunc()
		}

		if err := s.ipcLock.Release(context.Background()); err != nil {
			log.Printf("Warning: failed to release lock: %v\n", err)
		}

		s.wg.Wait()
		close(s.jobs)

		log.Println("Graceful shutdown: aggregator stopped")
	})
	return stopErr
}

func (s *AggregatorService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}
