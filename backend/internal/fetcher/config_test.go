package fetcher

import (
	"testing"

	"github.com/0x2E/fusion/internal/model"
)

func boolPtr(b bool) *bool { return &b }

func TestShouldAutoFetchPriority(t *testing.T) {
	tests := []struct {
		name          string
		feed          *bool
		group         *bool
		systemDefault bool
		want          bool
	}{
		{"feed true wins over group false", boolPtr(true), boolPtr(false), false, true},
		{"feed false wins over group true", boolPtr(false), boolPtr(true), true, false},
		{"group true wins over system false", nil, boolPtr(true), false, true},
		{"group false wins over system true", nil, boolPtr(false), true, false},
		{"system default when unset", nil, nil, true, true},
		{"system default off when unset", nil, nil, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feed := &model.Feed{
				AutoFetchFullContent:      tt.feed,
				GroupAutoFetchFullContent: tt.group,
			}
			if got := ShouldAutoFetch(feed, tt.systemDefault); got != tt.want {
				t.Errorf("ShouldAutoFetch() = %v, want %v", got, tt.want)
			}
		})
	}
}
