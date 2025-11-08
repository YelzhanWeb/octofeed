package http

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"rsshub/internal/domain"
)

type RSSFetcher struct {
	client *http.Client
}

func NewRSSFetcher() *RSSFetcher {
	return &RSSFetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (f *RSSFetcher) Fetch(ctx context.Context, url string) (*domain.RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "RSSHub/1.0")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSS feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var rssFeed domain.RSSFeed
	if err := xml.Unmarshal(body, &rssFeed); err != nil {
		return nil, fmt.Errorf("failed to parse RSS feed: %w", err)
	}

	return &rssFeed, nil
}
