package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmployeeNotFound              = errors.New("employee not found")
	ErrEducationNotFound             = errors.New("education not found")
	ErrExperienceNotFound            = errors.New("experience not found")
	ErrQualificationNotFound         = errors.New("qualification not found")
	ErrEmployeeAuthorizationNotFound = errors.New("employee authorization not found")
	ErrInvalidDateOfBirth            = errors.New("invalid date of birth format")
	ErrInvalidContractDate           = errors.New("invalid contract date format")
	ErrInvalidAttachmentID           = errors.New("invalid attachment ID")
	ErrEmployeeCreateFailed          = errors.New("failed to create employee")
	ErrPasswordHashFailed            = errors.New("failed to hash password")
	ErrEmailDeliveryFailed           = errors.New("failed to enqueue email delivery")
	ErrContractChangeInvalid         = errors.New("invalid contract change request")
	ErrContractChangeNotFound        = errors.New("contract change not found")
	ErrContractHistoryExists         = errors.New(
		"contract history exists; use contract changes endpoint",
	)
	ErrContractBaselineMissingStartDate = errors.New(
		"contract_start_date is required to bootstrap contract history",
	)
	ErrContractChangeLeaveConflict = errors.New(
		"contract change would invalidate current leave usage",
	)
	ErrInvalidPasswordResetRequest = errors.New("invalid password reset request")
)

const (
	IrregularHoursProfileNone      = "none"
	IrregularHoursProfileRoster    = "roster"
	IrregularHoursProfileNonRoster = "non_roster"
)

// Employee is the lean domain struct for list queries.
type Employee struct {
	ID              uuid.UUID
	FirstName       string
	LastName        string
	Bsn             string
	ContractType    string
	DepartmentName  *string
	ContractEndDate *time.Time
	LocationAddress string
	LeaveStatus     string
}

// EmployeeAttachmentDetail combines employee_attachment + attachment_file for list views.
type EmployeeAttachmentDetail struct {
	ID           uuid.UUID
	EmployeeID   uuid.UUID
	AttachmentID uuid.UUID
	Category     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Name         string
	File         string
	Size         int32
	Tag          *string
}

// EmployeeDetail is the rich domain struct for get-by-id queries (with joins).
type EmployeeDetail struct {
	ID                         uuid.UUID
	UserID                     uuid.UUID
	FirstName                  string
	LastName                   string
	NameInUse                  string
	MaritalStatus              *string
	Bsn                        string
	Street                     string
	HouseNumber                string
	HouseNumberAddition        *string
	PostalCode                 string
	City                       string
	EmployeeNumber             *string
	EmploymentNumber           *string
	PrivateEmailAddress        *string
	WorkEmailAddress           *string
	PrivatePhoneNumber         *string
	WorkPhoneNumber            *string
	DateOfBirth                *time.Time
	HomeTelephoneNumber        *string
	CreatedAt                  time.Time
	Gender                     string
	LocationID                 *uuid.UUID
	DepartmentID               *uuid.UUID
	ManagerEmployeeID          *uuid.UUID
	OutOfService               *bool
	IsArchived                 bool
	ContractHours              *float64
	ContractEndDate            *time.Time
	ContractStartDate          *time.Time
	ContractType               string
	ContractRate               *float64
	ProfilePicture             *string
	DepartmentName             *string
	ManagerFirstName           *string
	ManagerLastName            *string
	RemainingLeaveBalanceHours int32
	HoursWorkedThisMonth       float64
	HoursPendingApproval       float64
	TotalHoursWorkedThisYear   float64
	LastPerformanceReviewScore *float64
	SalaryAssignment           *EmployeeSalaryAssignmentDetail
	Attachments                []EmployeeAttachmentDetail
	Qualifications             []Qualification
	Authorizations             []EmployeeAuthorization
}

type EmployeeContractDetail struct {
	ID                uuid.UUID
	JobTitle          string
	DepartmentID      uuid.UUID
	DepartmentName    *string
	LocationID        uuid.UUID
	LocationAddress   *string
	OrganizationalRoleID   *uuid.UUID
	OrganizationalRoleName *string
	ContractType      string
	ContractHoursType string
	StartDate              time.Time
	ContractEndDate        *time.Time
	EffectiveEndDate       *time.Time
	PreviousContractID *uuid.UUID
	ContractEventType  string
	IsActive           bool
	HoursPerWeek      *float64
	MinHoursPerWeek   *float64
	MaxHoursPerWeek   *float64
	RosterFreeDay     string
	WageTaxTable      *string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type EmployeeSalaryAssignmentDetail struct {
	ID                uuid.UUID
	ContractID        *uuid.UUID
	SalaryScaleStepID uuid.UUID
	CAOCode           string
	SalaryTableName   string
	Scale             int32
	Step              string
	IPNumber          *int32
	MonthlySalary     float64
	HourlyRate        float64
	EffectiveFrom     time.Time
	EffectiveTo       *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// Portal access types for frontend routing.
const (
	PortalAccessAdmin    = "admin"
	PortalAccessEmployee = "employee"
	PortalAccessBoth     = "both"
)

// Portal permission names used to compute portal_access.
const (
	PortalPermissionAdmin    = "PORTAL.ADMIN.ACCESS"
	PortalPermissionEmployee = "PORTAL.EMPLOYEE.ACCESS"
)

// EmployeeProfile is the domain struct for the current user's profile (with permissions).
type EmployeeProfile struct {
	UserID           uuid.UUID
	Email            string
	LastLogin        time.Time
	TwoFactorEnabled bool
	Role             string
	RoleID           uuid.UUID
	EmployeeID       uuid.UUID
	FirstName        string
	LastName         string
	Permissions      []Permission
	PortalAccess     string
}

type Permission struct {
	ID       uuid.UUID
	Name     string
	Resource string
	Method   string
}

// EmployeeCounts is the domain struct for employee count statistics.
type EmployeeCounts struct {
	TotalPermanent    int64
	TotalTemporary    int64
	TotalOnCall       int64
	TotalOutOfService int64
}

// EmployeeSearchResult is the domain struct for search results.
type EmployeeSearchResult struct {
	ID               uuid.UUID
	FirstName        string
	LastName         string
	WorkEmailAddress *string
}

// Education domain struct.
type Education struct {
	ID              uuid.UUID
	EmployeeID      uuid.UUID
	InstitutionName string
	Degree          string
	FieldOfStudy    string
	StartDate       time.Time
	EndDate         time.Time
	CreatedAt       time.Time
}

// Experience domain struct.
type Experience struct {
	ID          uuid.UUID
	EmployeeID  uuid.UUID
	JobTitle    string
	CompanyName string
	StartDate   time.Time
	EndDate     time.Time
	Description *string
	CreatedAt   time.Time
}

// Qualification domain struct.
type Qualification struct {
	ID                uuid.UUID
	EmployeeID        uuid.UUID
	QualificationID   uuid.UUID
	AchievedOn        time.Time
	ExpirationDate    *time.Time
	CertificateNumber *string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// EmployeeAuthorization domain struct.
type EmployeeAuthorization struct {
	ID              uuid.UUID
	EmployeeID      uuid.UUID
	AuthorizationID uuid.UUID
	GrantedDate     time.Time
	ExpiryDate      time.Time
	IsActive        bool
	Notes           *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// --- Params ---

type ListEmployeesParams struct {
	Limit               int32
	Offset              int32
	IncludeArchived     *bool
	IncludeOutOfService *bool
	LocationID          *uuid.UUID
	ContractType        *string
	Search              *string
}

type EmployeePage struct {
	Items      []Employee
	TotalCount int64
}

type CreateEmployeeContractParams struct {
	JobTitle             string
	DepartmentID         uuid.UUID
	LocationID           uuid.UUID
	OrganizationalRoleID *uuid.UUID
	ContractType         string
	ContractHoursType    string
	StartDate            time.Time
	ContractEndDate      *time.Time
	HoursPerWeek         *float64
	MinHoursPerWeek      *float64
	MaxHoursPerWeek      *float64
	RosterFreeDay        string
	WageTaxTable         *string
}

type EmployeeContractInfo struct {
	ID                uuid.UUID
	EmployeeID        uuid.UUID
	ContractType      string
	ContractHoursType string
	StartDate         time.Time
	ContractEndDate   *time.Time
	EffectiveEndDate  *time.Time
	HoursPerWeek      *float64
}

type CreateNewContractParams struct {
	JobTitle             string
	DepartmentID         uuid.UUID
	LocationID           uuid.UUID
	OrganizationalRoleID *uuid.UUID
	ContractType         string
	ContractHoursType    string
	StartDate            time.Time
	ContractEndDate      *time.Time
	HoursPerWeek         *float64
	MinHoursPerWeek      *float64
	MaxHoursPerWeek      *float64
	RosterFreeDay        string
	WageTaxTable         *string
}

type UpdateEmployeeContractParams struct {
	JobTitle             *string    `json:"job_title"`
	DepartmentID         *uuid.UUID `json:"department_id"`
	LocationID           *uuid.UUID `json:"location_id"`
	OrganizationalRoleID *uuid.UUID `json:"organizational_role_id"`
	ContractType         *string    `json:"contract_type"`
	ContractHoursType    *string    `json:"contract_hours_type"`
	StartDate            *time.Time `json:"start_date"`
	ContractEndDate      *time.Time `json:"contract_end_date"`
	HoursPerWeek         *float64   `json:"hours_per_week"`
	MinHoursPerWeek      *float64   `json:"min_hours_per_week"`
	MaxHoursPerWeek      *float64   `json:"max_hours_per_week"`
	RosterFreeDay        *string    `json:"roster_free_day"`
	WageTaxTable         *string    `json:"wage_tax_table"`
}

type CreateContractAmendmentParams struct {
	JobTitle             string
	DepartmentID         uuid.UUID
	LocationID           uuid.UUID
	OrganizationalRoleID *uuid.UUID
	ContractType         string
	ContractHoursType    string
	StartDate            time.Time
	ContractEndDate      *time.Time
	HoursPerWeek         *float64
	MinHoursPerWeek      *float64
	MaxHoursPerWeek      *float64
	RosterFreeDay        string
	WageTaxTable         *string
	ChangeReason         *string
}

type CreateEmployeeSalaryAssignmentParams struct {
	SalaryScaleStepID uuid.UUID
	EffectiveFrom     *time.Time
	EffectiveTo       *time.Time
}

type CreateEmployeeParams struct {
	FirstName           string
	LastName            string
	Bsn                 string
	Street              string
	HouseNumber         string
	HouseNumberAddition *string
	PostalCode          string
	City                string
	ManagerEmployeeID   *uuid.UUID
	EmployeeNumber      *string
	EmploymentNumber    *string
	PrivateEmailAddress *string
	WorkEmailAddress    *string
	WorkPhoneNumber     *string
	PrivatePhoneNumber  *string
	DateOfBirth         *time.Time
	HomeTelephoneNumber *string
	Gender              string
	RoleID              uuid.UUID
	UserEmail           string
	UserPassword        string

	Contract         *CreateEmployeeContractParams
	SalaryAssignment *CreateEmployeeSalaryAssignmentParams
	AttachmentIDs    []uuid.UUID
	Qualifications   []CreateQualificationParams
	Authorizations   []CreateEmployeeAuthorizationParams
}

type UpdateEmployeeParams struct {
	FirstName           *string
	LastName            *string
	ManagerEmployeeID   *uuid.UUID
	EmployeeNumber      *string
	EmploymentNumber    *string
	PrivateEmailAddress *string
	PrivatePhoneNumber  *string
	WorkPhoneNumber     *string
	DateOfBirth         *time.Time
	HomeTelephoneNumber *string
	Gender              *string
	OutOfService        *bool
	IsArchived          *bool
}

type ResetPasswordParams struct {
	Generated bool
	Password  *string
	SendEmail bool
}

type ResetPasswordResult struct {
	TemporaryPassword string
}

type CreateEducationParams struct {
	InstitutionName string
	Degree          string
	FieldOfStudy    string
	StartDate       time.Time
	EndDate         time.Time
}

type UpdateEducationParams struct {
	InstitutionName *string
	Degree          *string
	FieldOfStudy    *string
	StartDate       *time.Time
	EndDate         *time.Time
}

type CreateExperienceParams struct {
	JobTitle    string
	CompanyName string
	StartDate   time.Time
	EndDate     time.Time
	Description *string
}

type UpdateExperienceParams struct {
	JobTitle    *string
	CompanyName *string
	StartDate   *time.Time
	EndDate     *time.Time
	Description *string
}

type CreateQualificationParams struct {
	QualificationID   uuid.UUID
	AchievedOn        time.Time
	ExpirationDate    *time.Time
	CertificateNumber *string
}

type UpdateQualificationParams struct {
	QualificationID   *uuid.UUID
	AchievedOn        *time.Time
	ExpirationDate    *time.Time
	CertificateNumber *string
}

type CreateEmployeeAuthorizationParams struct {
	AuthorizationID uuid.UUID
	GrantedDate     time.Time
	ExpiryDate      time.Time
	Notes           *string
}

type UpdateEmployeeAuthorizationParams struct {
	AuthorizationID *uuid.UUID
	GrantedDate     *time.Time
	ExpiryDate      *time.Time
	IsActive        *bool
	Notes           *string
}

// --- Interfaces ---

type EmployeeTxRepository interface {
	CreateUser(ctx context.Context, email, password string) (uuid.UUID, error)
	CreateEmployeeProfile(ctx context.Context, userID uuid.UUID, params CreateEmployeeParams) (uuid.UUID, error)
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error
	AddEmployeeContractDetails(ctx context.Context, employeeID uuid.UUID, params CreateEmployeeContractParams) (uuid.UUID, error)
	CreateEmployeeSalaryAssignment(ctx context.Context, employeeID uuid.UUID, contractID *uuid.UUID, params CreateEmployeeSalaryAssignmentParams) (uuid.UUID, error)
	GetEmployeeByID(ctx context.Context, id uuid.UUID) (*EmployeeDetail, error)
	LinkEmployeeAttachments(ctx context.Context, employeeID uuid.UUID, attachmentIDs []uuid.UUID, category string) error
	UpdateAttachmentsUsed(ctx context.Context, ids []uuid.UUID, isUsed bool) error
	AddEmployeeQualificationsBatch(ctx context.Context, employeeID uuid.UUID, params []CreateQualificationParams) error
	AddEmployeeAuthorizationsBatch(ctx context.Context, employeeID uuid.UUID, params []CreateEmployeeAuthorizationParams) error
	EndEmployeeContractSegment(ctx context.Context, contractID uuid.UUID, endDate time.Time, updatedBy *uuid.UUID) error
	AddEmployeeContractAmendment(ctx context.Context, employeeID uuid.UUID, previousContractID uuid.UUID, params CreateContractAmendmentParams) (uuid.UUID, error)
	GetEmployeeContractAtDate(ctx context.Context, employeeID uuid.UUID, targetDate time.Time) (*EmployeeContractInfo, error)
	AddNewContract(ctx context.Context, employeeID uuid.UUID, previousContractID *uuid.UUID, params CreateNewContractParams) (uuid.UUID, error)
	UpdateEmployeeContract(ctx context.Context, employeeID, contractID uuid.UUID, params UpdateEmployeeContractParams) (*EmployeeContractDetail, error)
}

type EmployeeRepository interface {
	WithTx(ctx context.Context, fn func(tx EmployeeTxRepository) error) error
	// Profile CRUD
	GetEmployeeByID(ctx context.Context, id uuid.UUID) (*EmployeeDetail, error)
	GetEmployeeByUserID(ctx context.Context, userID uuid.UUID) (*EmployeeProfile, error)
	ListEmployees(ctx context.Context, params ListEmployeesParams) (*EmployeePage, error)
	CountEmployees(ctx context.Context, params ListEmployeesParams) (int64, error)
	CreateEmployee(ctx context.Context, params CreateEmployeeParams) (*EmployeeDetail, error)
	UpdateEmployee(
		ctx context.Context,
		id uuid.UUID,
		params UpdateEmployeeParams,
	) (*EmployeeDetail, error)
	GetEmployeeCounts(ctx context.Context) (*EmployeeCounts, error)
	SearchEmployeesByNameOrEmail(
		ctx context.Context,
		search *string,
	) ([]EmployeeSearchResult, error)

	// Education
	ListEducation(ctx context.Context, employeeID uuid.UUID) ([]Education, error)
	AddEducation(
		ctx context.Context,
		employeeID uuid.UUID,
		params CreateEducationParams,
	) (*Education, error)
	UpdateEducation(
		ctx context.Context,
		id uuid.UUID,
		params UpdateEducationParams,
	) (*Education, error)
	DeleteEducation(ctx context.Context, id uuid.UUID) (*Education, error)

	// Experience
	ListExperience(ctx context.Context, employeeID uuid.UUID) ([]Experience, error)
	AddExperience(
		ctx context.Context,
		employeeID uuid.UUID,
		params CreateExperienceParams,
	) (*Experience, error)
	UpdateExperience(
		ctx context.Context,
		id uuid.UUID,
		params UpdateExperienceParams,
	) (*Experience, error)
	DeleteExperience(ctx context.Context, id uuid.UUID) (*Experience, error)

	// Qualification
	ListQualifications(ctx context.Context, employeeID uuid.UUID) ([]Qualification, error)
	AddQualifications(
		ctx context.Context,
		employeeID uuid.UUID,
		params []CreateQualificationParams,
	) (int, error)
	UpdateQualification(
		ctx context.Context,
		id uuid.UUID,
		params UpdateQualificationParams,
	) (*Qualification, error)
	DeleteQualification(ctx context.Context, id uuid.UUID) (*Qualification, error)

	// Attachment
	ListEmployeeAttachments(ctx context.Context, employeeID uuid.UUID) ([]EmployeeAttachmentDetail, error)

	// Employee Authorization
	ListEmployeeAuthorizations(ctx context.Context, employeeID uuid.UUID) ([]EmployeeAuthorization, error)
	AddEmployeeAuthorizations(
		ctx context.Context,
		employeeID uuid.UUID,
		params []CreateEmployeeAuthorizationParams,
	) (int, error)
	UpdateEmployeeAuthorization(
		ctx context.Context,
		id uuid.UUID,
		params UpdateEmployeeAuthorizationParams,
	) (*EmployeeAuthorization, error)
	DeleteEmployeeAuthorization(ctx context.Context, id uuid.UUID) (*EmployeeAuthorization, error)

	// Password
	UpdatePassword(ctx context.Context, userID uuid.UUID, password string) error

	// Contract
	GetEmployeeContractByID(ctx context.Context, contractID uuid.UUID) (*EmployeeContractInfo, error)
	ListEmployeeContracts(ctx context.Context, employeeID uuid.UUID) ([]EmployeeContractDetail, error)
	UpdateEmployeeContract(ctx context.Context, employeeID, contractID uuid.UUID, params UpdateEmployeeContractParams) (*EmployeeContractDetail, error)
}

type EmployeeService interface {
	GetEmployeeByID(
		ctx context.Context,
		id uuid.UUID,
		currentUserID uuid.UUID,
	) (*EmployeeDetail, error)
	GetEmployeeProfile(ctx context.Context, userID uuid.UUID) (*EmployeeProfile, error)
	ListEmployees(ctx context.Context, params ListEmployeesParams) (*EmployeePage, error)
	CreateEmployee(ctx context.Context, params CreateEmployeeParams) (*EmployeeDetail, error)
	UpdateEmployee(
		ctx context.Context,
		id uuid.UUID,
		params UpdateEmployeeParams,
	) (*EmployeeDetail, error)
	GetEmployeeCounts(ctx context.Context) (*EmployeeCounts, error)
	SearchEmployeesByNameOrEmail(
		ctx context.Context,
		search *string,
	) ([]EmployeeSearchResult, error)

	ListEducation(ctx context.Context, employeeID uuid.UUID) ([]Education, error)
	AddEducation(
		ctx context.Context,
		employeeID uuid.UUID,
		params CreateEducationParams,
	) (*Education, error)
	UpdateEducation(
		ctx context.Context,
		id uuid.UUID,
		params UpdateEducationParams,
	) (*Education, error)
	DeleteEducation(ctx context.Context, id uuid.UUID) (*Education, error)

	ListExperience(ctx context.Context, employeeID uuid.UUID) ([]Experience, error)
	AddExperience(
		ctx context.Context,
		employeeID uuid.UUID,
		params CreateExperienceParams,
	) (*Experience, error)
	UpdateExperience(
		ctx context.Context,
		id uuid.UUID,
		params UpdateExperienceParams,
	) (*Experience, error)
	DeleteExperience(ctx context.Context, id uuid.UUID) (*Experience, error)

	ListQualifications(ctx context.Context, employeeID uuid.UUID) ([]Qualification, error)
	AddQualifications(
		ctx context.Context,
		employeeID uuid.UUID,
		params []CreateQualificationParams,
	) (int, error)
	UpdateQualification(
		ctx context.Context,
		id uuid.UUID,
		params UpdateQualificationParams,
	) (*Qualification, error)
	DeleteQualification(ctx context.Context, id uuid.UUID) (*Qualification, error)

	ListEmployeeAuthorizations(ctx context.Context, employeeID uuid.UUID) ([]EmployeeAuthorization, error)
	AddEmployeeAuthorizations(
		ctx context.Context,
		employeeID uuid.UUID,
		params []CreateEmployeeAuthorizationParams,
	) (int, error)
	UpdateEmployeeAuthorization(
		ctx context.Context,
		id uuid.UUID,
		params UpdateEmployeeAuthorizationParams,
	) (*EmployeeAuthorization, error)
	DeleteEmployeeAuthorization(ctx context.Context, id uuid.UUID) (*EmployeeAuthorization, error)

	// Password
	ResetPassword(
		ctx context.Context,
		employeeID uuid.UUID,
		params ResetPasswordParams,
	) (*ResetPasswordResult, error)

	// Contract
	ListEmployeeContracts(ctx context.Context, employeeID uuid.UUID) ([]EmployeeContractDetail, error)
	CreateContractAmendment(
		ctx context.Context,
		employeeID uuid.UUID,
		contractID uuid.UUID,
		params CreateContractAmendmentParams,
	) (*EmployeeDetail, error)
	CreateNewContract(
		ctx context.Context,
		employeeID uuid.UUID,
		params CreateNewContractParams,
	) (*EmployeeDetail, error)
	UpdateEmployeeContract(
		ctx context.Context,
		employeeID uuid.UUID,
		contractID uuid.UUID,
		params UpdateEmployeeContractParams,
	) (*EmployeeContractDetail, error)
}
