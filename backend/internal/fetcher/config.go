package fetcher

import "github.com/0x2E/fusion/internal/model"

// ShouldAutoFetch determines whether to automatically fetch full content
// based on the priority: Feed > Group > System
func ShouldAutoFetch(feed *model.Feed, systemDefault bool) bool {
	// Priority 1: Feed-level setting
	if feed.AutoFetchFullContent != nil {
		return *feed.AutoFetchFullContent
	}

	// Priority 2: Group-level setting
	// Note: We need to check if feed has group info loaded
	// For now, this requires the caller to ensure feed.Group is populated
	// This is a limitation that should be addressed in the caller

	// Priority 3: System-level setting
	return systemDefault
}
