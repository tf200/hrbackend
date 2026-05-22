package handler

import (
	"time"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"

	"github.com/google/uuid"
)

type createEmployeeContractRequest struct {
	JobTitle             string     `json:"job_title" binding:"required,oneof=youth_worker_d care_coordinator behavioral_scientist quality_officer pedagogical_worker team_lead manager administrative_employee"`
	DepartmentID         uuid.UUID  `json:"department_id" binding:"required"`
	LocationID           uuid.UUID  `json:"location_id" binding:"required"`
	OrganizationalRoleID *uuid.UUID `json:"organizational_role_id"`
	ContractType         string     `json:"contract_type" binding:"required,oneof=permanent temporary on_call"`
	ContractHoursType    string     `json:"contract_hours_type" binding:"required,oneof=fixed zero_hours min_max"`
	StartDate            string     `json:"start_date" binding:"required,datetime=2006-01-02"`
	ContractEndDate      *string    `json:"contract_end_date" binding:"omitempty,datetime=2006-01-02"`
	HoursPerWeek         *float64   `json:"hours_per_week" binding:"omitempty,min=0,max=40"`
	MinHoursPerWeek      *float64   `json:"min_hours_per_week" binding:"omitempty,min=0,max=40"`
	MaxHoursPerWeek      *float64   `json:"max_hours_per_week" binding:"omitempty,min=0,max=40"`
	RosterFreeDay        *int16     `json:"roster_free_day" binding:"omitempty,min=0,max=6"`
	WageTaxTable         *string    `json:"wage_tax_table" binding:"omitempty,oneof=white_table green_table"`
}

type createEmployeeSalaryAssignmentRequest struct {
	SalaryScaleStepID uuid.UUID `json:"salary_scale_step_id" binding:"required"`
	EffectiveFrom     *string   `json:"effective_from" binding:"omitempty,datetime=2006-01-02"`
	EffectiveTo       *string   `json:"effective_to" binding:"omitempty,datetime=2006-01-02"`
}

type createEmployeeRequest struct {
	EmployeeNumber      *string                                `json:"employee_number"`
	EmploymentNumber    *string                                `json:"employment_number"`
	FirstName           string                                 `json:"first_name"              binding:"required"`
	LastName            string                                 `json:"last_name"               binding:"required"`
	Bsn                 string                                 `json:"bsn"                     binding:"required"`
	Street              string                                 `json:"street"                  binding:"required"`
	HouseNumber         string                                 `json:"house_number"            binding:"required"`
	HouseNumberAddition *string                                `json:"house_number_addition"`
	PostalCode          string                                 `json:"postal_code"             binding:"required"`
	City                string                                 `json:"city"                    binding:"required"`
	ManagerEmployeeID   *uuid.UUID                             `json:"manager_employee_id"`
	PrivateEmailAddress *string                                `json:"private_email_address"`
	WorkEmailAddress    string                                 `json:"work_email_address"      binding:"required,email"`
	WorkPhoneNumber     *string                                `json:"work_phone_number"`
	PrivatePhoneNumber  *string                                `json:"private_phone_number"`
	DateOfBirth         *string                                `json:"date_of_birth" binding:"omitempty,datetime=2006-01-02"`
	HomeTelephoneNumber *string                                `json:"home_telephone_number"`
	Gender              string                                 `json:"gender"                  binding:"required,oneof=male female not_specified"`
	RoleID              uuid.UUID                              `json:"role_id"                 binding:"required"`
	Contract            *createEmployeeContractRequest         `json:"contract,omitempty"`
	SalaryAssignment    *createEmployeeSalaryAssignmentRequest `json:"salary_assignment,omitempty"`
}

type updateEmployeeRequest struct {
	FirstName           *string    `json:"first_name"`
	LastName            *string    `json:"last_name"`
	ManagerEmployeeID   *uuid.UUID `json:"manager_employee_id"`
	EmployeeNumber      *string    `json:"employee_number"`
	EmploymentNumber    *string    `json:"employment_number"`
	PrivateEmailAddress *string    `json:"private_email_address"`
	PrivatePhoneNumber  *string    `json:"private_phone_number"`
	WorkPhoneNumber     *string    `json:"work_phone_number"`
	DateOfBirth         *string    `json:"date_of_birth"`
	HomeTelephoneNumber *string    `json:"home_telephone_number"`
	Gender              *string    `json:"gender"`
	OutOfService        *bool      `json:"out_of_service"`
	IsArchived          *bool      `json:"is_archived"`
}

type listEmployeesRequest struct {
	httpapi.PageRequest
	IncludeArchived     *bool      `form:"is_archived"`
	IncludeOutOfService *bool      `form:"out_of_service"`
	LocationID          *uuid.UUID `form:"location_id,parser=encoding.TextUnmarshaler"`
	ContractType        *string    `form:"contract_type" binding:"omitempty,oneof=permanent temporary on_call"`
	Search              *string    `form:"search"`
}

type setProfilePictureRequest struct {
	AttachmentID string `json:"attachement_id" binding:"required"`
}

type updateIsSubcontractorRequest struct {
	IsSubcontractor *bool `json:"is_subcontractor" binding:"required"`
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

type createQualificationRequest struct {
	QualificationTypeCode string  `json:"qualification_type_code" binding:"required"`
	AchievedOn            string  `json:"achieved_on"             binding:"required"`
	ExpirationDate        *string `json:"expiration_date"`
	CertificateNumber     *string `json:"certificate_number"`
}

type updateQualificationRequest struct {
	QualificationTypeCode *string `json:"qualification_type_code"`
	AchievedOn            *string `json:"achieved_on"`
	ExpirationDate        *string `json:"expiration_date"`
	CertificateNumber     *string `json:"certificate_number"`
}

type searchEmployeesRequest struct {
	Search *string `form:"search" binding:"required"`
}

type resetPasswordRequest struct {
	Generated *bool   `json:"generated" binding:"required"`
	Password  *string `json:"password"`
	SendEmail bool    `json:"send_email"`
}

type resetPasswordResponse struct {
	TemporaryPassword string `json:"temporary_password"`
}

type employeeDetailResponse struct {
	ID                         uuid.UUID  `json:"id"`
	UserID                     uuid.UUID  `json:"user_id"`
	FirstName                  string     `json:"first_name"`
	LastName                   string     `json:"last_name"`
	Bsn                        string     `json:"bsn"`
	Street                     string     `json:"street"`
	HouseNumber                string     `json:"house_number"`
	HouseNumberAddition        *string    `json:"house_number_addition"`
	PostalCode                 string     `json:"postal_code"`
	City                       string     `json:"city"`
	Position                   *string    `json:"position"`
	EmployeeNumber             *string    `json:"employee_number"`
	EmploymentNumber           *string    `json:"employment_number"`
	PrivateEmailAddress        *string    `json:"private_email_address"`
	WorkEmailAddress           *string    `json:"work_email_address"`
	PrivatePhoneNumber         *string    `json:"private_phone_number"`
	WorkPhoneNumber            *string    `json:"work_phone_number"`
	DateOfBirth                *time.Time `json:"date_of_birth"`
	HomeTelephoneNumber        *string    `json:"home_telephone_number"`
	CreatedAt                  time.Time  `json:"created_at"`
	Gender                     string     `json:"gender"`
	LocationID                 *uuid.UUID `json:"location_id"`
	DepartmentID               *uuid.UUID `json:"department_id"`
	ManagerEmployeeID          *uuid.UUID `json:"manager_employee_id"`
	HasBorrowed                bool       `json:"has_borrowed"`
	OutOfService               *bool      `json:"out_of_service"`
	IsArchived                 bool       `json:"is_archived"`
	ContractHours              *float64   `json:"contract_hours"`
	ContractEndDate            *time.Time `json:"contract_end_date"`
	ContractStartDate          *time.Time `json:"contract_start_date"`
	ContractType               string     `json:"contract_type"`
	ContractRate               *float64   `json:"contract_rate"`
	IrregularHoursProfile      string     `json:"irregular_hours_profile"`
	ProfilePicture             *string    `json:"profile_picture"`
	DepartmentName             *string    `json:"department_name"`
	ManagerFirstName           *string    `json:"manager_first_name"`
	ManagerLastName            *string    `json:"manager_last_name"`
	RemainingLeaveBalanceHours int32      `json:"remaining_leave_balance_hours"`
	HoursWorkedThisMonth       float64    `json:"hours_worked_this_month"`
	HoursPendingApproval       float64    `json:"hours_pending_approval"`
	TotalHoursWorkedThisYear   float64    `json:"total_hours_worked_this_year"`
	LastPerformanceReviewScore *float64   `json:"last_performance_review_score"`
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
	Role             string               `json:"role"`
	RoleID           uuid.UUID            `json:"role_id"`
	PortalAccess     string               `json:"portal_access"`
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

type qualificationResponse struct {
	ID                    uuid.UUID  `json:"id"`
	EmployeeID            uuid.UUID  `json:"employee_id"`
	QualificationTypeCode string     `json:"qualification_type_code"`
	AchievedOn            time.Time  `json:"achieved_on"`
	ExpirationDate        *time.Time `json:"expiration_date"`
	CertificateNumber     *string    `json:"certificate_number"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type qualificationTypeResponse struct {
	Code              string    `json:"code"`
	OriginalDutchText string    `json:"original_dutch_text"`
	EnglishName       string    `json:"english_name"`
	AppContext        string    `json:"app_context"`
	CreatedAt         time.Time `json:"created_at"`
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

	var contract *domain.CreateEmployeeContractParams
	if req.Contract != nil {
		startDate, _ := parseDate(req.Contract.StartDate)
		contractEndDate, _ := parseDatePtr(req.Contract.ContractEndDate)
		contract = &domain.CreateEmployeeContractParams{
			JobTitle:             req.Contract.JobTitle,
			DepartmentID:         req.Contract.DepartmentID,
			LocationID:           req.Contract.LocationID,
			OrganizationalRoleID: req.Contract.OrganizationalRoleID,
			ContractType:         req.Contract.ContractType,
			ContractHoursType:    req.Contract.ContractHoursType,
			StartDate:            startDate,
			ContractEndDate:      contractEndDate,
			HoursPerWeek:         req.Contract.HoursPerWeek,
			MinHoursPerWeek:      req.Contract.MinHoursPerWeek,
			MaxHoursPerWeek:      req.Contract.MaxHoursPerWeek,
			RosterFreeDay:        req.Contract.RosterFreeDay,
			WageTaxTable:         req.Contract.WageTaxTable,
		}
	}

	var salaryAssignment *domain.CreateEmployeeSalaryAssignmentParams
	if req.SalaryAssignment != nil {
		effectiveFrom, _ := parseDatePtr(req.SalaryAssignment.EffectiveFrom)
		effectiveTo, _ := parseDatePtr(req.SalaryAssignment.EffectiveTo)
		salaryAssignment = &domain.CreateEmployeeSalaryAssignmentParams{
			SalaryScaleStepID: req.SalaryAssignment.SalaryScaleStepID,
			EffectiveFrom:     effectiveFrom,
			EffectiveTo:       effectiveTo,
		}
	}

	return domain.CreateEmployeeParams{
		FirstName:           req.FirstName,
		LastName:            req.LastName,
		Bsn:                 req.Bsn,
		Street:              req.Street,
		HouseNumber:         req.HouseNumber,
		HouseNumberAddition: req.HouseNumberAddition,
		PostalCode:          req.PostalCode,
		City:                req.City,
		ManagerEmployeeID:   req.ManagerEmployeeID,
		EmployeeNumber:      req.EmployeeNumber,
		EmploymentNumber:    req.EmploymentNumber,
		PrivateEmailAddress: req.PrivateEmailAddress,
		WorkEmailAddress:    &req.WorkEmailAddress,
		WorkPhoneNumber:     req.WorkPhoneNumber,
		PrivatePhoneNumber:  req.PrivatePhoneNumber,
		DateOfBirth:         dateOfBirth,
		HomeTelephoneNumber: req.HomeTelephoneNumber,
		Gender:              req.Gender,
		RoleID:              req.RoleID,
		UserEmail:           req.WorkEmailAddress,
		UserPassword:        "",
		Contract:            contract,
		SalaryAssignment:    salaryAssignment,
	}
}

func toUpdateEmployeeParams(req updateEmployeeRequest) domain.UpdateEmployeeParams {
	dateOfBirth, _ := parseDatePtr(req.DateOfBirth)

	return domain.UpdateEmployeeParams{
		FirstName:           req.FirstName,
		LastName:            req.LastName,
		ManagerEmployeeID:   req.ManagerEmployeeID,
		EmployeeNumber:      req.EmployeeNumber,
		EmploymentNumber:    req.EmploymentNumber,
		PrivateEmailAddress: req.PrivateEmailAddress,
		PrivatePhoneNumber:  req.PrivatePhoneNumber,
		WorkPhoneNumber:     req.WorkPhoneNumber,
		DateOfBirth:         dateOfBirth,
		HomeTelephoneNumber: req.HomeTelephoneNumber,
		Gender:              req.Gender,
		OutOfService:        req.OutOfService,
		IsArchived:          req.IsArchived,
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

func toCreateQualificationParams(req createQualificationRequest) domain.CreateQualificationParams {
	achievedOn, _ := parseDate(req.AchievedOn)
	expirationDate, _ := parseDatePtr(req.ExpirationDate)

	return domain.CreateQualificationParams{
		QualificationTypeCode: req.QualificationTypeCode,
		AchievedOn:            achievedOn,
		ExpirationDate:        expirationDate,
		CertificateNumber:     req.CertificateNumber,
	}
}

func toUpdateQualificationParams(req updateQualificationRequest) domain.UpdateQualificationParams {
	achievedOn, _ := parseDatePtr(req.AchievedOn)
	expirationDate, _ := parseDatePtr(req.ExpirationDate)

	return domain.UpdateQualificationParams{
		QualificationTypeCode: req.QualificationTypeCode,
		AchievedOn:            achievedOn,
		ExpirationDate:        expirationDate,
		CertificateNumber:     req.CertificateNumber,
	}
}

func toEmployeeDetailResponse(emp *domain.EmployeeDetail) employeeDetailResponse {
	return employeeDetailResponse{
		ID:                         emp.ID,
		UserID:                     emp.UserID,
		FirstName:                  emp.FirstName,
		LastName:                   emp.LastName,
		Bsn:                        emp.Bsn,
		Street:                     emp.Street,
		HouseNumber:                emp.HouseNumber,
		HouseNumberAddition:        emp.HouseNumberAddition,
		PostalCode:                 emp.PostalCode,
		City:                       emp.City,
		Position:                   emp.Position,
		EmployeeNumber:             emp.EmployeeNumber,
		EmploymentNumber:           emp.EmploymentNumber,
		PrivateEmailAddress:        emp.PrivateEmailAddress,
		WorkEmailAddress:           emp.WorkEmailAddress,
		PrivatePhoneNumber:         emp.PrivatePhoneNumber,
		WorkPhoneNumber:            emp.WorkPhoneNumber,
		DateOfBirth:                emp.DateOfBirth,
		HomeTelephoneNumber:        emp.HomeTelephoneNumber,
		CreatedAt:                  emp.CreatedAt,
		Gender:                     emp.Gender,
		LocationID:                 emp.LocationID,
		DepartmentID:               emp.DepartmentID,
		ManagerEmployeeID:          emp.ManagerEmployeeID,
		HasBorrowed:                emp.HasBorrowed,
		OutOfService:               emp.OutOfService,
		IsArchived:                 emp.IsArchived,
		ContractHours:              emp.ContractHours,
		ContractEndDate:            emp.ContractEndDate,
		ContractStartDate:          emp.ContractStartDate,
		ContractType:               emp.ContractType,
		ContractRate:               emp.ContractRate,
		IrregularHoursProfile:      emp.IrregularHoursProfile,
		ProfilePicture:             emp.ProfilePicture,
		DepartmentName:             emp.DepartmentName,
		ManagerFirstName:           emp.ManagerFirstName,
		ManagerLastName:            emp.ManagerLastName,
		RemainingLeaveBalanceHours: emp.RemainingLeaveBalanceHours,
		HoursWorkedThisMonth:       emp.HoursWorkedThisMonth,
		HoursPendingApproval:       emp.HoursPendingApproval,
		TotalHoursWorkedThisYear:   emp.TotalHoursWorkedThisYear,
		LastPerformanceReviewScore: emp.LastPerformanceReviewScore,
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
		Role:             profile.Role,
		RoleID:           profile.RoleID,
		PortalAccess:     profile.PortalAccess,
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

func toQualificationResponse(qualification *domain.Qualification) qualificationResponse {
	return qualificationResponse{
		ID:                    qualification.ID,
		EmployeeID:            qualification.EmployeeID,
		QualificationTypeCode: qualification.QualificationTypeCode,
		AchievedOn:            qualification.AchievedOn,
		ExpirationDate:        qualification.ExpirationDate,
		CertificateNumber:     qualification.CertificateNumber,
		CreatedAt:             qualification.CreatedAt,
		UpdatedAt:             qualification.UpdatedAt,
	}
}

func toQualificationTypeResponse(qt *domain.QualificationType) qualificationTypeResponse {
	return qualificationTypeResponse{
		Code:              qt.Code,
		OriginalDutchText: qt.OriginalDutchText,
		EnglishName:       qt.EnglishName,
		AppContext:        qt.AppContext,
		CreatedAt:         qt.CreatedAt,
	}
}

func toResetPasswordParams(req resetPasswordRequest) domain.ResetPasswordParams {
	generated := req.Generated != nil && *req.Generated
	return domain.ResetPasswordParams{
		Generated: generated,
		Password:  req.Password,
		SendEmail: req.SendEmail,
	}
}

func toResetPasswordResponse(result *domain.ResetPasswordResult) resetPasswordResponse {
	return resetPasswordResponse{
		TemporaryPassword: result.TemporaryPassword,
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
