package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type EmployeeRepository struct {
	store *db.Store
}

func NewEmployeeRepository(store *db.Store) domain.EmployeeRepository {
	return &EmployeeRepository{store: store}
}

func (r *EmployeeRepository) WithTx(
	ctx context.Context,
	fn func(tx domain.EmployeeTxRepository) error,
) error {
	return r.store.ExecTx(ctx, func(q *db.Queries) error {
		return fn(&employeeTxRepo{queries: q})
	})
}

type employeeTxRepo struct {
	queries *db.Queries
}

func (tx *employeeTxRepo) CreateUser(
	ctx context.Context,
	email, password string,
) (uuid.UUID, error) {
	user, err := tx.queries.CreateUser(ctx, db.CreateUserParams{
		Email:    email,
		Password: password,
		IsActive: true,
	})
	if err != nil {
		return uuid.Nil, err
	}
	return user.ID, nil
}

func (tx *employeeTxRepo) CreateEmployeeProfile(
	ctx context.Context,
	userID uuid.UUID,
	params domain.CreateEmployeeParams,
) (uuid.UUID, error) {
	empProfile, err := tx.queries.CreateEmployeeProfile(ctx, db.CreateEmployeeProfileParams{
		UserID:              userID,
		FirstName:           params.FirstName,
		LastName:            params.LastName,
		Bsn:                 params.Bsn,
		Street:              params.Street,
		HouseNumber:         params.HouseNumber,
		HouseNumberAddition: params.HouseNumberAddition,
		PostalCode:          params.PostalCode,
		City:                params.City,
		ManagerEmployeeID:   params.ManagerEmployeeID,
		EmployeeNumber:      params.EmployeeNumber,
		EmploymentNumber:    params.EmploymentNumber,
		PrivateEmailAddress: params.PrivateEmailAddress,
		WorkEmailAddress:    params.WorkEmailAddress,
		WorkPhoneNumber:     params.WorkPhoneNumber,
		PrivatePhoneNumber:  params.PrivatePhoneNumber,
		DateOfBirth:         pgDateFromPtr(params.DateOfBirth),
		HomeTelephoneNumber: params.HomeTelephoneNumber,
		Gender:              genderEnumFromString(params.Gender),
	})
	if err != nil {
		return uuid.Nil, err
	}
	return empProfile.ID, nil
}

func (tx *employeeTxRepo) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return tx.queries.AssignRoleToUser(ctx, db.AssignRoleToUserParams{
		UserID: userID,
		RoleID: roleID,
	})
}

func (tx *employeeTxRepo) AddEmployeeContractDetails(
	ctx context.Context,
	employeeID uuid.UUID,
	params domain.CreateEmployeeContractParams,
) (uuid.UUID, error) {
	contract, err := tx.queries.AddEmployeeContractDetails(ctx, db.AddEmployeeContractDetailsParams{
		EmployeeID:           employeeID,
		JobTitle:             employeeJobTitleEnumFromString(params.JobTitle),
		DepartmentID:         params.DepartmentID,
		LocationID:           params.LocationID,
		ContractType:         contractTypeFromString(params.ContractType),
		StartDate:            conv.PgDateFromTime(params.StartDate),
		ContractEndDate:      pgDateFromPtr(params.ContractEndDate),
		HoursPerWeek:         params.HoursPerWeek,
		RosterFreeDay:        weekdayEnumFromString(params.RosterFreeDay),
		WageTaxTable:         wageTaxTablePtrFromStringPtr(params.WageTaxTable),
		CreatedByEmployeeID:  nil,
	})
	if err != nil {
		return uuid.Nil, err
	}
	return contract.ID, nil
}

func (tx *employeeTxRepo) CreateEmployeeSalaryAssignment(
	ctx context.Context,
	employeeID uuid.UUID,
	contractID *uuid.UUID,
	params domain.CreateEmployeeSalaryAssignmentParams,
) (uuid.UUID, error) {
	salary, err := tx.queries.CreateEmployeeSalaryAssignment(
		ctx,
		db.CreateEmployeeSalaryAssignmentParams{
			EmployeeID:          employeeID,
			ContractID:          contractID,
			SalaryScaleStepID:   params.SalaryScaleStepID,
			EffectiveFrom:       pgDateFromPtr(params.EffectiveFrom),
			EffectiveTo:         pgDateFromPtr(params.EffectiveTo),
			CreatedByEmployeeID: nil,
		},
	)
	if err != nil {
		return uuid.Nil, err
	}
	return salary.ID, nil
}

func (tx *employeeTxRepo) GetActiveEmployeeSalaryAssignment(
	ctx context.Context,
	employeeID uuid.UUID,
	contractID *uuid.UUID,
	targetDate time.Time,
) (*domain.EmployeeSalaryAssignmentInfo, error) {
	row, err := tx.queries.GetActiveEmployeeSalaryAssignment(
		ctx,
		db.GetActiveEmployeeSalaryAssignmentParams{
			EmployeeID: employeeID,
			ContractID: contractID,
			TargetDate: conv.PgDateFromTime(targetDate),
		},
	)
	if err != nil {
		if isDBNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &domain.EmployeeSalaryAssignmentInfo{
		ID:                row.ID,
		ContractID:        row.ContractID,
		SalaryScaleStepID: row.SalaryScaleStepID,
		EffectiveFrom:     row.EffectiveFrom.Time,
		EffectiveTo:       conv.TimePtrFromPgDate(row.EffectiveTo),
	}, nil
}

func (tx *employeeTxRepo) EndEmployeeSalaryAssignment(
	ctx context.Context,
	assignmentID uuid.UUID,
	effectiveTo time.Time,
) error {
	return tx.queries.EndEmployeeSalaryAssignment(
		ctx,
		db.EndEmployeeSalaryAssignmentParams{
			ID:          assignmentID,
			EffectiveTo: conv.PgDateFromTime(effectiveTo),
		},
	)
}

func (tx *employeeTxRepo) GetEmployeeByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.EmployeeDetail, error) {
	row, err := tx.queries.GetEmployeeProfileByID(ctx, id)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrEmployeeNotFound
		}
		return nil, err
	}

	stats, err := tx.queries.GetEmployeeDetailStats(ctx, id)
	if err != nil {
		return nil, err
	}

	employee := toDomainEmployeeDetailFromGetEmployeeProfileByIDRow(row)
	if contractRow, err := tx.queries.GetActiveEmployeeContractDetail(ctx, id); err == nil {
		mapActiveContractScalars(employee, contractRow)
	} else if !isDBNotFound(err) {
		return nil, err
	}
	if salaryRow, err := tx.queries.GetLatestEmployeeSalaryAssignmentDetail(ctx, id); err == nil {
		applyEmployeeSalaryAssignmentDetail(employee, salaryRow)
	} else if !isDBNotFound(err) {
		return nil, err
	}
	applyEmployeeDetailStats(employee, stats)

	return employee, nil
}

func (tx *employeeTxRepo) LinkEmployeeAttachments(
	ctx context.Context,
	employeeID uuid.UUID,
	attachmentIDs []uuid.UUID,
	category string,
) error {
	return tx.queries.CreateEmployeeAttachments(ctx, db.CreateEmployeeAttachmentsParams{
		EmployeeID:    employeeID,
		AttachmentIds: attachmentIDs,
		Category:      category,
	})
}

func (tx *employeeTxRepo) UpdateAttachmentsUsed(
	ctx context.Context,
	ids []uuid.UUID,
	isUsed bool,
) error {
	return tx.queries.UpdateAttachmentsUsed(ctx, db.UpdateAttachmentsUsedParams{
		AttachmentIds: ids,
		IsUsed:        isUsed,
	})
}

func (tx *employeeTxRepo) AddEmployeeQualificationsBatch(
	ctx context.Context,
	employeeID uuid.UUID,
	params []domain.CreateQualificationParams,
) error {
	if len(params) == 0 {
		return nil
	}
	arg := make([]db.AddEmployeeQualificationsBatchParams, len(params))
	for i, p := range params {
		arg[i] = db.AddEmployeeQualificationsBatchParams{
			EmployeeID:        employeeID,
			QualificationID:   p.QualificationID,
			AchievedOn:        conv.PgDateFromTime(p.AchievedOn),
			ExpirationDate:    pgDateFromPtr(p.ExpirationDate),
			CertificateNumber: p.CertificateNumber,
		}
	}
	_, err := tx.queries.AddEmployeeQualificationsBatch(ctx, arg)
	return err
}

func (tx *employeeTxRepo) AddEmployeeAuthorizationsBatch(
	ctx context.Context,
	employeeID uuid.UUID,
	params []domain.CreateEmployeeAuthorizationParams,
) error {
	if len(params) == 0 {
		return nil
	}
	arg := make([]db.AddEmployeeAuthorizationsBatchParams, len(params))
	for i, p := range params {
		isActive := !p.ExpiryDate.Before(time.Now().UTC().Truncate(24 * time.Hour))
		arg[i] = db.AddEmployeeAuthorizationsBatchParams{
			EmployeeID:      employeeID,
			AuthorizationID: p.AuthorizationID,
			GrantedDate:     conv.PgDateFromTime(p.GrantedDate),
			ExpiryDate:      conv.PgDateFromTime(p.ExpiryDate),
			IsActive:        isActive,
			Notes:           p.Notes,
		}
	}
	_, err := tx.queries.AddEmployeeAuthorizationsBatch(ctx, arg)
	return err
}

func (r *EmployeeRepository) GetEmployeeByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.EmployeeDetail, error) {
	row, err := r.store.GetEmployeeProfileByID(ctx, id)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrEmployeeNotFound
		}
		return nil, err
	}

	stats, err := r.store.GetEmployeeDetailStats(ctx, id)
	if err != nil {
		return nil, err
	}

	employee := toDomainEmployeeDetailFromGetEmployeeProfileByIDRow(row)
	if contractRow, err := r.store.GetActiveEmployeeContractDetail(ctx, id); err == nil {
		mapActiveContractScalars(employee, contractRow)
	} else if !isDBNotFound(err) {
		return nil, err
	}
	if salaryRow, err := r.store.GetLatestEmployeeSalaryAssignmentDetail(ctx, id); err == nil {
		applyEmployeeSalaryAssignmentDetail(employee, salaryRow)
	} else if !isDBNotFound(err) {
		return nil, err
	}
	applyEmployeeDetailStats(employee, stats)

	return employee, nil
}

func (r *EmployeeRepository) GetEmployeeByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (*domain.EmployeeProfile, error) {
	row, err := r.store.GetEmployeeProfileByUserID(ctx, userID)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrEmployeeNotFound
		}
		return nil, err
	}

	return toDomainEmployeeProfile(row)
}

func (r *EmployeeRepository) GetEmployeeProfileDetails(
	ctx context.Context,
	userID uuid.UUID,
) (*domain.EmployeeProfileDetails, error) {
	row, err := r.store.GetEmployeeProfileDetailsByUserID(ctx, userID)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrEmployeeNotFound
		}
		return nil, err
	}
	return toDomainEmployeeProfileDetails(row)
}

func (r *EmployeeRepository) ListActiveSessions(
	ctx context.Context,
	userID uuid.UUID,
) ([]domain.EmployeeProfileActiveSession, error) {
	rows, err := r.store.ListActiveSessionsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	sessions := make([]domain.EmployeeProfileActiveSession, len(rows))
	for i, row := range rows {
		sessions[i] = domain.EmployeeProfileActiveSession{
			ID:        row.ID,
			UserAgent: row.UserAgent,
			ClientIP:  row.ClientIp,
			ExpiresAt: conv.TimeFromPgTimestamptz(row.ExpiresAt),
			CreatedAt: conv.TimeFromPgTimestamptz(row.CreatedAt),
		}
	}
	return sessions, nil
}

func (r *EmployeeRepository) ListEmployees(
	ctx context.Context,
	params domain.ListEmployeesParams,
) (*domain.EmployeePage, error) {
	rows, err := r.store.ListEmployeeProfile(ctx, db.ListEmployeeProfileParams{
		Limit:               params.Limit,
		Offset:              params.Offset,
		IncludeArchived:     params.IncludeArchived,
		IncludeOutOfService: params.IncludeOutOfService,
		LocationID:          params.LocationID,
		ContractType:        contractTypePtrFromStringPtr(params.ContractType),
		Search:              params.Search,
	})
	if err != nil {
		return nil, err
	}

	totalCount, err := r.CountEmployees(ctx, params)
	if err != nil {
		return nil, err
	}

	page := &domain.EmployeePage{
		Items:      make([]domain.Employee, 0, len(rows)),
		TotalCount: totalCount,
	}

	for _, row := range rows {
		page.Items = append(page.Items, toDomainEmployee(row))
	}

	return page, nil
}

func (r *EmployeeRepository) CountEmployees(
	ctx context.Context,
	params domain.ListEmployeesParams,
) (int64, error) {
	return r.store.CountEmployeeProfile(ctx, db.CountEmployeeProfileParams{
		IncludeArchived:     params.IncludeArchived,
		IncludeOutOfService: params.IncludeOutOfService,
		LocationID:          params.LocationID,
		ContractType:        contractTypePtrFromStringPtr(params.ContractType),
		Search:              params.Search,
	})
}

func (r *EmployeeRepository) CreateEmployee(
	ctx context.Context,
	params domain.CreateEmployeeParams,
) (*domain.EmployeeDetail, error) {
	var emp *domain.EmployeeDetail
	err := r.WithTx(ctx, func(tx domain.EmployeeTxRepository) error {
		userID, err := tx.CreateUser(ctx, params.UserEmail, params.UserPassword)
		if err != nil {
			return err
		}

		empID, err := tx.CreateEmployeeProfile(ctx, userID, params)
		if err != nil {
			return err
		}

		err = tx.AssignRoleToUser(ctx, userID, params.RoleID)
		if err != nil {
			return err
		}

		var contractID *uuid.UUID
		if params.Contract != nil {
			cid, err := tx.AddEmployeeContractDetails(ctx, empID, *params.Contract)
			if err != nil {
				return err
			}
			contractID = &cid
		}

		if params.SalaryAssignment != nil {
			salaryParams := *params.SalaryAssignment
			if salaryParams.EffectiveFrom == nil && params.Contract != nil {
				salaryParams.EffectiveFrom = &params.Contract.StartDate
			}
			_, err = tx.CreateEmployeeSalaryAssignment(ctx, empID, contractID, salaryParams)
			if err != nil {
				return err
			}
		}

		emp, err = tx.GetEmployeeByID(ctx, empID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return emp, nil
}

func (r *EmployeeRepository) UpdateEmployee(
	ctx context.Context,
	id uuid.UUID,
	params domain.UpdateEmployeeParams,
) (*domain.EmployeeDetail, error) {
	var emp *domain.EmployeeDetail
	err := r.store.ExecTx(ctx, func(q *db.Queries) error {
		row, err := q.UpdateEmployeeProfile(ctx, db.UpdateEmployeeProfileParams{
			FirstName:           params.FirstName,
			LastName:            params.LastName,
			Bsn:                 params.Bsn,
			Street:              params.Street,
			HouseNumber:         params.HouseNumber,
			HouseNumberAddition: params.HouseNumberAddition,
			PostalCode:          params.PostalCode,
			City:                params.City,
			ManagerEmployeeID:   params.ManagerEmployeeID,
			EmployeeNumber:      params.EmployeeNumber,
			EmploymentNumber:    params.EmploymentNumber,
			PrivateEmailAddress: params.PrivateEmailAddress,
			WorkEmailAddress:    params.WorkEmailAddress,
			PrivatePhoneNumber:  params.PrivatePhoneNumber,
			WorkPhoneNumber:     params.WorkPhoneNumber,
			DateOfBirth:         pgDateFromPtr(params.DateOfBirth),
			HomeTelephoneNumber: params.HomeTelephoneNumber,
			Gender:              genderEnumPtrFromStringPtr(params.Gender),
			OutOfService:        params.OutOfService,
			IsArchived:          params.IsArchived,
			ID:                  id,
		})
		if err != nil {
			if isDBNotFound(err) {
				return domain.ErrEmployeeNotFound
			}
			return err
		}

		if params.WorkEmailAddress != nil {
			if err := q.UpdateUserEmail(ctx, db.UpdateUserEmailParams{
				ID:    row.UserID,
				Email: params.WorkEmailAddress,
			}); err != nil {
				return err
			}
		}

		if params.RoleID != nil {
			if err := q.AssignRoleToUser(ctx, db.AssignRoleToUserParams{
				UserID: row.UserID,
				RoleID: *params.RoleID,
			}); err != nil {
				return err
			}
		}

		if params.SalaryAssignment != nil {
			var contractID *uuid.UUID
			if params.SalaryAssignment.EffectiveFrom != nil {
				contract, err := q.GetEmployeeContractAtDate(
					ctx,
					db.GetEmployeeContractAtDateParams{
						EmployeeID: id,
						TargetDate: pgDateFromPtr(params.SalaryAssignment.EffectiveFrom),
					},
				)
				if err != nil && !isDBNotFound(err) {
					return err
				}
				if err == nil {
					contractID = &contract.ID
				}
			}

			if _, err := q.CreateEmployeeSalaryAssignment(ctx, db.CreateEmployeeSalaryAssignmentParams{
				EmployeeID:          id,
				ContractID:          contractID,
				SalaryScaleStepID:   params.SalaryAssignment.SalaryScaleStepID,
				EffectiveFrom:       pgDateFromPtr(params.SalaryAssignment.EffectiveFrom),
				EffectiveTo:         pgDateFromPtr(params.SalaryAssignment.EffectiveTo),
				CreatedByEmployeeID: nil,
			}); err != nil {
				return err
			}
		}

		emp, err = (&employeeTxRepo{queries: q}).GetEmployeeByID(ctx, id)
		return err
	})
	if err != nil {
		return nil, err
	}

	return emp, nil
}

func (r *EmployeeRepository) GetEmployeeCounts(
	ctx context.Context,
) (*domain.EmployeeCounts, error) {
	row, err := r.store.GetEmployeeCounts(ctx)
	if err != nil {
		return nil, err
	}

	return toDomainEmployeeCounts(row), nil
}

func (r *EmployeeRepository) SearchEmployeesByNameOrEmail(
	ctx context.Context,
	search *string,
) ([]domain.EmployeeSearchResult, error) {
	rows, err := r.store.SearchEmployeesByNameOrEmail(ctx, search)
	if err != nil {
		return nil, err
	}

	result := make([]domain.EmployeeSearchResult, 0, len(rows))
	for _, row := range rows {
		result = append(result, toDomainEmployeeSearchResult(row))
	}

	return result, nil
}

func (r *EmployeeRepository) ListEducation(
	ctx context.Context,
	employeeID uuid.UUID,
) ([]domain.Education, error) {
	rows, err := r.store.ListEducations(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Education, 0, len(rows))
	for _, row := range rows {
		result = append(result, toDomainEducation(row))
	}

	return result, nil
}

func (r *EmployeeRepository) AddEducation(
	ctx context.Context,
	employeeID uuid.UUID,
	params domain.CreateEducationParams,
) (*domain.Education, error) {
	row, err := r.store.AddEducationToEmployeeProfile(ctx, db.AddEducationToEmployeeProfileParams{
		EmployeeID:      employeeID,
		InstitutionName: params.InstitutionName,
		Degree:          params.Degree,
		FieldOfStudy:    params.FieldOfStudy,
		StartDate:       conv.PgDateFromTime(params.StartDate),
		EndDate:         conv.PgDateFromTime(params.EndDate),
	})
	if err != nil {
		return nil, err
	}

	result := toDomainEducation(row)
	return &result, nil
}

func (r *EmployeeRepository) UpdateEducation(
	ctx context.Context,
	id uuid.UUID,
	params domain.UpdateEducationParams,
) (*domain.Education, error) {
	row, err := r.store.UpdateEmployeeEducation(ctx, db.UpdateEmployeeEducationParams{
		ID:              id,
		InstitutionName: params.InstitutionName,
		Degree:          params.Degree,
		FieldOfStudy:    params.FieldOfStudy,
		StartDate:       pgDateFromPtr(params.StartDate),
		EndDate:         pgDateFromPtr(params.EndDate),
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrEducationNotFound
		}
		return nil, err
	}

	result := toDomainEducation(row)
	return &result, nil
}

func (r *EmployeeRepository) DeleteEducation(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Education, error) {
	row, err := r.store.DeleteEmployeeEducation(ctx, id)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrEducationNotFound
		}
		return nil, err
	}

	result := toDomainEducation(row)
	return &result, nil
}

func (r *EmployeeRepository) ListExperience(
	ctx context.Context,
	employeeID uuid.UUID,
) ([]domain.Experience, error) {
	rows, err := r.store.ListEmployeeExperience(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Experience, 0, len(rows))
	for _, row := range rows {
		result = append(result, toDomainExperience(row))
	}

	return result, nil
}

func (r *EmployeeRepository) AddExperience(
	ctx context.Context,
	employeeID uuid.UUID,
	params domain.CreateExperienceParams,
) (*domain.Experience, error) {
	row, err := r.store.AddEmployeeExperience(ctx, db.AddEmployeeExperienceParams{
		EmployeeID:  employeeID,
		JobTitle:    params.JobTitle,
		CompanyName: params.CompanyName,
		StartDate:   conv.PgDateFromTime(params.StartDate),
		EndDate:     conv.PgDateFromTime(params.EndDate),
		Description: params.Description,
	})
	if err != nil {
		return nil, err
	}

	result := toDomainExperience(row)
	return &result, nil
}

func (r *EmployeeRepository) UpdateExperience(
	ctx context.Context,
	id uuid.UUID,
	params domain.UpdateExperienceParams,
) (*domain.Experience, error) {
	row, err := r.store.UpdateEmployeeExperience(ctx, db.UpdateEmployeeExperienceParams{
		ID:          id,
		JobTitle:    params.JobTitle,
		CompanyName: params.CompanyName,
		StartDate:   pgDateFromPtr(params.StartDate),
		EndDate:     pgDateFromPtr(params.EndDate),
		Description: params.Description,
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrExperienceNotFound
		}
		return nil, err
	}

	result := toDomainExperience(row)
	return &result, nil
}

func (r *EmployeeRepository) DeleteExperience(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Experience, error) {
	row, err := r.store.DeleteEmployeeExperience(ctx, id)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrExperienceNotFound
		}
		return nil, err
	}

	result := toDomainExperience(row)
	return &result, nil
}

func (r *EmployeeRepository) ListQualifications(
	ctx context.Context,
	employeeID uuid.UUID,
) ([]domain.Qualification, error) {
	rows, err := r.store.ListEmployeeQualifications(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Qualification, 0, len(rows))
	for _, row := range rows {
		result = append(result, toDomainQualification(row))
	}

	return result, nil
}

func (r *EmployeeRepository) AddQualifications(
	ctx context.Context,
	employeeID uuid.UUID,
	params []domain.CreateQualificationParams,
) (int, error) {
	if len(params) == 0 {
		return 0, nil
	}

	arg := make([]db.AddEmployeeQualificationsBatchParams, len(params))
	for i, p := range params {
		arg[i] = db.AddEmployeeQualificationsBatchParams{
			EmployeeID:        employeeID,
			QualificationID:   p.QualificationID,
			AchievedOn:        conv.PgDateFromTime(p.AchievedOn),
			ExpirationDate:    pgDateFromPtr(p.ExpirationDate),
			CertificateNumber: p.CertificateNumber,
		}
	}

	n, err := r.store.AddEmployeeQualificationsBatch(ctx, arg)
	if err != nil {
		return 0, err
	}

	return int(n), nil
}

func (r *EmployeeRepository) UpdateQualification(
	ctx context.Context,
	id uuid.UUID,
	params domain.UpdateQualificationParams,
) (*domain.Qualification, error) {
	row, err := r.store.UpdateEmployeeQualification(ctx, db.UpdateEmployeeQualificationParams{
		ID:                id,
		QualificationID:   params.QualificationID,
		AchievedOn:        pgDateFromPtr(params.AchievedOn),
		ExpirationDate:    pgDateFromPtr(params.ExpirationDate),
		CertificateNumber: params.CertificateNumber,
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrQualificationNotFound
		}
		return nil, err
	}

	result := toDomainQualification(row)
	return &result, nil
}

func (r *EmployeeRepository) DeleteQualification(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Qualification, error) {
	row, err := r.store.DeleteEmployeeQualification(ctx, id)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrQualificationNotFound
		}
		return nil, err
	}

	result := toDomainQualification(row)
	return &result, nil
}

func (r *EmployeeRepository) ListEmployeeAuthorizations(
	ctx context.Context,
	employeeID uuid.UUID,
) ([]domain.EmployeeAuthorization, error) {
	rows, err := r.store.ListEmployeeAuthorizations(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.EmployeeAuthorization, 0, len(rows))
	for _, row := range rows {
		result = append(result, toDomainEmployeeAuthorization(row))
	}

	return result, nil
}

func (r *EmployeeRepository) ListEmployeeAttachments(
	ctx context.Context,
	employeeID uuid.UUID,
) ([]domain.EmployeeAttachmentDetail, error) {
	rows, err := r.store.ListEmployeeAttachments(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.EmployeeAttachmentDetail, 0, len(rows))
	for _, row := range rows {
		result = append(result, toDomainEmployeeAttachmentDetail(row))
	}

	return result, nil
}

func (r *EmployeeRepository) AddEmployeeAuthorizations(
	ctx context.Context,
	employeeID uuid.UUID,
	params []domain.CreateEmployeeAuthorizationParams,
) (int, error) {
	if len(params) == 0 {
		return 0, nil
	}

	arg := make([]db.AddEmployeeAuthorizationsBatchParams, len(params))
	for i, p := range params {
		isActive := !p.ExpiryDate.Before(time.Now().UTC().Truncate(24 * time.Hour))
		arg[i] = db.AddEmployeeAuthorizationsBatchParams{
			EmployeeID:      employeeID,
			AuthorizationID: p.AuthorizationID,
			GrantedDate:     conv.PgDateFromTime(p.GrantedDate),
			ExpiryDate:      conv.PgDateFromTime(p.ExpiryDate),
			IsActive:        isActive,
			Notes:           p.Notes,
		}
	}

	n, err := r.store.AddEmployeeAuthorizationsBatch(ctx, arg)
	if err != nil {
		return 0, err
	}

	return int(n), nil
}

func (r *EmployeeRepository) UpdateEmployeeAuthorization(
	ctx context.Context,
	id uuid.UUID,
	params domain.UpdateEmployeeAuthorizationParams,
) (*domain.EmployeeAuthorization, error) {
	var grantedDate pgtype.Date
	if params.GrantedDate != nil {
		grantedDate = conv.PgDateFromTime(*params.GrantedDate)
	}
	var expiryDate pgtype.Date
	if params.ExpiryDate != nil {
		expiryDate = conv.PgDateFromTime(*params.ExpiryDate)
	}

	row, err := r.store.UpdateEmployeeAuthorization(ctx, db.UpdateEmployeeAuthorizationParams{
		ID:              id,
		AuthorizationID: params.AuthorizationID,
		GrantedDate:     grantedDate,
		ExpiryDate:      expiryDate,
		IsActive:        params.IsActive,
		Notes:           params.Notes,
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrEmployeeAuthorizationNotFound
		}
		return nil, err
	}

	result := toDomainEmployeeAuthorization(row)
	return &result, nil
}

func (r *EmployeeRepository) DeleteEmployeeAuthorization(
	ctx context.Context,
	id uuid.UUID,
) (*domain.EmployeeAuthorization, error) {
	row, err := r.store.DeleteEmployeeAuthorization(ctx, id)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrEmployeeAuthorizationNotFound
		}
		return nil, err
	}

	result := toDomainEmployeeAuthorization(row)
	return &result, nil
}

func toDomainEmployee(row db.ListEmployeeProfileRow) domain.Employee {
	return domain.Employee{
		ID:              row.ID,
		FirstName:       row.FirstName,
		LastName:        row.LastName,
		Bsn:             row.Bsn,
		ContractType:    contractTypePtrToString(row.ContractType),
		DepartmentName:  row.DepartmentName,
		ContractEndDate: conv.TimePtrFromPgDate(row.ContractEndDate),
		LocationAddress: row.LocationAddress,
		LeaveStatus:     row.LeaveStatus,
	}
}

func toDomainEmployeeDetailFromGetEmployeeProfileByIDRow(
	row db.GetEmployeeProfileByIDRow,
) *domain.EmployeeDetail {
	employee := &domain.EmployeeDetail{
		ID:                  row.ID,
		UserID:              row.UserID,
		FirstName:           row.FirstName,
		LastName:            row.LastName,
		NameInUse:           string(row.NameInUse),
		MaritalStatus:       maritalStatusEnumPtrToStringPtr(row.MaritalStatus),
		Bsn:                 row.Bsn,
		Street:              row.Street,
		HouseNumber:         row.HouseNumber,
		HouseNumberAddition: row.HouseNumberAddition,
		PostalCode:          row.PostalCode,
		City:                row.City,
		EmployeeNumber:      row.EmployeeNumber,
		EmploymentNumber:    row.EmploymentNumber,
		PrivateEmailAddress: row.PrivateEmailAddress,
		WorkEmailAddress:    row.WorkEmailAddress,
		PrivatePhoneNumber:  row.PrivatePhoneNumber,
		WorkPhoneNumber:     row.WorkPhoneNumber,
		DateOfBirth:         conv.TimePtrFromPgDate(row.DateOfBirth),
		HomeTelephoneNumber: row.HomeTelephoneNumber,
		CreatedAt:           conv.TimeFromPgTimestamptz(row.CreatedAt),
		Gender:              string(row.Gender),
		ManagerEmployeeID:   row.ManagerEmployeeID,
		OutOfService:        row.OutOfService,
		IsArchived:          row.IsArchived,
		ProfilePicture:      row.ProfilePicture,
		DepartmentName:      row.DepartmentName,
		ManagerFirstName:    row.ManagerFirstName,
		ManagerLastName:     row.ManagerLastName,
	}

	return employee
}

func mapActiveContractScalars(
	employee *domain.EmployeeDetail,
	row db.GetActiveEmployeeContractDetailRow,
) {
	employee.LocationID = &row.LocationID
	employee.DepartmentID = &row.DepartmentID
	employee.ContractType = string(row.ContractType)
	employee.ContractStartDate = conv.TimePtrFromPgDate(row.StartDate)
	employee.ContractEndDate = conv.TimePtrFromPgDate(row.ContractEndDate)
	employee.ContractHours = row.HoursPerWeek
}

func toDomainContractDetail(row db.ListEmployeeContractDetailsRow) domain.EmployeeContractDetail {
	return domain.EmployeeContractDetail{
		ID:                     row.ID,
		JobTitle:               string(row.JobTitle),
		DepartmentID:           row.DepartmentID,
		DepartmentName:         &row.DepartmentName,
		LocationID:             row.LocationID,
		LocationAddress:        &row.LocationAddress,
		ContractType:           string(row.ContractType),
		StartDate:              conv.TimeFromPgDate(row.StartDate),
		ContractEndDate:        conv.TimePtrFromPgDate(row.ContractEndDate),
		EffectiveEndDate:       conv.TimePtrFromPgDate(row.EffectiveEndDate),
		PreviousContractID:     row.PreviousContractID,
		ContractEventType:      string(row.ContractEventType),
		IsActive: isContractActiveNow(
			row.StartDate,
			row.EffectiveEndDate,
			row.ContractEndDate,
		),
		HoursPerWeek:  row.HoursPerWeek,
		RosterFreeDay: string(row.RosterFreeDay),
		WageTaxTable:  wageTaxTablePtrToStringPtr(row.WageTaxTable),
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}
}

func toDomainContractDetailFromRow(row db.EmployeeContract) domain.EmployeeContractDetail {
	return domain.EmployeeContractDetail{
		ID:                   row.ID,
		JobTitle:             string(row.JobTitle),
		DepartmentID:         row.DepartmentID,
		LocationID:           row.LocationID,
		ContractType:         string(row.ContractType),
		StartDate:            conv.TimeFromPgDate(row.StartDate),
		ContractEndDate:      conv.TimePtrFromPgDate(row.ContractEndDate),
		EffectiveEndDate:     conv.TimePtrFromPgDate(row.EffectiveEndDate),
		PreviousContractID:   row.PreviousContractID,
		ContractEventType:    string(row.ContractEventType),
		IsActive: isContractActiveNow(
			row.StartDate,
			row.EffectiveEndDate,
			row.ContractEndDate,
		),
		HoursPerWeek:  row.HoursPerWeek,
		RosterFreeDay: string(row.RosterFreeDay),
		WageTaxTable:  wageTaxTablePtrToStringPtr(row.WageTaxTable),
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}
}

func isContractActiveNow(
	startDate pgtype.Date,
	effectiveEndDate pgtype.Date,
	contractEndDate pgtype.Date,
) bool {
	now := time.Now()
	if !startDate.Valid || startDate.Time.After(now) {
		return false
	}
	if effectiveEndDate.Valid && effectiveEndDate.Time.Before(now) {
		return false
	}
	if contractEndDate.Valid && contractEndDate.Time.Before(now) {
		return false
	}
	return true
}

func applyEmployeeSalaryAssignmentDetail(
	employee *domain.EmployeeDetail,
	row db.GetLatestEmployeeSalaryAssignmentDetailRow,
) {
	employee.ContractRate = &row.HourlyRate
	employee.SalaryAssignment = &domain.EmployeeSalaryAssignmentDetail{
		ID:                row.ID,
		ContractID:        row.ContractID,
		SalaryScaleStepID: row.SalaryScaleStepID,
		CAOCode:           row.CaoCode,
		SalaryTableName:   row.SalaryTableName,
		Scale:             row.Scale,
		Step:              row.Step,
		IPNumber:          row.IpNumber,
		MonthlySalary:     row.MonthlySalary,
		HourlyRate:        row.HourlyRate,
		EffectiveFrom:     conv.TimeFromPgDate(row.EffectiveFrom),
		EffectiveTo:       conv.TimePtrFromPgDate(row.EffectiveTo),
		CreatedAt:         conv.TimeFromPgTimestamptz(row.CreatedAt),
		UpdatedAt:         conv.TimeFromPgTimestamptz(row.UpdatedAt),
	}
}

func toDomainSalaryAssignmentDetail(
	row db.GetEmployeeSalaryAssignmentDetailByIDRow,
) *domain.EmployeeSalaryAssignmentDetail {
	return &domain.EmployeeSalaryAssignmentDetail{
		ID:                row.ID,
		ContractID:        row.ContractID,
		SalaryScaleStepID: row.SalaryScaleStepID,
		CAOCode:           row.CaoCode,
		SalaryTableName:   row.SalaryTableName,
		Scale:             row.Scale,
		Step:              row.Step,
		IPNumber:          row.IpNumber,
		MonthlySalary:     row.MonthlySalary,
		HourlyRate:        row.HourlyRate,
		EffectiveFrom:     conv.TimeFromPgDate(row.EffectiveFrom),
		EffectiveTo:       conv.TimePtrFromPgDate(row.EffectiveTo),
		CreatedAt:         conv.TimeFromPgTimestamptz(row.CreatedAt),
		UpdatedAt:         conv.TimeFromPgTimestamptz(row.UpdatedAt),
	}
}

func applyEmployeeDetailStats(
	employee *domain.EmployeeDetail,
	stats db.GetEmployeeDetailStatsRow,
) {
	employee.RemainingLeaveBalanceMinutes = stats.RemainingLeaveBalanceMinutes
	employee.HoursWorkedThisMonth = stats.HoursWorkedThisMonth
	employee.HoursPendingApproval = stats.HoursPendingApproval
	employee.TotalHoursWorkedThisYear = stats.TotalHoursWorkedThisYear
	employee.LastPerformanceReviewScore = stats.LastPerformanceReviewScore
}

func toDomainEmployeeDetailFromEmployeeProfile(row db.EmployeeProfile) *domain.EmployeeDetail {
	return &domain.EmployeeDetail{
		ID:                  row.ID,
		UserID:              row.UserID,
		FirstName:           row.FirstName,
		LastName:            row.LastName,
		NameInUse:           string(row.NameInUse),
		MaritalStatus:       maritalStatusEnumPtrToStringPtr(row.MaritalStatus),
		Bsn:                 row.Bsn,
		Street:              row.Street,
		HouseNumber:         row.HouseNumber,
		HouseNumberAddition: row.HouseNumberAddition,
		PostalCode:          row.PostalCode,
		City:                row.City,
		EmployeeNumber:      row.EmployeeNumber,
		EmploymentNumber:    row.EmploymentNumber,
		PrivateEmailAddress: row.PrivateEmailAddress,
		WorkEmailAddress:    row.WorkEmailAddress,
		PrivatePhoneNumber:  row.PrivatePhoneNumber,
		WorkPhoneNumber:     row.WorkPhoneNumber,
		DateOfBirth:         conv.TimePtrFromPgDate(row.DateOfBirth),
		HomeTelephoneNumber: row.HomeTelephoneNumber,
		CreatedAt:           conv.TimeFromPgTimestamptz(row.CreatedAt),
		Gender:              string(row.Gender),
		ManagerEmployeeID:   row.ManagerEmployeeID,
		OutOfService:        row.OutOfService,
		IsArchived:          row.IsArchived,
	}
}

func toDomainEmployeeProfile(
	row db.GetEmployeeProfileByUserIDRow,
) (*domain.EmployeeProfile, error) {
	permissions := make([]domain.Permission, 0)
	if len(row.Permissions) > 0 {
		if err := json.Unmarshal(row.Permissions, &permissions); err != nil {
			return nil, err
		}
	}

	var roleID uuid.UUID
	if row.RoleID != nil {
		roleID = *row.RoleID
	}

	return &domain.EmployeeProfile{
		UserID:           row.UserID,
		Email:            row.Email,
		LastLogin:        conv.TimeFromPgTimestamptz(row.LastLogin),
		TwoFactorEnabled: row.TwoFactorEnabled,
		Role:             row.Role,
		RoleID:           roleID,
		EmployeeID:       row.EmployeeID,
		FirstName:        row.FirstName,
		LastName:         row.LastName,
		Permissions:      permissions,
	}, nil
}

func toDomainEmployeeProfileDetails(
	row db.GetEmployeeProfileDetailsByUserIDRow,
) (*domain.EmployeeProfileDetails, error) {
	roles := make([]domain.EmployeeProfileRole, 0)
	if len(row.Roles) > 0 {
		if err := json.Unmarshal(row.Roles, &roles); err != nil {
			return nil, err
		}
	}

	var contract *domain.EmployeeProfileDetailsContract
	if row.ContractID != nil {
		contract = &domain.EmployeeProfileDetailsContract{
			ID:           row.ContractID,
			Position:     employeeJobTitleEnumPtrToStringPtr(row.Position),
			Department:   row.Department,
			LocationID:   row.LocationID,
			LocationName: row.LocationName,
			Type:         contractTypeEnumPtrToStringPtr(row.ContractType),
			Hours:        row.ContractHours,
			StartDate:    conv.TimePtrFromPgDate(row.ContractStartDate),
			EndDate:      conv.TimePtrFromPgDate(row.ContractEndDate),
			Rate:         row.ContractRate,
		}
	}

	return &domain.EmployeeProfileDetails{
		UserID:              row.UserID,
		EmployeeID:          row.EmployeeID,
		Email:               row.Email,
		TwoFactorEnabled:    row.TwoFactorEnabled,
		LastLogin:           conv.TimeFromPgTimestamptz(row.LastLogin),
		FirstName:           row.FirstName,
		LastName:            row.LastName,
		Roles:               roles,
		Street:              row.Street,
		HouseNumber:         row.HouseNumber,
		HouseNumberAddition: row.HouseNumberAddition,
		PostalCode:          row.PostalCode,
		City:                row.City,
		EmployeeNumber:      row.EmployeeNumber,
		EmploymentNumber:    row.EmploymentNumber,
		PrivateEmailAddress: row.PrivateEmailAddress,
		WorkEmailAddress:    row.WorkEmailAddress,
		PrivatePhoneNumber:  row.PrivatePhoneNumber,
		WorkPhoneNumber:     row.WorkPhoneNumber,
		HomeTelephoneNumber: row.HomeTelephoneNumber,
		DateOfBirth:         conv.TimePtrFromPgDate(row.DateOfBirth),
		Gender:              string(row.Gender),
		OutOfService:        row.OutOfService,
		IsArchived:          row.IsArchived,
		Contract:            contract,
	}, nil
}

func toDomainEmployeeCounts(row db.GetEmployeeCountsRow) *domain.EmployeeCounts {
	return &domain.EmployeeCounts{
		TotalPermanent:    row.TotalPermanent,
		TotalTemporary:    row.TotalTemporary,
		TotalOnCall:       row.TotalOnCall,
		TotalOutOfService: row.TotalOutOfService,
	}
}

func toDomainEmployeeSearchResult(
	row db.SearchEmployeesByNameOrEmailRow,
) domain.EmployeeSearchResult {
	return domain.EmployeeSearchResult{
		ID:               row.ID,
		FirstName:        row.FirstName,
		LastName:         row.LastName,
		WorkEmailAddress: row.WorkEmailAddress,
	}
}

func toDomainEducation(row db.EmployeeEducation) domain.Education {
	return domain.Education{
		ID:              row.ID,
		EmployeeID:      row.EmployeeID,
		InstitutionName: row.InstitutionName,
		Degree:          row.Degree,
		FieldOfStudy:    row.FieldOfStudy,
		StartDate:       conv.TimeFromPgDate(row.StartDate),
		EndDate:         conv.TimeFromPgDate(row.EndDate),
		CreatedAt:       conv.TimeFromPgTimestamptz(row.CreatedAt),
	}
}

func toDomainExperience(row db.EmployeeExperience) domain.Experience {
	return domain.Experience{
		ID:          row.ID,
		EmployeeID:  row.EmployeeID,
		JobTitle:    row.JobTitle,
		CompanyName: row.CompanyName,
		StartDate:   conv.TimeFromPgDate(row.StartDate),
		EndDate:     conv.TimeFromPgDate(row.EndDate),
		Description: row.Description,
		CreatedAt:   conv.TimeFromPgTimestamptz(row.CreatedAt),
	}
}

func toDomainQualification(row db.EmployeeQualification) domain.Qualification {
	return domain.Qualification{
		ID:                row.ID,
		EmployeeID:        row.EmployeeID,
		QualificationID:   row.QualificationID,
		AchievedOn:        conv.TimeFromPgDate(row.AchievedOn),
		ExpirationDate:    conv.TimePtrFromPgDate(row.ExpirationDate),
		CertificateNumber: row.CertificateNumber,
		CreatedAt:         conv.TimeFromPgTimestamptz(row.CreatedAt),
		UpdatedAt:         conv.TimeFromPgTimestamptz(row.UpdatedAt),
	}
}

func toDomainEmployeeAuthorization(row db.EmployeeAuthorization) domain.EmployeeAuthorization {
	return domain.EmployeeAuthorization{
		ID:              row.ID,
		EmployeeID:      row.EmployeeID,
		AuthorizationID: row.AuthorizationID,
		GrantedDate:     conv.TimeFromPgDate(row.GrantedDate),
		ExpiryDate:      conv.TimeFromPgDate(row.ExpiryDate),
		IsActive:        row.IsActive,
		Notes:           row.Notes,
		CreatedAt:       conv.TimeFromPgTimestamptz(row.CreatedAt),
		UpdatedAt:       conv.TimeFromPgTimestamptz(row.UpdatedAt),
	}
}

func toDomainEmployeeAttachmentDetail(
	row db.ListEmployeeAttachmentsRow,
) domain.EmployeeAttachmentDetail {
	return domain.EmployeeAttachmentDetail{
		ID:           row.ID,
		EmployeeID:   row.EmployeeID,
		AttachmentID: row.AttachmentID,
		Category:     row.Category,
		CreatedAt:    conv.TimeFromPgTimestamptz(row.CreatedAt),
		UpdatedAt:    conv.TimeFromPgTimestamptz(row.UpdatedAt),
		Name:         row.Name,
		File:         row.File,
		Size:         row.Size,
		Tag:          row.Tag,
	}
}

func isDBNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows)
}

func pgDateFromPtr(value *time.Time) pgtype.Date {
	if value == nil {
		return pgtype.Date{}
	}

	return conv.PgDateFromTime(*value)
}

func genderEnumFromString(value string) db.GenderEnum {
	switch db.GenderEnum(value) {
	case db.GenderEnumMale, db.GenderEnumFemale, db.GenderEnumOther, db.GenderEnumUnknown:
		return db.GenderEnum(value)
	default:
		return db.GenderEnumUnknown
	}
}

func genderEnumPtrFromStringPtr(value *string) *db.GenderEnum {
	if value == nil {
		return nil
	}

	return enumPtr(genderEnumFromString(*value))
}

func maritalStatusEnumPtrToStringPtr(v *db.MaritalStatusEnum) *string {
	if v == nil {
		return nil
	}
	s := string(*v)
	return &s
}

func contractTypePtrToString(ct *db.EmployeeContractTypeEnum) string {
	if ct == nil {
		return ""
	}
	return string(*ct)
}

func contractTypeEnumPtrToStringPtr(ct *db.EmployeeContractTypeEnum) *string {
	if ct == nil {
		return nil
	}
	value := string(*ct)
	return &value
}

func contractTypeFromString(value string) db.EmployeeContractTypeEnum {
	switch db.EmployeeContractTypeEnum(value) {
	case db.EmployeeContractTypeEnumPermanent,
		db.EmployeeContractTypeEnumTemporary,
		db.EmployeeContractTypeEnumOnCall:
		return db.EmployeeContractTypeEnum(value)
	default:
		return db.EmployeeContractTypeEnumPermanent
	}
}

func contractTypePtrFromStringPtr(value *string) *db.EmployeeContractTypeEnum {
	if value == nil {
		return nil
	}

	return enumPtr(contractTypeFromString(*value))
}

func employeeJobTitleEnumFromStringPtr(value *string) *db.EmployeeJobTitleEnum {
	if value == nil {
		return nil
	}
	return enumPtr(employeeJobTitleEnumFromString(*value))
}

func employeeJobTitleEnumPtrToStringPtr(value *db.EmployeeJobTitleEnum) *string {
	if value == nil {
		return nil
	}
	result := string(*value)
	return &result
}

func employeeJobTitleEnumFromString(value string) db.EmployeeJobTitleEnum {
	switch db.EmployeeJobTitleEnum(value) {
	case db.EmployeeJobTitleEnumYouthWorkerD,
		db.EmployeeJobTitleEnumCareCoordinator,
		db.EmployeeJobTitleEnumBehavioralScientist,
		db.EmployeeJobTitleEnumQualityOfficer,
		db.EmployeeJobTitleEnumPedagogicalWorker,
		db.EmployeeJobTitleEnumTeamLead,
		db.EmployeeJobTitleEnumManager,
		db.EmployeeJobTitleEnumAdministrativeEmployee:
		return db.EmployeeJobTitleEnum(value)
	default:
		return ""
	}
}

func wageTaxTablePtrFromStringPtr(value *string) *db.WageTaxTableEnum {
	if value == nil {
		return nil
	}
	switch db.WageTaxTableEnum(*value) {
	case db.WageTaxTableEnumWhiteTable, db.WageTaxTableEnumGreenTable:
		v := db.WageTaxTableEnum(*value)
		return &v
	default:
		return nil
	}
}

func wageTaxTablePtrToStringPtr(value *db.WageTaxTableEnum) *string {
	if value == nil {
		return nil
	}
	s := string(*value)
	return &s
}

func weekdayEnumFromString(value string) db.WeekdayEnum {
	switch db.WeekdayEnum(value) {
	case db.WeekdayEnumMonday,
		db.WeekdayEnumTuesday,
		db.WeekdayEnumWednesday,
		db.WeekdayEnumThursday,
		db.WeekdayEnumFriday,
		db.WeekdayEnumSaturday,
		db.WeekdayEnumSunday:
		return db.WeekdayEnum(value)
	default:
		return ""
	}
}

func weekdayEnumPtrFromStringPtr(value *string) *db.WeekdayEnum {
	if value == nil {
		return nil
	}
	return enumPtr(weekdayEnumFromString(*value))
}

func enumPtr[T any](value T) *T {
	return &value
}

func (r *EmployeeRepository) GetEmployeeContractByID(
	ctx context.Context,
	contractID uuid.UUID,
) (*domain.EmployeeContractInfo, error) {
	row, err := r.store.GetEmployeeContractByID(ctx, contractID)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrEmployeeNotFound
		}
		return nil, err
	}
	return &domain.EmployeeContractInfo{
		ID:               row.ID,
		EmployeeID:       row.EmployeeID,
		ContractType:     string(row.ContractType),
		StartDate:        conv.TimeFromPgDate(row.StartDate),
		ContractEndDate:  conv.TimePtrFromPgDate(row.ContractEndDate),
		EffectiveEndDate: conv.TimePtrFromPgDate(row.EffectiveEndDate),
		HoursPerWeek:     row.HoursPerWeek,
	}, nil
}

func (r *EmployeeRepository) ListEmployeeContracts(
	ctx context.Context,
	employeeID uuid.UUID,
) ([]domain.EmployeeContractDetail, error) {
	rows, err := r.store.ListEmployeeContractDetails(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	contracts := make([]domain.EmployeeContractDetail, len(rows))
	for i, row := range rows {
		contracts[i] = toDomainContractDetail(row)
	}
	return contracts, nil
}

func (tx *employeeTxRepo) EndEmployeeContractSegment(
	ctx context.Context,
	contractID uuid.UUID,
	endDate time.Time,
	updatedBy *uuid.UUID,
) error {
	_, err := tx.queries.EndEmployeeContractSegment(ctx, db.EndEmployeeContractSegmentParams{
		ID:                  contractID,
		EffectiveEndDate:    conv.PgDateFromTime(endDate),
		UpdatedByEmployeeID: updatedBy,
	})
	return err
}

func (tx *employeeTxRepo) GetEmployeeContractAtDate(
	ctx context.Context,
	employeeID uuid.UUID,
	targetDate time.Time,
) (*domain.EmployeeContractInfo, error) {
	contract, err := tx.queries.GetEmployeeContractAtDate(ctx, db.GetEmployeeContractAtDateParams{
		EmployeeID: employeeID,
		TargetDate: conv.PgDateFromTime(targetDate),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.EmployeeContractInfo{
		ID:               contract.ID,
		EmployeeID:       contract.EmployeeID,
		ContractType:     string(contract.ContractType),
		StartDate:        contract.StartDate.Time,
		ContractEndDate:  conv.TimePtrFromPgDate(contract.ContractEndDate),
		EffectiveEndDate: conv.TimePtrFromPgDate(contract.EffectiveEndDate),
		HoursPerWeek:     contract.HoursPerWeek,
	}, nil
}

func (tx *employeeTxRepo) AddNewContract(
	ctx context.Context,
	employeeID uuid.UUID,
	previousContractID *uuid.UUID,
	params domain.CreateNewContractParams,
) (uuid.UUID, error) {
	contract, err := tx.queries.AddEmployeeNewContract(ctx, db.AddEmployeeNewContractParams{
		EmployeeID:           employeeID,
		JobTitle:             employeeJobTitleEnumFromString(params.JobTitle),
		DepartmentID:         params.DepartmentID,
		LocationID:           params.LocationID,
		ContractType:         contractTypeFromString(params.ContractType),
		StartDate:            conv.PgDateFromTime(params.StartDate),
		ContractEndDate:      pgDateFromPtr(params.ContractEndDate),
		HoursPerWeek:         params.HoursPerWeek,
		RosterFreeDay:        weekdayEnumFromString(params.RosterFreeDay),
		WageTaxTable:         wageTaxTablePtrFromStringPtr(params.WageTaxTable),
		PreviousContractID:   previousContractID,
		CreatedByEmployeeID:  nil,
	})
	if err != nil {
		return uuid.Nil, err
	}
	return contract.ID, nil
}

func (tx *employeeTxRepo) UpdateEmployeeContract(
	ctx context.Context,
	employeeID, contractID uuid.UUID,
	params domain.UpdateEmployeeContractParams,
) (*domain.EmployeeContractDetail, error) {
	row, err := tx.queries.UpdateEmployeeContract(ctx, db.UpdateEmployeeContractParams{
		JobTitle:             employeeJobTitleEnumFromStringPtr(params.JobTitle),
		DepartmentID:         params.DepartmentID,
		LocationID:           params.LocationID,
		ContractType:         contractTypePtrFromStringPtr(params.ContractType),
		StartDate:            pgDateFromPtr(params.StartDate),
		ContractEndDate:      pgDateFromPtr(params.ContractEndDate),
		HoursPerWeek:         params.HoursPerWeek,
		RosterFreeDay:        weekdayEnumPtrFromStringPtr(params.RosterFreeDay),
		WageTaxTable:         wageTaxTablePtrFromStringPtr(params.WageTaxTable),
		ID:                   contractID,
		EmployeeID:           employeeID,
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrEmployeeNotFound
		}
		return nil, err
	}
	detail := toDomainContractDetailFromRow(row)
	return &detail, nil
}

func (r *EmployeeRepository) UpdateEmployeeContract(
	ctx context.Context,
	employeeID, contractID uuid.UUID,
	params domain.UpdateEmployeeContractParams,
) (*domain.EmployeeContractDetail, error) {
	row, err := r.store.UpdateEmployeeContract(ctx, db.UpdateEmployeeContractParams{
		JobTitle:             employeeJobTitleEnumFromStringPtr(params.JobTitle),
		DepartmentID:         params.DepartmentID,
		LocationID:           params.LocationID,
		ContractType:         contractTypePtrFromStringPtr(params.ContractType),
		StartDate:            pgDateFromPtr(params.StartDate),
		ContractEndDate:      pgDateFromPtr(params.ContractEndDate),
		HoursPerWeek:         params.HoursPerWeek,
		RosterFreeDay:        weekdayEnumPtrFromStringPtr(params.RosterFreeDay),
		WageTaxTable:         wageTaxTablePtrFromStringPtr(params.WageTaxTable),
		ID:                   contractID,
		EmployeeID:           employeeID,
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrEmployeeNotFound
		}
		return nil, err
	}
	detail := toDomainContractDetailFromRow(row)
	return &detail, nil
}

func (r *EmployeeRepository) UpdateEmployeeContractSalary(
	ctx context.Context,
	employeeID, contractID uuid.UUID,
	params domain.UpdateEmployeeContractSalaryParams,
) (*domain.EmployeeSalaryAssignmentDetail, error) {
	var salaryID uuid.UUID
	err := r.store.ExecTx(ctx, func(q *db.Queries) error {
		contract, err := q.GetEmployeeContractByID(ctx, contractID)
		if err != nil {
			if isDBNotFound(err) {
				return domain.ErrEmployeeNotFound
			}
			return err
		}
		if contract.EmployeeID != employeeID {
			return domain.ErrContractChangeInvalid
		}

		salary, err := q.GetEmployeeSalaryAssignmentByContract(
			ctx,
			db.GetEmployeeSalaryAssignmentByContractParams{
				EmployeeID: employeeID,
				ContractID: &contractID,
			},
		)
		if err != nil {
			if !isDBNotFound(err) {
				return err
			}
			salary, err = q.CreateEmployeeSalaryAssignment(
				ctx,
				db.CreateEmployeeSalaryAssignmentParams{
					EmployeeID:        employeeID,
					ContractID:        &contractID,
					SalaryScaleStepID: params.SalaryScaleStepID,
					EffectiveFrom:     contract.StartDate,
				},
			)
			if err != nil {
				return err
			}
			salaryID = salary.ID
			return nil
		}

		salary, err = q.UpdateEmployeeSalaryAssignmentScaleStep(
			ctx,
			db.UpdateEmployeeSalaryAssignmentScaleStepParams{
				ID:                salary.ID,
				SalaryScaleStepID: params.SalaryScaleStepID,
			},
		)
		if err != nil {
			return err
		}
		salaryID = salary.ID
		return nil
	})
	if err != nil {
		return nil, err
	}

	row, err := r.store.GetEmployeeSalaryAssignmentDetailByID(ctx, salaryID)
	if err != nil {
		return nil, err
	}
	return toDomainSalaryAssignmentDetail(row), nil
}

func (tx *employeeTxRepo) AddEmployeeContractAmendment(
	ctx context.Context,
	employeeID uuid.UUID,
	previousContractID uuid.UUID,
	params domain.CreateContractAmendmentParams,
) (uuid.UUID, error) {
	contract, err := tx.queries.AddEmployeeContractAmendment(
		ctx,
		db.AddEmployeeContractAmendmentParams{
			EmployeeID:           employeeID,
			JobTitle:             employeeJobTitleEnumFromString(params.JobTitle),
			DepartmentID:         params.DepartmentID,
			LocationID:           params.LocationID,
			ContractType:         contractTypeFromString(params.ContractType),
			StartDate:            conv.PgDateFromTime(params.StartDate),
			ContractEndDate:      pgDateFromPtr(params.ContractEndDate),
			HoursPerWeek:         params.HoursPerWeek,
			RosterFreeDay:        weekdayEnumFromString(params.RosterFreeDay),
			WageTaxTable:         wageTaxTablePtrFromStringPtr(params.WageTaxTable),
			PreviousContractID:   &previousContractID,
			ChangeReason:         params.ChangeReason,
			CreatedByEmployeeID:  nil,
		},
	)
	if err != nil {
		return uuid.Nil, err
	}
	return contract.ID, nil
}

func (r *EmployeeRepository) UpdatePassword(
	ctx context.Context,
	userID uuid.UUID,
	password string,
) error {
	return r.store.UpdatePassword(ctx, db.UpdatePasswordParams{
		ID:       userID,
		Password: password,
	})
}

var _ domain.EmployeeRepository = (*EmployeeRepository)(nil)
