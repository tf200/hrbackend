package handler

import (
	"strings"
	"time"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"
	"hrbackend/pkg/ptr"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const performanceDateLayout = "2006-01-02"

type createPerformanceAssessmentRequest struct {
	EmployeeID     uuid.UUID                                 `json:"employee_id"     binding:"required"`
	AssessmentDate string                                    `json:"assessment_date" binding:"required,datetime=2006-01-02"`
	Notes          *string                                   `json:"notes"`
	Scores         []createPerformanceAssessmentScoreRequest `json:"scores"          binding:"required,min=1,dive"`
}

type createPerformanceAssessmentScoreRequest struct {
	QuestionCode string  `json:"question_code" binding:"required"`
	Rating       float64 `json:"rating"        binding:"required"`
	Remarks      *string `json:"remarks"`
}

type performanceAssessmentCatalogDomainResponse struct {
	Code      string                                         `json:"code"`
	NameNL    string                                         `json:"name_nl"`
	NameEN    string                                         `json:"name_en"`
	SortOrder int32                                          `json:"sort_order"`
	Questions []performanceAssessmentCatalogQuestionResponse `json:"questions"`
}

type performanceAssessmentCatalogQuestionResponse struct {
	Code          string `json:"code"`
	DomainCode    string `json:"domain_code"`
	TitleNL       string `json:"title_nl"`
	TitleEN       string `json:"title_en"`
	DescriptionNL string `json:"description_nl"`
	DescriptionEN string `json:"description_en"`
	SortOrder     int32  `json:"sort_order"`
}

type listPerformanceAssessmentsRequest struct {
	httpapi.PageRequest
	Search   *string `form:"search"`
	Status   *string `form:"status"`
	FromDate *string `form:"from_date" binding:"omitempty,datetime=2006-01-02"`
	ToDate   *string `form:"to_date"   binding:"omitempty,datetime=2006-01-02"`
}

type listPerformanceWorkAssignmentsRequest struct {
	httpapi.PageRequest
	EmployeeID *uuid.UUID `form:"employee_id,parser=encoding.TextUnmarshaler"`
	Status     *string    `form:"status"`
	DueBefore  *string    `form:"due_before" binding:"omitempty,datetime=2006-01-02"`
	DueAfter   *string    `form:"due_after"  binding:"omitempty,datetime=2006-01-02"`
}

type decidePerformanceWorkAssignmentRequest struct {
	Decision string  `json:"decision" binding:"required,oneof=approve request_revision"`
	Feedback *string `json:"feedback"`
}

type listPerformanceUpcomingRequest struct {
	WindowDays *int `form:"window_days" binding:"omitempty,min=1,max=180"`
}

type sendPerformanceUpcomingInvitationsRequest struct {
	EmployeeIDs []uuid.UUID `json:"employee_ids" binding:"required,min=1,dive"`
	Message     *string     `json:"message"`
}

type performanceAssessmentResponse struct {
	ID             uuid.UUID `json:"id"`
	Employee       gin.H     `json:"employee"`
	AssessmentDate string    `json:"assessment_date"`
	TotalScore     *float64  `json:"total_score"`
	Status         string    `json:"status"`
	Notes          *string   `json:"notes"`
	CreatedAt      string    `json:"created_at"`
}

type performanceAssessmentScoreResponse struct {
	ID            uuid.UUID `json:"id"`
	AssessmentID  uuid.UUID `json:"assessment_id"`
	QuestionCode  string    `json:"question_code"`
	DomainCode    string    `json:"domain_code"`
	TitleNL       string    `json:"title_nl"`
	TitleEN       string    `json:"title_en"`
	DescriptionNL string    `json:"description_nl"`
	DescriptionEN string    `json:"description_en"`
	Rating        float64   `json:"rating"`
	Remarks       *string   `json:"remarks"`
}

type performanceWorkAssignmentResponse struct {
	ID                    uuid.UUID `json:"id"`
	AssessmentID          uuid.UUID `json:"assessment_id"`
	Employee              gin.H     `json:"employee"`
	QuestionCode          string    `json:"question_code"`
	DomainCode            string    `json:"domain_code"`
	QuestionTextNL        string    `json:"question_text_nl"`
	QuestionTextEN        string    `json:"question_text_en"`
	Score                 float64   `json:"score"`
	AssignmentDescription string    `json:"assignment_description"`
	ImprovementNotes      *string   `json:"improvement_notes"`
	Expectations          *string   `json:"expectations"`
	Advice                *string   `json:"advice"`
	DueDate               *string   `json:"due_date"`
	Status                string    `json:"status"`
	SubmittedAt           *string   `json:"submitted_at"`
	SubmissionText        *string   `json:"submission_text"`
	Feedback              *string   `json:"feedback"`
	ReviewedAt            *string   `json:"reviewed_at"`
}

type performanceUpcomingResponse struct {
	EmployeeID         uuid.UUID `json:"employee_id"`
	EmployeeName       string    `json:"employee_name"`
	LastAssessmentDate *string   `json:"last_assessment_date"`
	NextAssessmentDate string    `json:"next_assessment_date"`
	IsOverdue          bool      `json:"is_overdue"`
	IsDueSoon          bool      `json:"is_due_soon"`
	DaysUntilDue       int       `json:"days_until_due"`
	IsFirstReview      bool      `json:"is_first_review"`
}

type performanceStatsResponse struct {
	TotalEmployees     int64    `json:"total_employees"`
	CompletedCount     int64    `json:"completed_count"`
	CompletedThisMonth int64    `json:"completed_this_month"`
	AverageScore       *float64 `json:"average_score"`
	CoveragePercent    int32    `json:"coverage_percent"`
	CoveredCount       int64    `json:"covered_count"`
}

type sendPerformanceUpcomingInvitationsResponse struct {
	SentCount int `json:"sent_count"`
}

func toCreatePerformanceAssessmentParams(
	req createPerformanceAssessmentRequest,
) (domain.CreatePerformanceAssessmentParams, error) {
	assessmentDate, err := time.Parse(performanceDateLayout, req.AssessmentDate)
	if err != nil {
		return domain.CreatePerformanceAssessmentParams{}, err
	}

	scores := make([]domain.CreatePerformanceAssessmentScoreParams, len(req.Scores))
	for i, score := range req.Scores {
		scores[i] = domain.CreatePerformanceAssessmentScoreParams{
			QuestionCode: strings.TrimSpace(score.QuestionCode),
			Rating:       score.Rating,
			Remarks:      ptr.TrimString(score.Remarks),
		}
	}

	return domain.CreatePerformanceAssessmentParams{
		EmployeeID:     req.EmployeeID,
		AssessmentDate: assessmentDate.UTC(),
		Notes:          ptr.TrimString(req.Notes),
		Scores:         scores,
	}, nil
}

func toPerformanceAssessmentCatalogResponse(
	items []domain.PerformanceDomain,
) []performanceAssessmentCatalogDomainResponse {
	results := make([]performanceAssessmentCatalogDomainResponse, len(items))
	for i, item := range items {
		questions := make([]performanceAssessmentCatalogQuestionResponse, len(item.Questions))
		for j, question := range item.Questions {
			questions[j] = performanceAssessmentCatalogQuestionResponse{
				Code:          question.Code,
				DomainCode:    question.DomainCode,
				TitleNL:       question.TitleNL,
				TitleEN:       question.TitleEN,
				DescriptionNL: question.DescriptionNL,
				DescriptionEN: question.DescriptionEN,
				SortOrder:     question.SortOrder,
			}
		}

		results[i] = performanceAssessmentCatalogDomainResponse{
			Code:      item.Code,
			NameNL:    item.NameNL,
			NameEN:    item.NameEN,
			SortOrder: item.SortOrder,
			Questions: questions,
		}
	}
	return results
}

func toListPerformanceAssessmentsParams(
	req listPerformanceAssessmentsRequest,
) (domain.ListPerformanceAssessmentsParams, error) {
	fromDate, err := parsePerformanceDatePtr(req.FromDate)
	if err != nil {
		return domain.ListPerformanceAssessmentsParams{}, err
	}
	toDate, err := parsePerformanceDatePtr(req.ToDate)
	if err != nil {
		return domain.ListPerformanceAssessmentsParams{}, err
	}

	return domain.ListPerformanceAssessmentsParams{
		Limit:    req.PageSize,
		Offset:   (req.Page - 1) * req.PageSize,
		Search:   ptr.TrimString(req.Search),
		Status:   ptr.TrimString(req.Status),
		FromDate: fromDate,
		ToDate:   toDate,
	}, nil
}

func toListPerformanceWorkAssignmentsParams(
	req listPerformanceWorkAssignmentsRequest,
) (domain.ListPerformanceWorkAssignmentsParams, error) {
	dueBefore, err := parsePerformanceDatePtr(req.DueBefore)
	if err != nil {
		return domain.ListPerformanceWorkAssignmentsParams{}, err
	}
	dueAfter, err := parsePerformanceDatePtr(req.DueAfter)
	if err != nil {
		return domain.ListPerformanceWorkAssignmentsParams{}, err
	}

	return domain.ListPerformanceWorkAssignmentsParams{
		Limit:      req.PageSize,
		Offset:     (req.Page - 1) * req.PageSize,
		EmployeeID: req.EmployeeID,
		Status:     ptr.TrimString(req.Status),
		DueBefore:  dueBefore,
		DueAfter:   dueAfter,
	}, nil
}

func toDecidePerformanceWorkAssignmentParams(
	req decidePerformanceWorkAssignmentRequest,
) domain.DecidePerformanceWorkAssignmentParams {
	return domain.DecidePerformanceWorkAssignmentParams{
		Decision: strings.TrimSpace(strings.ToLower(req.Decision)),
		Feedback: ptr.TrimString(req.Feedback),
	}
}

func toPerformanceAssessmentResponse(item *domain.PerformanceAssessment) performanceAssessmentResponse {
	return performanceAssessmentResponse{
		ID:             item.ID,
		Employee:       gin.H{"id": item.EmployeeID, "name": item.EmployeeName},
		AssessmentDate: item.AssessmentDate.Format(performanceDateLayout),
		TotalScore:     item.TotalScore,
		Status:         item.Status,
		Notes:          item.Notes,
		CreatedAt:      item.CreatedAt.Format(time.RFC3339),
	}
}

func toPerformanceAssessmentResponses(items []domain.PerformanceAssessment) []performanceAssessmentResponse {
	results := make([]performanceAssessmentResponse, len(items))
	for i, item := range items {
		results[i] = toPerformanceAssessmentResponse(&item)
	}
	return results
}

func toPerformanceAssessmentScoreResponses(
	items []domain.PerformanceAssessmentScore,
) []performanceAssessmentScoreResponse {
	results := make([]performanceAssessmentScoreResponse, len(items))
	for i, item := range items {
		results[i] = performanceAssessmentScoreResponse{
			ID:            item.ID,
			AssessmentID:  item.AssessmentID,
			QuestionCode:  item.QuestionCode,
			DomainCode:    item.DomainCode,
			TitleNL:       item.TitleNL,
			TitleEN:       item.TitleEN,
			DescriptionNL: item.DescriptionNL,
			DescriptionEN: item.DescriptionEN,
			Rating:        item.Rating,
			Remarks:       item.Remarks,
		}
	}
	return results
}

func toPerformanceWorkAssignmentResponse(
	item *domain.PerformanceWorkAssignment,
) performanceWorkAssignmentResponse {
	return performanceWorkAssignmentResponse{
		ID:                    item.ID,
		AssessmentID:          item.AssessmentID,
		Employee:              gin.H{"id": item.EmployeeID, "name": item.EmployeeName},
		QuestionCode:          item.QuestionCode,
		DomainCode:            item.DomainCode,
		QuestionTextNL:        item.QuestionTextNL,
		QuestionTextEN:        item.QuestionTextEN,
		Score:                 item.Score,
		AssignmentDescription: item.AssignmentDescription,
		ImprovementNotes:      item.ImprovementNotes,
		Expectations:          item.Expectations,
		Advice:                item.Advice,
		DueDate:               formatDatePtr(item.DueDate),
		Status:                item.Status,
		SubmittedAt:           formatTimestampPtr(item.SubmittedAt),
		SubmissionText:        item.SubmissionText,
		Feedback:              item.Feedback,
		ReviewedAt:            formatTimestampPtr(item.ReviewedAt),
	}
}

func toPerformanceWorkAssignmentResponses(
	items []domain.PerformanceWorkAssignment,
) []performanceWorkAssignmentResponse {
	results := make([]performanceWorkAssignmentResponse, len(items))
	for i, item := range items {
		results[i] = toPerformanceWorkAssignmentResponse(&item)
	}
	return results
}

func toPerformanceUpcomingResponses(
	items []domain.PerformanceUpcomingItem,
) []performanceUpcomingResponse {
	results := make([]performanceUpcomingResponse, len(items))
	for i, item := range items {
		var lastDateStr *string
		if item.LastAssessmentDate != nil {
			s := item.LastAssessmentDate.Format(performanceDateLayout)
			lastDateStr = &s
		}
		results[i] = performanceUpcomingResponse{
			EmployeeID:         item.EmployeeID,
			EmployeeName:       item.EmployeeName,
			LastAssessmentDate: lastDateStr,
			NextAssessmentDate: item.NextAssessmentDate.Format(performanceDateLayout),
			IsOverdue:          item.IsOverdue,
			IsDueSoon:          item.IsDueSoon,
			DaysUntilDue:       item.DaysUntilDue,
			IsFirstReview:      item.IsFirstReview,
		}
	}
	return results
}

func toPerformanceStatsResponse(stats *domain.PerformanceStats) performanceStatsResponse {
	return performanceStatsResponse{
		TotalEmployees:     stats.TotalEmployees,
		CompletedCount:     stats.CompletedCount,
		CompletedThisMonth: stats.CompletedThisMonth,
		AverageScore:       stats.AverageScore,
		CoveragePercent:    stats.CoveragePercent,
		CoveredCount:       stats.CoveredCount,
	}
}

func parsePerformanceDatePtr(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := time.Parse(performanceDateLayout, strings.TrimSpace(*value))
	if err != nil {
		return nil, err
	}
	utc := parsed.UTC()
	return &utc, nil
}

func formatDatePtr(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format(performanceDateLayout)
	return &formatted
}

func formatTimestampPtr(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format(time.RFC3339)
	return &formatted
}
