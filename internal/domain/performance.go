package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrPerformanceNotFound       = errors.New("performance resource not found")
	ErrPerformanceInvalidRequest = errors.New("invalid performance request")
	ErrPerformanceStateInvalid   = errors.New("performance resource is in invalid state")
)

const (
	PerformanceAssessmentStatusDraft     = "draft"
	PerformanceAssessmentStatusCompleted = "completed"

	PerformanceWorkAssignmentStatusOpen           = "open"
	PerformanceWorkAssignmentStatusSubmitted      = "submitted"
	PerformanceWorkAssignmentStatusApproved       = "approved"
	PerformanceWorkAssignmentStatusRevisionNeeded = "revision_needed"
)

type PerformanceAssessment struct {
	ID             uuid.UUID
	EmployeeID     uuid.UUID
	EmployeeName   string
	AssessmentDate time.Time
	TotalScore     *float64
	Status         string
	Notes          *string
	CreatedAt      time.Time
}

type PerformanceAssessmentScore struct {
	ID            uuid.UUID
	AssessmentID  uuid.UUID
	QuestionCode  string
	DomainCode    string
	TitleNL       string
	TitleEN       string
	DescriptionNL string
	DescriptionEN string
	Rating        float64
	Remarks       *string
}

type PerformanceDomain struct {
	Code      string
	NameNL    string
	NameEN    string
	SortOrder int32
	Questions []PerformanceQuestion
}

type PerformanceQuestion struct {
	Code          string
	DomainCode    string
	TitleNL       string
	TitleEN       string
	DescriptionNL string
	DescriptionEN string
	SortOrder     int32
}

type PerformanceEmployeeName struct {
	FirstName string
	LastName  string
}

type PerformanceWorkAssignment struct {
	ID                    uuid.UUID
	AssessmentID          uuid.UUID
	EmployeeID            uuid.UUID
	EmployeeName          string
	QuestionCode          string
	DomainCode            string
	QuestionTextNL        string
	QuestionTextEN        string
	Score                 float64
	AssignmentDescription string
	ImprovementNotes      *string
	Expectations          *string
	Advice                *string
	DueDate               *time.Time
	Status                string
	SubmittedAt           *time.Time
	SubmissionText        *string
	Feedback              *string
	ReviewedAt            *time.Time
}

const PerformanceReviewIntervalDays = 42

type PerformanceUpcomingItem struct {
	EmployeeID         uuid.UUID
	EmployeeName       string
	LastAssessmentDate *time.Time
	NextAssessmentDate time.Time
	IsOverdue          bool
	IsDueSoon          bool
	DaysUntilDue       int
	IsFirstReview      bool
}

type PerformanceStats struct {
	TotalEmployees     int64
	CompletedCount     int64
	CompletedThisMonth int64
	AverageScore       *float64
	CoveragePercent    int32
	CoveredCount       int64
}

type PerformanceAssessmentPage struct {
	Items      []PerformanceAssessment
	TotalCount int64
}

type PerformanceWorkAssignmentPage struct {
	Items      []PerformanceWorkAssignment
	TotalCount int64
}

type CreatePerformanceAssessmentScoreParams struct {
	QuestionCode string
	Rating       float64
	Remarks      *string
}

type CreatePerformanceAssessmentParams struct {
	EmployeeID     uuid.UUID
	AssessmentDate time.Time
	Notes          *string
	Scores         []CreatePerformanceAssessmentScoreParams
}

type CreatePerformanceAssessmentRecordParams struct {
	EmployeeID     uuid.UUID
	AssessmentDate time.Time
	TotalScore     *float64
	Notes          *string
}

type ListPerformanceAssessmentsParams struct {
	Limit    int32
	Offset   int32
	Search   *string
	Status   *string
	FromDate *time.Time
	ToDate   *time.Time
}

type ListPerformanceWorkAssignmentsParams struct {
	Limit      int32
	Offset     int32
	EmployeeID *uuid.UUID
	Status     *string
	DueBefore  *time.Time
	DueAfter   *time.Time
}

type DecidePerformanceWorkAssignmentParams struct {
	Decision string
	Feedback *string
}

type PerformanceTxRepository interface {
	GetActiveEmployeeName(ctx context.Context, employeeID uuid.UUID) (*PerformanceEmployeeName, error)
	CreateAssessment(
		ctx context.Context,
		params CreatePerformanceAssessmentRecordParams,
		employeeName PerformanceEmployeeName,
	) (*PerformanceAssessment, error)
	CreateAssessmentScore(ctx context.Context, assessmentID uuid.UUID, score CreatePerformanceAssessmentScoreParams) error
	GetWorkAssignmentStatusForUpdate(ctx context.Context, id uuid.UUID) (string, error)
	UpdateWorkAssignmentDecision(ctx context.Context, id uuid.UUID, status string, feedback *string) error
}

type PerformanceRepository interface {
	WithTx(ctx context.Context, fn func(tx PerformanceTxRepository) error) error
	ListAssessmentCatalog(ctx context.Context) ([]PerformanceDomain, error)
	ListAssessments(
		ctx context.Context,
		params ListPerformanceAssessmentsParams,
	) (*PerformanceAssessmentPage, error)
	GetAssessmentByID(ctx context.Context, id uuid.UUID) (*PerformanceAssessment, error)
	DeleteAssessment(ctx context.Context, id uuid.UUID) (bool, error)
	ListAssessmentScores(ctx context.Context, assessmentID uuid.UUID) ([]PerformanceAssessmentScore, error)
	ListWorkAssignments(
		ctx context.Context,
		params ListPerformanceWorkAssignmentsParams,
	) (*PerformanceWorkAssignmentPage, error)
	GetWorkAssignmentByID(ctx context.Context, id uuid.UUID) (*PerformanceWorkAssignment, error)
	ListUpcoming(ctx context.Context, windowDays int) ([]PerformanceUpcomingItem, error)
	GetStats(ctx context.Context) (*PerformanceStats, error)
}

type PerformanceService interface {
	ListAssessmentCatalog(ctx context.Context) ([]PerformanceDomain, error)
	CreateAssessment(ctx context.Context, params CreatePerformanceAssessmentParams) (*PerformanceAssessment, error)
	ListAssessments(
		ctx context.Context,
		params ListPerformanceAssessmentsParams,
	) (*PerformanceAssessmentPage, error)
	GetAssessmentByID(ctx context.Context, id uuid.UUID) (*PerformanceAssessment, error)
	DeleteAssessment(ctx context.Context, id uuid.UUID) (bool, error)
	ListAssessmentScores(ctx context.Context, assessmentID uuid.UUID) ([]PerformanceAssessmentScore, error)
	ListWorkAssignments(
		ctx context.Context,
		params ListPerformanceWorkAssignmentsParams,
	) (*PerformanceWorkAssignmentPage, error)
	GetWorkAssignmentByID(ctx context.Context, id uuid.UUID) (*PerformanceWorkAssignment, error)
	DecideWorkAssignment(
		ctx context.Context,
		id uuid.UUID,
		params DecidePerformanceWorkAssignmentParams,
	) (*PerformanceWorkAssignment, error)
	ListUpcoming(ctx context.Context, windowDays int) ([]PerformanceUpcomingItem, error)
	SendUpcomingInvitations(
		ctx context.Context,
		employeeIDs []uuid.UUID,
		message *string,
	) (int, error)
	GetStats(ctx context.Context) (*PerformanceStats, error)
}
