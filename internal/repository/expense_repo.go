package repository

import (
	"context"
	"strings"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"
	"hrbackend/pkg/ptr"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type ExpenseRepository struct {
	store *db.Store
}

func NewExpenseRepository(store *db.Store) domain.ExpenseRepository {
	return &ExpenseRepository{store: store}
}

func (r *ExpenseRepository) WithTx(
	ctx context.Context,
	fn func(tx domain.ExpenseTxRepository) error,
) error {
	return r.store.ExecTx(ctx, func(q *db.Queries) error {
		return fn(&expenseTxRepo{queries: q})
	})
}

func (r *ExpenseRepository) CreateExpenseRequest(
	ctx context.Context,
	params domain.CreateExpenseRequestParams,
) (*domain.ExpenseRequest, error) {
	category, ok := toDBExpenseCategory(params.Category)
	if !ok {
		return nil, domain.ErrExpenseRequestInvalidRequest
	}

	row, err := r.store.CreateExpenseRequest(ctx, db.CreateExpenseRequestParams{
		EmployeeID:          params.EmployeeID,
		CreatedByEmployeeID: params.CreatedByEmployeeID,
		Category:            category,
		ExpenseDate:         conv.PgDateFromTime(params.ExpenseDate),
		MerchantName:        params.MerchantName,
		Description:         params.Description,
		BusinessPurpose:     params.BusinessPurpose,
		Currency:            strings.TrimSpace(strings.ToUpper(params.Currency)),
		ClaimedAmount:       params.ClaimedAmount,
		TravelMode:          params.TravelMode,
		TravelFrom:          params.TravelFrom,
		TravelTo:            params.TravelTo,
		DistanceKm:          params.DistanceKm,
		RequestNote:         params.RequestNote,
	})
	if err != nil {
		return nil, err
	}

	model := toDomainExpenseRequestFromRow(row)
	return &model, nil
}

func (r *ExpenseRepository) GetExpenseRequestByID(
	ctx context.Context,
	expenseRequestID uuid.UUID,
) (*domain.ExpenseRequest, error) {
	row, err := r.store.GetExpenseRequestByID(ctx, expenseRequestID)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrExpenseRequestNotFound
		}
		return nil, err
	}

	model := toDomainExpenseRequest(
		row.ID,
		row.EmployeeID,
		fullName(row.EmployeeFirstName, row.EmployeeLastName),
		row.CreatedByEmployeeID,
		string(row.Category),
		row.ExpenseDate,
		row.MerchantName,
		row.Description,
		row.BusinessPurpose,
		row.Currency,
		row.ClaimedAmount,
		row.ApprovedAmount,
		row.TravelMode,
		row.TravelFrom,
		row.TravelTo,
		row.DistanceKm,
		string(row.Status),
		row.RequestNote,
		row.DecisionNote,
		row.DecidedByEmployeeID,
		row.ReimbursedByEmployeeID,
		row.RequestedAt,
		row.DecidedAt,
		row.ReimbursedAt,
		row.CancelledAt,
		row.CreatedAt,
		row.UpdatedAt,
	)
	return &model, nil
}

func (r *ExpenseRepository) ListMyExpenseRequests(
	ctx context.Context,
	params domain.ListMyExpenseRequestsParams,
) (*domain.ExpenseRequestPage, error) {
	rows, err := r.store.ListMyExpenseRequestsPaginated(ctx, db.ListMyExpenseRequestsPaginatedParams{
		EmployeeID: params.EmployeeID,
		Status:     toDBNullExpenseStatus(params.Status),
		Category:   toDBNullExpenseCategory(params.Category),
		Limit:      params.Limit,
		Offset:     params.Offset,
	})
	if err != nil {
		return nil, err
	}

	page := &domain.ExpenseRequestPage{
		Items: make([]domain.ExpenseRequest, 0, len(rows)),
	}
	if len(rows) > 0 {
		page.TotalCount = rows[0].TotalCount
	}

	for _, row := range rows {
		page.Items = append(page.Items, toDomainExpenseRequest(
			row.ID,
			row.EmployeeID,
			fullName(row.EmployeeFirstName, row.EmployeeLastName),
			row.CreatedByEmployeeID,
			string(row.Category),
			row.ExpenseDate,
			row.MerchantName,
			row.Description,
			row.BusinessPurpose,
			row.Currency,
			row.ClaimedAmount,
			row.ApprovedAmount,
			row.TravelMode,
			row.TravelFrom,
			row.TravelTo,
			row.DistanceKm,
			string(row.Status),
			row.RequestNote,
			row.DecisionNote,
			row.DecidedByEmployeeID,
			row.ReimbursedByEmployeeID,
			row.RequestedAt,
			row.DecidedAt,
			row.ReimbursedAt,
			row.CancelledAt,
			row.CreatedAt,
			row.UpdatedAt,
		))
	}

	return page, nil
}

func (r *ExpenseRepository) ListExpenseRequests(
	ctx context.Context,
	params domain.ListExpenseRequestsParams,
) (*domain.ExpenseRequestPage, error) {
	rows, err := r.store.ListExpenseRequestsPaginated(ctx, db.ListExpenseRequestsPaginatedParams{
		Status:         toDBNullExpenseStatus(params.Status),
		Category:       toDBNullExpenseCategory(params.Category),
		EmployeeSearch: ptr.TrimString(params.EmployeeSearch),
		Limit:          params.Limit,
		Offset:         params.Offset,
	})
	if err != nil {
		return nil, err
	}

	page := &domain.ExpenseRequestPage{
		Items: make([]domain.ExpenseRequest, 0, len(rows)),
	}
	if len(rows) > 0 {
		page.TotalCount = rows[0].TotalCount
	}

	for _, row := range rows {
		page.Items = append(page.Items, toDomainExpenseRequest(
			row.ID,
			row.EmployeeID,
			fullName(row.EmployeeFirstName, row.EmployeeLastName),
			row.CreatedByEmployeeID,
			string(row.Category),
			row.ExpenseDate,
			row.MerchantName,
			row.Description,
			row.BusinessPurpose,
			row.Currency,
			row.ClaimedAmount,
			row.ApprovedAmount,
			row.TravelMode,
			row.TravelFrom,
			row.TravelTo,
			row.DistanceKm,
			string(row.Status),
			row.RequestNote,
			row.DecisionNote,
			row.DecidedByEmployeeID,
			row.ReimbursedByEmployeeID,
			row.RequestedAt,
			row.DecidedAt,
			row.ReimbursedAt,
			row.CancelledAt,
			row.CreatedAt,
			row.UpdatedAt,
		))
	}

	return page, nil
}

type expenseTxRepo struct {
	queries *db.Queries
}

func (r *expenseTxRepo) GetExpenseRequestForUpdate(
	ctx context.Context,
	expenseRequestID uuid.UUID,
) (*domain.ExpenseRequest, error) {
	row, err := r.queries.LockExpenseRequestByID(ctx, expenseRequestID)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrExpenseRequestNotFound
		}
		return nil, err
	}

	model := toDomainExpenseRequestFromRow(row)
	return &model, nil
}

func (r *expenseTxRepo) UpdateExpenseRequestEditableFields(
	ctx context.Context,
	expenseRequestID uuid.UUID,
	params domain.UpdateExpenseRequestParams,
) (*domain.ExpenseRequest, error) {
	row, err := r.queries.UpdateExpenseRequestEditableFields(ctx, db.UpdateExpenseRequestEditableFieldsParams{
		ID: expenseRequestID,
		Category: func() db.NullExpenseRequestCategoryEnum {
			if params.Category == nil {
				return db.NullExpenseRequestCategoryEnum{}
			}
			category, ok := toDBExpenseCategory(*params.Category)
			if !ok {
				return db.NullExpenseRequestCategoryEnum{}
			}
			return db.NullExpenseRequestCategoryEnum{
				ExpenseRequestCategoryEnum: category,
				Valid:                      true,
			}
		}(),
		ExpenseDate: func() pgtype.Date {
			if params.ExpenseDate == nil {
				return pgtype.Date{}
			}
			return conv.PgDateFromTime(*params.ExpenseDate)
		}(),
		MerchantName:    params.MerchantName,
		Description:     params.Description,
		BusinessPurpose: params.BusinessPurpose,
		Currency: func() *string {
			if params.Currency == nil {
				return nil
			}
			value := strings.TrimSpace(strings.ToUpper(*params.Currency))
			return &value
		}(),
		ClaimedAmount: params.ClaimedAmount,
		TravelMode:    params.TravelMode,
		TravelFrom:    params.TravelFrom,
		TravelTo:      params.TravelTo,
		DistanceKm:    params.DistanceKm,
		RequestNote:   params.RequestNote,
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrExpenseRequestNotFound
		}
		return nil, err
	}

	model := toDomainExpenseRequestFromRow(row)
	return &model, nil
}

func (r *expenseTxRepo) ApproveExpenseRequest(
	ctx context.Context,
	expenseRequestID, decidedByEmployeeID uuid.UUID,
	approvedAmount float64,
	decisionNote *string,
) (*domain.ExpenseRequest, error) {
	row, err := r.queries.ApproveExpenseRequest(ctx, db.ApproveExpenseRequestParams{
		ID:                  expenseRequestID,
		ApprovedAmount:      approvedAmount,
		DecisionNote:        decisionNote,
		DecidedByEmployeeID: &decidedByEmployeeID,
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrExpenseRequestNotFound
		}
		return nil, err
	}

	model := toDomainExpenseRequestFromRow(row)
	return &model, nil
}

func (r *expenseTxRepo) RejectExpenseRequest(
	ctx context.Context,
	expenseRequestID, decidedByEmployeeID uuid.UUID,
	decisionNote *string,
) (*domain.ExpenseRequest, error) {
	row, err := r.queries.RejectExpenseRequest(ctx, db.RejectExpenseRequestParams{
		ID:                  expenseRequestID,
		DecisionNote:        decisionNote,
		DecidedByEmployeeID: &decidedByEmployeeID,
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrExpenseRequestNotFound
		}
		return nil, err
	}

	model := toDomainExpenseRequestFromRow(row)
	return &model, nil
}

func (r *expenseTxRepo) MarkExpenseRequestReimbursed(
	ctx context.Context,
	expenseRequestID, reimbursedByEmployeeID uuid.UUID,
) (*domain.ExpenseRequest, error) {
	row, err := r.queries.MarkExpenseRequestReimbursed(ctx, db.MarkExpenseRequestReimbursedParams{
		ID:                     expenseRequestID,
		ReimbursedByEmployeeID: &reimbursedByEmployeeID,
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrExpenseRequestNotFound
		}
		return nil, err
	}

	model := toDomainExpenseRequestFromRow(row)
	return &model, nil
}

func (r *expenseTxRepo) CancelExpenseRequest(
	ctx context.Context,
	expenseRequestID uuid.UUID,
) (*domain.ExpenseRequest, error) {
	row, err := r.queries.CancelExpenseRequest(ctx, expenseRequestID)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrExpenseRequestNotFound
		}
		return nil, err
	}

	model := toDomainExpenseRequestFromRow(row)
	return &model, nil
}

func toDomainExpenseRequest(
	id uuid.UUID,
	employeeID uuid.UUID,
	employeeName string,
	createdByEmployeeID uuid.UUID,
	category string,
	expenseDate pgtype.Date,
	merchantName *string,
	description string,
	businessPurpose string,
	currency string,
	claimedAmount float64,
	approvedAmount *float64,
	travelMode *string,
	travelFrom *string,
	travelTo *string,
	distanceKm *float64,
	status string,
	requestNote *string,
	decisionNote *string,
	decidedByEmployeeID *uuid.UUID,
	reimbursedByEmployeeID *uuid.UUID,
	requestedAt pgtype.Timestamptz,
	decidedAt pgtype.Timestamptz,
	reimbursedAt pgtype.Timestamptz,
	cancelledAt pgtype.Timestamptz,
	createdAt pgtype.Timestamptz,
	updatedAt pgtype.Timestamptz,
) domain.ExpenseRequest {
	return domain.ExpenseRequest{
		ID:                     id,
		EmployeeID:             employeeID,
		EmployeeName:           employeeName,
		CreatedByEmployeeID:    createdByEmployeeID,
		Category:               category,
		ExpenseDate:            conv.TimeFromPgDate(expenseDate),
		MerchantName:           merchantName,
		Description:            description,
		BusinessPurpose:        businessPurpose,
		Currency:               strings.TrimSpace(currency),
		ClaimedAmount:          claimedAmount,
		ApprovedAmount:         approvedAmount,
		TravelMode:             travelMode,
		TravelFrom:             travelFrom,
		TravelTo:               travelTo,
		DistanceKm:             distanceKm,
		Status:                 status,
		RequestNote:            requestNote,
		DecisionNote:           decisionNote,
		DecidedByEmployeeID:    decidedByEmployeeID,
		ReimbursedByEmployeeID: reimbursedByEmployeeID,
		RequestedAt:            conv.TimeFromPgTimestamptz(requestedAt),
		DecidedAt:              timePtrFromPgTimestamptz(decidedAt),
		ReimbursedAt:           timePtrFromPgTimestamptz(reimbursedAt),
		CancelledAt:            timePtrFromPgTimestamptz(cancelledAt),
		CreatedAt:              conv.TimeFromPgTimestamptz(createdAt),
		UpdatedAt:              conv.TimeFromPgTimestamptz(updatedAt),
	}
}

func toDomainExpenseRequestFromRow(row db.ExpenseRequest) domain.ExpenseRequest {
	return toDomainExpenseRequest(
		row.ID,
		row.EmployeeID,
		"",
		row.CreatedByEmployeeID,
		string(row.Category),
		row.ExpenseDate,
		row.MerchantName,
		row.Description,
		row.BusinessPurpose,
		row.Currency,
		row.ClaimedAmount,
		row.ApprovedAmount,
		row.TravelMode,
		row.TravelFrom,
		row.TravelTo,
		row.DistanceKm,
		string(row.Status),
		row.RequestNote,
		row.DecisionNote,
		row.DecidedByEmployeeID,
		row.ReimbursedByEmployeeID,
		row.RequestedAt,
		row.DecidedAt,
		row.ReimbursedAt,
		row.CancelledAt,
		row.CreatedAt,
		row.UpdatedAt,
	)
}

func toDBExpenseCategory(value string) (db.ExpenseRequestCategoryEnum, bool) {
	switch db.ExpenseRequestCategoryEnum(strings.TrimSpace(value)) {
	case db.ExpenseRequestCategoryEnumTravel,
		db.ExpenseRequestCategoryEnumMeal,
		db.ExpenseRequestCategoryEnumAccommodation,
		db.ExpenseRequestCategoryEnumOfficeSupplies,
		db.ExpenseRequestCategoryEnumTraining,
		db.ExpenseRequestCategoryEnumClientEntertainment,
		db.ExpenseRequestCategoryEnumOther:
		return db.ExpenseRequestCategoryEnum(strings.TrimSpace(value)), true
	default:
		return "", false
	}
}

func toDBExpenseStatus(value string) (db.ExpenseRequestStatusEnum, bool) {
	switch db.ExpenseRequestStatusEnum(strings.TrimSpace(value)) {
	case db.ExpenseRequestStatusEnumPending,
		db.ExpenseRequestStatusEnumApproved,
		db.ExpenseRequestStatusEnumRejected,
		db.ExpenseRequestStatusEnumReimbursed,
		db.ExpenseRequestStatusEnumCancelled:
		return db.ExpenseRequestStatusEnum(strings.TrimSpace(value)), true
	default:
		return "", false
	}
}

func toDBNullExpenseCategory(value *string) db.NullExpenseRequestCategoryEnum {
	if value == nil {
		return db.NullExpenseRequestCategoryEnum{}
	}
	parsed, ok := toDBExpenseCategory(*value)
	if !ok {
		return db.NullExpenseRequestCategoryEnum{}
	}
	return db.NullExpenseRequestCategoryEnum{
		ExpenseRequestCategoryEnum: parsed,
		Valid:                      true,
	}
}

func toDBNullExpenseStatus(value *string) db.NullExpenseRequestStatusEnum {
	if value == nil {
		return db.NullExpenseRequestStatusEnum{}
	}
	parsed, ok := toDBExpenseStatus(*value)
	if !ok {
		return db.NullExpenseRequestStatusEnum{}
	}
	return db.NullExpenseRequestStatusEnum{
		ExpenseRequestStatusEnum: parsed,
		Valid:                    true,
	}
}

var _ domain.ExpenseRepository = (*ExpenseRepository)(nil)
