package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"
)

type EmployeeSignatureProfileRepository struct {
	store *db.Store
}

func NewEmployeeSignatureProfileRepository(
	store *db.Store,
) domain.EmployeeSignatureProfileRepository {
	return &EmployeeSignatureProfileRepository{store: store}
}

func (r *EmployeeSignatureProfileRepository) GetByEmployeeID(
	ctx context.Context,
	employeeID uuid.UUID,
) (*domain.EmployeeSignatureProfile, error) {
	row, err := r.store.GetEmployeeSignatureProfileByEmployeeID(ctx, employeeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEmployeeSignatureProfileNotFound
		}
		return nil, err
	}
	model := toDomainEmployeeSignatureProfile(row)
	return &model, nil
}

func (r *EmployeeSignatureProfileRepository) Upsert(
	ctx context.Context,
	employeeID uuid.UUID,
	params domain.UpsertEmployeeSignatureProfileParams,
) (*domain.EmployeeSignatureProfile, error) {
	row, err := r.store.UpsertEmployeeSignatureProfile(
		ctx,
		db.UpsertEmployeeSignatureProfileParams{
			EmployeeID:   employeeID,
			Type:         db.EmployeeSignatureTypeEnum(params.Type),
			TypedName:    params.TypedName,
			ImageFileKey: params.ImageFileKey,
		},
	)
	if err != nil {
		return nil, err
	}
	model := toDomainEmployeeSignatureProfile(row)
	return &model, nil
}

func (r *EmployeeSignatureProfileRepository) DeleteByEmployeeID(
	ctx context.Context,
	employeeID uuid.UUID,
) error {
	return r.store.DeleteEmployeeSignatureProfileByEmployeeID(ctx, employeeID)
}
