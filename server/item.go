package server

import (
	"context"
	"log/slog"

	"github.com/0x2e/fusion/model"
	"github.com/0x2e/fusion/repo"
	"github.com/0x2e/fusion/service/fetcher"
)

type ItemRepo interface {
	List(filter repo.ItemFilter, page, pageSize int) ([]*model.Item, int, error)
	Get(id uint) (*model.Item, error)
	Delete(id uint) error
	UpdateUnread(ids []uint, unread *bool) error
	UpdateBookmark(id uint, bookmark *bool) error
	UpdateFullContent(id uint, fullContent *string) error
}

type Item struct {
	repo ItemRepo
}

func NewItem(repo ItemRepo) *Item {
	return &Item{
		repo: repo,
	}
}

func (i Item) List(ctx context.Context, req *ReqItemList) (*RespItemList, error) {
	filter := repo.ItemFilter{
		Keyword:  req.Keyword,
		FeedID:   req.FeedID,
		GroupID:  req.GroupID,
		Unread:   req.Unread,
		Bookmark: req.Bookmark,
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	data, total, err := i.repo.List(filter, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	items := make([]*ItemForm, 0, len(data))
	for _, v := range data {
		items = append(items, &ItemForm{
			ID:          v.ID,
			GUID:        v.GUID,
			Title:       v.Title,
			Link:        v.Link,
			FullContent: v.FullContent,
			Unread:      v.Unread,
			Bookmark:    v.Bookmark,
			PubDate:     v.PubDate,
			UpdatedAt:   &v.UpdatedAt,
			Feed: ItemFeed{
				ID:   v.Feed.ID,
				Name: v.Feed.Name,
				Link: v.Feed.Link,
			},
		})
	}
	return &RespItemList{
		Total: &total,
		Items: items,
	}, nil
}

func (i Item) Get(ctx context.Context, req *ReqItemGet) (*RespItemGet, error) {
	data, err := i.repo.Get(req.ID)
	if err != nil {
		return nil, err
	}

	// Default fetch to true if not specified
	shouldFetch := req.Fetch == nil || *req.Fetch

	// Fetch full content if requested and not already available
	if shouldFetch && data.Link != nil && (data.FullContent == nil || *data.FullContent == "") {
		slog.Info("Fetching full content for item", "id", req.ID, "link", *data.Link)
		result := fetcher.FetchFullContent(fetcher.FetchOptions{
			URL: *data.Link,
		})

		if result.Error == nil && result.Content != "" {
			data.FullContent = &result.Content
			// Save to database asynchronously
			go func() {
				if err := i.repo.UpdateFullContent(data.ID, data.FullContent); err != nil {
					slog.Error("Failed to save full content", "id", data.ID, "error", err)
				}
			}()
		} else if result.Error != nil {
			slog.Warn("Failed to fetch full content, will use RSS content", "id", req.ID, "error", result.Error)
		}
	}

	return &RespItemGet{
		ID:          data.ID,
		GUID:        data.GUID,
		Title:       data.Title,
		Link:        data.Link,
		Content:     data.Content,
		FullContent: data.FullContent,
		Unread:      data.Unread,
		Bookmark:    data.Bookmark,
		PubDate:     data.PubDate,
		UpdatedAt:   &data.UpdatedAt,
		Feed: ItemFeed{
			ID:   data.Feed.ID,
			Name: data.Feed.Name,
			Link: data.Feed.Link,
		},
	}, nil
}

func (i Item) Delete(ctx context.Context, req *ReqItemDelete) error {
	return i.repo.Delete(req.ID)
}

func (i Item) UpdateUnread(ctx context.Context, req *ReqItemUpdateUnread) error {
	return i.repo.UpdateUnread(req.IDs, req.Unread)
}

func (i Item) UpdateBookmark(ctx context.Context, req *ReqItemUpdateBookmark) error {
	return i.repo.UpdateBookmark(req.ID, req.Bookmark)
}
