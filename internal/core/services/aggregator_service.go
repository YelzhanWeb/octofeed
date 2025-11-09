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
		fmt.Println("Background process is already running")
		return nil
	}

	s.running = true
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.ticker = time.NewTicker(s.interval)
	s.workerContexts = make([]context.CancelFunc, 0)

	s.startWorkers(s.workersCount)

	s.wg.Add(2)
	go s.fetchLoop()
	go s.commandListener()

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

func (s *AggregatorService) SetInterval(d time.Duration) error {
	err := s.ipcLock.SetCommand(context.Background(), "set_interval", d.String())
	if err != nil {
		return fmt.Errorf("failed to set interval command: %w", err)
	}
	return nil
}

func (s *AggregatorService) Resize(workers int) error {
	err := s.ipcLock.SetCommand(context.Background(), "set_workers", fmt.Sprintf("%d", workers))
	if err != nil {
		return fmt.Errorf("failed to set workers command: %w", err)
	}
	return nil
}

func (s *AggregatorService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

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
			s.applySetInterval(d)
			s.ipcLock.SetCommand(ctx, "set_interval", "") // Очищаем команду
		}
	}

	if workersStr, err := s.ipcLock.GetCommand(ctx, "set_workers"); err == nil && workersStr != "" {
		var workers int
		if _, err := fmt.Sscanf(workersStr, "%d", &workers); err == nil {
			s.applyResize(workers)
			s.ipcLock.SetCommand(ctx, "set_workers", "") // Очищаем команду
		}
	}
}

func (s *AggregatorService) applySetInterval(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldInterval := s.interval
	s.interval = d

	if s.ticker != nil {
		s.ticker.Stop()
		s.ticker = time.NewTicker(d)
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

func (s *AggregatorService) startWorkers(count int) {
	for i := 0; i < count; i++ {
		workerCtx, workerCancel := context.WithCancel(s.ctx)
		s.workerContexts = append(s.workerContexts, workerCancel)

		s.wg.Add(1)
		go s.worker(workerCtx, len(s.workerContexts))
	}
}

func (s *AggregatorService) fetchLoop() {
	defer s.wg.Done()

	s.processBatch()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.ticker.C:
			s.processBatch()
		}
	}
}

func (s *AggregatorService) processBatch() {
	s.mu.RLock()
	workersCount := s.workersCount
	s.mu.RUnlock()

	feeds, err := s.feedRepo.GetMostOutdated(context.Background(), workersCount*2)
	if err != nil {
		log.Printf("Error fetching outdated feeds: %v\n", err)
		return
	}

	for _, feed := range feeds {
		select {
		case <-s.ctx.Done():
			return
		case s.jobs <- feed:
		default:
			log.Printf("Job queue full, skipping feed %s\n", feed.Name)
		}
	}
}

func (s *AggregatorService) worker(ctx context.Context, id int) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case feed, ok := <-s.jobs:
			if !ok {
				return
			}
			s.processFeed(feed)
		}
	}
}

func (s *AggregatorService) processFeed(feed *domain.Feed) {
	log.Printf("Worker processing feed: %s (%s)\n", feed.Name, feed.URL)

	rssFeed, err := s.rssFetcher.Fetch(context.Background(), feed.URL)
	if err != nil {
		log.Printf("Error fetching RSS feed %s: %v\n", feed.Name, err)
		return
	}

	var newArticles []*domain.Article
	for _, item := range rssFeed.Channel.Items {
		exists, err := s.articleRepo.Exists(context.Background(), item.Link, feed.ID)
		if err != nil {
			log.Printf("Error checking article existence: %v\n", err)
			continue
		}
		if exists {
			continue
		}

		pubDate := item.ParsePubDate()
		article := domain.NewArticle(item.Title, item.Link, item.Description, pubDate, feed.ID)
		newArticles = append(newArticles, article)
	}

	if len(newArticles) > 0 {
		if err := s.articleRepo.CreateBatch(context.Background(), newArticles); err != nil {
			log.Printf("Error saving articles for feed %s: %v\n", feed.Name, err)
			return
		}
		log.Printf("Saved %d new articles for feed %s\n", len(newArticles), feed.Name)
	}

	feed.MarkAsFetched()
	if err := s.feedRepo.Update(context.Background(), feed); err != nil {
		log.Printf("Error updating feed %s: %v\n", feed.Name, err)
	}
}
