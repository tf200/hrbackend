package repository

import (
	"context"
	"strings"
	"time"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"
	"hrbackend/pkg/ptr"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type PayoutRepository struct {
	store *db.Store
}

func NewPayoutRepository(store *db.Store) domain.PayoutRepository {
	return &PayoutRepository{store: store}
}

func (r *PayoutRepository) WithTx(
	ctx context.Context,
	fn func(tx domain.PayoutTxRepository) error,
) error {
	return r.store.ExecTx(ctx, func(q *db.Queries) error {
		return fn(&payoutTxRepo{queries: q})
	})
}

func (r *PayoutRepository) ListMyPayoutRequests(
	ctx context.Context,
	params domain.ListMyPayoutRequestsParams,
) (*domain.PayoutRequestPage, error) {
	rows, err := r.store.ListMyPayoutRequestsPaginated(ctx, db.ListMyPayoutRequestsPaginatedParams{
		EmployeeID: params.EmployeeID,
		Status:     toDBPayoutStatusPtr(params.Status),
		Limit:      params.Limit,
		Offset:     params.Offset,
	})
	if err != nil {
		return nil, err
	}

	page := &domain.PayoutRequestPage{
		Items: make([]domain.PayoutRequest, 0, len(rows)),
	}
	if len(rows) > 0 {
		page.TotalCount = rows[0].TotalCount
	}

	for _, row := range rows {
		page.Items = append(page.Items, toDomainPayoutRequest(
			row.ID,
			row.EmployeeID,
			strings.TrimSpace(row.EmployeeFirstName+" "+row.EmployeeLastName),
			row.CreatedByEmployeeID,
			row.RequestedHours,
			row.BalanceYear,
			row.HourlyRate,
			row.GrossAmount,
			row.PayPeriodStart,
			string(row.Status),
			row.RequestNote,
			row.DecisionNote,
			row.DecidedByEmployeeID,
			row.PaidByEmployeeID,
			row.RequestedAt,
			row.DecidedAt,
			row.PaidAt,
			row.CreatedAt,
			row.UpdatedAt,
		))
	}

	return page, nil
}

func (r *PayoutRepository) ListPayoutRequests(
	ctx context.Context,
	params domain.ListPayoutRequestsParams,
) (*domain.PayoutRequestPage, error) {
	rows, err := r.store.ListPayoutRequestsPaginated(ctx, db.ListPayoutRequestsPaginatedParams{
		Status:         toDBPayoutStatusPtr(params.Status),
		EmployeeSearch: ptr.TrimString(params.EmployeeSearch),
		Limit:          params.Limit,
		Offset:         params.Offset,
	})
	if err != nil {
		return nil, err
	}

	page := &domain.PayoutRequestPage{
		Items: make([]domain.PayoutRequest, 0, len(rows)),
	}
	if len(rows) > 0 {
		page.TotalCount = rows[0].TotalCount
	}

	for _, row := range rows {
		page.Items = append(page.Items, toDomainPayoutRequest(
			row.ID,
			row.EmployeeID,
			strings.TrimSpace(row.EmployeeFirstName+" "+row.EmployeeLastName),
			row.CreatedByEmployeeID,
			row.RequestedHours,
			row.BalanceYear,
			row.HourlyRate,
			row.GrossAmount,
			row.PayPeriodStart,
			string(row.Status),
			row.RequestNote,
			row.DecisionNote,
			row.DecidedByEmployeeID,
			row.PaidByEmployeeID,
			row.RequestedAt,
			row.DecidedAt,
			row.PaidAt,
			row.CreatedAt,
			row.UpdatedAt,
		))
	}

	return page, nil
}

type payoutTxRepo struct {
	queries *db.Queries
}

func (r *payoutTxRepo) GetEmployeePayoutContract(
	ctx context.Context,
	employeeID uuid.UUID,
) (*domain.PayoutContract, error) {
	row, err := r.queries.GetEmployeePayoutContract(ctx, employeeID)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrPayoutRequestNotFound
		}
		return nil, err
	}

	contractRate := row.ContractRate
	return &domain.PayoutContract{
		ContractType: string(row.ContractType),
		ContractRate: &contractRate,
	}, nil
}

func (r *payoutTxRepo) CreatePayoutRequest(
	ctx context.Context,
	params domain.CreatePayoutRequestTxParams,
) (*domain.PayoutRequest, error) {
	row, err := r.queries.CreatePayoutRequest(ctx, db.CreatePayoutRequestParams{
		EmployeeID:          params.EmployeeID,
		CreatedByEmployeeID: params.CreatedByEmployeeID,
		RequestedHours:      params.RequestedHours,
		BalanceYear:         params.BalanceYear,
		HourlyRate:          params.HourlyRate,
		GrossAmount:         params.GrossAmount,
		RequestNote:         params.RequestNote,
	})
	if err != nil {
		return nil, err
	}
	model := toDomainPayoutRequestFromRow(row)
	return &model, nil
}

func (r *payoutTxRepo) GetPayoutRequestForUpdate(
	ctx context.Context,
	payoutRequestID uuid.UUID,
) (*domain.PayoutRequest, error) {
	row, err := r.queries.LockPayoutRequestByID(ctx, payoutRequestID)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrPayoutRequestNotFound
		}
		return nil, err
	}
	model := toDomainPayoutRequestFromRow(row)
	return &model, nil
}

func (r *payoutTxRepo) ApprovePayoutRequest(
	ctx context.Context,
	payoutRequestID, decidedByEmployeeID uuid.UUID,
	payPeriodStart time.Time,
	decisionNote *string,
) (*domain.PayoutRequest, error) {
	row, err := r.queries.ApprovePayoutRequest(ctx, db.ApprovePayoutRequestParams{
		ID:                  payoutRequestID,
		DecisionNote:        decisionNote,
		DecidedByEmployeeID: &decidedByEmployeeID,
		PayPeriodStart:      conv.PgDateFromTime(payPeriodStart),
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrPayoutRequestNotFound
		}
		return nil, err
	}
	model := toDomainPayoutRequestFromRow(row)
	return &model, nil
}

func (r *payoutTxRepo) RejectPayoutRequest(
	ctx context.Context,
	payoutRequestID, decidedByEmployeeID uuid.UUID,
	decisionNote *string,
) (*domain.PayoutRequest, error) {
	row, err := r.queries.RejectPayoutRequest(ctx, db.RejectPayoutRequestParams{
		ID:                  payoutRequestID,
		DecisionNote:        decisionNote,
		DecidedByEmployeeID: &decidedByEmployeeID,
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrPayoutRequestNotFound
		}
		return nil, err
	}
	model := toDomainPayoutRequestFromRow(row)
	return &model, nil
}

func (r *payoutTxRepo) MarkPayoutRequestPaid(
	ctx context.Context,
	payoutRequestID, paidByEmployeeID uuid.UUID,
) (*domain.PayoutRequest, error) {
	row, err := r.queries.MarkPayoutRequestPaid(ctx, db.MarkPayoutRequestPaidParams{
		ID:               payoutRequestID,
		PaidByEmployeeID: &paidByEmployeeID,
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrPayoutRequestNotFound
		}
		return nil, err
	}
	model := toDomainPayoutRequestFromRow(row)
	return &model, nil
}

func (r *payoutTxRepo) UpdatePayoutRequest(
	ctx context.Context,
	payoutRequestID uuid.UUID,
	params domain.UpdatePayoutRequestTxParams,
) (*domain.PayoutRequest, error) {
	row, err := r.queries.UpdatePayoutRequest(ctx, db.UpdatePayoutRequestParams{
		ID:             payoutRequestID,
		RequestedHours: params.RequestedHours,
		BalanceYear:    params.BalanceYear,
		GrossAmount:    params.GrossAmount,
		RequestNote:    params.RequestNote,
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrPayoutRequestNotFound
		}
		return nil, err
	}
	model := toDomainPayoutRequestFromRow(row)
	return &model, nil
}

func toDomainPayoutRequestFromRow(row db.LeavePayoutRequest) domain.PayoutRequest {
	return toDomainPayoutRequest(
		row.ID,
		row.EmployeeID,
		"",
		row.CreatedByEmployeeID,
		row.RequestedHours,
		row.BalanceYear,
		row.HourlyRate,
		row.GrossAmount,
		row.PayPeriodStart,
		string(row.Status),
		row.RequestNote,
		row.DecisionNote,
		row.DecidedByEmployeeID,
		row.PaidByEmployeeID,
		row.RequestedAt,
		row.DecidedAt,
		row.PaidAt,
		row.CreatedAt,
		row.UpdatedAt,
	)
}

func stringFromDBValue(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case []byte:
		return strings.TrimSpace(string(v))
	default:
		return ""
	}
}

func toDomainPayoutRequest(
	id uuid.UUID,
	employeeID uuid.UUID,
	employeeName string,
	createdByEmployeeID uuid.UUID,
	requestedHours int32,
	balanceYear int32,
	hourlyRate float64,
	grossAmount float64,
	payPeriodStart pgtype.Date,
	status string,
	requestNote *string,
	decisionNote *string,
	decidedByEmployeeID *uuid.UUID,
	paidByEmployeeID *uuid.UUID,
	requestedAt pgtype.Timestamptz,
	decidedAt pgtype.Timestamptz,
	paidAt pgtype.Timestamptz,
	createdAt pgtype.Timestamptz,
	updatedAt pgtype.Timestamptz,
) domain.PayoutRequest {
	return domain.PayoutRequest{
		ID:                  id,
		EmployeeID:          employeeID,
		EmployeeName:        employeeName,
		CreatedByEmployeeID: createdByEmployeeID,
		RequestedHours:      requestedHours,
		BalanceYear:         balanceYear,
		HourlyRate:          hourlyRate,
		GrossAmount:         grossAmount,
		PayPeriodStart:      conv.TimePtrFromPgDate(payPeriodStart),
		Status:              status,
		RequestNote:         requestNote,
		DecisionNote:        decisionNote,
		DecidedByEmployeeID: decidedByEmployeeID,
		PaidByEmployeeID:    paidByEmployeeID,
		RequestedAt:         conv.TimeFromPgTimestamptz(requestedAt),
		DecidedAt:           timePtrFromPgTimestamptz(decidedAt),
		PaidAt:              timePtrFromPgTimestamptz(paidAt),
		CreatedAt:           conv.TimeFromPgTimestamptz(createdAt),
		UpdatedAt:           conv.TimeFromPgTimestamptz(updatedAt),
	}
}

func toDBPayoutStatus(value string) (db.PayoutRequestStatusEnum, bool) {
	switch db.PayoutRequestStatusEnum(strings.TrimSpace(value)) {
	case db.PayoutRequestStatusEnumPending,
		db.PayoutRequestStatusEnumApproved,
		db.PayoutRequestStatusEnumRejected,
		db.PayoutRequestStatusEnumPaid:
		return db.PayoutRequestStatusEnum(strings.TrimSpace(value)), true
	default:
		return "", false
	}
}

func toDBPayoutStatusPtr(value *string) *db.PayoutRequestStatusEnum {
	if value == nil {
		return nil
	}
	parsed, ok := toDBPayoutStatus(*value)
	if !ok {
		return nil
	}
	return enumPtr(parsed)
}

func nilStringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

var _ domain.PayoutRepository = (*PayoutRepository)(nil)
