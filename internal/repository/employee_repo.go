package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type EmployeeRepository struct {
	store *db.Store
}

func NewEmployeeRepository(store *db.Store) domain.EmployeeRepository {
	return &EmployeeRepository{store: store}
}

func (r *EmployeeRepository) GetEmployeeByID(ctx context.Context, id uuid.UUID) (*domain.EmployeeDetail, error) {
	row, err := r.store.GetEmployeeProfileByID(ctx, id)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrEmployeeNotFound
		}
		return nil, err
	}

	return toDomainEmployeeDetailFromGetEmployeeProfileByIDRow(row), nil
}

func (r *EmployeeRepository) GetEmployeeByUserID(ctx context.Context, userID uuid.UUID) (*domain.EmployeeProfile, error) {
	row, err := r.store.GetEmployeeProfileByUserID(ctx, userID)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrEmployeeNotFound
		}
		return nil, err
	}

	return toDomainEmployeeProfile(row)
}

func (r *EmployeeRepository) ListEmployees(ctx context.Context, params domain.ListEmployeesParams) (*domain.EmployeePage, error) {
	rows, err := r.store.ListEmployeeProfile(ctx, db.ListEmployeeProfileParams{
		Limit:               params.Limit,
		Offset:              params.Offset,
		IncludeArchived:     params.IncludeArchived,
		IncludeOutOfService: params.IncludeOutOfService,
		LocationID:          params.LocationID,
		ContractType:        nullContractTypeFromPtr(params.ContractType),
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

func (r *EmployeeRepository) CountEmployees(ctx context.Context, params domain.ListEmployeesParams) (int64, error) {
	return r.store.CountEmployeeProfile(ctx, db.CountEmployeeProfileParams{
		IncludeArchived:     params.IncludeArchived,
		IncludeOutOfService: params.IncludeOutOfService,
		LocationID:          params.LocationID,
		ContractType:        nullContractTypeFromPtr(params.ContractType),
	})
}

func (r *EmployeeRepository) CreateEmployee(ctx context.Context, params domain.CreateEmployeeParams) (*domain.EmployeeDetail, error) {
	result, err := r.store.CreateEmployeeWithAccountTx(ctx, db.CreateEmployeeWithAccountTxParams{
		CreateUserParams: db.CreateUserParams{
			Password: params.UserPassword,
			Email:    params.UserEmail,
			IsActive: true,
		},
		CreateEmployeeParams: db.CreateEmployeeProfileParams{
			FirstName:             params.FirstName,
			LastName:              params.LastName,
			Bsn:                   params.Bsn,
			Street:                params.Street,
			HouseNumber:           params.HouseNumber,
			HouseNumberAddition:   params.HouseNumberAddition,
			PostalCode:            params.PostalCode,
			City:                  params.City,
			Position:              params.Position,
			DepartmentID:          params.DepartmentID,
			ManagerEmployeeID:     params.ManagerEmployeeID,
			EmployeeNumber:        params.EmployeeNumber,
			EmploymentNumber:      params.EmploymentNumber,
			PrivateEmailAddress:   params.PrivateEmailAddress,
			WorkEmailAddress:      params.WorkEmailAddress,
			WorkPhoneNumber:       params.WorkPhoneNumber,
			PrivatePhoneNumber:    params.PrivatePhoneNumber,
			DateOfBirth:           pgDateFromPtr(params.DateOfBirth),
			HomeTelephoneNumber:   params.HomeTelephoneNumber,
			Gender:                genderEnumFromString(params.Gender),
			LocationID:            params.LocationID,
			ContractHours:         params.ContractHours,
			ContractEndDate:       pgDateFromPtr(params.ContractEndDate),
			ContractStartDate:     pgDateFromPtr(params.ContractStartDate),
			ContractType:          contractTypeFromString(params.ContractType),
			ContractRate:          params.ContractRate,
			IrregularHoursProfile: irregularHoursProfileFromString(params.IrregularHoursProfile),
		},
		RoleID: params.RoleID,
	})
	if err != nil {
		return nil, err
	}

	return toDomainEmployeeDetailFromEmployeeProfile(result.Employee), nil
}

func (r *EmployeeRepository) UpdateEmployee(ctx context.Context, id uuid.UUID, params domain.UpdateEmployeeParams) (*domain.EmployeeDetail, error) {
	row, err := r.store.UpdateEmployeeProfile(ctx, db.UpdateEmployeeProfileParams{
		FirstName:             params.FirstName,
		LastName:              params.LastName,
		Position:              params.Position,
		DepartmentID:          params.DepartmentID,
		ManagerEmployeeID:     params.ManagerEmployeeID,
		EmployeeNumber:        params.EmployeeNumber,
		EmploymentNumber:      params.EmploymentNumber,
		PrivateEmailAddress:   params.PrivateEmailAddress,
		WorkEmailAddress:      nil,
		PrivatePhoneNumber:    params.PrivatePhoneNumber,
		WorkPhoneNumber:       params.WorkPhoneNumber,
		DateOfBirth:           pgDateFromPtr(params.DateOfBirth),
		HomeTelephoneNumber:   params.HomeTelephoneNumber,
		Gender:                nullGenderEnumFromPtr(params.Gender),
		LocationID:            params.LocationID,
		IrregularHoursProfile: nullIrregularHoursProfileFromPtr(params.IrregularHoursProfile),
		HasBorrowed:           params.HasBorrowed,
		OutOfService:          params.OutOfService,
		IsArchived:            params.IsArchived,
		ID:                    id,
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrEmployeeNotFound
		}
		return nil, err
	}

	return toDomainEmployeeDetailFromEmployeeProfile(row), nil
}

func (r *EmployeeRepository) GetEmployeeCounts(ctx context.Context) (*domain.EmployeeCounts, error) {
	row, err := r.store.GetEmployeeCounts(ctx)
	if err != nil {
		return nil, err
	}

	return toDomainEmployeeCounts(row), nil
}

func (r *EmployeeRepository) SearchEmployeesByNameOrEmail(ctx context.Context, search *string) ([]domain.EmployeeSearchResult, error) {
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

func (r *EmployeeRepository) GetContractDetails(ctx context.Context, employeeID uuid.UUID) (*domain.ContractDetails, error) {
	row, err := r.store.GetEmployeeContractDetails(ctx, employeeID)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrEmployeeNotFound
		}
		return nil, err
	}

	return toDomainContractDetails(row), nil
}

func (r *EmployeeRepository) AddContractDetails(ctx context.Context, employeeID uuid.UUID, params domain.AddContractDetailsParams) (*domain.EmployeeDetail, error) {
	changeCount, err := r.store.CountEmployeeContractChanges(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	if changeCount > 0 {
		return nil, domain.ErrContractHistoryExists
	}

	row, err := r.store.AddEmployeeContractDetails(ctx, db.AddEmployeeContractDetailsParams{
		ID:                employeeID,
		ContractHours:     params.ContractHours,
		ContractStartDate: conv.PgDateFromTime(params.ContractStartDate),
		ContractEndDate:   conv.PgDateFromTime(params.ContractEndDate),
		ContractType:      db.NullEmployeeContractTypeEnum{},
		ContractRate:      params.ContractRate,
		IrregularHoursProfile: db.NullIrregularHoursProfileEnum{
			IrregularHoursProfileEnum: irregularHoursProfileFromString(params.IrregularHoursProfile),
			Valid:                     true,
		},
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrEmployeeNotFound
		}
		return nil, err
	}

	return toDomainEmployeeDetailFromEmployeeProfile(row), nil
}

func (r *EmployeeRepository) ListContractChanges(ctx context.Context, employeeID uuid.UUID) ([]domain.EmployeeContractChange, error) {
	if _, err := r.store.GetEmployeeContractSnapshotForContractChange(ctx, employeeID); err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrEmployeeNotFound
		}
		return nil, err
	}

	rows, err := r.store.ListEmployeeContractChanges(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	items := make([]domain.EmployeeContractChange, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.EmployeeContractChange{
			ID:                    row.ID,
			EmployeeID:            row.EmployeeID,
			EffectiveFrom:         conv.TimeFromPgDate(row.EffectiveFrom),
			EffectiveTo:           conv.TimePtrFromPgDate(row.EffectiveTo),
			ContractHours:         row.ContractHours,
			ContractType:          string(row.ContractType),
			ContractRate:          row.ContractRate,
			IrregularHoursProfile: string(row.IrregularHoursProfile),
			ContractEndDate:       conv.TimePtrFromPgDate(row.ContractEndDate),
			CreatedByEmployeeID:   row.CreatedByEmployeeID,
			CreatedAt:             conv.TimeFromPgTimestamptz(row.CreatedAt),
			UpdatedAt:             conv.TimeFromPgTimestamptz(row.UpdatedAt),
		})
	}
	return items, nil
}

func (r *EmployeeRepository) CreateContractChange(
	ctx context.Context,
	actorEmployeeID, employeeID uuid.UUID,
	params domain.CreateEmployeeContractChangeParams,
) (*domain.CreateEmployeeContractChangeResult, error) {
	var result *domain.CreateEmployeeContractChangeResult

	err := r.store.ExecTx(ctx, func(q *db.Queries) error {
		snapshot, err := q.GetEmployeeContractSnapshotForContractChange(ctx, employeeID)
		if err != nil {
			if isDBNotFound(err) {
				return domain.ErrEmployeeNotFound
			}
			return err
		}

		changeCount, err := q.CountEmployeeContractChanges(ctx, employeeID)
		if err != nil {
			return err
		}

		if changeCount == 0 {
			if !snapshot.ContractStartDate.Valid {
				return domain.ErrContractBaselineMissingStartDate
			}

			baselineDate := conv.TimeFromPgDate(snapshot.ContractStartDate).UTC()
			effectiveDate := dateOnly(params.EffectiveFrom)
			if !effectiveDate.Equal(dateOnly(baselineDate)) {
				_, err = q.CreateEmployeeContractChange(ctx, db.CreateEmployeeContractChangeParams{
					EmployeeID:            employeeID,
					EffectiveFrom:         conv.PgDateFromTime(baselineDate),
					ContractHours:         valueOrZero(snapshot.ContractHours),
					ContractType:          snapshot.ContractType,
					ContractRate:          snapshot.ContractRate,
					IrregularHoursProfile: snapshot.IrregularHoursProfile,
					ContractEndDate:       snapshot.ContractEndDate,
					CreatedByEmployeeID:   actorEmployeeID,
				})
				if err != nil {
					return mapContractChangeDBError(err)
				}
			}
		}

		created, err := q.CreateEmployeeContractChange(ctx, db.CreateEmployeeContractChangeParams{
			EmployeeID:            employeeID,
			EffectiveFrom:         conv.PgDateFromTime(dateOnly(params.EffectiveFrom)),
			ContractHours:         params.ContractHours,
			ContractType:          contractTypeFromString(params.ContractType),
			ContractRate:          params.ContractRate,
			IrregularHoursProfile: irregularHoursProfileFromString(params.IrregularHoursProfile),
			ContractEndDate:       pgDateFromPtr(params.ContractEndDate),
			CreatedByEmployeeID:   actorEmployeeID,
		})
		if err != nil {
			return mapContractChangeDBError(err)
		}

		if _, err := q.SyncEmployeeProfileContractFromLatestChange(ctx, employeeID); err != nil {
			return err
		}

		recalculations, err := recomputeLegalLeaveBalancesFromYear(ctx, q, actorEmployeeID, employeeID, int32(created.EffectiveFrom.Time.Year()), created.ID, conv.TimeFromPgDate(created.EffectiveFrom))
		if err != nil {
			return err
		}

		result = &domain.CreateEmployeeContractChangeResult{
			Change: domain.EmployeeContractChange{
				ID:                    created.ID,
				EmployeeID:            created.EmployeeID,
				EffectiveFrom:         conv.TimeFromPgDate(created.EffectiveFrom),
				EffectiveTo:           nil,
				ContractHours:         created.ContractHours,
				ContractType:          string(created.ContractType),
				ContractRate:          created.ContractRate,
				IrregularHoursProfile: string(created.IrregularHoursProfile),
				ContractEndDate:       conv.TimePtrFromPgDate(created.ContractEndDate),
				CreatedByEmployeeID:   created.CreatedByEmployeeID,
				CreatedAt:             conv.TimeFromPgTimestamptz(created.CreatedAt),
				UpdatedAt:             conv.TimeFromPgTimestamptz(created.UpdatedAt),
			},
			Recalculations: recalculations,
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *EmployeeRepository) UpdateIsSubcontractor(ctx context.Context, employeeID uuid.UUID, contractType string) (*domain.EmployeeDetail, error) {
	row, err := r.store.UpdateEmployeeIsSubcontractor(ctx, db.UpdateEmployeeIsSubcontractorParams{
		ID:           employeeID,
		ContractType: contractTypeFromString(contractType),
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrEmployeeNotFound
		}
		return nil, err
	}

	return toDomainEmployeeDetailFromEmployeeProfile(row), nil
}

func (r *EmployeeRepository) ListEducation(ctx context.Context, employeeID uuid.UUID) ([]domain.Education, error) {
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

func (r *EmployeeRepository) AddEducation(ctx context.Context, employeeID uuid.UUID, params domain.CreateEducationParams) (*domain.Education, error) {
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

func (r *EmployeeRepository) UpdateEducation(ctx context.Context, id uuid.UUID, params domain.UpdateEducationParams) (*domain.Education, error) {
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

func (r *EmployeeRepository) DeleteEducation(ctx context.Context, id uuid.UUID) (*domain.Education, error) {
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

func (r *EmployeeRepository) ListExperience(ctx context.Context, employeeID uuid.UUID) ([]domain.Experience, error) {
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

func (r *EmployeeRepository) AddExperience(ctx context.Context, employeeID uuid.UUID, params domain.CreateExperienceParams) (*domain.Experience, error) {
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

func (r *EmployeeRepository) UpdateExperience(ctx context.Context, id uuid.UUID, params domain.UpdateExperienceParams) (*domain.Experience, error) {
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

func (r *EmployeeRepository) DeleteExperience(ctx context.Context, id uuid.UUID) (*domain.Experience, error) {
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

func (r *EmployeeRepository) ListCertification(ctx context.Context, employeeID uuid.UUID) ([]domain.Certification, error) {
	rows, err := r.store.ListEmployeeCertifications(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Certification, 0, len(rows))
	for _, row := range rows {
		result = append(result, toDomainCertification(row))
	}

	return result, nil
}

func (r *EmployeeRepository) AddCertification(ctx context.Context, employeeID uuid.UUID, params domain.CreateCertificationParams) (*domain.Certification, error) {
	row, err := r.store.AddEmployeeCertification(ctx, db.AddEmployeeCertificationParams{
		EmployeeID: employeeID,
		Name:       params.Name,
		IssuedBy:   params.IssuedBy,
		DateIssued: conv.PgDateFromTime(params.DateIssued),
	})
	if err != nil {
		return nil, err
	}

	result := toDomainCertification(row)
	return &result, nil
}

func (r *EmployeeRepository) UpdateCertification(ctx context.Context, id uuid.UUID, params domain.UpdateCertificationParams) (*domain.Certification, error) {
	row, err := r.store.UpdateEmployeeCertification(ctx, db.UpdateEmployeeCertificationParams{
		ID:         id,
		Name:       params.Name,
		IssuedBy:   params.IssuedBy,
		DateIssued: pgDateFromPtr(params.DateIssued),
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrCertificationNotFound
		}
		return nil, err
	}

	result := toDomainCertification(row)
	return &result, nil
}

func (r *EmployeeRepository) DeleteCertification(ctx context.Context, id uuid.UUID) (*domain.Certification, error) {
	row, err := r.store.DeleteEmployeeCertification(ctx, id)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrCertificationNotFound
		}
		return nil, err
	}

	result := toDomainCertification(row)
	return &result, nil
}

func toDomainEmployee(row db.ListEmployeeProfileRow) domain.Employee {
	return domain.Employee{
		ID:              row.ID,
		FirstName:       row.FirstName,
		LastName:        row.LastName,
		Bsn:             row.Bsn,
		ContractType:    string(row.ContractType),
		DepartmentName:  row.DepartmentName,
		ContractEndDate: conv.TimePtrFromPgDate(row.ContractEndDate),
		LocationAddress: row.LocationAddress,
	}
}

func toDomainEmployeeDetailFromGetEmployeeProfileByIDRow(row db.GetEmployeeProfileByIDRow) *domain.EmployeeDetail {
	return &domain.EmployeeDetail{
		ID:                    row.ID,
		UserID:                row.UserID,
		FirstName:             row.FirstName,
		LastName:              row.LastName,
		Bsn:                   row.Bsn,
		Street:                row.Street,
		HouseNumber:           row.HouseNumber,
		HouseNumberAddition:   row.HouseNumberAddition,
		PostalCode:            row.PostalCode,
		City:                  row.City,
		Position:              row.Position,
		EmployeeNumber:        row.EmployeeNumber,
		EmploymentNumber:      row.EmploymentNumber,
		PrivateEmailAddress:   row.PrivateEmailAddress,
		WorkEmailAddress:      row.WorkEmailAddress,
		PrivatePhoneNumber:    row.PrivatePhoneNumber,
		WorkPhoneNumber:       row.WorkPhoneNumber,
		DateOfBirth:           conv.TimePtrFromPgDate(row.DateOfBirth),
		HomeTelephoneNumber:   row.HomeTelephoneNumber,
		CreatedAt:             conv.TimeFromPgTimestamptz(row.CreatedAt),
		Gender:                string(row.Gender),
		LocationID:            row.LocationID,
		DepartmentID:          row.DepartmentID,
		ManagerEmployeeID:     row.ManagerEmployeeID,
		HasBorrowed:           row.HasBorrowed,
		OutOfService:          row.OutOfService,
		IsArchived:            row.IsArchived,
		ContractHours:         row.ContractHours,
		ContractEndDate:       conv.TimePtrFromPgDate(row.ContractEndDate),
		ContractStartDate:     conv.TimePtrFromPgDate(row.ContractStartDate),
		ContractType:          string(row.ContractType),
		ContractRate:          row.ContractRate,
		IrregularHoursProfile: string(row.IrregularHoursProfile),
		ProfilePicture:        row.ProfilePicture,
		DepartmentName:        row.DepartmentName,
		ManagerFirstName:      row.ManagerFirstName,
		ManagerLastName:       row.ManagerLastName,
	}
}

func toDomainEmployeeDetailFromEmployeeProfile(row db.EmployeeProfile) *domain.EmployeeDetail {
	return &domain.EmployeeDetail{
		ID:                    row.ID,
		UserID:                row.UserID,
		FirstName:             row.FirstName,
		LastName:              row.LastName,
		Bsn:                   row.Bsn,
		Street:                row.Street,
		HouseNumber:           row.HouseNumber,
		HouseNumberAddition:   row.HouseNumberAddition,
		PostalCode:            row.PostalCode,
		City:                  row.City,
		Position:              row.Position,
		EmployeeNumber:        row.EmployeeNumber,
		EmploymentNumber:      row.EmploymentNumber,
		PrivateEmailAddress:   row.PrivateEmailAddress,
		WorkEmailAddress:      row.WorkEmailAddress,
		PrivatePhoneNumber:    row.PrivatePhoneNumber,
		WorkPhoneNumber:       row.WorkPhoneNumber,
		DateOfBirth:           conv.TimePtrFromPgDate(row.DateOfBirth),
		HomeTelephoneNumber:   row.HomeTelephoneNumber,
		CreatedAt:             conv.TimeFromPgTimestamptz(row.CreatedAt),
		Gender:                string(row.Gender),
		LocationID:            row.LocationID,
		DepartmentID:          row.DepartmentID,
		ManagerEmployeeID:     row.ManagerEmployeeID,
		HasBorrowed:           row.HasBorrowed,
		OutOfService:          row.OutOfService,
		IsArchived:            row.IsArchived,
		ContractHours:         row.ContractHours,
		ContractEndDate:       conv.TimePtrFromPgDate(row.ContractEndDate),
		ContractStartDate:     conv.TimePtrFromPgDate(row.ContractStartDate),
		ContractType:          string(row.ContractType),
		ContractRate:          row.ContractRate,
		IrregularHoursProfile: string(row.IrregularHoursProfile),
	}
}

func toDomainEmployeeProfile(row db.GetEmployeeProfileByUserIDRow) (*domain.EmployeeProfile, error) {
	permissions := make([]domain.Permission, 0)
	if len(row.Permissions) > 0 {
		if err := json.Unmarshal(row.Permissions, &permissions); err != nil {
			return nil, err
		}
	}

	return &domain.EmployeeProfile{
		UserID:           row.UserID,
		Email:            row.Email,
		LastLogin:        conv.TimeFromPgTimestamptz(row.LastLogin),
		TwoFactorEnabled: row.TwoFactorEnabled,
		EmployeeID:       row.EmployeeID,
		FirstName:        row.FirstName,
		LastName:         row.LastName,
		Permissions:      permissions,
	}, nil
}

func toDomainEmployeeCounts(row db.GetEmployeeCountsRow) *domain.EmployeeCounts {
	return &domain.EmployeeCounts{
		TotalEmployees:      row.TotalEmployees,
		TotalSubcontractors: row.TotalSubcontractors,
		TotalArchived:       row.TotalArchived,
		TotalOutOfService:   row.TotalOutOfService,
	}
}

func toDomainEmployeeSearchResult(row db.SearchEmployeesByNameOrEmailRow) domain.EmployeeSearchResult {
	return domain.EmployeeSearchResult{
		ID:               row.ID,
		FirstName:        row.FirstName,
		LastName:         row.LastName,
		WorkEmailAddress: row.WorkEmailAddress,
	}
}

func toDomainContractDetails(row db.GetEmployeeContractDetailsRow) *domain.ContractDetails {
	isSubcontractor := row.ContractType == db.EmployeeContractTypeEnumZZP

	return &domain.ContractDetails{
		ContractHours:         row.ContractHours,
		ContractStartDate:     conv.TimeFromPgDate(row.ContractStartDate),
		ContractEndDate:       conv.TimeFromPgDate(row.ContractEndDate),
		ContractType:          string(row.ContractType),
		ContractRate:          row.ContractRate,
		IrregularHoursProfile: string(row.IrregularHoursProfile),
		IsSubcontractor:       &isSubcontractor,
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

func toDomainCertification(row db.Certification) domain.Certification {
	return domain.Certification{
		ID:         row.ID,
		EmployeeID: row.EmployeeID,
		Name:       row.Name,
		IssuedBy:   row.IssuedBy,
		DateIssued: conv.TimeFromPgDate(row.DateIssued),
		CreatedAt:  conv.TimeFromPgTimestamptz(row.CreatedAt),
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

func nullGenderEnumFromPtr(value *string) db.NullGenderEnum {
	if value == nil {
		return db.NullGenderEnum{}
	}

	return db.NullGenderEnum{GenderEnum: genderEnumFromString(*value), Valid: true}
}

func contractTypeFromString(value string) db.EmployeeContractTypeEnum {
	switch db.EmployeeContractTypeEnum(value) {
	case db.EmployeeContractTypeEnumLoondienst, db.EmployeeContractTypeEnumZZP, db.EmployeeContractTypeEnumNone:
		return db.EmployeeContractTypeEnum(value)
	default:
		return db.EmployeeContractTypeEnumNone
	}
}

func nullContractTypeFromPtr(value *string) db.NullEmployeeContractTypeEnum {
	if value == nil {
		return db.NullEmployeeContractTypeEnum{}
	}

	return db.NullEmployeeContractTypeEnum{EmployeeContractTypeEnum: contractTypeFromString(*value), Valid: true}
}

func irregularHoursProfileFromString(value string) db.IrregularHoursProfileEnum {
	switch db.IrregularHoursProfileEnum(value) {
	case db.IrregularHoursProfileEnumNone, db.IrregularHoursProfileEnumRoster, db.IrregularHoursProfileEnumNonRoster:
		return db.IrregularHoursProfileEnum(value)
	default:
		return db.IrregularHoursProfileEnumNone
	}
}

func nullIrregularHoursProfileFromPtr(value *string) db.NullIrregularHoursProfileEnum {
	if value == nil {
		return db.NullIrregularHoursProfileEnum{}
	}

	return db.NullIrregularHoursProfileEnum{
		IrregularHoursProfileEnum: irregularHoursProfileFromString(*value),
		Valid:                     true,
	}
}

func valueOrZero(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}

func dateOnly(value time.Time) time.Time {
	return time.Date(value.UTC().Year(), value.UTC().Month(), value.UTC().Day(), 0, 0, 0, 0, time.UTC)
}

func recomputeLegalLeaveBalancesFromYear(
	ctx context.Context,
	q *db.Queries,
	actorEmployeeID uuid.UUID,
	employeeID uuid.UUID,
	startYear int32,
	contractChangeID uuid.UUID,
	effectiveFrom time.Time,
) ([]domain.LeaveRecalculationImpact, error) {
	currentYear := int32(time.Now().UTC().Year())
	if startYear > currentYear {
		return []domain.LeaveRecalculationImpact{}, nil
	}

	reasons := make([]domain.LeaveRecalculationImpact, 0, currentYear-startYear+1)

	for year := startYear; year <= currentYear; year++ {
		if err := q.EnsureLeaveBalanceForYear(ctx, db.EnsureLeaveBalanceForYearParams{
			EmployeeID: employeeID,
			Year:       year,
		}); err != nil {
			return nil, err
		}

		balance, err := q.LockLeaveBalanceByEmployeeYear(ctx, db.LockLeaveBalanceByEmployeeYearParams{
			EmployeeID: employeeID,
			Year:       year,
		})
		if err != nil {
			return nil, err
		}

		legalTotalAfter, err := q.ComputeLegalLeaveTotalForYear(ctx, db.ComputeLegalLeaveTotalForYearParams{
			EmployeeID: employeeID,
			Year:       year,
		})
		if err != nil {
			return nil, err
		}

		if balance.LegalUsedHours > legalTotalAfter {
			return nil, fmt.Errorf("%w: year %d", domain.ErrContractChangeLeaveConflict, year)
		}

		reasons = append(reasons, domain.LeaveRecalculationImpact{
			Year:        year,
			LegalBefore: balance.LegalTotalHours,
			LegalAfter:  legalTotalAfter,
			Delta:       legalTotalAfter - balance.LegalTotalHours,
		})

		legalDelta := legalTotalAfter - balance.LegalTotalHours
		if legalDelta == 0 {
			continue
		}

		updated, err := q.ApplyLeaveBalanceTotalAdjustment(ctx, db.ApplyLeaveBalanceTotalAdjustmentParams{
			ID:              balance.ID,
			LegalHoursDelta: legalDelta,
			ExtraHoursDelta: 0,
		})
		if err != nil {
			return nil, err
		}

		reason := fmt.Sprintf(
			"contract change %s effective %s",
			contractChangeID.String(),
			effectiveFrom.UTC().Format("2006-01-02"),
		)
		if _, err := q.CreateLeaveBalanceAdjustmentAudit(ctx, db.CreateLeaveBalanceAdjustmentAuditParams{
			LeaveBalanceID:        updated.ID,
			EmployeeID:            employeeID,
			Year:                  year,
			LegalHoursDelta:       legalDelta,
			ExtraHoursDelta:       0,
			Reason:                reason,
			AdjustedByEmployeeID:  actorEmployeeID,
			LegalTotalHoursBefore: balance.LegalTotalHours,
			ExtraTotalHoursBefore: balance.ExtraTotalHours,
			LegalTotalHoursAfter:  updated.LegalTotalHours,
			ExtraTotalHoursAfter:  updated.ExtraTotalHours,
		}); err != nil {
			return nil, err
		}
	}

	return reasons, nil
}

func mapContractChangeDBError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return fmt.Errorf("%w: duplicate effective_from for employee", domain.ErrContractChangeInvalid)
		}
	}
	return err
}

var _ domain.EmployeeRepository = (*EmployeeRepository)(nil)
