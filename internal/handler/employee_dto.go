package handler

import (
	"fmt"
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
	StartDate            string     `json:"start_date" binding:"required,datetime=2006-01-02"`
	ContractEndDate      *string    `json:"contract_end_date" binding:"omitempty,datetime=2006-01-02"`
	HoursPerWeek         *float64   `json:"hours_per_week" binding:"omitempty,min=0,max=40"`
	RosterFreeDay        string     `json:"roster_free_day" binding:"required,oneof=monday tuesday wednesday thursday friday saturday sunday"`
	WageTaxTable         *string    `json:"wage_tax_table" binding:"omitempty,oneof=white_table green_table"`
}

type updateContractRequest struct {
	JobTitle             *string    `json:"job_title" binding:"omitempty,oneof=youth_worker_d care_coordinator behavioral_scientist quality_officer pedagogical_worker team_lead manager administrative_employee"`
	DepartmentID         *uuid.UUID `json:"department_id"`
	LocationID           *uuid.UUID `json:"location_id"`
	OrganizationalRoleID *uuid.UUID `json:"organizational_role_id"`
	ContractType         *string    `json:"contract_type" binding:"omitempty,oneof=permanent temporary on_call"`
	StartDate            *string    `json:"start_date" binding:"omitempty,datetime=2006-01-02"`
	ContractEndDate      *string    `json:"contract_end_date" binding:"omitempty,datetime=2006-01-02"`
	HoursPerWeek         *float64   `json:"hours_per_week" binding:"omitempty,min=0,max=40"`
	RosterFreeDay        *string    `json:"roster_free_day" binding:"omitempty,oneof=monday tuesday wednesday thursday friday saturday sunday"`
	WageTaxTable         *string    `json:"wage_tax_table" binding:"omitempty,oneof=white_table green_table"`
}

type createContractRequest struct {
	JobTitle             string     `json:"job_title" binding:"required,oneof=youth_worker_d care_coordinator behavioral_scientist quality_officer pedagogical_worker team_lead manager administrative_employee"`
	DepartmentID         uuid.UUID  `json:"department_id" binding:"required"`
	LocationID           uuid.UUID  `json:"location_id" binding:"required"`
	OrganizationalRoleID *uuid.UUID `json:"organizational_role_id"`
	ContractType         string     `json:"contract_type" binding:"required,oneof=permanent temporary on_call"`
	StartDate            string     `json:"start_date" binding:"required,datetime=2006-01-02"`
	ContractEndDate      *string    `json:"contract_end_date" binding:"omitempty,datetime=2006-01-02"`
	HoursPerWeek         *float64   `json:"hours_per_week" binding:"omitempty,min=0,max=40"`
	RosterFreeDay        string     `json:"roster_free_day" binding:"required,oneof=monday tuesday wednesday thursday friday saturday sunday"`
	WageTaxTable         *string    `json:"wage_tax_table" binding:"omitempty,oneof=white_table green_table"`
}

type createContractAmendmentRequest struct {
	JobTitle             string     `json:"job_title" binding:"required,oneof=youth_worker_d care_coordinator behavioral_scientist quality_officer pedagogical_worker team_lead manager administrative_employee"`
	DepartmentID         uuid.UUID  `json:"department_id" binding:"required"`
	LocationID           uuid.UUID  `json:"location_id" binding:"required"`
	OrganizationalRoleID *uuid.UUID `json:"organizational_role_id"`
	ContractType         string     `json:"contract_type" binding:"required,oneof=permanent temporary on_call"`
	StartDate            string     `json:"start_date" binding:"required,datetime=2006-01-02"`
	ContractEndDate      *string    `json:"contract_end_date" binding:"omitempty,datetime=2006-01-02"`
	HoursPerWeek         *float64   `json:"hours_per_week" binding:"omitempty,min=0,max=40"`
	RosterFreeDay        string     `json:"roster_free_day" binding:"required,oneof=monday tuesday wednesday thursday friday saturday sunday"`
	WageTaxTable         *string    `json:"wage_tax_table" binding:"omitempty,oneof=white_table green_table"`
	ChangeReason         *string    `json:"change_reason"`
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
	AttachmentIDs       []uuid.UUID                            `json:"attachment_ids"          binding:"omitempty"`
	Qualifications      []createQualificationRequest           `json:"qualifications,omitempty" binding:"dive"`
	Authorizations      []createEmployeeAuthorizationRequest   `json:"authorizations,omitempty" binding:"dive"`
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
	QualificationID   string  `json:"qualification_id"   binding:"required"`
	AchievedOn        string  `json:"achieved_on"        binding:"required"`
	ExpirationDate    *string `json:"expiration_date"`
	CertificateNumber *string `json:"certificate_number"`
}

type updateQualificationRequest struct {
	QualificationID   *string `json:"qualification_id"`
	AchievedOn        *string `json:"achieved_on"`
	ExpirationDate    *string `json:"expiration_date"`
	CertificateNumber *string `json:"certificate_number"`
}

type createEmployeeAuthorizationRequest struct {
	AuthorizationID string  `json:"authorization_id" binding:"required"`
	GrantedDate     string  `json:"granted_date"     binding:"required,datetime=2006-01-02"`
	ExpiryDate      string  `json:"expiry_date"      binding:"required,datetime=2006-01-02"`
	Notes           *string `json:"notes"`
}

type updateEmployeeAuthorizationRequest struct {
	AuthorizationID *string `json:"authorization_id"`
	GrantedDate     *string `json:"granted_date"     binding:"omitempty,datetime=2006-01-02"`
	ExpiryDate      *string `json:"expiry_date"      binding:"omitempty,datetime=2006-01-02"`
	IsActive        *bool   `json:"is_active"`
	Notes           *string `json:"notes"`
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

type employeeAttachmentDetailResponse struct {
	ID           uuid.UUID `json:"id"`
	AttachmentID uuid.UUID `json:"attachment_id"`
	Category     string    `json:"category"`
	Name         string    `json:"name"`
	File         string    `json:"file"`
	Size         int32     `json:"size"`
	Tag          *string   `json:"tag"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type employeeDetailResponse struct {
	ID                           uuid.UUID                               `json:"id"`
	UserID                       uuid.UUID                               `json:"user_id"`
	FirstName                    string                                  `json:"first_name"`
	LastName                     string                                  `json:"last_name"`
	NameInUse                    string                                  `json:"name_in_use"`
	MaritalStatus                *string                                 `json:"marital_status"`
	Bsn                          string                                  `json:"bsn"`
	Street                       string                                  `json:"street"`
	HouseNumber                  string                                  `json:"house_number"`
	HouseNumberAddition          *string                                 `json:"house_number_addition"`
	PostalCode                   string                                  `json:"postal_code"`
	City                         string                                  `json:"city"`
	EmployeeNumber               *string                                 `json:"employee_number"`
	EmploymentNumber             *string                                 `json:"employment_number"`
	PrivateEmailAddress          *string                                 `json:"private_email_address"`
	WorkEmailAddress             *string                                 `json:"work_email_address"`
	PrivatePhoneNumber           *string                                 `json:"private_phone_number"`
	WorkPhoneNumber              *string                                 `json:"work_phone_number"`
	DateOfBirth                  *time.Time                              `json:"date_of_birth"`
	HomeTelephoneNumber          *string                                 `json:"home_telephone_number"`
	CreatedAt                    time.Time                               `json:"created_at"`
	Gender                       string                                  `json:"gender"`
	LocationID                   *uuid.UUID                              `json:"location_id"`
	ManagerEmployeeID            *uuid.UUID                              `json:"manager_employee_id"`
	OutOfService                 *bool                                   `json:"out_of_service"`
	IsArchived                   bool                                    `json:"is_archived"`
	ContractHours                *float64                                `json:"contract_hours"`
	ContractRate                 *float64                                `json:"contract_rate"`
	ProfilePicture               *string                                 `json:"profile_picture"`
	ManagerFirstName             *string                                 `json:"manager_first_name"`
	ManagerLastName              *string                                 `json:"manager_last_name"`
	RemainingLeaveBalanceMinutes int32                                   `json:"remaining_leave_balance_minutes"`
	HoursWorkedThisMonth         float64                                 `json:"hours_worked_this_month"`
	HoursPendingApproval         float64                                 `json:"hours_pending_approval"`
	TotalHoursWorkedThisYear     float64                                 `json:"total_hours_worked_this_year"`
	LastPerformanceReviewScore   *float64                                `json:"last_performance_review_score"`
	SalaryAssignment             *employeeSalaryAssignmentDetailResponse `json:"salary_assignment"`
	Attachments                  []employeeAttachmentDetailResponse      `json:"attachments"`
	Qualifications               []qualificationResponse                 `json:"qualifications"`
	Authorizations               []employeeAuthorizationResponse         `json:"authorizations"`
}

type employeeContractDetailResponse struct {
	ID                     uuid.UUID  `json:"id"`
	JobTitle               string     `json:"job_title"`
	DepartmentID           uuid.UUID  `json:"department_id"`
	DepartmentName         *string    `json:"department_name"`
	LocationID             uuid.UUID  `json:"location_id"`
	LocationAddress        *string    `json:"location_address"`
	OrganizationalRoleID   *uuid.UUID `json:"organizational_role_id"`
	OrganizationalRoleName *string    `json:"organizational_role_name"`
	ContractType           string     `json:"contract_type"`
	StartDate              time.Time  `json:"start_date"`
	ContractEndDate        *time.Time `json:"contract_end_date"`
	EffectiveEndDate       *time.Time `json:"effective_end_date"`
	PreviousContractID     *uuid.UUID `json:"previous_contract_id"`
	ContractEventType      string     `json:"contract_event_type"`
	IsActive               bool       `json:"is_active"`
	HoursPerWeek           *float64   `json:"hours_per_week"`
	RosterFreeDay          string     `json:"roster_free_day"`
	WageTaxTable           *string    `json:"wage_tax_table"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

type employeeSalaryAssignmentDetailResponse struct {
	ID                uuid.UUID  `json:"id"`
	ContractID        *uuid.UUID `json:"contract_id"`
	SalaryScaleStepID uuid.UUID  `json:"salary_scale_step_id"`
	CAOCode           string     `json:"cao_code"`
	SalaryTableName   string     `json:"salary_table_name"`
	Scale             int32      `json:"scale"`
	Step              string     `json:"step"`
	IPNumber          *int32     `json:"ip_number"`
	MonthlySalary     float64    `json:"monthly_salary"`
	HourlyRate        float64    `json:"hourly_rate"`
	EffectiveFrom     time.Time  `json:"effective_from"`
	EffectiveTo       *time.Time `json:"effective_to"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
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
	LeaveStatus     string     `json:"leave_status"`
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
	TotalPermanent    int64 `json:"total_permanent"`
	TotalTemporary    int64 `json:"total_temporary"`
	TotalOnCall       int64 `json:"total_on_call"`
	TotalOutOfService int64 `json:"total_out_of_service"`
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
	ID                uuid.UUID  `json:"id"`
	EmployeeID        uuid.UUID  `json:"employee_id"`
	QualificationID   uuid.UUID  `json:"qualification_id"`
	AchievedOn        time.Time  `json:"achieved_on"`
	ExpirationDate    *time.Time `json:"expiration_date"`
	CertificateNumber *string    `json:"certificate_number"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type employeeAuthorizationResponse struct {
	ID              uuid.UUID `json:"id"`
	EmployeeID      uuid.UUID `json:"employee_id"`
	AuthorizationID uuid.UUID `json:"authorization_id"`
	GrantedDate     time.Time `json:"granted_date"`
	ExpiryDate      time.Time `json:"expiry_date"`
	IsActive        bool      `json:"is_active"`
	Notes           *string   `json:"notes"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
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
			StartDate:            startDate,
			ContractEndDate:      contractEndDate,
			HoursPerWeek:         req.Contract.HoursPerWeek,
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

	var qualifications []domain.CreateQualificationParams
	if len(req.Qualifications) > 0 {
		qualifications = make([]domain.CreateQualificationParams, len(req.Qualifications))
		for i, q := range req.Qualifications {
			qualifications[i] = toCreateQualificationParams(q)
		}
	}

	var authorizations []domain.CreateEmployeeAuthorizationParams
	if len(req.Authorizations) > 0 {
		authorizations = make([]domain.CreateEmployeeAuthorizationParams, len(req.Authorizations))
		for i, a := range req.Authorizations {
			authorizations[i] = toCreateEmployeeAuthorizationParams(a)
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
		AttachmentIDs:       req.AttachmentIDs,
		Qualifications:      qualifications,
		Authorizations:      authorizations,
	}
}

func toUpdateEmployeeContractParams(req updateContractRequest) domain.UpdateEmployeeContractParams {
	startDate, _ := parseDatePtr(req.StartDate)
	contractEndDate, _ := parseDatePtr(req.ContractEndDate)

	return domain.UpdateEmployeeContractParams{
		JobTitle:             req.JobTitle,
		DepartmentID:         req.DepartmentID,
		LocationID:           req.LocationID,
		OrganizationalRoleID: req.OrganizationalRoleID,
		ContractType:         req.ContractType,
		StartDate:            startDate,
		ContractEndDate:      contractEndDate,
		HoursPerWeek:         req.HoursPerWeek,
		RosterFreeDay:        req.RosterFreeDay,
		WageTaxTable:         req.WageTaxTable,
	}
}

func toCreateNewContractParams(req createContractRequest) domain.CreateNewContractParams {
	startDate, _ := parseDate(req.StartDate)
	contractEndDate, _ := parseDatePtr(req.ContractEndDate)

	return domain.CreateNewContractParams{
		JobTitle:             req.JobTitle,
		DepartmentID:         req.DepartmentID,
		LocationID:           req.LocationID,
		OrganizationalRoleID: req.OrganizationalRoleID,
		ContractType:         req.ContractType,
		StartDate:            startDate,
		ContractEndDate:      contractEndDate,
		HoursPerWeek:         req.HoursPerWeek,
		RosterFreeDay:        req.RosterFreeDay,
		WageTaxTable:         req.WageTaxTable,
	}
}

func toCreateContractAmendmentParams(req createContractAmendmentRequest) domain.CreateContractAmendmentParams {
	startDate, _ := parseDate(req.StartDate)
	contractEndDate, _ := parseDatePtr(req.ContractEndDate)

	return domain.CreateContractAmendmentParams{
		JobTitle:             req.JobTitle,
		DepartmentID:         req.DepartmentID,
		LocationID:           req.LocationID,
		OrganizationalRoleID: req.OrganizationalRoleID,
		ContractType:         req.ContractType,
		StartDate:            startDate,
		ContractEndDate:      contractEndDate,
		HoursPerWeek:         req.HoursPerWeek,
		RosterFreeDay:        req.RosterFreeDay,
		WageTaxTable:         req.WageTaxTable,
		ChangeReason:         req.ChangeReason,
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
		QualificationID:   uuid.MustParse(req.QualificationID),
		AchievedOn:        achievedOn,
		ExpirationDate:    expirationDate,
		CertificateNumber: req.CertificateNumber,
	}
}

func toCreateQualificationsParams(req []createQualificationRequest) ([]domain.CreateQualificationParams, error) {
	out := make([]domain.CreateQualificationParams, len(req))
	for i, r := range req {
		qualificationID, err := uuid.Parse(r.QualificationID)
		if err != nil {
			return nil, fmt.Errorf("item %d: invalid qualification_id", i)
		}
		achievedOn, err := parseDate(r.AchievedOn)
		if err != nil {
			return nil, fmt.Errorf("item %d: invalid achieved_on", i)
		}
		expirationDate, err := parseDatePtr(r.ExpirationDate)
		if err != nil {
			return nil, fmt.Errorf("item %d: invalid expiration_date", i)
		}
		out[i] = domain.CreateQualificationParams{
			QualificationID:   qualificationID,
			AchievedOn:        achievedOn,
			ExpirationDate:    expirationDate,
			CertificateNumber: r.CertificateNumber,
		}
	}
	return out, nil
}

func toUpdateQualificationParams(req updateQualificationRequest) domain.UpdateQualificationParams {
	achievedOn, _ := parseDatePtr(req.AchievedOn)
	expirationDate, _ := parseDatePtr(req.ExpirationDate)

	return domain.UpdateQualificationParams{
		QualificationID:   uuidPtrFromStringPtr(req.QualificationID),
		AchievedOn:        achievedOn,
		ExpirationDate:    expirationDate,
		CertificateNumber: req.CertificateNumber,
	}
}

func toCreateEmployeeAuthorizationsParams(req []createEmployeeAuthorizationRequest) ([]domain.CreateEmployeeAuthorizationParams, error) {
	out := make([]domain.CreateEmployeeAuthorizationParams, len(req))
	for i, r := range req {
		authorizationID, err := uuid.Parse(r.AuthorizationID)
		if err != nil {
			return nil, fmt.Errorf("item %d: invalid authorization_id", i)
		}
		grantedDate, err := parseDate(r.GrantedDate)
		if err != nil {
			return nil, fmt.Errorf("item %d: invalid granted_date", i)
		}
		expiryDate, err := parseDate(r.ExpiryDate)
		if err != nil {
			return nil, fmt.Errorf("item %d: invalid expiry_date", i)
		}
		out[i] = domain.CreateEmployeeAuthorizationParams{
			AuthorizationID: authorizationID,
			GrantedDate:     grantedDate,
			ExpiryDate:      expiryDate,
			Notes:           r.Notes,
		}
	}
	return out, nil
}

func toCreateEmployeeAuthorizationParams(req createEmployeeAuthorizationRequest) domain.CreateEmployeeAuthorizationParams {
	grantedDate, _ := parseDate(req.GrantedDate)
	expiryDate, _ := parseDate(req.ExpiryDate)

	return domain.CreateEmployeeAuthorizationParams{
		AuthorizationID: uuid.MustParse(req.AuthorizationID),
		GrantedDate:     grantedDate,
		ExpiryDate:      expiryDate,
		Notes:           req.Notes,
	}
}

func toUpdateEmployeeAuthorizationParams(req updateEmployeeAuthorizationRequest) domain.UpdateEmployeeAuthorizationParams {
	var grantedDate *time.Time
	if req.GrantedDate != nil {
		d, _ := parseDate(*req.GrantedDate)
		grantedDate = &d
	}
	var expiryDate *time.Time
	if req.ExpiryDate != nil {
		d, _ := parseDate(*req.ExpiryDate)
		expiryDate = &d
	}

	return domain.UpdateEmployeeAuthorizationParams{
		AuthorizationID: uuidPtrFromStringPtr(req.AuthorizationID),
		GrantedDate:     grantedDate,
		ExpiryDate:      expiryDate,
		IsActive:        req.IsActive,
		Notes:           req.Notes,
	}
}

func toEmployeeAttachmentDetailResponse(a *domain.EmployeeAttachmentDetail) employeeAttachmentDetailResponse {
	return employeeAttachmentDetailResponse{
		ID:           a.ID,
		AttachmentID: a.AttachmentID,
		Category:     a.Category,
		Name:         a.Name,
		File:         a.File,
		Size:         a.Size,
		Tag:          a.Tag,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}
}

func toEmployeeDetailResponse(emp *domain.EmployeeDetail) employeeDetailResponse {
	attachments := make([]employeeAttachmentDetailResponse, len(emp.Attachments))
	for i, a := range emp.Attachments {
		attachments[i] = toEmployeeAttachmentDetailResponse(&a)
	}

	qualifications := make([]qualificationResponse, len(emp.Qualifications))
	for i, q := range emp.Qualifications {
		q := q
		qualifications[i] = toQualificationResponse(&q)
	}

	authorizations := make([]employeeAuthorizationResponse, len(emp.Authorizations))
	for i, a := range emp.Authorizations {
		a := a
		authorizations[i] = toEmployeeAuthorizationResponse(&a)
	}

	return employeeDetailResponse{
		ID:                           emp.ID,
		UserID:                       emp.UserID,
		FirstName:                    emp.FirstName,
		LastName:                     emp.LastName,
		NameInUse:                    emp.NameInUse,
		MaritalStatus:                emp.MaritalStatus,
		Bsn:                          emp.Bsn,
		Street:                       emp.Street,
		HouseNumber:                  emp.HouseNumber,
		HouseNumberAddition:          emp.HouseNumberAddition,
		PostalCode:                   emp.PostalCode,
		City:                         emp.City,
		EmployeeNumber:               emp.EmployeeNumber,
		EmploymentNumber:             emp.EmploymentNumber,
		PrivateEmailAddress:          emp.PrivateEmailAddress,
		WorkEmailAddress:             emp.WorkEmailAddress,
		PrivatePhoneNumber:           emp.PrivatePhoneNumber,
		WorkPhoneNumber:              emp.WorkPhoneNumber,
		DateOfBirth:                  emp.DateOfBirth,
		HomeTelephoneNumber:          emp.HomeTelephoneNumber,
		CreatedAt:                    emp.CreatedAt,
		Gender:                       emp.Gender,
		LocationID:                   emp.LocationID,
		ManagerEmployeeID:            emp.ManagerEmployeeID,
		OutOfService:                 emp.OutOfService,
		IsArchived:                   emp.IsArchived,
		ContractHours:                emp.ContractHours,
		ContractRate:                 emp.ContractRate,
		ProfilePicture:               emp.ProfilePicture,
		ManagerFirstName:             emp.ManagerFirstName,
		ManagerLastName:              emp.ManagerLastName,
		RemainingLeaveBalanceMinutes: emp.RemainingLeaveBalanceMinutes,
		HoursWorkedThisMonth:         emp.HoursWorkedThisMonth,
		HoursPendingApproval:         emp.HoursPendingApproval,
		TotalHoursWorkedThisYear:     emp.TotalHoursWorkedThisYear,
		LastPerformanceReviewScore:   emp.LastPerformanceReviewScore,
		SalaryAssignment:             toEmployeeSalaryAssignmentDetailResponse(emp.SalaryAssignment),
		Attachments:                  attachments,
		Qualifications:               qualifications,
		Authorizations:               authorizations,
	}
}

func toEmployeeContractDetailResponse(contract *domain.EmployeeContractDetail) *employeeContractDetailResponse {
	if contract == nil {
		return nil
	}
	return &employeeContractDetailResponse{
		ID:                     contract.ID,
		JobTitle:               contract.JobTitle,
		DepartmentID:           contract.DepartmentID,
		DepartmentName:         contract.DepartmentName,
		LocationID:             contract.LocationID,
		LocationAddress:        contract.LocationAddress,
		OrganizationalRoleID:   contract.OrganizationalRoleID,
		OrganizationalRoleName: contract.OrganizationalRoleName,
		ContractType:           contract.ContractType,
		StartDate:              contract.StartDate,
		ContractEndDate:        contract.ContractEndDate,
		EffectiveEndDate:       contract.EffectiveEndDate,
		PreviousContractID:     contract.PreviousContractID,
		ContractEventType:      contract.ContractEventType,
		IsActive:               contract.IsActive,
		HoursPerWeek:           contract.HoursPerWeek,
		RosterFreeDay:          contract.RosterFreeDay,
		WageTaxTable:           contract.WageTaxTable,
		CreatedAt:              contract.CreatedAt,
		UpdatedAt:              contract.UpdatedAt,
	}
}

func toEmployeeSalaryAssignmentDetailResponse(salary *domain.EmployeeSalaryAssignmentDetail) *employeeSalaryAssignmentDetailResponse {
	if salary == nil {
		return nil
	}
	return &employeeSalaryAssignmentDetailResponse{
		ID:                salary.ID,
		ContractID:        salary.ContractID,
		SalaryScaleStepID: salary.SalaryScaleStepID,
		CAOCode:           salary.CAOCode,
		SalaryTableName:   salary.SalaryTableName,
		Scale:             salary.Scale,
		Step:              salary.Step,
		IPNumber:          salary.IPNumber,
		MonthlySalary:     salary.MonthlySalary,
		HourlyRate:        salary.HourlyRate,
		EffectiveFrom:     salary.EffectiveFrom,
		EffectiveTo:       salary.EffectiveTo,
		CreatedAt:         salary.CreatedAt,
		UpdatedAt:         salary.UpdatedAt,
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
		LeaveStatus:     emp.LeaveStatus,
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
		TotalPermanent:    counts.TotalPermanent,
		TotalTemporary:    counts.TotalTemporary,
		TotalOnCall:       counts.TotalOnCall,
		TotalOutOfService: counts.TotalOutOfService,
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
		ID:                qualification.ID,
		EmployeeID:        qualification.EmployeeID,
		QualificationID:   qualification.QualificationID,
		AchievedOn:        qualification.AchievedOn,
		ExpirationDate:    qualification.ExpirationDate,
		CertificateNumber: qualification.CertificateNumber,
		CreatedAt:         qualification.CreatedAt,
		UpdatedAt:         qualification.UpdatedAt,
	}
}

func toEmployeeAuthorizationResponse(a *domain.EmployeeAuthorization) employeeAuthorizationResponse {
	return employeeAuthorizationResponse{
		ID:              a.ID,
		EmployeeID:      a.EmployeeID,
		AuthorizationID: a.AuthorizationID,
		GrantedDate:     a.GrantedDate,
		ExpiryDate:      a.ExpiryDate,
		IsActive:        a.IsActive,
		Notes:           a.Notes,
		CreatedAt:       a.CreatedAt,
		UpdatedAt:       a.UpdatedAt,
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

func uuidPtrFromStringPtr(s *string) *uuid.UUID {
	if s == nil {
		return nil
	}
	id := uuid.MustParse(*s)
	return &id
}
