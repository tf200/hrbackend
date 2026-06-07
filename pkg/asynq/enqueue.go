package asynq

import (
	"context"
	"fmt"
	"log"

	"github.com/goccy/go-json"
	hibikenasynq "github.com/hibiken/asynq"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"

	TypeEmailDelivery          = "email:deliver"
	TypeIncidentProcess        = "incident:process"
	TypeIncidentConfirmedEmail = "incident:confirmed_email"
	TypeNotificationSend       = "notification:send"
)

func (c *AsynqClient) EnqueueEmailDelivery(
	payload EmailDeliveryPayload,
	ctx context.Context,
	opts ...hibikenasynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json.Marshal failed: %v", err)
	}

	task := hibikenasynq.NewTask(TypeEmailDelivery, jsonPayload)
	info, err := c.client.EnqueueContext(ctx, task, opts...)
	if err != nil {
		return fmt.Errorf("client.EnqueueContext failed: %v", err)
	}

	log.Printf("task enqueued: id=%s queue=%s", info.ID, info.Queue)
	return nil
}
