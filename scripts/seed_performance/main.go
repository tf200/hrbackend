package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var seedNamespace = uuid.MustParse("4c286364-c481-4d8c-bd5f-5f6ad824f111")

type employee struct {
	ID   uuid.UUID
	Name string
}

type scoreSeed struct {
	DomainID string
	ItemID   string
	Rating   float64
	Remarks  *string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	dbSource := firstNonEmpty(os.Getenv("MIGRATION_DB_SOURCE"), os.Getenv("DB_SOURCE"))
	if dbSource == "" {
		exitErr(errors.New("MIGRATION_DB_SOURCE or DB_SOURCE is required"))
	}

	pool, err := pgxpool.New(ctx, dbSource)
	if err != nil {
		exitErr(fmt.Errorf("connect db: %w", err))
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		exitErr(fmt.Errorf("ping db: %w", err))
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		exitErr(fmt.Errorf("begin tx: %w", err))
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	employees, err := loadEmployees(ctx, tx, 6)
	if err != nil {
		exitErr(err)
	}
	if len(employees) == 0 {
		exitErr(errors.New("no active employees found; seed employees first"))
	}

	createdAssessments := 0
	createdScores := 0
	createdAssignments := 0

	baseDate := dateOnly(time.Now().UTC()).AddDate(0, 0, -84)

	for i, emp := range employees {
		assessmentDate := baseDate.AddDate(0, 0, i*13)
		assessmentID := deterministicID("assessment", emp.ID.String(), assessmentDate.Format("2006-01-02"))

		scores := buildScores(i)
		totalScore := avg(scores)
		status := "completed"
		notes := fmt.Sprintf("Automatisch gegenereerde seed beoordeling voor %s", emp.Name)

		if err := upsertAssessment(ctx, tx, assessmentID, emp.ID, assessmentDate, totalScore, status, notes); err != nil {
			exitErr(err)
		}
		createdAssessments++

		for _, score := range scores {
			scoreID := deterministicID("score", assessmentID.String(), score.DomainID, score.ItemID)
			if err := upsertScore(ctx, tx, scoreID, assessmentID, score); err != nil {
				exitErr(err)
			}
			createdScores++

			if score.Rating <= 5 {
				assignmentID := deterministicID("assignment", assessmentID.String(), score.ItemID)
				status := assignmentStatusForIndex(i)
				dueDate := assessmentDate.AddDate(0, 0, 14)
				if err := upsertAssignment(
					ctx,
					tx,
					assignmentID,
					assessmentID,
					emp.ID,
					score,
					dueDate,
					status,
				); err != nil {
					exitErr(err)
				}
				createdAssignments++
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		exitErr(fmt.Errorf("commit tx: %w", err))
	}

	fmt.Println("Performance seed completed.")
	fmt.Printf("Assessments upserted: %d\n", createdAssessments)
	fmt.Printf("Scores upserted: %d\n", createdScores)
	fmt.Printf("Work assignments upserted: %d\n", createdAssignments)
}

func loadEmployees(ctx context.Context, tx pgx.Tx, limit int) ([]employee, error) {
	rows, err := tx.Query(
		ctx,
		`SELECT id, first_name, last_name
		 FROM employee_profile
		 WHERE is_archived = FALSE
		   AND COALESCE(out_of_service, FALSE) = FALSE
		 ORDER BY created_at ASC
		 LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("load employees: %w", err)
	}
	defer rows.Close()

	items := make([]employee, 0)
	for rows.Next() {
		var id uuid.UUID
		var firstName, lastName string
		if err := rows.Scan(&id, &firstName, &lastName); err != nil {
			return nil, fmt.Errorf("scan employee: %w", err)
		}
		items = append(items, employee{ID: id, Name: strings.TrimSpace(firstName + " " + lastName)})
	}

	return items, rows.Err()
}

func upsertAssessment(
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

func upsertScore(
	ctx context.Context,
	tx pgx.Tx,
	id uuid.UUID,
	assessmentID uuid.UUID,
	score scoreSeed,
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

func upsertAssignment(
	ctx context.Context,
	tx pgx.Tx,
	id uuid.UUID,
	assessmentID uuid.UUID,
	employeeID uuid.UUID,
	score scoreSeed,
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
		txt := "Seed reflectie: ik heb concrete acties toegepast in de praktijk."
		submissionText = &txt
	}
	if status == "approved" || status == "revision_needed" {
		t := now.Add(-24 * time.Hour)
		reviewedAt = &t
		msg := "Seed feedback van leidinggevende."
		feedback = &msg
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

func buildScores(seed int) []scoreSeed {
	all := []scoreSeed{
		{DomainID: "veilig-stabiel-leefklimaat", ItemID: "vsl-1", Rating: 8},
		{DomainID: "veilig-stabiel-leefklimaat", ItemID: "vsl-2", Rating: 7},
		{DomainID: "adl-begeleiding", ItemID: "adl-2", Rating: 6},
		{DomainID: "stimuleren-ontwikkeling", ItemID: "so-2", Rating: 5, Remarks: strPtr("Meer focus op emotieregulatie.")},
		{DomainID: "opvoeden-begrenzen", ItemID: "ob-1", Rating: 4, Remarks: strPtr("Nog onvoldoende consistent in begrenzen.")},
		{DomainID: "individuele-begeleiding", ItemID: "ib-3", Rating: 6},
		{DomainID: "individuele-begeleiding", ItemID: "ib-5", Rating: 5, Remarks: strPtr("Systemisch werken vraagt meer structuur.")},
	}

	// small deterministic variance per employee
	for i := range all {
		if all[i].Rating >= 6 {
			continue
		}
		if (seed+i)%3 == 0 {
			all[i].Rating = 5
		}
	}

	sort.Slice(all, func(i, j int) bool {
		if all[i].DomainID == all[j].DomainID {
			return all[i].ItemID < all[j].ItemID
		}
		return all[i].DomainID < all[j].DomainID
	})

	return all
}

func assignmentStatusForIndex(i int) string {
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

func avg(scores []scoreSeed) float64 {
	if len(scores) == 0 {
		return 0
	}
	var sum float64
	for _, score := range scores {
		sum += score.Rating
	}
	return sum / float64(len(scores))
}

func deterministicID(parts ...string) uuid.UUID {
	return uuid.NewSHA1(seedNamespace, []byte(strings.Join(parts, "|")))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func dateOnly(value time.Time) time.Time {
	utc := value.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func strPtr(value string) *string {
	return &value
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
