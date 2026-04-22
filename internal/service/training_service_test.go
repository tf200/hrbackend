package service

import (
	"context"
	"testing"
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

func TestTrainingServiceListTrainingAssignmentsNormalizesFilters(t *testing.T) {
	repo := &fakeTrainingRepository{}
	svc := &TrainingService{repository: repo}

	status := " Completed "
	search := "  jane doe  "

	_, err := svc.ListTrainingAssignments(context.Background(), domain.ListTrainingAssignmentsParams{
		Limit:          20,
		Offset:         40,
		EmployeeSearch: &search,
		Status:         &status,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if repo.lastListTrainingAssignmentsParams.EmployeeSearch == nil {
		t.Fatalf("expected employee search to be preserved")
	}
	if got := *repo.lastListTrainingAssignmentsParams.EmployeeSearch; got != "jane doe" {
		t.Fatalf("expected normalized employee search, got %q", got)
	}
	if repo.lastListTrainingAssignmentsParams.Status == nil {
		t.Fatalf("expected status filter to be preserved")
	}
	if got := *repo.lastListTrainingAssignmentsParams.Status; got != "completed" {
		t.Fatalf("expected normalized status, got %q", got)
	}
}

func TestTrainingServiceListTrainingAssignmentsAllowsDefaultCurrentView(t *testing.T) {
	repo := &fakeTrainingRepository{}
	svc := &TrainingService{repository: repo}

	status := "   "

	_, err := svc.ListTrainingAssignments(context.Background(), domain.ListTrainingAssignmentsParams{
		Limit:  10,
		Offset: 0,
		Status: &status,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if repo.lastListTrainingAssignmentsParams.Status != nil {
		t.Fatalf("expected blank status to normalize to nil, got %#v", repo.lastListTrainingAssignmentsParams.Status)
	}
}

func TestTrainingServiceListTrainingAssignmentsRejectsInvalidStatus(t *testing.T) {
	svc := &TrainingService{repository: &fakeTrainingRepository{}}

	status := "paused"

	_, err := svc.ListTrainingAssignments(context.Background(), domain.ListTrainingAssignmentsParams{
		Limit:  10,
		Offset: 0,
		Status: &status,
	})
	if err != domain.ErrTrainingInvalidRequest {
		t.Fatalf("expected %v, got %v", domain.ErrTrainingInvalidRequest, err)
	}
}

func TestTrainingServiceCancelTrainingAssignmentNormalizesReason(t *testing.T) {
	repo := &fakeTrainingRepository{}
	svc := &TrainingService{repository: repo}

	reason := "  duplicate assignment  "
	assignmentID := uuid.New()

	_, err := svc.CancelTrainingAssignment(context.Background(), domain.CancelTrainingAssignmentParams{
		AssignmentID:       assignmentID,
		CancellationReason: &reason,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if repo.lastCancelTrainingAssignmentParams.AssignmentID != assignmentID {
		t.Fatalf("expected assignment id %v, got %v", assignmentID, repo.lastCancelTrainingAssignmentParams.AssignmentID)
	}
	if repo.lastCancelTrainingAssignmentParams.CancellationReason == nil {
		t.Fatalf("expected cancellation reason to be preserved")
	}
	if got := *repo.lastCancelTrainingAssignmentParams.CancellationReason; got != "duplicate assignment" {
		t.Fatalf("expected trimmed reason, got %q", got)
	}
}

func TestTrainingServiceCancelTrainingAssignmentBlankReasonBecomesNil(t *testing.T) {
	repo := &fakeTrainingRepository{}
	svc := &TrainingService{repository: repo}

	reason := "   "

	_, err := svc.CancelTrainingAssignment(context.Background(), domain.CancelTrainingAssignmentParams{
		AssignmentID:       uuid.New(),
		CancellationReason: &reason,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if repo.lastCancelTrainingAssignmentParams.CancellationReason != nil {
		t.Fatalf("expected blank reason to normalize to nil, got %#v", repo.lastCancelTrainingAssignmentParams.CancellationReason)
	}
}

func TestTrainingServiceCancelTrainingAssignmentRejectsNilID(t *testing.T) {
	svc := &TrainingService{repository: &fakeTrainingRepository{}}

	_, err := svc.CancelTrainingAssignment(context.Background(), domain.CancelTrainingAssignmentParams{})
	if err != domain.ErrTrainingInvalidRequest {
		t.Fatalf("expected %v, got %v", domain.ErrTrainingInvalidRequest, err)
	}
}

type fakeTrainingRepository struct {
	lastListTrainingAssignmentsParams  domain.ListTrainingAssignmentsParams
	lastCancelTrainingAssignmentParams domain.CancelTrainingAssignmentParams
}

func (f *fakeTrainingRepository) AssignTrainingToEmployee(
	_ context.Context,
	_ domain.AssignTrainingToEmployeeParams,
) (*domain.EmployeeTrainingAssignment, error) {
	return &domain.EmployeeTrainingAssignment{
		ID:         uuid.New(),
		EmployeeID: uuid.New(),
		TrainingID: uuid.New(),
		DueAt:      time.Now().UTC(),
	}, nil
}

func (f *fakeTrainingRepository) ListTrainingAssignments(
	_ context.Context,
	params domain.ListTrainingAssignmentsParams,
) (*domain.TrainingAssignmentPage, error) {
	f.lastListTrainingAssignmentsParams = params
	return &domain.TrainingAssignmentPage{
		Items:      []domain.TrainingAssignmentListItem{},
		TotalCount: 0,
	}, nil
}

func (f *fakeTrainingRepository) CancelTrainingAssignment(
	_ context.Context,
	params domain.CancelTrainingAssignmentParams,
) (*domain.EmployeeTrainingAssignment, error) {
	f.lastCancelTrainingAssignmentParams = params
	now := time.Now().UTC()
	return &domain.EmployeeTrainingAssignment{
		ID:                 params.AssignmentID,
		EmployeeID:         uuid.New(),
		TrainingID:         uuid.New(),
		Status:             "cancelled",
		DueAt:              now,
		CancelledAt:        &now,
		CancellationReason: params.CancellationReason,
		UpdatedAt:          now,
	}, nil
}

func (f *fakeTrainingRepository) CreateTrainingCatalogItem(
	_ context.Context,
	_ domain.CreateTrainingCatalogItemParams,
) (*domain.TrainingCatalogItem, error) {
	return nil, nil
}

func (f *fakeTrainingRepository) ListTrainingCatalogItems(
	_ context.Context,
	_ domain.ListTrainingCatalogItemsParams,
) (*domain.TrainingCatalogItemPage, error) {
	return nil, nil
}

var _ domain.TrainingRepository = (*fakeTrainingRepository)(nil)
