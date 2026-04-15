package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/0x2E/fusion/internal/fetcher"
	"github.com/0x2E/fusion/internal/fetcherext"
	"github.com/0x2E/fusion/internal/store"
	"github.com/gin-gonic/gin"
)

const maxListLimit = 100
const maxBatchUpdateIDs = 1000

type markItemsReadRequest struct {
	IDs []int64 `json:"ids" binding:"required"`
}

func (h *Handler) listItems(c *gin.Context) {
	params := store.ListItemsParams{}

	if feedID := c.Query("feed_id"); feedID != "" {
		id, err := strconv.ParseInt(feedID, 10, 64)
		if err != nil {
			badRequestError(c, "invalid feed_id")
			return
		}
		params.FeedID = &id
	}

	if groupID := c.Query("group_id"); groupID != "" {
		id, err := strconv.ParseInt(groupID, 10, 64)
		if err != nil {
			badRequestError(c, "invalid group_id")
			return
		}
		params.GroupID = &id
	}

	if unread := c.Query("unread"); unread != "" {
		val, err := strconv.ParseBool(unread)
		if err != nil {
			badRequestError(c, "invalid unread")
			return
		}
		params.Unread = &val
	}

	if limit := c.Query("limit"); limit != "" {
		val, err := strconv.Atoi(limit)
		if err != nil || val <= 0 {
			badRequestError(c, "invalid limit")
			return
		}
		if val > maxListLimit {
			val = maxListLimit
		}
		params.Limit = val
	} else {
		params.Limit = 10
	}

	if offset := c.Query("offset"); offset != "" {
		val, err := strconv.Atoi(offset)
		if err != nil || val < 0 {
			badRequestError(c, "invalid offset")
			return
		}
		params.Offset = val
	}

	if orderBy := c.Query("order_by"); orderBy != "" {
		params.OrderBy = orderBy
	} else {
		params.OrderBy = "pub_date"
	}

	items, err := h.store.ListItems(params)
	if err != nil {
		internalError(c, err, "list items")
		return
	}

	total, err := h.store.CountItems(params)
	if err != nil {
		internalError(c, err, "count items")
		return
	}

	listResponse(c, items, total)
}

func (h *Handler) getItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequestError(c, "invalid id")
		return
	}

	item, err := h.store.GetItem(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			notFoundError(c, "item")
			return
		}
		internalError(c, err, "get item")
		return
	}

	// Default fetch to true if not specified
	shouldFetch := true
	if fetch := c.Query("fetch"); fetch != "" {
		shouldFetch, err = strconv.ParseBool(fetch)
		if err != nil {
			badRequestError(c, "invalid fetch")
			return
		}
	}

	// Fetch full content if requested and not already available
	if shouldFetch && item.Link != "" && item.FullContent == "" {
		slog.Info("Fetching full content for item", "id", id, "link", item.Link)
		result := fetcherext.FetchFullContentWithRandomUA(fetcher.FetchOptions{
			URL: item.Link,
		})

		if result.Error == nil && result.Content != "" {
			item.FullContent = result.Content
			item.Content = result.Content
			go func() {
				if err := h.store.UpdateFullContent(item.ID, item.FullContent); err != nil {
					slog.Error("Failed to save full content", "id", item.ID, "error", err)
				}
			}()
		} else if result.Error != nil {
			slog.Warn("Failed to fetch full content, will use RSS content", "id", id, "error", result.Error)
		}
	} else if item.FullContent != "" {
		item.Content = item.FullContent
	}

	dataResponse(c, item)
}

func (h *Handler) markItemsRead(c *gin.Context) {
	var req markItemsReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestError(c, "invalid request")
		return
	}
	if len(req.IDs) == 0 || len(req.IDs) > maxBatchUpdateIDs {
		badRequestError(c, "invalid ids")
		return
	}

	if err := h.store.BatchUpdateItemsUnread(req.IDs, false); err != nil {
		internalError(c, err, "mark items as read")
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) markItemsUnread(c *gin.Context) {
	var req markItemsReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestError(c, "invalid request")
		return
	}
	if len(req.IDs) == 0 || len(req.IDs) > maxBatchUpdateIDs {
		badRequestError(c, "invalid ids")
		return
	}

	if err := h.store.BatchUpdateItemsUnread(req.IDs, true); err != nil {
		internalError(c, err, "mark items as unread")
		return
	}

	c.Status(http.StatusNoContent)
}
