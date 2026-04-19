package handler

import (
	"time"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"

	"github.com/google/uuid"
)

type createEmployeeRequest struct {
	EmployeeNumber        *string    `json:"employee_number"`
	EmploymentNumber      *string    `json:"employment_number"`
	LocationID            *uuid.UUID `json:"location_id"`
	FirstName             string     `json:"first_name"              binding:"required"`
	LastName              string     `json:"last_name"               binding:"required"`
	Bsn                   string     `json:"bsn"                     binding:"required"`
	Street                string     `json:"street"                  binding:"required"`
	HouseNumber           string     `json:"house_number"            binding:"required"`
	HouseNumberAddition   *string    `json:"house_number_addition"`
	PostalCode            string     `json:"postal_code"             binding:"required"`
	City                  string     `json:"city"                    binding:"required"`
	Position              *string    `json:"position"`
	DepartmentID          *uuid.UUID `json:"department_id"`
	ManagerEmployeeID     *uuid.UUID `json:"manager_employee_id"`
	PrivateEmailAddress   *string    `json:"private_email_address"`
	WorkEmailAddress      string     `json:"work_email_address"      binding:"required,email"`
	WorkPhoneNumber       *string    `json:"work_phone_number"`
	PrivatePhoneNumber    *string    `json:"private_phone_number"`
	DateOfBirth           *string    `json:"date_of_birth"`
	HomeTelephoneNumber   *string    `json:"home_telephone_number"`
	Gender                string     `json:"gender"                  binding:"required,oneof=male female not_specified"`
	ContractHours         *float64   `json:"contract_hours"`
	ContractStartDate     *string    `json:"contract_start_date"`
	ContractEndDate       *string    `json:"contract_end_date"`
	ContractType          string     `json:"contract_type"           binding:"required,oneof=loondienst ZZP none"`
	ContractRate          *float64   `json:"contract_rate"`
	IrregularHoursProfile string     `json:"irregular_hours_profile" binding:"required,oneof=none roster non_roster"`
	RoleID                uuid.UUID  `json:"role_id"                 binding:"required"`
}

type updateEmployeeRequest struct {
	FirstName             *string    `json:"first_name"`
	LastName              *string    `json:"last_name"`
	Position              *string    `json:"position"`
	DepartmentID          *uuid.UUID `json:"department_id"`
	ManagerEmployeeID     *uuid.UUID `json:"manager_employee_id"`
	EmployeeNumber        *string    `json:"employee_number"`
	EmploymentNumber      *string    `json:"employment_number"`
	PrivateEmailAddress   *string    `json:"private_email_address"`
	PrivatePhoneNumber    *string    `json:"private_phone_number"`
	WorkPhoneNumber       *string    `json:"work_phone_number"`
	DateOfBirth           *string    `json:"date_of_birth"`
	HomeTelephoneNumber   *string    `json:"home_telephone_number"`
	Gender                *string    `json:"gender"`
	LocationID            *uuid.UUID `json:"location_id"`
	IrregularHoursProfile *string    `json:"irregular_hours_profile" binding:"omitempty,oneof=none roster non_roster"`
	HasBorrowed           *bool      `json:"has_borrowed"`
	OutOfService          *bool      `json:"out_of_service"`
	IsArchived            *bool      `json:"is_archived"`
}

type listEmployeesRequest struct {
	httpapi.PageRequest
	IncludeArchived     *bool      `form:"is_archived"`
	IncludeOutOfService *bool      `form:"out_of_service"`
	LocationID          *uuid.UUID `form:"location_id,parser=encoding.TextUnmarshaler"`
	ContractType        *string    `form:"contract_type" binding:"omitempty,oneof=loondienst ZZP none"`
	Search              *string    `form:"search"`
}

type setProfilePictureRequest struct {
	AttachmentID string `json:"attachement_id" binding:"required"`
}

type updateIsSubcontractorRequest struct {
	IsSubcontractor *bool `json:"is_subcontractor" binding:"required"`
}

type addContractDetailsRequest struct {
	ContractHours         *float64 `json:"contract_hours"          binding:"required"`
	ContractStartDate     *string  `json:"contract_start_date"     binding:"required"`
	ContractEndDate       *string  `json:"contract_end_date"       binding:"required"`
	ContractRate          *float64 `json:"contract_rate"`
	IrregularHoursProfile string   `json:"irregular_hours_profile" binding:"required,oneof=none roster non_roster"`
}

type createContractChangeRequest struct {
	EffectiveFrom         string   `json:"effective_from"          binding:"required,datetime=2006-01-02"`
	ContractHours         float64  `json:"contract_hours"          binding:"required"`
	ContractType          string   `json:"contract_type"           binding:"required,oneof=loondienst ZZP none"`
	ContractRate          *float64 `json:"contract_rate"`
	IrregularHoursProfile string   `json:"irregular_hours_profile" binding:"required,oneof=none roster non_roster"`
	ContractEndDate       *string  `json:"contract_end_date"       binding:"omitempty,datetime=2006-01-02"`
}

type createEducationRequest struct {
	InstitutionName string `json:"institution_name" binding:"required"`
	Degree          string `json:"degree"           binding:"required"`
	FieldOfStudy    string `json:"field_of_study"   binding:"required"`
	StartDate       string `json:"start_date"       binding:"required"`
	EndDate         string `json:"end_date"         binding:"required"`
}

type updateEducationRequest struct {
	InstitutionName *string `json:"institution_name"`
	Degree          *string `json:"degree"`
	FieldOfStudy    *string `json:"field_of_study"`
	StartDate       *string `json:"start_date"`
	EndDate         *string `json:"end_date"`
}

type createExperienceRequest struct {
	JobTitle    string  `json:"job_title"    binding:"required"`
	CompanyName string  `json:"company_name" binding:"required"`
	StartDate   string  `json:"start_date"   binding:"required"`
	EndDate     string  `json:"end_date"     binding:"required"`
	Description *string `json:"description"`
}

type updateExperienceRequest struct {
	JobTitle    *string `json:"job_title"`
	CompanyName *string `json:"company_name"`
	StartDate   *string `json:"start_date"`
	EndDate     *string `json:"end_date"`
	Description *string `json:"description"`
}

type createCertificationRequest struct {
	Name       string `json:"name"        binding:"required"`
	IssuedBy   string `json:"issued_by"   binding:"required"`
	DateIssued string `json:"date_issued" binding:"required"`
}

type updateCertificationRequest struct {
	Name       *string `json:"name"`
	IssuedBy   *string `json:"issued_by"`
	DateIssued *string `json:"date_issued"`
}

type searchEmployeesRequest struct {
	Search *string `form:"search" binding:"required"`
}

type employeeDetailResponse struct {
	ID                    uuid.UUID  `json:"id"`
	UserID                uuid.UUID  `json:"user_id"`
	FirstName             string     `json:"first_name"`
	LastName              string     `json:"last_name"`
	Bsn                   string     `json:"bsn"`
	Street                string     `json:"street"`
	HouseNumber           string     `json:"house_number"`
	HouseNumberAddition   *string    `json:"house_number_addition"`
	PostalCode            string     `json:"postal_code"`
	City                  string     `json:"city"`
	Position              *string    `json:"position"`
	EmployeeNumber        *string    `json:"employee_number"`
	EmploymentNumber      *string    `json:"employment_number"`
	PrivateEmailAddress   *string    `json:"private_email_address"`
	WorkEmailAddress      *string    `json:"work_email_address"`
	PrivatePhoneNumber    *string    `json:"private_phone_number"`
	WorkPhoneNumber       *string    `json:"work_phone_number"`
	DateOfBirth           *time.Time `json:"date_of_birth"`
	HomeTelephoneNumber   *string    `json:"home_telephone_number"`
	CreatedAt             time.Time  `json:"created_at"`
	Gender                string     `json:"gender"`
	LocationID            *uuid.UUID `json:"location_id"`
	DepartmentID          *uuid.UUID `json:"department_id"`
	ManagerEmployeeID     *uuid.UUID `json:"manager_employee_id"`
	HasBorrowed           bool       `json:"has_borrowed"`
	OutOfService          *bool      `json:"out_of_service"`
	IsArchived            bool       `json:"is_archived"`
	ContractHours         *float64   `json:"contract_hours"`
	ContractEndDate       *time.Time `json:"contract_end_date"`
	ContractStartDate     *time.Time `json:"contract_start_date"`
	ContractType          string     `json:"contract_type"`
	ContractRate          *float64   `json:"contract_rate"`
	IrregularHoursProfile string     `json:"irregular_hours_profile"`
	ProfilePicture        *string    `json:"profile_picture"`
	DepartmentName        *string    `json:"department_name"`
	ManagerFirstName      *string    `json:"manager_first_name"`
	ManagerLastName       *string    `json:"manager_last_name"`
}

type employeeListItemResponse struct {
	ID              uuid.UUID  `json:"id"`
	FirstName       string     `json:"first_name"`
	LastName        string     `json:"last_name"`
	Bsn             string     `json:"bsn"`
	ContractType    string     `json:"contract_type"`
	DepartmentName  *string    `json:"department_name"`
	LocationAddress string     `json:"location_address"`
	ContractEndDate *time.Time `json:"contract_end_date"`
}

type permissionResponse struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Resource string    `json:"resource"`
	Method   string    `json:"method"`
}

type employeeProfileResponse struct {
	UserID           uuid.UUID            `json:"user_id"`
	Email            string               `json:"email"`
	LastLogin        time.Time            `json:"last_login"`
	TwoFactorEnabled bool                 `json:"two_factor_enabled"`
	EmployeeID       uuid.UUID            `json:"employee_id"`
	FirstName        string               `json:"first_name"`
	LastName         string               `json:"last_name"`
	Permissions      []permissionResponse `json:"permissions"`
}

type employeeCountsResponse struct {
	TotalEmployees      int64 `json:"total_employees"`
	TotalSubcontractors int64 `json:"total_subcontractors"`
	TotalArchived       int64 `json:"total_archived"`
	TotalOutOfService   int64 `json:"total_out_of_service"`
}

type setProfilePictureResponse struct {
	ID             uuid.UUID `json:"id"`
	Email          string    `json:"email"`
	ProfilePicture *string   `json:"profile_picture"`
}

type contractDetailsResponse struct {
	ContractHours         *float64  `json:"contract_hours"`
	ContractStartDate     time.Time `json:"contract_start_date"`
	ContractEndDate       time.Time `json:"contract_end_date"`
	ContractType          string    `json:"contract_type"`
	ContractRate          *float64  `json:"contract_rate"`
	IrregularHoursProfile string    `json:"irregular_hours_profile"`
	IsSubcontractor       *bool     `json:"is_subcontractor"`
}

type contractChangeResponse struct {
	ID                    uuid.UUID  `json:"id"`
	EmployeeID            uuid.UUID  `json:"employee_id"`
	EffectiveFrom         time.Time  `json:"effective_from"`
	EffectiveTo           *time.Time `json:"effective_to,omitempty"`
	ContractHours         float64    `json:"contract_hours"`
	ContractType          string     `json:"contract_type"`
	ContractRate          *float64   `json:"contract_rate,omitempty"`
	IrregularHoursProfile string     `json:"irregular_hours_profile"`
	ContractEndDate       *time.Time `json:"contract_end_date,omitempty"`
	CreatedByEmployeeID   uuid.UUID  `json:"created_by_employee_id"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type leaveRecalculationImpactResponse struct {
	Year        int32 `json:"year"`
	LegalBefore int32 `json:"legal_before"`
	LegalAfter  int32 `json:"legal_after"`
	Delta       int32 `json:"delta"`
}

type createContractChangeResponse struct {
	Change         contractChangeResponse             `json:"change"`
	Recalculations []leaveRecalculationImpactResponse `json:"recalculations"`
}

type educationResponse struct {
	ID              uuid.UUID `json:"id"`
	EmployeeID      uuid.UUID `json:"employee_id"`
	InstitutionName string    `json:"institution_name"`
	Degree          string    `json:"degree"`
	FieldOfStudy    string    `json:"field_of_study"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	CreatedAt       time.Time `json:"created_at"`
}

type experienceResponse struct {
	ID          uuid.UUID `json:"id"`
	EmployeeID  uuid.UUID `json:"employee_id"`
	JobTitle    string    `json:"job_title"`
	CompanyName string    `json:"company_name"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type certificationResponse struct {
	ID         uuid.UUID `json:"id"`
	EmployeeID uuid.UUID `json:"employee_id"`
	Name       string    `json:"name"`
	IssuedBy   string    `json:"issued_by"`
	DateIssued time.Time `json:"date_issued"`
	CreatedAt  time.Time `json:"created_at"`
}

type employeeSearchResultResponse struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     *string   `json:"email"`
}

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

func parseDatePtr(s *string) (*time.Time, error) {
	if s == nil {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func toCreateEmployeeParams(req createEmployeeRequest) domain.CreateEmployeeParams {
	dateOfBirth, _ := parseDatePtr(req.DateOfBirth)
	contractStartDate, _ := parseDatePtr(req.ContractStartDate)
	contractEndDate, _ := parseDatePtr(req.ContractEndDate)

	return domain.CreateEmployeeParams{
		FirstName:             req.FirstName,
		LastName:              req.LastName,
		Bsn:                   req.Bsn,
		Street:                req.Street,
		HouseNumber:           req.HouseNumber,
		HouseNumberAddition:   req.HouseNumberAddition,
		PostalCode:            req.PostalCode,
		City:                  req.City,
		Position:              req.Position,
		DepartmentID:          req.DepartmentID,
		ManagerEmployeeID:     req.ManagerEmployeeID,
		EmployeeNumber:        req.EmployeeNumber,
		EmploymentNumber:      req.EmploymentNumber,
		PrivateEmailAddress:   req.PrivateEmailAddress,
		WorkEmailAddress:      &req.WorkEmailAddress,
		WorkPhoneNumber:       req.WorkPhoneNumber,
		PrivatePhoneNumber:    req.PrivatePhoneNumber,
		DateOfBirth:           dateOfBirth,
		HomeTelephoneNumber:   req.HomeTelephoneNumber,
		Gender:                req.Gender,
		LocationID:            req.LocationID,
		ContractHours:         req.ContractHours,
		ContractType:          req.ContractType,
		ContractStartDate:     contractStartDate,
		ContractEndDate:       contractEndDate,
		ContractRate:          req.ContractRate,
		IrregularHoursProfile: req.IrregularHoursProfile,
		RoleID:                req.RoleID,
		UserEmail:             req.WorkEmailAddress,
		UserPassword:          "",
	}
}

func toUpdateEmployeeParams(req updateEmployeeRequest) domain.UpdateEmployeeParams {
	dateOfBirth, _ := parseDatePtr(req.DateOfBirth)

	return domain.UpdateEmployeeParams{
		FirstName:             req.FirstName,
		LastName:              req.LastName,
		Position:              req.Position,
		DepartmentID:          req.DepartmentID,
		ManagerEmployeeID:     req.ManagerEmployeeID,
		EmployeeNumber:        req.EmployeeNumber,
		EmploymentNumber:      req.EmploymentNumber,
		PrivateEmailAddress:   req.PrivateEmailAddress,
		PrivatePhoneNumber:    req.PrivatePhoneNumber,
		WorkPhoneNumber:       req.WorkPhoneNumber,
		DateOfBirth:           dateOfBirth,
		HomeTelephoneNumber:   req.HomeTelephoneNumber,
		Gender:                req.Gender,
		LocationID:            req.LocationID,
		IrregularHoursProfile: req.IrregularHoursProfile,
		HasBorrowed:           req.HasBorrowed,
		OutOfService:          req.OutOfService,
		IsArchived:            req.IsArchived,
	}
}

func toListEmployeesParams(req listEmployeesRequest) domain.ListEmployeesParams {
	return domain.ListEmployeesParams{
		Limit:               req.PageSize,
		Offset:              (req.Page - 1) * req.PageSize,
		IncludeArchived:     req.IncludeArchived,
		IncludeOutOfService: req.IncludeOutOfService,
		LocationID:          req.LocationID,
		ContractType:        req.ContractType,
		Search:              req.Search,
	}
}

func toCreateEducationParams(req createEducationRequest) domain.CreateEducationParams {
	startDate, _ := parseDate(req.StartDate)
	endDate, _ := parseDate(req.EndDate)

	return domain.CreateEducationParams{
		InstitutionName: req.InstitutionName,
		Degree:          req.Degree,
		FieldOfStudy:    req.FieldOfStudy,
		StartDate:       startDate,
		EndDate:         endDate,
	}
}

func toUpdateEducationParams(req updateEducationRequest) domain.UpdateEducationParams {
	startDate, _ := parseDatePtr(req.StartDate)
	endDate, _ := parseDatePtr(req.EndDate)

	return domain.UpdateEducationParams{
		InstitutionName: req.InstitutionName,
		Degree:          req.Degree,
		FieldOfStudy:    req.FieldOfStudy,
		StartDate:       startDate,
		EndDate:         endDate,
	}
}

func toCreateExperienceParams(req createExperienceRequest) domain.CreateExperienceParams {
	startDate, _ := parseDate(req.StartDate)
	endDate, _ := parseDate(req.EndDate)

	return domain.CreateExperienceParams{
		JobTitle:    req.JobTitle,
		CompanyName: req.CompanyName,
		StartDate:   startDate,
		EndDate:     endDate,
		Description: req.Description,
	}
}

func toUpdateExperienceParams(req updateExperienceRequest) domain.UpdateExperienceParams {
	startDate, _ := parseDatePtr(req.StartDate)
	endDate, _ := parseDatePtr(req.EndDate)

	return domain.UpdateExperienceParams{
		JobTitle:    req.JobTitle,
		CompanyName: req.CompanyName,
		StartDate:   startDate,
		EndDate:     endDate,
		Description: req.Description,
	}
}

func toCreateCertificationParams(req createCertificationRequest) domain.CreateCertificationParams {
	dateIssued, _ := parseDate(req.DateIssued)

	return domain.CreateCertificationParams{
		Name:       req.Name,
		IssuedBy:   req.IssuedBy,
		DateIssued: dateIssued,
	}
}

func toUpdateCertificationParams(req updateCertificationRequest) domain.UpdateCertificationParams {
	dateIssued, _ := parseDatePtr(req.DateIssued)

	return domain.UpdateCertificationParams{
		Name:       req.Name,
		IssuedBy:   req.IssuedBy,
		DateIssued: dateIssued,
	}
}

func toAddContractDetailsParams(req addContractDetailsRequest) domain.AddContractDetailsParams {
	contractStartDate, _ := parseDate(*req.ContractStartDate)
	contractEndDate, _ := parseDate(*req.ContractEndDate)

	return domain.AddContractDetailsParams{
		ContractHours:         req.ContractHours,
		ContractStartDate:     contractStartDate,
		ContractEndDate:       contractEndDate,
		ContractRate:          req.ContractRate,
		IrregularHoursProfile: req.IrregularHoursProfile,
	}
}

func toCreateContractChangeParams(
	req createContractChangeRequest,
) (domain.CreateEmployeeContractChangeParams, error) {
	effectiveFrom, err := parseDate(req.EffectiveFrom)
	if err != nil {
		return domain.CreateEmployeeContractChangeParams{}, err
	}
	contractEndDate, err := parseDatePtr(req.ContractEndDate)
	if err != nil {
		return domain.CreateEmployeeContractChangeParams{}, err
	}

	return domain.CreateEmployeeContractChangeParams{
		EffectiveFrom:         effectiveFrom,
		ContractHours:         req.ContractHours,
		ContractType:          req.ContractType,
		ContractRate:          req.ContractRate,
		IrregularHoursProfile: req.IrregularHoursProfile,
		ContractEndDate:       contractEndDate,
	}, nil
}

func toEmployeeDetailResponse(emp *domain.EmployeeDetail) employeeDetailResponse {
	return employeeDetailResponse{
		ID:                    emp.ID,
		UserID:                emp.UserID,
		FirstName:             emp.FirstName,
		LastName:              emp.LastName,
		Bsn:                   emp.Bsn,
		Street:                emp.Street,
		HouseNumber:           emp.HouseNumber,
		HouseNumberAddition:   emp.HouseNumberAddition,
		PostalCode:            emp.PostalCode,
		City:                  emp.City,
		Position:              emp.Position,
		EmployeeNumber:        emp.EmployeeNumber,
		EmploymentNumber:      emp.EmploymentNumber,
		PrivateEmailAddress:   emp.PrivateEmailAddress,
		WorkEmailAddress:      emp.WorkEmailAddress,
		PrivatePhoneNumber:    emp.PrivatePhoneNumber,
		WorkPhoneNumber:       emp.WorkPhoneNumber,
		DateOfBirth:           emp.DateOfBirth,
		HomeTelephoneNumber:   emp.HomeTelephoneNumber,
		CreatedAt:             emp.CreatedAt,
		Gender:                emp.Gender,
		LocationID:            emp.LocationID,
		DepartmentID:          emp.DepartmentID,
		ManagerEmployeeID:     emp.ManagerEmployeeID,
		HasBorrowed:           emp.HasBorrowed,
		OutOfService:          emp.OutOfService,
		IsArchived:            emp.IsArchived,
		ContractHours:         emp.ContractHours,
		ContractEndDate:       emp.ContractEndDate,
		ContractStartDate:     emp.ContractStartDate,
		ContractType:          emp.ContractType,
		ContractRate:          emp.ContractRate,
		IrregularHoursProfile: emp.IrregularHoursProfile,
		ProfilePicture:        emp.ProfilePicture,
		DepartmentName:        emp.DepartmentName,
		ManagerFirstName:      emp.ManagerFirstName,
		ManagerLastName:       emp.ManagerLastName,
	}
}

func toEmployeeListItemResponse(emp domain.Employee) employeeListItemResponse {
	return employeeListItemResponse{
		ID:              emp.ID,
		FirstName:       emp.FirstName,
		LastName:        emp.LastName,
		Bsn:             emp.Bsn,
		ContractType:    emp.ContractType,
		DepartmentName:  emp.DepartmentName,
		LocationAddress: emp.LocationAddress,
		ContractEndDate: emp.ContractEndDate,
	}
}

func toEmployeeProfileResponse(profile *domain.EmployeeProfile) employeeProfileResponse {
	permissions := make([]permissionResponse, len(profile.Permissions))
	for i, permission := range profile.Permissions {
		permissions[i] = permissionResponse{
			ID:       permission.ID,
			Name:     permission.Name,
			Resource: permission.Resource,
			Method:   permission.Method,
		}
	}

	return employeeProfileResponse{
		UserID:           profile.UserID,
		Email:            profile.Email,
		LastLogin:        profile.LastLogin,
		TwoFactorEnabled: profile.TwoFactorEnabled,
		EmployeeID:       profile.EmployeeID,
		FirstName:        profile.FirstName,
		LastName:         profile.LastName,
		Permissions:      permissions,
	}
}

func toEmployeeCountsResponse(counts *domain.EmployeeCounts) employeeCountsResponse {
	return employeeCountsResponse{
		TotalEmployees:      counts.TotalEmployees,
		TotalSubcontractors: counts.TotalSubcontractors,
		TotalArchived:       counts.TotalArchived,
		TotalOutOfService:   counts.TotalOutOfService,
	}
}

func toContractDetailsResponse(details *domain.ContractDetails) contractDetailsResponse {
	return contractDetailsResponse{
		ContractHours:         details.ContractHours,
		ContractStartDate:     details.ContractStartDate,
		ContractEndDate:       details.ContractEndDate,
		ContractType:          details.ContractType,
		ContractRate:          details.ContractRate,
		IrregularHoursProfile: details.IrregularHoursProfile,
		IsSubcontractor:       details.IsSubcontractor,
	}
}

func toContractChangeResponse(item domain.EmployeeContractChange) contractChangeResponse {
	return contractChangeResponse{
		ID:                    item.ID,
		EmployeeID:            item.EmployeeID,
		EffectiveFrom:         item.EffectiveFrom,
		EffectiveTo:           item.EffectiveTo,
		ContractHours:         item.ContractHours,
		ContractType:          item.ContractType,
		ContractRate:          item.ContractRate,
		IrregularHoursProfile: item.IrregularHoursProfile,
		ContractEndDate:       item.ContractEndDate,
		CreatedByEmployeeID:   item.CreatedByEmployeeID,
		CreatedAt:             item.CreatedAt,
		UpdatedAt:             item.UpdatedAt,
	}
}

func toContractChangeResponses(items []domain.EmployeeContractChange) []contractChangeResponse {
	results := make([]contractChangeResponse, len(items))
	for i, item := range items {
		results[i] = toContractChangeResponse(item)
	}
	return results
}

func toCreateContractChangeResponse(
	result *domain.CreateEmployeeContractChangeResult,
) createContractChangeResponse {
	recalcs := make([]leaveRecalculationImpactResponse, len(result.Recalculations))
	for i, impact := range result.Recalculations {
		recalcs[i] = leaveRecalculationImpactResponse{
			Year:        impact.Year,
			LegalBefore: impact.LegalBefore,
			LegalAfter:  impact.LegalAfter,
			Delta:       impact.Delta,
		}
	}

	return createContractChangeResponse{
		Change:         toContractChangeResponse(result.Change),
		Recalculations: recalcs,
	}
}

func toEducationResponse(education *domain.Education) educationResponse {
	return educationResponse{
		ID:              education.ID,
		EmployeeID:      education.EmployeeID,
		InstitutionName: education.InstitutionName,
		Degree:          education.Degree,
		FieldOfStudy:    education.FieldOfStudy,
		StartDate:       education.StartDate,
		EndDate:         education.EndDate,
		CreatedAt:       education.CreatedAt,
	}
}

func toExperienceResponse(experience *domain.Experience) experienceResponse {
	return experienceResponse{
		ID:          experience.ID,
		EmployeeID:  experience.EmployeeID,
		JobTitle:    experience.JobTitle,
		CompanyName: experience.CompanyName,
		StartDate:   experience.StartDate,
		EndDate:     experience.EndDate,
		Description: experience.Description,
		CreatedAt:   experience.CreatedAt,
	}
}

func toCertificationResponse(certification *domain.Certification) certificationResponse {
	return certificationResponse{
		ID:         certification.ID,
		EmployeeID: certification.EmployeeID,
		Name:       certification.Name,
		IssuedBy:   certification.IssuedBy,
		DateIssued: certification.DateIssued,
		CreatedAt:  certification.CreatedAt,
	}
}

func toEmployeeSearchResultResponse(
	result domain.EmployeeSearchResult,
) employeeSearchResultResponse {
	return employeeSearchResultResponse{
		ID:        result.ID,
		FirstName: result.FirstName,
		LastName:  result.LastName,
		Email:     result.WorkEmailAddress,
	}
}
