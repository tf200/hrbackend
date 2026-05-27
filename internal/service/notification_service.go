package service

import (
	"context"
	"sync"
	"time"

	"hrbackend/internal/domain"
	"hrbackend/internal/repository"
	"hrbackend/internal/ws"

	"github.com/google/uuid"
)

type NotificationService struct {
	repository domain.NotificationRepository
	sender     domain.RealtimeSender
	logger     domain.Logger
}

func NewNotificationService(
	repository domain.NotificationRepository,
	sender domain.RealtimeSender,
	logger domain.Logger,
) domain.NotificationService {
	return &NotificationService{
		repository: repository,
		sender:     sender,
		logger:     logger,
	}
}

func (s *NotificationService) Notify(ctx context.Context, req domain.NotificationRequest) {
	go s.worker(context.Background(), req)
}

func (s *NotificationService) worker(ctx context.Context, req domain.NotificationRequest) {
	workerCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	userIDs, err := s.resolveRecipients(workerCtx, req.Recipients)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(workerCtx, "NotificationService.worker", "failed to resolve recipients", err)
		}
		return
	}
	if len(userIDs) == 0 {
		return
	}

	now := req.CreatedAt
	if now == nil {
		t := time.Now()
		now = &t
	}

	dataJSON, err := repository.MarshalNotificationData(req.Data)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(workerCtx, "NotificationService.worker", "failed to marshal notification data", err)
		}
		return
	}

	notifications, err := s.repository.CreateNotifications(workerCtx, domain.CreateNotificationsParams{
		UserIDs:   userIDs,
		Type:      req.Type,
		Message:   req.Message,
		Data:      dataJSON,
		CreatedAt: *now,
	})
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(workerCtx, "NotificationService.worker", "failed to create notifications", err)
		}
		return
	}

	var wg sync.WaitGroup
	for i := range notifications {
		wg.Add(1)
		n := notifications[i]
		go s.deliver(workerCtx, n, &wg)
	}
	wg.Wait()
}

func (s *NotificationService) deliver(ctx context.Context, n domain.Notification, wg *sync.WaitGroup) {
	defer wg.Done()

	payload, err := ws.MarshalEvent(ws.EventNotificationCreated, n)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(ctx, "NotificationService.deliver", "failed to marshal ws event", err)
		}
		return
	}

	s.sender.SendToUser(n.UserID, payload)
}

func (s *NotificationService) resolveRecipients(
	ctx context.Context,
	recipients domain.NotificationRecipients,
) ([]uuid.UUID, error) {
	seen := make(map[uuid.UUID]struct{})
	var result []uuid.UUID

	addIfAbsent := func(id uuid.UUID) {
		if id == uuid.Nil {
			return
		}
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			result = append(result, id)
		}
	}

	for _, id := range recipients.UserIDs {
		addIfAbsent(id)
	}

	if len(recipients.EmployeeIDs) > 0 {
		ids, err := s.repository.ListUserIDsByEmployeeIDs(ctx, recipients.EmployeeIDs)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			addIfAbsent(id)
		}
	}

	if len(recipients.Roles) > 0 {
		ids, err := s.repository.ListUserIDsByRoles(ctx, recipients.Roles)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			addIfAbsent(id)
		}
	}

	if len(recipients.Permissions) > 0 {
		ids, err := s.repository.ListUserIDsByPermissions(ctx, recipients.Permissions)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			addIfAbsent(id)
		}
	}

	return result, nil
}
