package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"rsshub/internal/domain"
	"rsshub/internal/ports"
)

type AggregatorService struct {
	feedRepo    ports.FeedRepository
	articleRepo ports.ArticleRepository
	rssFetcher  ports.RSSFetcher

	mu             sync.RWMutex
	interval       atomic.Value // time.Duration
	workersCount   atomic.Int32
	running        atomic.Bool
	ticker         *time.Ticker
	jobs           chan *domain.Feed
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	stopOnce       sync.Once
	resizeMu       sync.Mutex
	currentWorkers int32
}

func NewAggregatorService(
	feedRepo ports.FeedRepository,
	articleRepo ports.ArticleRepository,
	rssFetcher ports.RSSFetcher,
	defaultInterval time.Duration,
	defaultWorkers int,
) ports.AggregatorPort {
	svc := &AggregatorService{
		feedRepo:    feedRepo,
		articleRepo: articleRepo,
		rssFetcher:  rssFetcher,
		jobs:        make(chan *domain.Feed, 100),
	}
	svc.interval.Store(defaultInterval)
	svc.workersCount.Store(int32(defaultWorkers))
	return svc
}

func (s *AggregatorService) Start(ctx context.Context) error {
	if s.running.Load() {
		return fmt.Errorf("aggregator is already running")
	}

	s.running.Store(true)
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	interval := s.interval.Load().(time.Duration)
	s.ticker = time.NewTicker(interval)

	workersCount := int(s.workersCount.Load())
	for i := 0; i < workersCount; i++ {
		s.wg.Add(1)
		go s.worker(ctx, i+1)
	}
	s.currentWorkers = int32(workersCount)

	s.wg.Add(1)
	go s.fetchLoop(ctx)

	log.Printf("The background process for fetching feeds has started (interval = %v, workers = %d)\n",
		interval, workersCount)

	return nil
}

func (s *AggregatorService) Stop() error {
	var stopErr error
	s.stopOnce.Do(func() {
		if !s.running.Load() {
			stopErr = fmt.Errorf("aggregator is not running")
			return
		}

		s.running.Store(false)
		if s.cancel != nil {
			s.cancel()
		}
		if s.ticker != nil {
			s.ticker.Stop()
		}

		s.wg.Wait()
		close(s.jobs)

		log.Println("Graceful shutdown: aggregator stopped")
	})
	return stopErr
}

func (s *AggregatorService) SetInterval(d time.Duration) error {
	if !s.running.Load() {
		return fmt.Errorf("aggregator is not running")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	oldInterval := s.interval.Load().(time.Duration)
	s.interval.Store(d)

	if s.ticker != nil {
		s.ticker.Stop()
		s.ticker = time.NewTicker(d)
	}

	log.Printf("Interval of fetching feeds changed from %v to %v\n", oldInterval, d)
	return nil
}

func (s *AggregatorService) Resize(workers int) error {
	if !s.running.Load() {
		return fmt.Errorf("aggregator is not running")
	}

	s.resizeMu.Lock()
	defer s.resizeMu.Unlock()

	oldCount := int(s.currentWorkers)
	newCount := workers

	if newCount == oldCount {
		return nil
	}

	s.workersCount.Store(int32(newCount))

	ctx := context.Background()
	if s.cancel != nil {
		ctx = context.Background()
	}

	if newCount > oldCount {
		for i := oldCount; i < newCount; i++ {
			s.wg.Add(1)
			go s.worker(ctx, i+1)
		}
	}

	s.currentWorkers = int32(newCount)

	log.Printf("Number of workers changed from %d to %d\n", oldCount, newCount)
	return nil
}

func (s *AggregatorService) IsRunning() bool {
	return s.running.Load()
}

func (s *AggregatorService) GetInterval() time.Duration {
	return s.interval.Load().(time.Duration)
}

func (s *AggregatorService) GetWorkersCount() int {
	return int(s.workersCount.Load())
}

func (s *AggregatorService) fetchLoop(ctx context.Context) {
	defer s.wg.Done()

	s.processBatch(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.ticker.C:
			s.processBatch(ctx)
		}
	}
}

func (s *AggregatorService) processBatch(ctx context.Context) {
	workersCount := int(s.workersCount.Load())
	feeds, err := s.feedRepo.GetMostOutdated(ctx, workersCount*2)
	if err != nil {
		log.Printf("Error fetching outdated feeds: %v\n", err)
		return
	}

	for _, feed := range feeds {
		select {
		case <-ctx.Done():
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

			currentWorkers := int(s.currentWorkers)
			if id > currentWorkers {
				return
			}

			s.processFeed(ctx, feed)
		}
	}
}

func (s *AggregatorService) processFeed(ctx context.Context, feed *domain.Feed) {
	log.Printf("Worker processing feed: %s (%s)\n", feed.Name, feed.URL)

	rssFeed, err := s.rssFetcher.Fetch(ctx, feed.URL)
	if err != nil {
		log.Printf("Error fetching RSS feed %s: %v\n", feed.Name, err)
		return
	}

	var newArticles []*domain.Article
	for _, item := range rssFeed.Channel.Items {
		exists, err := s.articleRepo.Exists(ctx, item.Link, feed.ID)
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
		if err := s.articleRepo.CreateBatch(ctx, newArticles); err != nil {
			log.Printf("Error saving articles for feed %s: %v\n", feed.Name, err)
			return
		}
		log.Printf("Saved %d new articles for feed %s\n", len(newArticles), feed.Name)
	}

	feed.MarkAsFetched()
	if err := s.feedRepo.Update(ctx, feed); err != nil {
		log.Printf("Error updating feed %s: %v\n", feed.Name, err)
	}
}
