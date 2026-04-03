package service

import (
	"context"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DepartmentService struct {
	repository domain.DepartmentRepository
	logger     domain.Logger
}

func NewDepartmentService(
	repository domain.DepartmentRepository,
	logger domain.Logger,
) domain.DepartmentService {
	return &DepartmentService{
		repository: repository,
		logger:     logger,
	}
}

func (s *DepartmentService) CreateDepartment(
	ctx context.Context,
	params domain.CreateDepartmentParams,
) (*domain.Department, error) {
	department, err := s.repository.CreateDepartment(ctx, params)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"DepartmentService.CreateDepartment",
				"failed to create department",
				err,
				zap.String("name", params.Name),
			)
		}
		return nil, err
	}

	return department, nil
}

func (s *DepartmentService) GetDepartmentByID(
	ctx context.Context,
	departmentID uuid.UUID,
) (*domain.Department, error) {
	department, err := s.repository.GetDepartmentByID(ctx, departmentID)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"DepartmentService.GetDepartmentByID",
				"failed to get department by id",
				err,
				zap.String("department_id", departmentID.String()),
			)
		}
		return nil, err
	}

	return department, nil
}

func (s *DepartmentService) UpdateDepartment(
	ctx context.Context,
	departmentID uuid.UUID,
	params domain.UpdateDepartmentParams,
) (*domain.Department, error) {
	department, err := s.repository.UpdateDepartment(ctx, departmentID, params)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"DepartmentService.UpdateDepartment",
				"failed to update department",
				err,
				zap.String("department_id", departmentID.String()),
			)
		}
		return nil, err
	}

	return department, nil
}

func (s *DepartmentService) DeleteDepartment(ctx context.Context, departmentID uuid.UUID) error {
	if err := s.repository.DeleteDepartment(ctx, departmentID); err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"DepartmentService.DeleteDepartment",
				"failed to delete department",
				err,
				zap.String("department_id", departmentID.String()),
			)
		}
		return err
	}

	return nil
}

func (s *DepartmentService) ListDepartments(
	ctx context.Context,
	params domain.ListDepartmentsParams,
) (*domain.DepartmentPage, error) {
	page, err := s.repository.ListDepartments(ctx, params)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"DepartmentService.ListDepartments",
				"failed to list departments",
				err,
				zap.String("search", params.Search),
				zap.Int32("limit", params.Limit),
				zap.Int32("offset", params.Offset),
			)
		}
		return nil, err
	}

	return page, nil
}

var _ domain.DepartmentService = (*DepartmentService)(nil)
