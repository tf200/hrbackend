package seed

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var performanceSeedNamespace = uuid.MustParse("4c286364-c481-4d8c-bd5f-5f6ad824f111")

type PerformanceSeed struct {
	EmployeeAlias string
}

type PerformanceSeeder struct {
	Assessments []PerformanceSeed
}

type performanceEmployee struct {
	ID    uuid.UUID
	Name  string
	Alias string
}

type performanceScoreSeed struct {
	DomainID string
	ItemID   string
	Rating   float64
	Remarks  *string
}

func (s PerformanceSeeder) Name() string {
	return "performance"
}

func (s PerformanceSeeder) Seed(ctx context.Context, env Env) error {
	if len(s.Assessments) == 0 {
		return nil
	}
	if env.State == nil {
		return fmt.Errorf("seed performance: state is required")
	}

	tx, ok := env.DB.(pgx.Tx)
	if !ok {
		return fmt.Errorf("seed performance: env DB must be pgx.Tx")
	}

	employees, err := s.loadEmployees(ctx, tx, env)
	if err != nil {
		return err
	}

	baseDate := performanceDateOnly(time.Now().UTC()).AddDate(0, 0, -84)
	for i, employee := range employees {
		assessmentDate := baseDate.AddDate(0, 0, i*13)
		assessmentID := performanceDeterministicID(
			"assessment",
			employee.ID.String(),
			assessmentDate.Format("2006-01-02"),
		)
		scores := buildPerformanceScores(i)

		if err := upsertPerformanceAssessment(
			ctx,
			tx,
			assessmentID,
			employee.ID,
			assessmentDate,
			averagePerformanceScore(scores),
			"completed",
			fmt.Sprintf("Automatisch gegenereerde seed beoordeling voor %s", employee.Name),
		); err != nil {
			return fmt.Errorf("seed performance[%s]: %w", employee.Alias, err)
		}

		for _, score := range scores {
			scoreID := performanceDeterministicID(
				"score",
				assessmentID.String(),
				score.DomainID,
				score.ItemID,
			)
			if err := upsertPerformanceScore(ctx, tx, scoreID, assessmentID, score); err != nil {
				return fmt.Errorf("seed performance[%s]: %w", employee.Alias, err)
			}

			if score.Rating > 5 {
				continue
			}

			assignmentID := performanceDeterministicID(
				"assignment",
				assessmentID.String(),
				score.ItemID,
			)
			if err := upsertPerformanceAssignment(
				ctx,
				tx,
				assignmentID,
				assessmentID,
				employee.ID,
				score,
				assessmentDate.AddDate(0, 0, 14),
				performanceAssignmentStatusForIndex(i),
			); err != nil {
				return fmt.Errorf("seed performance[%s]: %w", employee.Alias, err)
			}
		}
	}

	return nil
}

func (s PerformanceSeeder) loadEmployees(
	ctx context.Context,
	tx pgx.Tx,
	env Env,
) ([]performanceEmployee, error) {
	items := make([]performanceEmployee, 0, len(s.Assessments))
	for _, assessment := range s.Assessments {
		alias := strings.TrimSpace(assessment.EmployeeAlias)
		if alias == "" {
			return nil, fmt.Errorf("seed performance: employee alias is required")
		}

		employeeID, ok := env.State.EmployeeID(alias)
		if !ok {
			return nil, fmt.Errorf("seed performance[%s]: employee alias not found in state", alias)
		}

		var firstName string
		var lastName string
		if err := tx.QueryRow(
			ctx,
			`SELECT first_name, last_name
			 FROM employee_profile
			 WHERE id = $1`,
			employeeID,
		).Scan(&firstName, &lastName); err != nil {
			return nil, fmt.Errorf("seed performance[%s]: load employee: %w", alias, err)
		}

		items = append(items, performanceEmployee{
			ID:    employeeID,
			Name:  strings.TrimSpace(firstName + " " + lastName),
			Alias: alias,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Alias < items[j].Alias
	})

	return items, nil
}

func upsertPerformanceAssessment(
	ctx context.Context,
	tx pgx.Tx,
	id uuid.UUID,
	employeeID uuid.UUID,
	assessmentDate time.Time,
	totalScore float64,
	status string,
	notes string,
) error {
	_, err := tx.Exec(
		ctx,
		`INSERT INTO performance_assessments (
			id,
			employee_id,
			assessment_date,
			total_score,
			status,
			notes
		)
		VALUES ($1, $2, $3, $4, $5::performance_assessment_status_enum, $6)
		ON CONFLICT (id)
		DO UPDATE SET
			total_score = EXCLUDED.total_score,
			status = EXCLUDED.status,
			notes = EXCLUDED.notes,
			updated_at = NOW()`,
		id,
		employeeID,
		assessmentDate,
		totalScore,
		status,
		notes,
	)
	if err != nil {
		return fmt.Errorf("upsert assessment: %w", err)
	}

	return nil
}

func upsertPerformanceScore(
	ctx context.Context,
	tx pgx.Tx,
	id uuid.UUID,
	assessmentID uuid.UUID,
	score performanceScoreSeed,
) error {
	_, err := tx.Exec(
		ctx,
		`INSERT INTO performance_assessment_scores (
			id,
			assessment_id,
			domain_id,
			item_id,
			rating,
			remarks
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id)
		DO UPDATE SET
			rating = EXCLUDED.rating,
			remarks = EXCLUDED.remarks,
			updated_at = NOW()`,
		id,
		assessmentID,
		score.DomainID,
		score.ItemID,
		score.Rating,
		score.Remarks,
	)
	if err != nil {
		return fmt.Errorf("upsert score: %w", err)
	}

	return nil
}

func upsertPerformanceAssignment(
	ctx context.Context,
	tx pgx.Tx,
	id uuid.UUID,
	assessmentID uuid.UUID,
	employeeID uuid.UUID,
	score performanceScoreSeed,
	dueDate time.Time,
	status string,
) error {
	var submittedAt *time.Time
	var reviewedAt *time.Time
	var submissionText *string
	var feedback *string

	now := time.Now().UTC()
	if status == "submitted" || status == "approved" || status == "revision_needed" {
		t := now.Add(-72 * time.Hour)
		submittedAt = &t
		text := "Seed reflectie: ik heb concrete acties toegepast in de praktijk."
		submissionText = &text
	}
	if status == "approved" || status == "revision_needed" {
		t := now.Add(-24 * time.Hour)
		reviewedAt = &t
		text := "Seed feedback van leidinggevende."
		feedback = &text
	}

	_, err := tx.Exec(
		ctx,
		`INSERT INTO performance_work_assignments (
			id,
			assessment_id,
			employee_id,
			question_id,
			domain_id,
			question_text,
			score,
			assignment_description,
			due_date,
			status,
			submitted_at,
			submission_text,
			feedback,
			reviewed_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::performance_work_assignment_status_enum, $11, $12, $13, $14)
		ON CONFLICT (id)
		DO UPDATE SET
			status = EXCLUDED.status,
			submitted_at = EXCLUDED.submitted_at,
			submission_text = EXCLUDED.submission_text,
			feedback = EXCLUDED.feedback,
			reviewed_at = EXCLUDED.reviewed_at,
			updated_at = NOW()`,
		id,
		assessmentID,
		employeeID,
		score.ItemID,
		score.DomainID,
		fmt.Sprintf("Reflectiepunt %s", score.ItemID),
		score.Rating,
		fmt.Sprintf("Werk aan verbetering voor %s (%s)", score.DomainID, score.ItemID),
		dueDate,
		status,
		submittedAt,
		submissionText,
		feedback,
		reviewedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert assignment: %w", err)
	}

	return nil
}

func buildPerformanceScores(seed int) []performanceScoreSeed {
	items := []performanceScoreSeed{
		{DomainID: "veilig-stabiel-leefklimaat", ItemID: "vsl-1", Rating: 8},
		{DomainID: "veilig-stabiel-leefklimaat", ItemID: "vsl-2", Rating: 7},
		{DomainID: "adl-begeleiding", ItemID: "adl-2", Rating: 6},
		{DomainID: "stimuleren-ontwikkeling", ItemID: "so-2", Rating: 5, Remarks: strPtr("Meer focus op emotieregulatie.")},
		{DomainID: "opvoeden-begrenzen", ItemID: "ob-1", Rating: 4, Remarks: strPtr("Nog onvoldoende consistent in begrenzen.")},
		{DomainID: "individuele-begeleiding", ItemID: "ib-3", Rating: 6},
		{DomainID: "individuele-begeleiding", ItemID: "ib-5", Rating: 5, Remarks: strPtr("Systemisch werken vraagt meer structuur.")},
	}

	for i := range items {
		if items[i].Rating >= 6 {
			continue
		}
		if (seed+i)%3 == 0 {
			items[i].Rating = 5
		}
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].DomainID == items[j].DomainID {
			return items[i].ItemID < items[j].ItemID
		}
		return items[i].DomainID < items[j].DomainID
	})

	return items
}

func performanceAssignmentStatusForIndex(i int) string {
	switch i % 4 {
	case 0:
		return "open"
	case 1:
		return "submitted"
	case 2:
		return "approved"
	default:
		return "revision_needed"
	}
}

func averagePerformanceScore(scores []performanceScoreSeed) float64 {
	if len(scores) == 0 {
		return 0
	}

	var sum float64
	for _, score := range scores {
		sum += score.Rating
	}

	return sum / float64(len(scores))
}

func performanceDeterministicID(parts ...string) uuid.UUID {
	return uuid.NewSHA1(performanceSeedNamespace, []byte(strings.Join(parts, "|")))
}

func performanceDateOnly(value time.Time) time.Time {
	utc := value.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}
