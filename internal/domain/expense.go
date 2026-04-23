package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrExpenseRequestInvalidRequest = errors.New("invalid expense request")
	ErrExpenseRequestNotFound       = errors.New("expense request not found")
	ErrExpenseRequestForbidden      = errors.New("expense request is not accessible by the actor")
	ErrExpenseRequestStateInvalid   = errors.New("expense request is not in a valid state for this operation")
)

const (
	ExpenseRequestStatusPending    = "pending"
	ExpenseRequestStatusApproved   = "approved"
	ExpenseRequestStatusRejected   = "rejected"
	ExpenseRequestStatusReimbursed = "reimbursed"
	ExpenseRequestStatusCancelled  = "cancelled"
)

type ExpenseRequest struct {
	ID                     uuid.UUID
	EmployeeID             uuid.UUID
	EmployeeName           string
	CreatedByEmployeeID    uuid.UUID
	Category               string
	ExpenseDate            time.Time
	MerchantName           *string
	Description            string
	BusinessPurpose        string
	Currency               string
	ClaimedAmount          float64
	ApprovedAmount         *float64
	TravelMode             *string
	TravelFrom             *string
	TravelTo               *string
	DistanceKm             *float64
	Status                 string
	RequestNote            *string
	DecisionNote           *string
	DecidedByEmployeeID    *uuid.UUID
	ReimbursedByEmployeeID *uuid.UUID
	RequestedAt            time.Time
	DecidedAt              *time.Time
	ReimbursedAt           *time.Time
	CancelledAt            *time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type ExpenseRequestPage struct {
	Items      []ExpenseRequest
	TotalCount int64
}

type CreateExpenseRequestParams struct {
	EmployeeID          uuid.UUID
	CreatedByEmployeeID uuid.UUID
	Category            string
	ExpenseDate         time.Time
	MerchantName        *string
	Description         string
	BusinessPurpose     string
	Currency            string
	ClaimedAmount       float64
	TravelMode          *string
	TravelFrom          *string
	TravelTo            *string
	DistanceKm          *float64
	RequestNote         *string
}

type UpdateExpenseRequestParams struct {
	Category        *string
	ExpenseDate     *time.Time
	MerchantName    *string
	Description     *string
	BusinessPurpose *string
	Currency        *string
	ClaimedAmount   *float64
	TravelMode      *string
	TravelFrom      *string
	TravelTo        *string
	DistanceKm      *float64
	RequestNote     *string
}

type DecideExpenseRequestParams struct {
	Decision       string
	ApprovedAmount *float64
	DecisionNote   *string
}

type ListMyExpenseRequestsParams struct {
	EmployeeID uuid.UUID
	Limit      int32
	Offset     int32
	Status     *string
	Category   *string
}

type ListExpenseRequestsParams struct {
	Limit          int32
	Offset         int32
	Status         *string
	Category       *string
	EmployeeSearch *string
}

type ExpenseTxRepository interface {
	GetExpenseRequestForUpdate(ctx context.Context, expenseRequestID uuid.UUID) (*ExpenseRequest, error)
	UpdateExpenseRequestEditableFields(
		ctx context.Context,
		expenseRequestID uuid.UUID,
		params UpdateExpenseRequestParams,
	) (*ExpenseRequest, error)
	ApproveExpenseRequest(
		ctx context.Context,
		expenseRequestID, decidedByEmployeeID uuid.UUID,
		approvedAmount float64,
		decisionNote *string,
	) (*ExpenseRequest, error)
	RejectExpenseRequest(
		ctx context.Context,
		expenseRequestID, decidedByEmployeeID uuid.UUID,
		decisionNote *string,
	) (*ExpenseRequest, error)
	MarkExpenseRequestReimbursed(
		ctx context.Context,
		expenseRequestID, reimbursedByEmployeeID uuid.UUID,
	) (*ExpenseRequest, error)
	CancelExpenseRequest(ctx context.Context, expenseRequestID uuid.UUID) (*ExpenseRequest, error)
}

type ExpenseRepository interface {
	WithTx(ctx context.Context, fn func(tx ExpenseTxRepository) error) error
	CreateExpenseRequest(ctx context.Context, params CreateExpenseRequestParams) (*ExpenseRequest, error)
	GetExpenseRequestByID(ctx context.Context, expenseRequestID uuid.UUID) (*ExpenseRequest, error)
	ListMyExpenseRequests(
		ctx context.Context,
		params ListMyExpenseRequestsParams,
	) (*ExpenseRequestPage, error)
	ListExpenseRequests(
		ctx context.Context,
		params ListExpenseRequestsParams,
	) (*ExpenseRequestPage, error)
}

type ExpenseService interface {
	CreateExpenseRequestByAdmin(
		ctx context.Context,
		adminEmployeeID uuid.UUID,
		params CreateExpenseRequestParams,
	) (*ExpenseRequest, error)
	GetExpenseRequestByID(ctx context.Context, expenseRequestID uuid.UUID) (*ExpenseRequest, error)
	ListExpenseRequests(
		ctx context.Context,
		params ListExpenseRequestsParams,
	) (*ExpenseRequestPage, error)
	UpdateExpenseRequestByAdmin(
		ctx context.Context,
		adminEmployeeID, expenseRequestID uuid.UUID,
		params UpdateExpenseRequestParams,
	) (*ExpenseRequest, error)
	DecideExpenseRequestByAdmin(
		ctx context.Context,
		adminEmployeeID, expenseRequestID uuid.UUID,
		params DecideExpenseRequestParams,
	) (*ExpenseRequest, error)
	CancelExpenseRequestByAdmin(
		ctx context.Context,
		adminEmployeeID, expenseRequestID uuid.UUID,
	) (*ExpenseRequest, error)
	MarkExpenseRequestReimbursedByAdmin(
		ctx context.Context,
		adminEmployeeID, expenseRequestID uuid.UUID,
	) (*ExpenseRequest, error)
}
