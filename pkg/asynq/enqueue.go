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

	TypeEmailDelivery                = "email:deliver"
	TypeIncidentProcess              = "incident:process"
	TypeIncidentConfirmedEmail       = "incident:confirmed_email"
	TypeNotificationSend             = "notification:send"
	TypeAcceptedRegistration         = "accepted:registration"
	TypeProcessRegistrationFormEmail = "email:process_registration_form"
)

func (c *AsynqClient) EnqueueProcessRegistrationFormEmail(
	ctx context.Context,
	payload ProcessRegistrationFormEmailPayload,
	opts ...hibikenasynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("EnqueueProcessRegistrationFormEmail: json.Marshal failed: %w", err)
	}

	if len(opts) == 0 {
		opts = append(opts, hibikenasynq.Queue(QueueDefault), hibikenasynq.MaxRetry(5))
	}

	task := hibikenasynq.NewTask(TypeProcessRegistrationFormEmail, jsonPayload)
	info, err := c.client.EnqueueContext(ctx, task, opts...)
	if err != nil {
		return fmt.Errorf("EnqueueProcessRegistrationFormEmail: client.EnqueueContext failed: %w", err)
	}

	log.Printf("Process Registration Form Email task enqueued: id=%s queue=%s", info.ID, info.Queue)
	return nil
}

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

func (c *AsynqClient) EnqueueIncident(
	payload IncidentPayload,
	ctx context.Context,
	opts ...hibikenasynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json.Marshal failed: %v", err)
	}

	task := hibikenasynq.NewTask(TypeIncidentProcess, jsonPayload)
	info, err := c.client.EnqueueContext(ctx, task, opts...)
	if err != nil {
		return fmt.Errorf("client.EnqueueContext failed: %v", err)
	}

	log.Printf("task enqueued: id=%s queue=%s", info.ID, info.Queue)
	return nil
}

func (c *AsynqClient) EnqueueIncidentConfirmedEmail(
	ctx context.Context,
	payload IncidentConfirmedEmailPayload,
	opts ...hibikenasynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("EnqueueIncidentConfirmedEmail: json.Marshal failed: %w", err)
	}

	if len(opts) == 0 {
		opts = append(opts, hibikenasynq.Queue(QueueDefault), hibikenasynq.MaxRetry(5))
	}

	task := hibikenasynq.NewTask(TypeIncidentConfirmedEmail, jsonPayload)
	info, err := c.client.EnqueueContext(ctx, task, opts...)
	if err != nil {
		return fmt.Errorf("EnqueueIncidentConfirmedEmail: client.EnqueueContext failed: %w", err)
	}

	log.Printf("Incident confirmed email task enqueued: id=%s queue=%s", info.ID, info.Queue)
	return nil
}

func (c *AsynqClient) EnqueueNotificationTask(
	ctx context.Context,
	payload NotificationPayload,
	opts ...hibikenasynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("EnqueueNotificationTask: json.Marshal failed: %w", err)
	}

	if len(opts) == 0 {
		opts = append(opts, hibikenasynq.Queue(QueueDefault), hibikenasynq.MaxRetry(5))
	}

	task := hibikenasynq.NewTask(TypeNotificationSend, jsonPayload)
	info, err := c.client.EnqueueContext(ctx, task, opts...)
	if err != nil {
		return fmt.Errorf("EnqueueNotificationTask: client.EnqueueContext failed: %w", err)
	}

	log.Printf("Notification task enqueued: id=%s queue=%s", info.ID, info.Queue)
	return nil
}

func (c *AsynqClient) EnqueueAcceptedRegistration(
	ctx context.Context,
	payload AcceptedRegistrationFormPayload,
	opts ...hibikenasynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("EnqueueAcceptedRegistration: json.Marshal failed: %w", err)
	}

	if len(opts) == 0 {
		opts = append(opts, hibikenasynq.Queue(QueueDefault), hibikenasynq.MaxRetry(5))
	}

	task := hibikenasynq.NewTask(TypeAcceptedRegistration, jsonPayload)
	info, err := c.client.EnqueueContext(ctx, task, opts...)
	if err != nil {
		return fmt.Errorf("EnqueueAcceptedRegistration: client.EnqueueContext failed: %w", err)
	}

	log.Printf("Accepted Registration task enqueued: id=%s queue=%s", info.ID, info.Queue)
	return nil
}
