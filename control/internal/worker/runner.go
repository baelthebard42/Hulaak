package worker

import (
	"context"

	"github.com/baelthebard42/Hulaak/control/internal/outbox"
)

type Runner struct {
	repo *outbox.Repository
}

func NewRunner(repo *outbox.Repository) *Runner {
	return &Runner{repo: repo}
}

func (r *Runner) processJob(ctx context.Context, job outbox.WebhookDetails)
