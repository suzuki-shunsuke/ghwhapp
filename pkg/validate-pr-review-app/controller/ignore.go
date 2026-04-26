package controller

import (
	"log/slog"

	"github.com/suzuki-shunsuke/ghwhapp/pkg/github"
)

func ignore(logger *slog.Logger, ev *github.Event) bool {
	// For pull_request events, only process "synchronize" action.
	if ev.EventType == github.EventTypePullRequest {
		if ev.Action != "synchronize" {
			logger.Debug("ignore the pull_request event because the action is not 'synchronize'", "action", ev.Action)
			return true
		}
		return false
	}
	if ev.Action == "edited" {
		logger.Info("ignore the event because the action is 'edited'")
		return true
	}
	state := ev.ReviewState
	if state == "commented" || state == "pending" {
		logger.Info("ignore the event because the state is '" + state + "'")
		return true
	}
	return false
}
