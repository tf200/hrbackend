package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmployeeNotFound       = errors.New("employee not found")
	ErrEducationNotFound      = errors.New("education not found")
	ErrExperienceNotFound     = errors.New("experience not found")
	ErrQualificationNotFound  = errors.New("qualification not found")
	ErrInvalidDateOfBirth     = errors.New("invalid date of birth format")
	ErrInvalidContractDate    = errors.New("invalid contract date format")
	ErrInvalidAttachmentID    = errors.New("invalid attachment ID")
	ErrEmployeeCreateFailed   = errors.New("failed to create employee")
	ErrPasswordHashFailed     = errors.New("failed to hash password")
	ErrEmailDeliveryFailed    = errors.New("failed to enqueue email delivery")
	ErrContractChangeInvalid  = errors.New("invalid contract change request")
	ErrContractChangeNotFound = errors.New("contract change not found")
	ErrContractHistoryExists  = errors.New(
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
}

// EmployeeDetail is the rich domain struct for get-by-id queries (with joins).
type EmployeeDetail struct {
	ID                         uuid.UUID
	UserID                     uuid.UUID
	FirstName                  string
	LastName                   string
	Bsn                        string
	Street                     string
	HouseNumber                string
	HouseNumberAddition        *string
	PostalCode                 string
	City                       string
	Position                   *string
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
	HasBorrowed                bool
	OutOfService               *bool
	IsArchived                 bool
	ContractHours              *float64
	ContractEndDate            *time.Time
	ContractStartDate          *time.Time
	ContractType               string
	ContractRate               *float64
	IrregularHoursProfile      string
	ProfilePicture             *string
	DepartmentName             *string
	ManagerFirstName           *string
	ManagerLastName            *string
	RemainingLeaveBalanceHours int32
	HoursWorkedThisMonth       float64
	HoursPendingApproval       float64
	TotalHoursWorkedThisYear   float64
	LastPerformanceReviewScore *float64
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
	TotalEmployees      int64
	TotalSubcontractors int64
	TotalArchived       int64
	TotalOutOfService   int64
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

// QualificationType domain struct.
type QualificationType struct {
	Code              string
	OriginalDutchText string
	EnglishName       string
	AppContext        string
	IsActive          bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// Qualification domain struct.
type Qualification struct {
	ID                    uuid.UUID
	EmployeeID            uuid.UUID
	QualificationTypeCode string
	AchievedOn            time.Time
	ExpirationDate        *time.Time
	CertificateNumber     *string
	CreatedAt             time.Time
	UpdatedAt             time.Time
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
	RosterFreeDay        *int16
	WageTaxTable         *string
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

type UpdateIsSubcontractorParams struct {
	IsSubcontractor bool
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
	QualificationTypeCode string
	AchievedOn            time.Time
	ExpirationDate        *time.Time
	CertificateNumber     *string
}

type UpdateQualificationParams struct {
	QualificationTypeCode *string
	AchievedOn            *time.Time
	ExpirationDate        *time.Time
	CertificateNumber     *string
}

// --- Interfaces ---

type EmployeeRepository interface {
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

	// Contract
	UpdateIsSubcontractor(
		ctx context.Context,
		employeeID uuid.UUID,
		contractType string,
	) (*EmployeeDetail, error)

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
	AddQualification(
		ctx context.Context,
		employeeID uuid.UUID,
		params CreateQualificationParams,
	) (*Qualification, error)
	UpdateQualification(
		ctx context.Context,
		id uuid.UUID,
		params UpdateQualificationParams,
	) (*Qualification, error)
	DeleteQualification(ctx context.Context, id uuid.UUID) (*Qualification, error)
	ListQualificationTypes(ctx context.Context) ([]QualificationType, error)

	// Password
	UpdatePassword(ctx context.Context, userID uuid.UUID, password string) error
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

	UpdateIsSubcontractor(
		ctx context.Context,
		employeeID uuid.UUID,
		params UpdateIsSubcontractorParams,
	) (*EmployeeDetail, error)

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
	AddQualification(
		ctx context.Context,
		employeeID uuid.UUID,
		params CreateQualificationParams,
	) (*Qualification, error)
	UpdateQualification(
		ctx context.Context,
		id uuid.UUID,
		params UpdateQualificationParams,
	) (*Qualification, error)
	DeleteQualification(ctx context.Context, id uuid.UUID) (*Qualification, error)
	ListQualificationTypes(ctx context.Context) ([]QualificationType, error)

	// Password
	ResetPassword(
		ctx context.Context,
		employeeID uuid.UUID,
		params ResetPasswordParams,
	) (*ResetPasswordResult, error)
}
