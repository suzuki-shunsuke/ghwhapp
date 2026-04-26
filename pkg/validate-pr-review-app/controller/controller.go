package controller

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/shurcooL/githubv4"
	"github.com/suzuki-shunsuke/ghwhapp/pkg/config"
	"github.com/suzuki-shunsuke/ghwhapp/pkg/github"
	"github.com/suzuki-shunsuke/ghwhapp/pkg/validate-pr-review-app/validation"
)

type Controller struct {
	input           *InputNew
	gh              GitHub
	validator       Validator
	webhookVerifier WebhookVerifier
}

func New(input *InputNew) (*Controller, error) {
	// Create GitHub client
	gh, err := github.New(&github.ParamNewApp{
		AppID:          input.Config.AppID,
		InstallationID: input.Config.InstallationID,
		KeyFile:        input.GitHubAppPrivateKey,
		Logger:         input.Logger,
	})
	if err != nil {
		return nil, fmt.Errorf("create GitHub client: %w", err)
	}
	return &Controller{
		input:     input,
		gh:        gh,
		validator: validation.New(&validation.InputNew{}),
		webhookVerifier: github.NewWebhookVerifier(&github.ParamNewWebhookVerifier{
			ValidateSignature: github.ValidateSignature,
			WebhookSecret:     input.WebhookSecret,
		}),
	}, nil
}

type WebhookVerifier interface {
	Verify(logger *slog.Logger, req *github.Request) *github.Event
}

type InputNew struct {
	Config              *config.Config
	Version             string
	WebhookSecret       []byte
	GitHubAppPrivateKey string
	Logger              *slog.Logger
}

type Validator interface {
	Run(logger *slog.Logger, input *validation.Input) *validation.Result
}

type GitHub interface {
	GetPR(ctx context.Context, owner, name string, number int) (*github.PullRequest, error)
	CreateCheckRun(ctx context.Context, input githubv4.CreateCheckRunInput) error
	CompareCommits(ctx context.Context, owner, repo, base, head string) ([]string, error)
	IsAncestor(ctx context.Context, owner, repo, ancestor, descendant string) (bool, error)
}
