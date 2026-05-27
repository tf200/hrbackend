package app

import (
	"context"

	"hrbackend/internal/domain"
	pkgasynq "hrbackend/pkg/asynq"

	hibikenasynq "github.com/hibiken/asynq"
)

type taskQueueAdapter struct {
	client *pkgasynq.AsynqClient
}

func (a *taskQueueAdapter) EnqueueEmailDelivery(
	ctx context.Context,
	payload domain.EmailDeliveryTaskPayload,
	opts *domain.TaskEnqueueOptions,
) error {
	return a.client.EnqueueEmailDelivery(
		toEmailDeliveryPayload(payload),
		ctx,
		toAsynqOptions(opts)...)
}

func (a *taskQueueAdapter) Close() error {
	return a.client.Close()
}

func toAsynqOptions(opts *domain.TaskEnqueueOptions) []hibikenasynq.Option {
	if opts == nil {
		return nil
	}

	result := make([]hibikenasynq.Option, 0, 2)
	if opts.Queue != "" {
		result = append(result, hibikenasynq.Queue(opts.Queue))
	}
	if opts.MaxRetry > 0 {
		result = append(result, hibikenasynq.MaxRetry(opts.MaxRetry))
	}
	return result
}

func toEmailDeliveryPayload(payload domain.EmailDeliveryTaskPayload) pkgasynq.EmailDeliveryPayload {
	return pkgasynq.EmailDeliveryPayload{
		To:           payload.To,
		Name:         payload.Name,
		UserEmail:    payload.UserEmail,
		UserPassword: payload.UserPassword,
	}
}
