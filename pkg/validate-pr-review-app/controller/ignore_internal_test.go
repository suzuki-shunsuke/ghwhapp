//nolint:funlen
package controller

import (
	"log/slog"
	"testing"

	"github.com/suzuki-shunsuke/ghwhapp/pkg/github"
)

func Test_ignore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		event    *github.Event
		expected bool
	}{
		{
			name: "ignore edited action",
			event: &github.Event{
				Action:      "edited",
				ReviewState: "approved",
			},
			expected: true,
		},
		{
			name: "ignore commented state",
			event: &github.Event{
				Action:      "submitted",
				ReviewState: "commented",
			},
			expected: true,
		},
		{
			name: "ignore pending state",
			event: &github.Event{
				Action:      "submitted",
				ReviewState: "pending",
			},
			expected: true,
		},
		{
			name: "do not ignore approved state",
			event: &github.Event{
				Action:      "submitted",
				ReviewState: "approved",
			},
			expected: false,
		},
		{
			name: "do not ignore changes_requested state",
			event: &github.Event{
				Action:      "submitted",
				ReviewState: "changes_requested",
			},
			expected: false,
		},
		{
			name: "do not ignore dismissed state",
			event: &github.Event{
				Action:      "dismissed",
				ReviewState: "dismissed",
			},
			expected: false,
		},
		{
			name: "handle empty action",
			event: &github.Event{
				Action:      "",
				ReviewState: "approved",
			},
			expected: false,
		},
		{
			name: "handle empty review state",
			event: &github.Event{
				Action:      "submitted",
				ReviewState: "",
			},
			expected: false,
		},
		{
			name: "handle empty fields",
			event: &github.Event{
				Action:      "",
				ReviewState: "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := slog.Default()
			result := ignore(logger, tt.event)

			if result != tt.expected {
				t.Errorf("ignore() = %v, want %v", result, tt.expected)
			}
		})
	}
}
