package repository

import (
	"context"
	"encoding/json"
	"time"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type NotificationRepository struct {
	queries db.Querier
}

func NewNotificationRepository(queries db.Querier) domain.NotificationRepository {
	return &NotificationRepository{queries: queries}
}

func (r *NotificationRepository) CreateNotifications(
	ctx context.Context,
	params domain.CreateNotificationsParams,
) ([]domain.Notification, error) {
	rows, err := r.queries.CreateNotifications(ctx, db.CreateNotificationsParams{
		UserIds: params.UserIDs,
		Type:    params.Type,
		Message: params.Message,
		Data:    params.Data,
		CreatedAt: pgtype.Timestamptz{
			Time:  params.CreatedAt,
			Valid: true,
		},
	})
	if err != nil {
		return nil, err
	}

	items := make([]domain.Notification, 0, len(rows))
	for _, row := range rows {
		n, err := notificationFromDB(row)
		if err != nil {
			return nil, err
		}
		items = append(items, *n)
	}
	return items, nil
}

func (r *NotificationRepository) ListUserIDsByEmployeeIDs(
	ctx context.Context,
	employeeIDs []uuid.UUID,
) ([]uuid.UUID, error) {
	return r.queries.ListNotificationUserIDsByEmployeeIDs(ctx, employeeIDs)
}

func (r *NotificationRepository) ListUserIDsByRoles(
	ctx context.Context,
	roleNames []string,
) ([]uuid.UUID, error) {
	return r.queries.ListNotificationUserIDsByRoles(ctx, roleNames)
}

func (r *NotificationRepository) ListUserIDsByPermissions(
	ctx context.Context,
	permissionNames []string,
) ([]uuid.UUID, error) {
	return r.queries.ListNotificationUserIDsByPermissions(ctx, permissionNames)
}

func notificationFromDB(row db.Notification) (*domain.Notification, error) {
	data, err := UnmarshalNotificationData(row.Data)
	if err != nil {
		return nil, err
	}

	n := &domain.Notification{
		ID:        row.ID,
		UserID:    row.UserID,
		Type:      row.Type,
		Message:   row.Message,
		IsRead:    row.IsRead,
		Data:      data,
		ReadAt:    timestamptzPtr(row.ReadAt),
		CreatedAt: conv.TimeFromPgTimestamptz(row.CreatedAt),
	}
	return n, nil
}

func timestamptzPtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func MarshalNotificationData(data domain.NotificationData) ([]byte, error) {
	return json.Marshal(data)
}

func UnmarshalNotificationData(data []byte) (domain.NotificationData, error) {
	if len(data) == 0 {
		return domain.NotificationData{}, nil
	}
	var out domain.NotificationData
	if err := json.Unmarshal(data, &out); err != nil {
		return domain.NotificationData{}, err
	}
	return out, nil
}
