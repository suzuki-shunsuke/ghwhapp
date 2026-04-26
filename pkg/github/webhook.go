package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"path"
	"strconv"
	"strings"

	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

var (
	errHeaderXHubSignatureIsRequired           = errors.New("header X-HUB-SIGNATURE is required")
	errGHReadonlyQueueMissingPR                = errors.New("gh-readonly-queue branch is not a valid format. missing 'pr-'")
	errGHReadonlyQueueMissingDashAfterPRNumber = errors.New("gh-readonly-queue branch is not a valid format. missing '-' after the PR number")
)

const (
	headerXGitHubHookInstallationTargetID = "X-GITHUB-HOOK-INSTALLATION-TARGET-ID"
	headerXHubSignature                   = "X-HUB-SIGNATURE"
	headerXGitHubEvent                    = "X-GITHUB-EVENT"

	EventTypePullRequestReview = "pull_request_review"
	EventTypePullRequest       = "pull_request"
	EventTypeInstallation      = "installation"
	EventTypeCheckSuite        = "check_suite"
)

type Request struct {
	// Generate template > Method request passthrough
	Body      string            `json:"body"`
	Headers   map[string]string `json:"header"`
	RequestID string            `json:"requestid"`
}

type WebhookVerifier struct {
	validateSignature func(signature string, payload, secretToken []byte) error
	webhookSecret     []byte
}

type ParamNewWebhookVerifier struct {
	ValidateSignature func(signature string, payload, secretToken []byte) error
	WebhookSecret     []byte
}

func NewWebhookVerifier(param *ParamNewWebhookVerifier) *WebhookVerifier {
	return &WebhookVerifier{
		validateSignature: param.ValidateSignature,
		webhookSecret:     param.WebhookSecret,
	}
}

func (c *WebhookVerifier) Verify(logger *slog.Logger, req *Request) *Event { //nolint:cyclop
	headers := c.normalizeHeaders(req.Headers)
	body := []byte(req.Body)
	if err := c.verifySignature(body, headers); err != nil {
		slogerr.WithError(logger, err).Warn("validate the webhook signature")
		return nil
	}

	evType, ok := headers[headerXGitHubEvent]
	if !ok {
		logger.Warn("header X-GITHUB-EVENT is required")
		return nil
	}
	switch evType {
	case EventTypePullRequestReview:
		payload := &PullRequestReviewEvent{}
		if err := json.Unmarshal(body, payload); err != nil {
			logger.Warn("parse a webhook payload", "error", err)
			return nil
		}
		return newPullRequestReviewEvent(payload)
	case EventTypePullRequest:
		payload := &PullRequestEvent{}
		if err := json.Unmarshal(body, payload); err != nil {
			logger.Warn("parse a webhook payload", "error", err)
			return nil
		}
		return newPullRequestEvent(payload)
	case EventTypeCheckSuite:
		payload := &CheckSuiteEvent{}
		if err := json.Unmarshal(body, payload); err != nil {
			logger.Warn("parse a webhook payload", "error", err)
			return nil
		}
		ev, err := newCheckSuiteEvent(logger, payload)
		if err != nil {
			slogerr.WithError(logger, err).Warn("create event from check suite event")
		}
		return ev
	case EventTypeInstallation:
		logger.Info("ignore the event", "event_type", evType)
		return nil
	default:
		logger.Warn("ignore the event", "event_type", evType)
		return nil
	}
}

func (c *WebhookVerifier) verifySignature(body []byte, headers map[string]string) error {
	sig, ok := headers[headerXHubSignature]
	if !ok {
		return errHeaderXHubSignatureIsRequired
	}
	return c.validateSignature(sig, body, c.webhookSecret)
}

func (c *WebhookVerifier) normalizeHeaders(headers map[string]string) map[string]string {
	hs := make(map[string]string, len(headers))
	for k, v := range headers {
		hs[strings.ToUpper(k)] = v
	}
	return hs
}

type Event struct {
	EventType    string
	Action       string
	RepoFullName string
	RepoOwner    string
	RepoName     string
	PRNumber     int
	ReviewState  string
	RepoID       string
	HeadSHA      string
}

func newPullRequestReviewEvent(ev *PullRequestReviewEvent) *Event {
	return &Event{
		EventType:    EventTypePullRequestReview,
		Action:       ev.GetAction(),
		RepoFullName: ev.GetRepo().GetFullName(),
		RepoOwner:    ev.GetRepo().GetOwner().GetLogin(),
		RepoName:     ev.GetRepo().GetName(),
		PRNumber:     ev.GetPullRequest().GetNumber(),
		ReviewState:  ev.GetReview().GetState(),
		RepoID:       ev.GetRepo().GetNodeID(),
		HeadSHA:      ev.GetPullRequest().GetHead().GetSHA(),
	}
}

func newPullRequestEvent(ev *PullRequestEvent) *Event {
	return &Event{
		EventType:    EventTypePullRequest,
		Action:       ev.GetAction(),
		RepoFullName: ev.GetRepo().GetFullName(),
		RepoOwner:    ev.GetRepo().GetOwner().GetLogin(),
		RepoName:     ev.GetRepo().GetName(),
		PRNumber:     ev.GetPullRequest().GetNumber(),
		RepoID:       ev.GetRepo().GetNodeID(),
		HeadSHA:      ev.GetPullRequest().GetHead().GetSHA(),
	}
}

func getPRNumberFromBranch(logger *slog.Logger, branch string) (int, error) {
	branch2, ok := strings.CutPrefix(branch, "gh-readonly-queue/")
	if !ok {
		logger.Debug("the branch is not a gh-readonly-queue", "branch", branch)
		return 0, nil
	}
	// e.g. pr-24-a9d10f59f8c051673f45263c42aca8346614e716
	branch3, ok := strings.CutPrefix(path.Base(branch2), "pr-")
	if !ok {
		logger.Debug("the branch is not a gh-readonly-queue", "branch", branch)
		return 0, errGHReadonlyQueueMissingPR
	}

	s, _, ok := strings.Cut(branch3, "-")
	if !ok {
		return 0, errGHReadonlyQueueMissingDashAfterPRNumber
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("parse pull request number in gh-readonly-queue branch as number: %w", err)
	}
	return n, nil
}

func newCheckSuiteEvent(logger *slog.Logger, ev *CheckSuiteEvent) (*Event, error) {
	// e.g. refs/heads/gh-readonly-queue/main/pr-24-a9d10f59f8c051673f45263c42aca8346614e716
	prNumber, err := getPRNumberFromBranch(logger, ev.GetCheckSuite().GetHeadBranch())
	if err != nil {
		return nil, fmt.Errorf("get a pull request number from the branch name: %w", err)
	}
	if prNumber == 0 {
		// Ignore webhook events not from gh-readonly-queue branches
		return nil, nil //nolint:nilnil
	}
	return &Event{
		EventType:    EventTypeCheckSuite,
		Action:       ev.GetAction(),
		RepoFullName: ev.GetRepo().GetFullName(),
		RepoOwner:    ev.GetRepo().GetOwner().GetLogin(),
		RepoName:     ev.GetRepo().GetName(),
		PRNumber:     prNumber,
		RepoID:       ev.GetRepo().GetNodeID(),
		HeadSHA:      ev.GetCheckSuite().GetHeadSHA(),
	}, nil
}
