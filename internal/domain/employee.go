package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmployeeNotFound                 = errors.New("employee not found")
	ErrEducationNotFound                = errors.New("education not found")
	ErrExperienceNotFound               = errors.New("experience not found")
	ErrCertificationNotFound            = errors.New("certification not found")
	ErrInvalidDateOfBirth               = errors.New("invalid date of birth format")
	ErrInvalidContractDate              = errors.New("invalid contract date format")
	ErrInvalidAttachmentID              = errors.New("invalid attachment ID")
	ErrEmployeeCreateFailed             = errors.New("failed to create employee")
	ErrPasswordHashFailed               = errors.New("failed to hash password")
	ErrEmailDeliveryFailed              = errors.New("failed to enqueue email delivery")
	ErrContractChangeInvalid            = errors.New("invalid contract change request")
	ErrContractChangeNotFound           = errors.New("contract change not found")
	ErrContractHistoryExists            = errors.New("contract history exists; use contract changes endpoint")
	ErrContractBaselineMissingStartDate = errors.New("contract_start_date is required to bootstrap contract history")
	ErrContractChangeLeaveConflict      = errors.New("contract change would invalidate current leave usage")
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
	ID                    uuid.UUID
	UserID                uuid.UUID
	FirstName             string
	LastName              string
	Bsn                   string
	Street                string
	HouseNumber           string
	HouseNumberAddition   *string
	PostalCode            string
	City                  string
	Position              *string
	EmployeeNumber        *string
	EmploymentNumber      *string
	PrivateEmailAddress   *string
	WorkEmailAddress      *string
	PrivatePhoneNumber    *string
	WorkPhoneNumber       *string
	DateOfBirth           *time.Time
	HomeTelephoneNumber   *string
	CreatedAt             time.Time
	Gender                string
	LocationID            *uuid.UUID
	DepartmentID          *uuid.UUID
	ManagerEmployeeID     *uuid.UUID
	HasBorrowed           bool
	OutOfService          *bool
	IsArchived            bool
	ContractHours         *float64
	ContractEndDate       *time.Time
	ContractStartDate     *time.Time
	ContractType          string
	ContractRate          *float64
	IrregularHoursProfile string
	ProfilePicture        *string
	DepartmentName        *string
	ManagerFirstName      *string
	ManagerLastName       *string
}

// EmployeeProfile is the domain struct for the current user's profile (with permissions).
type EmployeeProfile struct {
	UserID           uuid.UUID
	Email            string
	LastLogin        time.Time
	TwoFactorEnabled bool
	EmployeeID       uuid.UUID
	FirstName        string
	LastName         string
	Permissions      []Permission
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

// Certification domain struct.
type Certification struct {
	ID         uuid.UUID
	EmployeeID uuid.UUID
	Name       string
	IssuedBy   string
	DateIssued time.Time
	CreatedAt  time.Time
}

// ContractDetails domain struct.
type ContractDetails struct {
	ContractHours         *float64
	ContractStartDate     time.Time
	ContractEndDate       time.Time
	ContractType          string
	ContractRate          *float64
	IrregularHoursProfile string
	IsSubcontractor       *bool
}

type EmployeeContractChange struct {
	ID                    uuid.UUID
	EmployeeID            uuid.UUID
	EffectiveFrom         time.Time
	EffectiveTo           *time.Time
	ContractHours         float64
	ContractType          string
	ContractRate          *float64
	IrregularHoursProfile string
	ContractEndDate       *time.Time
	CreatedByEmployeeID   uuid.UUID
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type LeaveRecalculationImpact struct {
	Year        int32
	LegalBefore int32
	LegalAfter  int32
	Delta       int32
}

type CreateEmployeeContractChangeResult struct {
	Change         EmployeeContractChange
	Recalculations []LeaveRecalculationImpact
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

type CreateEmployeeParams struct {
	FirstName             string
	LastName              string
	Bsn                   string
	Street                string
	HouseNumber           string
	HouseNumberAddition   *string
	PostalCode            string
	City                  string
	Position              *string
	DepartmentID          *uuid.UUID
	ManagerEmployeeID     *uuid.UUID
	EmployeeNumber        *string
	EmploymentNumber      *string
	PrivateEmailAddress   *string
	WorkEmailAddress      *string
	WorkPhoneNumber       *string
	PrivatePhoneNumber    *string
	DateOfBirth           *time.Time
	HomeTelephoneNumber   *string
	Gender                string
	LocationID            *uuid.UUID
	ContractHours         *float64
	ContractType          string
	ContractStartDate     *time.Time
	ContractEndDate       *time.Time
	ContractRate          *float64
	IrregularHoursProfile string
	RoleID                uuid.UUID
	UserEmail             string
	UserPassword          string
}

type UpdateEmployeeParams struct {
	FirstName             *string
	LastName              *string
	Position              *string
	DepartmentID          *uuid.UUID
	ManagerEmployeeID     *uuid.UUID
	EmployeeNumber        *string
	EmploymentNumber      *string
	PrivateEmailAddress   *string
	PrivatePhoneNumber    *string
	WorkPhoneNumber       *string
	DateOfBirth           *time.Time
	HomeTelephoneNumber   *string
	Gender                *string
	LocationID            *uuid.UUID
	IrregularHoursProfile *string
	HasBorrowed           *bool
	OutOfService          *bool
	IsArchived            *bool
}

type AddContractDetailsParams struct {
	ContractHours         *float64
	ContractStartDate     time.Time
	ContractEndDate       time.Time
	ContractRate          *float64
	IrregularHoursProfile string
}

type UpdateIsSubcontractorParams struct {
	IsSubcontractor bool
}

type CreateEmployeeContractChangeParams struct {
	EffectiveFrom         time.Time
	ContractHours         float64
	ContractType          string
	ContractRate          *float64
	IrregularHoursProfile string
	ContractEndDate       *time.Time
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

type CreateCertificationParams struct {
	Name       string
	IssuedBy   string
	DateIssued time.Time
}

type UpdateCertificationParams struct {
	Name       *string
	IssuedBy   *string
	DateIssued *time.Time
}

// --- Interfaces ---

type EmployeeRepository interface {
	// Profile CRUD
	GetEmployeeByID(ctx context.Context, id uuid.UUID) (*EmployeeDetail, error)
	GetEmployeeByUserID(ctx context.Context, userID uuid.UUID) (*EmployeeProfile, error)
	ListEmployees(ctx context.Context, params ListEmployeesParams) (*EmployeePage, error)
	CountEmployees(ctx context.Context, params ListEmployeesParams) (int64, error)
	CreateEmployee(ctx context.Context, params CreateEmployeeParams) (*EmployeeDetail, error)
	UpdateEmployee(ctx context.Context, id uuid.UUID, params UpdateEmployeeParams) (*EmployeeDetail, error)
	GetEmployeeCounts(ctx context.Context) (*EmployeeCounts, error)
	SearchEmployeesByNameOrEmail(ctx context.Context, search *string) ([]EmployeeSearchResult, error)

	// Contract
	GetContractDetails(ctx context.Context, employeeID uuid.UUID) (*ContractDetails, error)
	AddContractDetails(ctx context.Context, employeeID uuid.UUID, params AddContractDetailsParams) (*EmployeeDetail, error)
	UpdateIsSubcontractor(ctx context.Context, employeeID uuid.UUID, contractType string) (*EmployeeDetail, error)
	ListContractChanges(ctx context.Context, employeeID uuid.UUID) ([]EmployeeContractChange, error)
	CreateContractChange(ctx context.Context, actorEmployeeID, employeeID uuid.UUID, params CreateEmployeeContractChangeParams) (*CreateEmployeeContractChangeResult, error)

	// Education
	ListEducation(ctx context.Context, employeeID uuid.UUID) ([]Education, error)
	AddEducation(ctx context.Context, employeeID uuid.UUID, params CreateEducationParams) (*Education, error)
	UpdateEducation(ctx context.Context, id uuid.UUID, params UpdateEducationParams) (*Education, error)
	DeleteEducation(ctx context.Context, id uuid.UUID) (*Education, error)

	// Experience
	ListExperience(ctx context.Context, employeeID uuid.UUID) ([]Experience, error)
	AddExperience(ctx context.Context, employeeID uuid.UUID, params CreateExperienceParams) (*Experience, error)
	UpdateExperience(ctx context.Context, id uuid.UUID, params UpdateExperienceParams) (*Experience, error)
	DeleteExperience(ctx context.Context, id uuid.UUID) (*Experience, error)

	// Certification
	ListCertification(ctx context.Context, employeeID uuid.UUID) ([]Certification, error)
	AddCertification(ctx context.Context, employeeID uuid.UUID, params CreateCertificationParams) (*Certification, error)
	UpdateCertification(ctx context.Context, id uuid.UUID, params UpdateCertificationParams) (*Certification, error)
	DeleteCertification(ctx context.Context, id uuid.UUID) (*Certification, error)
}

type EmployeeService interface {
	GetEmployeeByID(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID) (*EmployeeDetail, error)
	GetEmployeeProfile(ctx context.Context, userID uuid.UUID) (*EmployeeProfile, error)
	ListEmployees(ctx context.Context, params ListEmployeesParams) (*EmployeePage, error)
	CreateEmployee(ctx context.Context, params CreateEmployeeParams) (*EmployeeDetail, error)
	UpdateEmployee(ctx context.Context, id uuid.UUID, params UpdateEmployeeParams) (*EmployeeDetail, error)
	GetEmployeeCounts(ctx context.Context) (*EmployeeCounts, error)
	SearchEmployeesByNameOrEmail(ctx context.Context, search *string) ([]EmployeeSearchResult, error)

	GetContractDetails(ctx context.Context, employeeID uuid.UUID) (*ContractDetails, error)
	AddContractDetails(ctx context.Context, employeeID uuid.UUID, params AddContractDetailsParams) (*EmployeeDetail, error)
	UpdateIsSubcontractor(ctx context.Context, employeeID uuid.UUID, params UpdateIsSubcontractorParams) (*EmployeeDetail, error)
	ListContractChanges(ctx context.Context, employeeID uuid.UUID) ([]EmployeeContractChange, error)
	CreateContractChange(ctx context.Context, actorEmployeeID, employeeID uuid.UUID, params CreateEmployeeContractChangeParams) (*CreateEmployeeContractChangeResult, error)

	ListEducation(ctx context.Context, employeeID uuid.UUID) ([]Education, error)
	AddEducation(ctx context.Context, employeeID uuid.UUID, params CreateEducationParams) (*Education, error)
	UpdateEducation(ctx context.Context, id uuid.UUID, params UpdateEducationParams) (*Education, error)
	DeleteEducation(ctx context.Context, id uuid.UUID) (*Education, error)

	ListExperience(ctx context.Context, employeeID uuid.UUID) ([]Experience, error)
	AddExperience(ctx context.Context, employeeID uuid.UUID, params CreateExperienceParams) (*Experience, error)
	UpdateExperience(ctx context.Context, id uuid.UUID, params UpdateExperienceParams) (*Experience, error)
	DeleteExperience(ctx context.Context, id uuid.UUID) (*Experience, error)

	ListCertification(ctx context.Context, employeeID uuid.UUID) ([]Certification, error)
	AddCertification(ctx context.Context, employeeID uuid.UUID, params CreateCertificationParams) (*Certification, error)
	UpdateCertification(ctx context.Context, id uuid.UUID, params UpdateCertificationParams) (*Certification, error)
	DeleteCertification(ctx context.Context, id uuid.UUID) (*Certification, error)
}
