package handler

import (
	"log/slog"
	"time"

	"github.com/0x2E/fusion/internal/pullext"
	"github.com/gin-gonic/gin"
)

type backfillFullContentRequest struct {
	Days            int   `json:"days"`
	Limit           int   `json:"limit"`
	FeedID          int64 `json:"feed_id"`
	IntervalSeconds int   `json:"interval_seconds"`
}

func (h *Handler) backfillFullContent(c *gin.Context) {
	if pullext.BackfillRunning() {
		c.JSON(409, gin.H{"error": "backfill already running"})
		return
	}

	var req backfillFullContentRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			badRequestError(c, "invalid request body")
			return
		}
	}

	if req.Days < 0 || req.Days > 3650 {
		badRequestError(c, "days must be between 0 and 3650")
		return
	}
	if req.Limit < 0 || req.Limit > 10000 {
		badRequestError(c, "limit must be between 0 and 10000")
		return
	}
	if req.FeedID < 0 {
		badRequestError(c, "feed_id must be non-negative")
		return
	}
	if req.IntervalSeconds < 0 || req.IntervalSeconds > 60 {
		badRequestError(c, "interval_seconds must be between 0 and 60")
		return
	}

	limit := req.Limit
	if limit == 0 {
		limit = 500
	}
	interval := time.Duration(req.IntervalSeconds) * time.Second
	if req.IntervalSeconds == 0 {
		interval = time.Second
	}

	go pullext.BackfillFullContent(h.store, pullext.BackfillParams{
		Days:     req.Days,
		Limit:    limit,
		FeedID:   req.FeedID,
		Interval: interval,
	}, slog.Default())

	c.JSON(202, gin.H{"data": gin.H{"started": true}})
}

func (h *Handler) backfillFullContentStatus(c *gin.Context) {
	dataResponse(c, pullext.BackfillStatusSnapshot())
}
