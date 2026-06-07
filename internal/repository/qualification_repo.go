package repository

import (
	"context"
	"errors"

	"hrbackend/internal/domain"
	"hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type QualificationRepository struct {
	store *db.Store
}

func NewQualificationRepository(store *db.Store) domain.QualificationTypeRepository {
	return &QualificationRepository{store: store}
}

func (r *QualificationRepository) ListQualificationTypes(
	ctx context.Context,
) ([]domain.QualificationType, error) {
	rows, err := r.store.ListQualificationTypes(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]domain.QualificationType, 0, len(rows))
	for _, row := range rows {
		result = append(result, toDomainQualificationType(row))
	}

	return result, nil
}

func (r *QualificationRepository) CreateQualificationType(
	ctx context.Context,
	params domain.CreateQualificationTypeParams,
) (*domain.QualificationType, error) {
	row, err := r.store.CreateQualificationType(ctx, db.CreateQualificationTypeParams{
		Code: params.Code,
		Name: params.Name,
	})
	if err != nil {
		return nil, err
	}

	result := toDomainQualificationType(row)
	return &result, nil
}

func (r *QualificationRepository) UpdateQualificationType(
	ctx context.Context,
	id uuid.UUID,
	params domain.CreateQualificationTypeParams,
) (*domain.QualificationType, error) {
	row, err := r.store.UpdateQualificationType(ctx, db.UpdateQualificationTypeParams{
		ID:   id,
		Code: params.Code,
		Name: params.Name,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrQualificationTypeNotFound
		}
		return nil, err
	}

	result := toDomainQualificationType(row)
	return &result, nil
}

func toDomainQualificationType(row db.Qualification) domain.QualificationType {
	return domain.QualificationType{
		ID:         row.ID,
		Code:       row.Code,
		Name:       row.Name,
		AppContext: row.AppContext,
		IsActive:   row.IsActive,
		CreatedAt:  conv.TimeFromPgTimestamptz(row.CreatedAt),
		UpdatedAt:  conv.TimeFromPgTimestamptz(row.UpdatedAt),
	}
}
