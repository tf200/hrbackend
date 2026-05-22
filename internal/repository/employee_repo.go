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
	})
}

func (r *EmployeeRepository) CreateEmployee(
	ctx context.Context,
	params domain.CreateEmployeeParams,
) (*domain.EmployeeDetail, error) {
	var contractParams *db.AddEmployeeContractDetailsParams
	if params.Contract != nil {
		cp := db.AddEmployeeContractDetailsParams{
			EmployeeID:           uuid.Nil,
			JobTitle:             employeeJobTitleEnumFromString(params.Contract.JobTitle),
			DepartmentID:         params.Contract.DepartmentID,
			LocationID:           params.Contract.LocationID,
			OrganizationalRoleID: params.Contract.OrganizationalRoleID,
			ContractType:         contractTypeFromString(params.Contract.ContractType),
			ContractHoursType:    contractHoursTypeFromString(params.Contract.ContractHoursType),
			StartDate:            conv.PgDateFromTime(params.Contract.StartDate),
			ContractEndDate:      pgDateFromPtr(params.Contract.ContractEndDate),
			HoursPerWeek:         params.Contract.HoursPerWeek,
			MinHoursPerWeek:      params.Contract.MinHoursPerWeek,
			MaxHoursPerWeek:      params.Contract.MaxHoursPerWeek,
			RosterFreeDay:        params.Contract.RosterFreeDay,
			WageTaxTable:         wageTaxTablePtrFromStringPtr(params.Contract.WageTaxTable),
			CreatedByEmployeeID:  nil,
		}
		contractParams = &cp
	}

	var salaryParams *db.CreateEmployeeSalaryAssignmentParams
	if params.SalaryAssignment != nil {
		sp := db.CreateEmployeeSalaryAssignmentParams{
			EmployeeID:          uuid.Nil,
			ContractID:          nil,
			SalaryScaleStepID:   params.SalaryAssignment.SalaryScaleStepID,
			EffectiveFrom:       pgDateFromPtr(params.SalaryAssignment.EffectiveFrom),
			EffectiveTo:         pgDateFromPtr(params.SalaryAssignment.EffectiveTo),
			CreatedByEmployeeID: nil,
		}
		if !sp.EffectiveFrom.Valid && contractParams != nil {
			sp.EffectiveFrom = contractParams.StartDate
		}
		salaryParams = &sp
	}

	result, err := r.store.CreateEmployeeWithAccountTx(ctx, db.CreateEmployeeWithAccountTxParams{
		CreateUserParams: db.CreateUserParams{
			Password: params.UserPassword,
			Email:    params.UserEmail,
			IsActive: true,
		},
		CreateEmployeeParams: db.CreateEmployeeProfileParams{
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
		},
		RoleID:              params.RoleID,
		Contract:            contractParams,
		SalaryAssignment:    salaryParams,
		CreatedByEmployeeID: nil,
	})
	if err != nil {
		return nil, err
	}

	return toDomainEmployeeDetailFromEmployeeProfile(result.Employee), nil
}

func (r *EmployeeRepository) UpdateEmployee(
	ctx context.Context,
	id uuid.UUID,
	params domain.UpdateEmployeeParams,
) (*domain.EmployeeDetail, error) {
	row, err := r.store.UpdateEmployeeProfile(ctx, db.UpdateEmployeeProfileParams{
		FirstName:           params.FirstName,
		LastName:            params.LastName,
		ManagerEmployeeID:   params.ManagerEmployeeID,
		EmployeeNumber:      params.EmployeeNumber,
		EmploymentNumber:    params.EmploymentNumber,
		PrivateEmailAddress: params.PrivateEmailAddress,
		WorkEmailAddress:    nil,
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
			return nil, domain.ErrEmployeeNotFound
		}
		return nil, err
	}

	return toDomainEmployeeDetailFromEmployeeProfile(row), nil
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

func (r *EmployeeRepository) AddQualification(
	ctx context.Context,
	employeeID uuid.UUID,
	params domain.CreateQualificationParams,
) (*domain.Qualification, error) {
	row, err := r.store.AddEmployeeQualification(ctx, db.AddEmployeeQualificationParams{
		EmployeeID:        employeeID,
		QualificationID:   params.QualificationID,
		AchievedOn:        conv.PgDateFromTime(params.AchievedOn),
		ExpirationDate:    pgDateFromPtr(params.ExpirationDate),
		CertificateNumber: params.CertificateNumber,
	})
	if err != nil {
		return nil, err
	}

	result := toDomainQualification(row)
	return &result, nil
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

func (r *EmployeeRepository) ListQualificationTypes(
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
	}
}

func toDomainEmployeeDetailFromGetEmployeeProfileByIDRow(
	row db.GetEmployeeProfileByIDRow,
) *domain.EmployeeDetail {
	return &domain.EmployeeDetail{
		ID:                  row.ID,
		UserID:              row.UserID,
		FirstName:           row.FirstName,
		LastName:            row.LastName,
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
}

func applyEmployeeDetailStats(
	employee *domain.EmployeeDetail,
	stats db.GetEmployeeDetailStatsRow,
) {
	employee.RemainingLeaveBalanceHours = stats.RemainingLeaveBalanceHours
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

func toDomainEmployeeCounts(row db.GetEmployeeCountsRow) *domain.EmployeeCounts {
	return &domain.EmployeeCounts{
		TotalEmployees:      row.TotalEmployees,
		TotalSubcontractors: row.TotalSubcontractors,
		TotalArchived:       row.TotalArchived,
		TotalOutOfService:   row.TotalOutOfService,
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

func toDomainQualificationType(row db.Qualification) domain.QualificationType {
	return domain.QualificationType{
		ID:                row.ID,
		Code:              row.Code,
		OriginalDutchText: row.OriginalDutchText,
		EnglishName:       row.EnglishName,
		AppContext:        row.AppContext,
		IsActive:          row.IsActive,
		CreatedAt:         conv.TimeFromPgTimestamptz(row.CreatedAt),
		UpdatedAt:         conv.TimeFromPgTimestamptz(row.UpdatedAt),
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

func contractTypePtrToString(ct *db.EmployeeContractTypeEnum) string {
	if ct == nil {
		return ""
	}
	return string(*ct)
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

func contractHoursTypeFromString(value string) db.ContractHoursTypeEnum {
	switch db.ContractHoursTypeEnum(value) {
	case db.ContractHoursTypeEnumFixed,
		db.ContractHoursTypeEnumZeroHours,
		db.ContractHoursTypeEnumMinMax:
		return db.ContractHoursTypeEnum(value)
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

func enumPtr[T any](value T) *T {
	return &value
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
