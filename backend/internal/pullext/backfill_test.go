package pullext

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/0x2E/fusion/internal/store"
)

func newTestStore(t *testing.T) *store.Store {
	t.Helper()

	s, err := store.New(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("failed to create test store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func waitForCondition(t *testing.T, timeout time.Duration, condition func() bool, message string) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for: %s", message)
}

// TestBackfillFullContentSingletonGuard verifies that a second call is a no-op
// while a backfill is running, using a hanging local server as the fetch
// target so the first run stays in progress.
func TestBackfillFullContentSingletonGuard(t *testing.T) {
	s := newTestStore(t)

	release := make(chan struct{})
	hanging := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-release
		http.Error(w, "gone", http.StatusInternalServerError)
	}))
	defer hanging.Close()

	group, err := s.CreateGroup("test")
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err)
	}
	feed, err := s.CreateFeed(group.ID, "feed", "http://example.com/feed", "", "", false)
	if err != nil {
		t.Fatalf("CreateFeed() failed: %v", err)
	}
	if _, err := s.CreateItem(feed.ID, "guid-1", "item", hanging.URL, "summary", time.Now().Unix()); err != nil {
		t.Fatalf("CreateItem() failed: %v", err)
	}

	go BackfillFullContent(s, BackfillParams{Interval: time.Millisecond}, slog.Default())

	waitForCondition(t, 5*time.Second, BackfillRunning, "backfill to start")

	// A second call while the first is still running must return immediately
	// as a no-op and must not reset the running status.
	BackfillFullContent(s, BackfillParams{Interval: time.Millisecond}, slog.Default())

	status := BackfillStatusSnapshot()
	if !status.Running {
		t.Error("expected backfill to still be running after second call")
	}
	if status.Done != 0 {
		t.Errorf("expected done=0 while first run is in progress, got %d", status.Done)
	}

	// Unblock the in-flight fetch so the first run fails the item and
	// completes.
	close(release)
	waitForCondition(t, 10*time.Second, func() bool {
		return !BackfillRunning()
	}, "backfill to finish")
}

// TestBackfillFullContentNoCandidates verifies that a run over an empty
// candidate list completes with zero counts.
func TestBackfillFullContentNoCandidates(t *testing.T) {
	s := newTestStore(t)

	BackfillFullContent(s, BackfillParams{Interval: time.Millisecond}, slog.Default())

	status := BackfillStatusSnapshot()
	if status.Running {
		t.Error("expected backfill to be finished")
	}
	if status.Done != 0 || status.Failed != 0 || status.Skipped != 0 {
		t.Errorf("expected zero counts, got %+v", status)
	}
}
