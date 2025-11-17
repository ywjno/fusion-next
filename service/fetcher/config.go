package fetcher

import "github.com/0x2e/fusion/model"

// ShouldAutoFetch determines whether to automatically fetch full content
// based on the priority: Feed > Group > System
func ShouldAutoFetch(feed *model.Feed, systemDefault bool) bool {
	// Priority 1: Feed-level setting
	if feed.AutoFetchFullContent != nil {
		return *feed.AutoFetchFullContent
	}

	// Priority 2: Group-level setting
	if feed.Group.AutoFetchFullContent != nil {
		return *feed.Group.AutoFetchFullContent
	}

	// Priority 3: System-level setting
	return systemDefault
}
