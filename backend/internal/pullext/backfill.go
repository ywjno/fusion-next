package pullext

import (
	"log/slog"
	"sync"
	"time"

	"github.com/0x2E/fusion/internal/fetcher"
	"github.com/0x2E/fusion/internal/fetcherext"
	"github.com/0x2E/fusion/internal/store"
)

type BackfillParams struct {
	Days     int
	Limit    int
	FeedID   int64
	Interval time.Duration
}

type BackfillStatus struct {
	Running bool `json:"running"`
	Done    int  `json:"done"`
	Failed  int  `json:"failed"`
	Skipped int  `json:"skipped"`
}

var backfillState struct {
	mu     sync.Mutex
	status BackfillStatus
}

// BackfillRunning reports whether a full-content backfill is in progress.
func BackfillRunning() bool {
	backfillState.mu.Lock()
	defer backfillState.mu.Unlock()
	return backfillState.status.Running
}

// BackfillStatusSnapshot returns a copy of the current backfill status.
func BackfillStatusSnapshot() BackfillStatus {
	backfillState.mu.Lock()
	defer backfillState.mu.Unlock()
	return backfillState.status
}

func setBackfillStatus(status BackfillStatus) {
	backfillState.mu.Lock()
	defer backfillState.mu.Unlock()
	backfillState.status = status
}

// BackfillFullContent fetches full content for all candidate items and stores
// it via UpdateFullContent. It runs synchronously; the caller is expected to
// launch it in a goroutine. A second call while one is running is a no-op.
// Failed items are not tracked anywhere, so the next run retries them.
func BackfillFullContent(s *store.Store, params BackfillParams, logger *slog.Logger) {
	backfillState.mu.Lock()
	if backfillState.status.Running {
		backfillState.mu.Unlock()
		return
	}
	backfillState.status = BackfillStatus{Running: true}
	backfillState.mu.Unlock()

	status := BackfillStatus{Running: true}
	defer func() {
		status.Running = false
		setBackfillStatus(status)
	}()

	candidates, err := s.ListFullContentBackfillCandidates(params.Days, params.Limit, params.FeedID)
	if err != nil {
		logger.Error("full content backfill: failed to list candidates", "error", err)
		return
	}
	if len(candidates) == 0 {
		logger.Info("full content backfill: no candidates found")
		return
	}

	logger.Info("full content backfill: started", "candidates", len(candidates))

	for i, candidate := range candidates {
		result := fetcherext.FetchFullContentWithRandomUA(fetcher.FetchOptions{
			URL:     candidate.Link,
			Timeout: 30 * time.Second,
		})
		switch {
		case result.Error == nil && result.Content != "":
			if err := s.UpdateFullContent(candidate.ID, result.Content); err != nil {
				logger.Warn("full content backfill: failed to save content", "item_id", candidate.ID, "error", err)
				status.Failed++
			} else {
				status.Done++
			}
		case result.Error == nil:
			status.Skipped++
		default:
			status.Failed++
		}

		setBackfillStatus(status)

		if (i+1)%25 == 0 {
			logger.Info("full content backfill: progress",
				"processed", i+1,
				"total", len(candidates),
				"done", status.Done,
				"failed", status.Failed,
				"skipped", status.Skipped,
			)
		}

		if params.Interval > 0 && i < len(candidates)-1 {
			time.Sleep(params.Interval)
		}
	}

	logger.Info("full content backfill: completed",
		"total", len(candidates),
		"done", status.Done,
		"failed", status.Failed,
		"skipped", status.Skipped,
	)
}
