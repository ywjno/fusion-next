package fetcher

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	readability "github.com/go-shiori/go-readability"
)

type FetchOptions struct {
	URL     string
	Timeout time.Duration
}

type FetchResult struct {
	Content string
	Title   string
	Error   error
}

func FetchFullContent(options FetchOptions) *FetchResult {
	if options.Timeout == 0 {
		options.Timeout = 30 * time.Second
	}

	client := &http.Client{
		Timeout: options.Timeout,
	}

	req, err := http.NewRequest("GET", options.URL, nil)
	if err != nil {
		slog.Warn("Failed to create request", "url", options.URL, "error", err)
		return &FetchResult{Error: fmt.Errorf("failed to create request: %w", err)}
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; FusionRSS/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		slog.Warn("Failed to fetch URL", "url", options.URL, "error", err)
		return &FetchResult{Error: fmt.Errorf("failed to fetch URL: %w", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("Non-200 status code", "url", options.URL, "status", resp.StatusCode)
		return &FetchResult{Error: fmt.Errorf("non-200 status code: %d", resp.StatusCode)}
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(strings.ToLower(contentType), "text/html") {
		slog.Warn("Non-HTML content type", "url", options.URL, "content_type", contentType)
		return &FetchResult{Error: fmt.Errorf("non-HTML content type: %s", contentType)}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Warn("Failed to read response body", "url", options.URL, "error", err)
		return &FetchResult{Error: fmt.Errorf("failed to read response body: %w", err)}
	}

	parsedURL, err := url.Parse(options.URL)
	if err != nil {
		slog.Warn("Failed to parse URL", "url", options.URL, "error", err)
		return &FetchResult{Error: fmt.Errorf("failed to parse URL: %w", err)}
	}

	article, err := readability.FromReader(strings.NewReader(string(body)), parsedURL)
	if err != nil {
		slog.Warn("Failed to parse content with readability", "url", options.URL, "error", err)
		return &FetchResult{Error: fmt.Errorf("failed to parse content: %w", err)}
	}

	slog.Info("Successfully fetched full content", "url", options.URL, "title", article.Title)

	return &FetchResult{
		Content: article.Content,
		Title:   article.Title,
		Error:   nil,
	}
}
