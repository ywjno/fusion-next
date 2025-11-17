package pull

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/0x2e/fusion/model"
	"github.com/0x2e/fusion/pkg/ptr"
	"github.com/0x2e/fusion/service/fetcher"
	"github.com/0x2e/fusion/service/pull/client"
)

// ReadFeedItemsFn is responsible for reading a feed from an HTTP server and
// converting the result to fusion-native data types. The error return value
// is for request errors (e.g. HTTP errors).
type ReadFeedItemsFn func(ctx context.Context, feedURL string, options model.FeedRequestOptions) (client.FetchItemsResult, error)

// UpdateFeedInStoreFn is responsible for saving the result of a feed fetch to a data
// store. If the fetch failed, it records that in the data store. If the fetch
// succeeds, it stores the latest build time in the data store and adds any new
// feed items to the datastore.
type UpdateFeedInStoreFn func(feedID uint, items []*model.Item, lastBuild *time.Time, requestError error) error

// SingleFeedRepo represents a datastore for storing information about a feed.
type SingleFeedRepo interface {
	InsertItems(items []*model.Item) error
	RecordSuccess(lastBuild *time.Time) error
	RecordFailure(readErr error) error
}

type SingleFeedPuller struct {
	readFeed ReadFeedItemsFn
	repo     SingleFeedRepo
}

// NewSingleFeedPuller creates a new SingleFeedPuller with the given ReadFeedItemsFn and repository.
func NewSingleFeedPuller(readFeed ReadFeedItemsFn, repo SingleFeedRepo) SingleFeedPuller {
	return SingleFeedPuller{
		readFeed: readFeed,
		repo:     repo,
	}
}

// defaultSingleFeedRepo is the default implementation of SingleFeedRepo
type defaultSingleFeedRepo struct {
	feedID          uint
	feedRepo        FeedRepo
	itemRepo        ItemRepo
	systemAutoFetch bool
}

func (r *defaultSingleFeedRepo) InsertItems(items []*model.Item) error {
	// Set the correct feed ID for all items.
	for _, item := range items {
		item.FeedID = r.feedID
	}

	err := r.itemRepo.Insert(items)
	if err != nil {
		return err
	}

	// Auto-fetch full content if enabled
	go r.autoFetchFullContent(items)

	return nil
}

func (r *defaultSingleFeedRepo) autoFetchFullContent(items []*model.Item) {
	feed, err := r.feedRepo.Get(r.feedID)
	if err != nil {
		slog.Error("Failed to get feed for auto-fetch", "feed_id", r.feedID, "error", err)
		return
	}

	if !fetcher.ShouldAutoFetch(feed, r.systemAutoFetch) {
		return
	}

	if len(items) == 0 {
		return
	}

	slog.Info("Auto-fetching full content", "feed_id", r.feedID, "items_count", len(items))

	// Limit concurrency to avoid overwhelming the server
	const maxConcurrency = 3
	routinePool := make(chan struct{}, maxConcurrency)
	defer close(routinePool)

	// Collect results for batch update
	var updatesMutex sync.Mutex
	updates := make(map[uint]string)

	wg := sync.WaitGroup{}
	for _, item := range items {
		// Skip items without link or ID (ID=0 means item already exists and wasn't inserted)
		if item.ID == 0 || item.Link == nil || *item.Link == "" {
			continue
		}

		routinePool <- struct{}{}
		wg.Add(1)

		go func(item *model.Item) {
			defer func() {
				wg.Done()
				<-routinePool
			}()

			result := fetcher.FetchFullContent(fetcher.FetchOptions{
				URL: *item.Link,
			})

			if result.Error != nil {
				slog.Warn("Failed to auto-fetch full content",
					"item_id", item.ID,
					"link", *item.Link,
					"error", result.Error)
				return
			}

			if result.Content != "" {
				updatesMutex.Lock()
				updates[item.ID] = result.Content
				updatesMutex.Unlock()

				slog.Debug("Successfully auto-fetched full content",
					"item_id", item.ID,
					"title", item.Title)
			}
		}(item)
	}

	wg.Wait()

	// Batch update all items at once in a single transaction
	if len(updates) > 0 {
		slog.Info("Batch updating full content", "feed_id", r.feedID, "count", len(updates))
		if err := r.itemRepo.BatchUpdateFullContent(updates); err != nil {
			slog.Error("Failed to batch update full content", "feed_id", r.feedID, "error", err)
		} else {
			slog.Info("Completed auto-fetching full content", "feed_id", r.feedID, "success_count", len(updates))
		}
	} else {
		slog.Info("No full content to update", "feed_id", r.feedID)
	}
}

func (r *defaultSingleFeedRepo) RecordSuccess(lastBuild *time.Time) error {
	return r.feedRepo.Update(r.feedID, &model.Feed{
		LastBuild:           lastBuild,
		Failure:             ptr.To(""),
		ConsecutiveFailures: 0,
	})
}

func (r *defaultSingleFeedRepo) RecordFailure(readErr error) error {
	feed, err := r.feedRepo.Get(r.feedID)
	if err != nil {
		return err
	}

	return r.feedRepo.Update(r.feedID, &model.Feed{
		Failure:             ptr.To(readErr.Error()),
		ConsecutiveFailures: feed.ConsecutiveFailures + 1,
	})
}

func (p SingleFeedPuller) Pull(ctx context.Context, feed *model.Feed) error {
	logger := slog.With("feed_id", feed.ID, "feed_link", ptr.From(feed.Link))

	// We don't exit on error, as we want to record any error in the data store.
	fetchResult, readErr := p.readFeed(ctx, *feed.Link, feed.FeedRequestOptions)
	if readErr == nil {
		logger.Info(fmt.Sprintf("fetched %d items", len(fetchResult.Items)))
	} else {
		logger.Warn("failed to fetch feed", "error", readErr)
	}

	return p.updateFeedInStore(feed.ID, fetchResult.Items, fetchResult.LastBuild, readErr)
}

// updateFeedInStore saves the result of a feed fetch to the data store.
// If the fetch failed, it records that in the data store.
// If the fetch succeeds, it stores the latest build time and adds any new feed items.
func (p SingleFeedPuller) updateFeedInStore(feedID uint, items []*model.Item, lastBuild *time.Time, requestError error) error {
	if requestError != nil {
		return p.repo.RecordFailure(requestError)
	}

	if err := p.repo.InsertItems(items); err != nil {
		return err
	}

	return p.repo.RecordSuccess(lastBuild)
}
