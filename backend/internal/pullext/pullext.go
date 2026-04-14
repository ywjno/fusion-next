package pullext

import (
	"log/slog"
	"time"

	"github.com/0x2E/fusion/internal/fetcher"
	"github.com/0x2E/fusion/internal/fetcherext"
	"github.com/0x2E/fusion/internal/model"
)

// MaybeFetchFullContent attempts to fetch full content for an item when the
// feed or global configuration enables it.
func MaybeFetchFullContent(
	feed *model.Feed,
	link string,
	content string,
	timeout time.Duration,
	globalAutoFetch bool,
	logger *slog.Logger,
) string {
	if !fetcher.ShouldAutoFetch(feed, globalAutoFetch) || link == "" {
		return content
	}

	result := fetcherext.FetchFullContentWithRandomUA(fetcher.FetchOptions{
		URL:     link,
		Timeout: timeout,
	})
	if result.Error == nil && result.Content != "" {
		return result.Content
	}
	if result.Error != nil {
		logger.Debug("failed to fetch full content", "feed_id", feed.ID, "item_link", link, "error", result.Error)
	}
	return content
}
