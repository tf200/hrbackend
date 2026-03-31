package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrActiveHandbookNotFound          = errors.New("no active handbook assigned")
	ErrDraftTemplateAlreadyExists      = errors.New("a draft template already exists for this department")
	ErrTemplateNotFound                = errors.New("template not found")
	ErrTemplateNotDraft                = errors.New("template is not in draft status")
	ErrTemplateNotPublished            = errors.New("template is not in published status")
	ErrTemplateHasNoSteps              = errors.New("template must contain at least one step before publishing")
	ErrStepNotFound                    = errors.New("step not found")
	ErrInvalidStepReorder              = errors.New("ordered_step_ids must match the template steps exactly")
	ErrInvalidStepContent              = errors.New("invalid step content")
	ErrEmployeeHandbookNotFound        = errors.New("employee handbook not found")
	ErrEmployeeHandbookNotActive       = errors.New("employee handbook is not active")
	ErrInvalidAssignmentStatusFilter   = errors.New("invalid assignment status filter")
	ErrHandbookInvalidRequest          = errors.New("invalid handbook request")
	ErrEligibleEmployeePermissionCheck = errors.New("failed to check eligible employee permissions")
)

type HandbookTemplate struct {
	ID           uuid.UUID
	DepartmentID uuid.UUID
	Title        string
	Description  *string
	Version      int32
	Status       string
	PublishedAt  *time.Time
	ArchivedAt   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type HandbookStep struct {
	ID         uuid.UUID
	TemplateID uuid.UUID
	SortOrder  int32
	Kind       string
	Title      string
	Body       *string
	Content    any
	IsRequired bool
	UpdatedAt  time.Time
}

type MyHandbookStep struct {
	StepID      uuid.UUID
	SortOrder   int32
	Kind        string
	Title       string
	Body        *string
	Content     any
	IsRequired  bool
	Status      string
	StartedAt   *time.Time
	CompletedAt *time.Time
	Response    any
}

type MyActiveHandbook struct {
	HandbookID      uuid.UUID
	EmployeeID      uuid.UUID
	Status          string
	AssignedAt      time.Time
	StartedAt       *time.Time
	CompletedAt     *time.Time
	DueAt           *time.Time
	TemplateID      uuid.UUID
	TemplateTitle   string
	TemplateDesc    *string
	TemplateVersion int32
	DepartmentID    uuid.UUID
	DepartmentName  string
	Steps           []MyHandbookStep
}

type StartedHandbook struct {
	HandbookID uuid.UUID
	Status     string
	StartedAt  *time.Time
}

type CompletedHandbookStep struct {
	HandbookID     uuid.UUID
	StepID         uuid.UUID
	StepStatus     string
	CompletedAt    time.Time
	HandbookStatus string
}

type EmployeeHandbookAssignment struct {
	EmployeeHandbookID uuid.UUID
	EmployeeID         uuid.UUID
	TemplateID         uuid.UUID
	TemplateVersion    int32
	AssignedAt         time.Time
	StartedAt          *time.Time
	CompletedAt        *time.Time
	DueAt              *time.Time
	Status             string
}

type WaivedEmployeeHandbook struct {
	EmployeeHandbookID uuid.UUID
	EmployeeID         uuid.UUID
	Status             string
	CompletedAt        *time.Time
}

type HandbookAssignmentHistoryEntry struct {
	ID                 uuid.UUID
	EmployeeHandbookID *uuid.UUID
	EmployeeID         uuid.UUID
	TemplateID         uuid.UUID
	TemplateVersion    int32
	Event              string
	ActorEmployeeID    *uuid.UUID
	Metadata           any
	CreatedAt          time.Time
}

type EmployeeHandbookAssignmentSummary struct {
	EmployeeID             uuid.UUID
	FirstName              string
	LastName               string
	DepartmentID           *uuid.UUID
	DepartmentName         *string
	EmployeeHandbookID     *uuid.UUID
	TemplateID             *uuid.UUID
	TemplateTitle          *string
	TemplateVersion        *int32
	HandbookStatus         string
	AssignedAt             *time.Time
	StartedAt              *time.Time
	CompletedAt            *time.Time
	DueAt                  *time.Time
	RequiredStepsTotal     int32
	RequiredStepsCompleted int32
}

type EmployeeHandbookAssignmentPage struct {
	Items      []EmployeeHandbookAssignmentSummary
	TotalCount int64
}

type EligibleEmployee struct {
	EmployeeID     uuid.UUID
	FirstName      string
	LastName       string
	DepartmentID   *uuid.UUID
	DepartmentName *string
}

type EligibleEmployeePage struct {
	Items      []EligibleEmployee
	TotalCount int64
}

type EmployeeHandbookDetails struct {
	EmployeeHandbookID uuid.UUID
	EmployeeID         uuid.UUID
	FirstName          string
	LastName           string
	Status             string
	AssignedAt         time.Time
	StartedAt          *time.Time
	CompletedAt        *time.Time
	DueAt              *time.Time
	TemplateID         uuid.UUID
	TemplateTitle      string
	TemplateDesc       *string
	TemplateVersion    int32
	DepartmentID       uuid.UUID
	DepartmentName     string
	Steps              []MyHandbookStep
}

type HandbookEmployeeProfile struct {
	ID           uuid.UUID
	DepartmentID *uuid.UUID
}

type CreateTemplateForDepartmentParams struct {
	DepartmentID uuid.UUID
	Title        string
	Description  *string
}

type CloneTemplateToDraftParams struct {
	SourceTemplateID uuid.UUID
}

type UpdateTemplateParams struct {
	TemplateID     uuid.UUID
	Title          *string
	SetTitle       bool
	Description    *string
	SetDescription bool
}

type PublishTemplateParams struct {
	TemplateID uuid.UUID
}

type CreateStepParams struct {
	TemplateID uuid.UUID
	SortOrder  int32
	Kind       string
	Title      string
	Body       *string
	Content    []byte
	IsRequired *bool
}

type UpdateStepParams struct {
	StepID          uuid.UUID
	Title           *string
	SetTitle        bool
	Body            *string
	SetBody         bool
	Content         []byte
	ContentProvided bool
	IsRequired      *bool
	SetIsRequired   bool
}

type DeleteStepParams struct {
	StepID uuid.UUID
}

type ReorderStepsParams struct {
	TemplateID     uuid.UUID
	OrderedStepIDs []uuid.UUID
}

type AssignTemplateToEmployeeParams struct {
	EmployeeID uuid.UUID
	TemplateID uuid.UUID
}

type WaiveEmployeeHandbookParams struct {
	EmployeeHandbookID uuid.UUID
	Reason             *string
}

type ListEmployeeHandbookAssignmentsParams struct {
	Limit        int32
	Offset       int32
	DepartmentID *uuid.UUID
	Search       *string
	Status       *string
}

type ListEligibleEmployeesParams struct {
	Limit        int32
	Offset       int32
	DepartmentID *uuid.UUID
	Search       *string
}

type CompleteHandbookStepParams struct {
	EmployeeHandbookID uuid.UUID
	StepID             uuid.UUID
	Response           []byte
}

type CreateAssignmentHistoryParams struct {
	EmployeeHandbookID *uuid.UUID
	EmployeeID         uuid.UUID
	TemplateID         uuid.UUID
	TemplateVersion    int32
	Event              string
	ActorEmployeeID    *uuid.UUID
	Metadata           []byte
}

type HandbookRepository interface {
	WithTx(ctx context.Context, fn func(tx HandbookRepository) error) error
	GetActiveEmployeeHandbookByEmployeeID(ctx context.Context, employeeID uuid.UUID) (*MyActiveHandbook, error)
	ListEmployeeHandbookStepsByHandbookID(ctx context.Context, handbookID uuid.UUID) ([]MyHandbookStep, error)
	MarkEmployeeHandbookStarted(ctx context.Context, handbookID uuid.UUID) (*EmployeeHandbookAssignment, error)
	CompleteEmployeeHandbookStep(ctx context.Context, params CompleteHandbookStepParams) (*CompletedHandbookStep, error)
	CountRemainingRequiredHandbookSteps(ctx context.Context, handbookID uuid.UUID) (int32, error)
	MarkEmployeeHandbookCompleted(ctx context.Context, handbookID uuid.UUID) (*EmployeeHandbookAssignment, error)
	CreateEmployeeHandbookAssignmentHistory(ctx context.Context, params CreateAssignmentHistoryParams) error
	CreateHandbookTemplateForDepartment(ctx context.Context, actorEmployeeID uuid.UUID, params CreateTemplateForDepartmentParams) (*HandbookTemplate, error)
	CloneHandbookTemplateToDraft(ctx context.Context, actorEmployeeID uuid.UUID, params CloneTemplateToDraftParams) (*HandbookTemplate, error)
	GetHandbookTemplateByID(ctx context.Context, templateID uuid.UUID) (*HandbookTemplate, error)
	UpdateHandbookTemplateMetadata(ctx context.Context, params UpdateTemplateParams) (*HandbookTemplate, error)
	CountHandbookStepsByTemplateID(ctx context.Context, templateID uuid.UUID) (int32, error)
	PublishHandbookTemplate(ctx context.Context, actorEmployeeID uuid.UUID, params PublishTemplateParams) (*HandbookTemplate, error)
	ListHandbookTemplatesByDepartment(ctx context.Context, departmentID uuid.UUID) ([]HandbookTemplate, error)
	CreateHandbookStep(ctx context.Context, params CreateStepParams) (*HandbookStep, error)
	ListHandbookStepsByTemplate(ctx context.Context, templateID uuid.UUID) ([]HandbookStep, error)
	GetHandbookStepByID(ctx context.Context, stepID uuid.UUID) (*HandbookStep, error)
	UpdateHandbookStepByID(ctx context.Context, params UpdateStepParams) (*HandbookStep, error)
	DeleteHandbookStepByID(ctx context.Context, stepID uuid.UUID) error
	UpdateHandbookStepSortOrder(ctx context.Context, stepID uuid.UUID, sortOrder int32) error
	WaiveActiveEmployeeHandbooksByEmployeeID(ctx context.Context, employeeID uuid.UUID) error
	CreateEmployeeHandbookFromTemplate(ctx context.Context, actorEmployeeID uuid.UUID, params AssignTemplateToEmployeeParams) (*EmployeeHandbookAssignment, error)
	GetEmployeeHandbookByID(ctx context.Context, handbookID uuid.UUID) (*EmployeeHandbookAssignment, error)
	WaiveEmployeeHandbookByID(ctx context.Context, handbookID uuid.UUID) (*WaivedEmployeeHandbook, error)
	ListEmployeeHandbookAssignmentHistoryByEmployeeID(ctx context.Context, employeeID uuid.UUID, limit, offset int32) ([]HandbookAssignmentHistoryEntry, error)
	ListEmployeeHandbookAssignments(ctx context.Context, params ListEmployeeHandbookAssignmentsParams) (*EmployeeHandbookAssignmentPage, error)
	GetEmployeeHandbookDetailsByID(ctx context.Context, handbookID uuid.UUID) (*EmployeeHandbookDetails, error)
	GetUserIDByEmployeeID(ctx context.Context, employeeID uuid.UUID) (uuid.UUID, error)
	CheckUserPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error)
	GetEmployeeProfileByID(ctx context.Context, employeeID uuid.UUID) (*HandbookEmployeeProfile, error)
	ListEligibleEmployeesForHandbookAssignment(ctx context.Context, params ListEligibleEmployeesParams) (*EligibleEmployeePage, error)
}

type HandbookService interface {
	GetMyActiveHandbook(ctx context.Context, employeeID uuid.UUID) (*MyActiveHandbook, error)
	StartMyHandbook(ctx context.Context, employeeID uuid.UUID) (*StartedHandbook, error)
	CompleteMyHandbookStep(ctx context.Context, employeeID, stepID uuid.UUID, response []byte) (*CompletedHandbookStep, error)
	CreateTemplateForDepartment(ctx context.Context, actorEmployeeID uuid.UUID, params CreateTemplateForDepartmentParams) (*HandbookTemplate, error)
	CloneTemplateToDraft(ctx context.Context, actorEmployeeID uuid.UUID, params CloneTemplateToDraftParams) (*HandbookTemplate, error)
	UpdateTemplate(ctx context.Context, params UpdateTemplateParams) (*HandbookTemplate, error)
	PublishTemplate(ctx context.Context, actorEmployeeID uuid.UUID, params PublishTemplateParams) (*HandbookTemplate, error)
	ListTemplatesByDepartment(ctx context.Context, departmentID uuid.UUID) ([]HandbookTemplate, error)
	CreateStep(ctx context.Context, params CreateStepParams) (*HandbookStep, error)
	UpdateStep(ctx context.Context, params UpdateStepParams) (*HandbookStep, error)
	DeleteStep(ctx context.Context, params DeleteStepParams) error
	ReorderTemplateSteps(ctx context.Context, params ReorderStepsParams) ([]HandbookStep, error)
	ListStepsByTemplate(ctx context.Context, templateID uuid.UUID) ([]HandbookStep, error)
	AssignTemplateToEmployee(ctx context.Context, actorEmployeeID uuid.UUID, params AssignTemplateToEmployeeParams) (*EmployeeHandbookAssignment, error)
	WaiveEmployeeHandbook(ctx context.Context, actorEmployeeID uuid.UUID, params WaiveEmployeeHandbookParams) (*WaivedEmployeeHandbook, error)
	ListEmployeeHandbookHistory(ctx context.Context, employeeID uuid.UUID) ([]HandbookAssignmentHistoryEntry, error)
	ListEligibleEmployees(ctx context.Context, actorEmployeeID uuid.UUID, params ListEligibleEmployeesParams) (*EligibleEmployeePage, error)
	ListEmployeeHandbookAssignments(ctx context.Context, params ListEmployeeHandbookAssignmentsParams) (*EmployeeHandbookAssignmentPage, error)
	GetEmployeeHandbookDetails(ctx context.Context, handbookID uuid.UUID) (*EmployeeHandbookDetails, error)
}
