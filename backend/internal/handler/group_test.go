package handler

import (
	"net/http"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/0x2E/fusion/internal/config"
	"github.com/0x2E/fusion/internal/model"
	"github.com/0x2E/fusion/internal/store"
)

func newGroupTestHandler(t *testing.T) (*Handler, *store.Store) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	st, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}

	cfg := &config.Config{
		Password:       "secret",
		FeverUsername:  "fusion",
		PullTimeout:    30,
		LoginRateLimit: 10,
		LoginWindow:    60,
		LoginBlock:     300,
	}

	h, err := New(st, cfg, noopPuller{})
	if err != nil {
		_ = st.Close()
		t.Fatalf("new handler: %v", err)
	}

	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Errorf("close store: %v", err)
		}
	})

	return h, st
}

func TestCreateGroupWithAutoFetch(t *testing.T) {
	h, st := newGroupTestHandler(t)

	h.sessions["test-session"] = time.Now().Add(time.Minute).Unix()

	r := newTestRouter()
	auth := r.Group("/api")
	auth.Use(h.authMiddleware())
	auth.POST("/groups", h.createGroup)

	body := mustJSONBody(t, map[string]any{
		"name":                    "Test Group",
		"auto_fetch_full_content": true,
	})
	w := performRequest(
		r,
		http.MethodPost,
		"/api/groups",
		body,
		map[string]string{"Content-Type": "application/json"},
		&http.Cookie{Name: "session", Value: "test-session"},
	)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	groups, err := st.ListGroups()
	if err != nil {
		t.Fatalf("list groups: %v", err)
	}
	var created *model.Group
	for _, g := range groups {
		if g.Name == "Test Group" {
			created = g
			break
		}
	}
	if created == nil {
		t.Fatal("created group not found")
	}
	if created.AutoFetchFullContent == nil || !*created.AutoFetchFullContent {
		t.Fatalf("expected auto_fetch_full_content to be true, got %v", created.AutoFetchFullContent)
	}
}

func TestUpdateGroupAutoFetchWithoutName(t *testing.T) {
	h, st := newGroupTestHandler(t)

	group, err := st.CreateGroup("Test Group")
	if err != nil {
		t.Fatalf("create group: %v", err)
	}

	h.sessions["test-session"] = time.Now().Add(time.Minute).Unix()

	r := newTestRouter()
	auth := r.Group("/api")
	auth.Use(h.authMiddleware())
	auth.PATCH("/groups/:id", h.updateGroup)

	body := mustJSONBody(t, map[string]any{
		"auto_fetch_full_content": true,
	})
	w := performRequest(
		r,
		http.MethodPatch,
		"/api/groups/"+strconv.FormatInt(group.ID, 10),
		body,
		map[string]string{"Content-Type": "application/json"},
		&http.Cookie{Name: "session", Value: "test-session"},
	)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	updated, err := st.GetGroup(group.ID)
	if err != nil {
		t.Fatalf("get group: %v", err)
	}
	if updated.AutoFetchFullContent == nil || !*updated.AutoFetchFullContent {
		t.Fatalf("expected auto_fetch_full_content to be true, got %v", updated.AutoFetchFullContent)
	}
	if updated.Name != "Test Group" {
		t.Fatalf("expected name to remain 'Test Group', got %q", updated.Name)
	}
}
