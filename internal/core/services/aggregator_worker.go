package services

import (
	"context"
	"log"

	"rsshub/internal/domain"
)

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

	log.Printf("DEBUG: Found %d outdated feeds to process", len(feeds))

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
