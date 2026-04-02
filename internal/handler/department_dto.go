package handler

import (
	"time"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"

	"github.com/google/uuid"
)

type listDepartmentsRequest struct {
	httpapi.PageRequest
	Search *string `form:"search"`
}

type createDepartmentRequest struct {
	Name                     string     `json:"name" binding:"required"`
	Description              *string    `json:"description"`
	DepartmentHeadEmployeeID *uuid.UUID `json:"department_head_employee_id"`
}

type updateDepartmentRequest struct {
	Name                     *string    `json:"name"`
	Description              *string    `json:"description"`
	DepartmentHeadEmployeeID *uuid.UUID `json:"department_head_employee_id"`
}

type departmentResponse struct {
	ID                       uuid.UUID  `json:"id"`
	Name                     string     `json:"name"`
	Description              *string    `json:"description"`
	DepartmentHeadEmployeeID *uuid.UUID `json:"department_head_employee_id"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
}

type deleteDepartmentResponse struct {
	ID uuid.UUID `json:"id"`
}

func toCreateDepartmentParams(req createDepartmentRequest) domain.CreateDepartmentParams {
	return domain.CreateDepartmentParams{
		Name:                     req.Name,
		Description:              req.Description,
		DepartmentHeadEmployeeID: req.DepartmentHeadEmployeeID,
	}
}

func toUpdateDepartmentParams(req updateDepartmentRequest) domain.UpdateDepartmentParams {
	return domain.UpdateDepartmentParams{
		Name:                     req.Name,
		Description:              req.Description,
		DepartmentHeadEmployeeID: req.DepartmentHeadEmployeeID,
	}
}

func toListDepartmentsParams(req listDepartmentsRequest) domain.ListDepartmentsParams {
	search := ""
	if req.Search != nil {
		search = *req.Search
	}

	return domain.ListDepartmentsParams{
		Limit:  req.PageSize,
		Offset: (req.Page - 1) * req.PageSize,
		Search: search,
	}
}

func toDepartmentResponse(department *domain.Department) departmentResponse {
	return departmentResponse{
		ID:                       department.ID,
		Name:                     department.Name,
		Description:              department.Description,
		DepartmentHeadEmployeeID: department.DepartmentHeadEmployeeID,
		CreatedAt:                department.CreatedAt,
		UpdatedAt:                department.UpdatedAt,
	}
}

func toDepartmentItemResponse(department domain.Department) departmentResponse {
	return departmentResponse{
		ID:                       department.ID,
		Name:                     department.Name,
		Description:              department.Description,
		DepartmentHeadEmployeeID: department.DepartmentHeadEmployeeID,
		CreatedAt:                department.CreatedAt,
		UpdatedAt:                department.UpdatedAt,
	}
}
