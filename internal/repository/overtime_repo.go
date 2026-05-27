package repository

import (
	"context"
	"errors"
	"strings"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type OvertimeRepository struct {
	store *db.Store
}

func NewOvertimeRepository(store *db.Store) domain.OvertimeRepository {
	return &OvertimeRepository{store: store}
}

func (r *OvertimeRepository) WithTx(
	ctx context.Context,
	fn func(tx domain.OvertimeTxRepository) error,
) error {
	return r.store.ExecTx(ctx, func(q *db.Queries) error {
		return fn(&overtimeTxRepo{queries: q})
	})
}

func (r *OvertimeRepository) CreateOvertimeEntry(
	ctx context.Context,
	params domain.CreateOvertimeEntryParams,
) (*domain.OvertimeEntry, error) {

	//[code_quality]: this is buisness logique needs to be refactored
	if params.ScheduleID != nil {
		schedule, err := r.store.GetScheduleById(ctx, *params.ScheduleID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, domain.ErrOvertimeInvalidRequest
			}
			return nil, err
		}
		if schedule.EmployeeID != params.EmployeeID {
			return nil, domain.ErrOvertimeInvalidRequest
		}
	}

	row, err := r.store.CreateOvertimeEntry(ctx, db.CreateOvertimeEntryParams{
		EmployeeID:  params.EmployeeID,
		ScheduleID:  params.ScheduleID,
		EntryDate:   conv.PgDateFromTime(params.EntryDate),
		Minutes:     params.Minutes,
		Reason:      toDBOvertimeReason(params.Reason),
		Description: params.Description,
	})
	if err != nil {
		return nil, err
	}

	result := toDomainOvertimeEntryFromCreateRow(row)
	return &result, nil
}

func (r *OvertimeRepository) GetOvertimeEntryByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.OvertimeEntry, error) {
	row, err := r.store.GetOvertimeEntryByID(ctx, id)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrOvertimeNotFound
		}
		return nil, err
	}

	result := toDomainOvertimeEntryFromGetRow(row)
	return &result, nil
}

func (r *OvertimeRepository) ListOvertimeEntries(
	ctx context.Context,
	params domain.ListOvertimeEntriesParams,
) (*domain.OvertimeEntryPage, error) {
	rows, err := r.store.ListOvertimeEntriesPaginated(ctx, db.ListOvertimeEntriesPaginatedParams{
		Status:      toDBOvertimeStatusPtr(params.Status),
		LimitCount:  params.Limit,
		OffsetCount: params.Offset,
	})
	if err != nil {
		return nil, err
	}

	page := &domain.OvertimeEntryPage{
		Items: make([]domain.OvertimeEntry, 0, len(rows)),
	}
	if len(rows) > 0 {
		page.TotalCount = rows[0].TotalCount
	}

	for _, row := range rows {
		page.Items = append(page.Items, buildDomainOvertimeEntry(
			row.ID,
			row.EmployeeID,
			row.ScheduleID,
			row.PaidPeriodID,
			row.EntryDate,
			row.Minutes,
			string(row.Reason),
			row.Description,
			string(row.Status),
			row.SubmittedAt,
			row.ApprovedAt,
			row.ApprovedByEmployeeID,
			row.RejectionReason,
			row.CreatedAt,
			row.UpdatedAt,
			row.EmployeeFirstName,
			row.EmployeeLastName,
			row.ApprovedByFirstName,
			row.ApprovedByLastName,
		))
	}

	return page, nil
}

func (r *OvertimeRepository) ListMyOvertimeEntries(
	ctx context.Context,
	params domain.ListMyOvertimeEntriesParams,
) (*domain.OvertimeEntryPage, error) {
	rows, err := r.store.ListMyOvertimeEntriesPaginated(ctx, db.ListMyOvertimeEntriesPaginatedParams{
		EmployeeID:  params.EmployeeID,
		Status:      toDBOvertimeStatusPtr(params.Status),
		LimitCount:  params.Limit,
		OffsetCount: params.Offset,
	})
	if err != nil {
		return nil, err
	}

	page := &domain.OvertimeEntryPage{
		Items: make([]domain.OvertimeEntry, 0, len(rows)),
	}
	if len(rows) > 0 {
		page.TotalCount = rows[0].TotalCount
	}

	for _, row := range rows {
		page.Items = append(page.Items, buildDomainOvertimeEntry(
			row.ID,
			row.EmployeeID,
			row.ScheduleID,
			row.PaidPeriodID,
			row.EntryDate,
			row.Minutes,
			string(row.Reason),
			row.Description,
			string(row.Status),
			row.SubmittedAt,
			row.ApprovedAt,
			row.ApprovedByEmployeeID,
			row.RejectionReason,
			row.CreatedAt,
			row.UpdatedAt,
			row.EmployeeFirstName,
			row.EmployeeLastName,
			row.ApprovedByFirstName,
			row.ApprovedByLastName,
		))
	}

	return page, nil
}

func (r *OvertimeRepository) GetCurrentMonthOvertimeStats(
	ctx context.Context,
) (*domain.OvertimeStats, error) {
	row, err := r.store.GetCurrentMonthOvertimeStats(ctx)
	if err != nil {
		return nil, err
	}

	return &domain.OvertimeStats{
		TotalApprovedMinutes:  row.TotalApprovedMinutes,
		TotalAwaitingApproval: row.TotalAwaitingApproval,
		TotalApproved:         row.TotalApproved,
		TotalSubmitted:        row.TotalSubmitted,
	}, nil
}

func (r *OvertimeRepository) GetMyCurrentMonthOvertimeStats(
	ctx context.Context,
	employeeID uuid.UUID,
) (*domain.OvertimeStats, error) {
	row, err := r.store.GetMyCurrentMonthOvertimeStats(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	return &domain.OvertimeStats{
		TotalApprovedMinutes:  row.TotalApprovedMinutes,
		TotalAwaitingApproval: row.TotalAwaitingApproval,
		TotalApproved:         row.TotalApproved,
		TotalSubmitted:        row.TotalSubmitted,
	}, nil
}

type overtimeTxRepo struct {
	queries *db.Queries
}

func (r *overtimeTxRepo) GetOvertimeEntryForUpdate(
	ctx context.Context,
	overtimeEntryID uuid.UUID,
) (*domain.OvertimeEntry, error) {
	row, err := r.queries.LockOvertimeEntryByID(ctx, overtimeEntryID)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrOvertimeNotFound
		}
		return nil, err
	}

	model := toDomainOvertimeEntryFromLockRow(row)
	return &model, nil
}

func (r *overtimeTxRepo) ApproveOvertimeEntry(
	ctx context.Context,
	overtimeEntryID, approvedByEmployeeID uuid.UUID,
) (*domain.OvertimeEntry, error) {
	row, err := r.queries.ApproveOvertimeEntry(ctx, db.ApproveOvertimeEntryParams{
		ID:                   overtimeEntryID,
		ApprovedByEmployeeID: &approvedByEmployeeID,
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrOvertimeNotFound
		}
		return nil, err
	}

	model := toDomainOvertimeEntryFromApproveRow(row)
	return &model, nil
}

func (r *overtimeTxRepo) RejectOvertimeEntry(
	ctx context.Context,
	overtimeEntryID uuid.UUID,
	rejectionReason *string,
) (*domain.OvertimeEntry, error) {
	row, err := r.queries.RejectOvertimeEntry(ctx, db.RejectOvertimeEntryParams{
		ID:              overtimeEntryID,
		RejectionReason: rejectionReason,
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrOvertimeNotFound
		}
		return nil, err
	}

	model := toDomainOvertimeEntryFromRejectRow(row)
	return &model, nil
}

func (r *overtimeTxRepo) UpdateOvertimeEntryByAdmin(
	ctx context.Context,
	overtimeEntryID uuid.UUID,
	params domain.UpdateOvertimeEntryParams,
) (*domain.OvertimeEntry, error) {
	if params.ScheduleID != nil {
		schedule, err := r.queries.GetScheduleById(ctx, *params.ScheduleID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, domain.ErrOvertimeInvalidRequest
			}
			return nil, err
		}
		if schedule.EmployeeID != params.EmployeeID {
			return nil, domain.ErrOvertimeInvalidRequest
		}
	}

	row, err := r.queries.UpdateOvertimeEntryByAdmin(ctx, db.UpdateOvertimeEntryByAdminParams{
		ScheduleID:  params.ScheduleID,
		EntryDate:   toNullablePgDate(params.EntryDate),
		Minutes:     params.Minutes,
		Reason:      toDBOvertimeReasonPtr(params.Reason),
		Description: params.Description,
		ID:          overtimeEntryID,
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrOvertimeNotFound
		}
		return nil, err
	}

	model := toDomainOvertimeEntryFromUpdateByAdminRow(row)
	return &model, nil
}

func buildDomainOvertimeEntry(
	id uuid.UUID,
	employeeID uuid.UUID,
	scheduleID *uuid.UUID,
	paidPeriodID *uuid.UUID,
	entryDate pgtype.Date,
	minutes int32,
	reason string,
	description *string,
	status string,
	submittedAt pgtype.Timestamptz,
	approvedAt pgtype.Timestamptz,
	approvedByEmployeeID *uuid.UUID,
	rejectionReason *string,
	createdAt pgtype.Timestamptz,
	updatedAt pgtype.Timestamptz,
	employeeFirstName string,
	employeeLastName string,
	approvedByFirstName *string,
	approvedByLastName *string,
) domain.OvertimeEntry {
	return domain.OvertimeEntry{
		ID:                   id,
		EmployeeID:           employeeID,
		EmployeeName:         fullName(employeeFirstName, employeeLastName),
		ScheduleID:           scheduleID,
		PaidPeriodID:         paidPeriodID,
		EntryDate:            conv.TimeFromPgDate(entryDate),
		Minutes:              minutes,
		Reason:               reason,
		Description:          description,
		Status:               status,
		SubmittedAt:          conv.TimeFromPgTimestamptz(submittedAt),
		ApprovedAt:           timePtrFromPgTimestamptz(approvedAt),
		ApprovedByEmployeeID: approvedByEmployeeID,
		ApprovedByName:       nullableFullName(approvedByFirstName, approvedByLastName),
		RejectionReason:      rejectionReason,
		CreatedAt:            conv.TimeFromPgTimestamptz(createdAt),
		UpdatedAt:            conv.TimeFromPgTimestamptz(updatedAt),
	}
}

func toDomainOvertimeEntryFromCreateRow(row db.CreateOvertimeEntryRow) domain.OvertimeEntry {
	return buildDomainOvertimeEntry(
		row.ID,
		row.EmployeeID,
		row.ScheduleID,
		row.PaidPeriodID,
		row.EntryDate,
		row.Minutes,
		string(row.Reason),
		row.Description,
		string(row.Status),
		row.SubmittedAt,
		row.ApprovedAt,
		row.ApprovedByEmployeeID,
		row.RejectionReason,
		row.CreatedAt,
		row.UpdatedAt,
		row.EmployeeFirstName,
		row.EmployeeLastName,
		row.ApprovedByFirstName,
		row.ApprovedByLastName,
	)
}

func toDomainOvertimeEntryFromGetRow(row db.GetOvertimeEntryByIDRow) domain.OvertimeEntry {
	return buildDomainOvertimeEntry(
		row.ID,
		row.EmployeeID,
		row.ScheduleID,
		row.PaidPeriodID,
		row.EntryDate,
		row.Minutes,
		string(row.Reason),
		row.Description,
		string(row.Status),
		row.SubmittedAt,
		row.ApprovedAt,
		row.ApprovedByEmployeeID,
		row.RejectionReason,
		row.CreatedAt,
		row.UpdatedAt,
		row.EmployeeFirstName,
		row.EmployeeLastName,
		row.ApprovedByFirstName,
		row.ApprovedByLastName,
	)
}

func toDomainOvertimeEntryFromLockRow(row db.OvertimeEntry) domain.OvertimeEntry {
	return buildDomainOvertimeEntry(
		row.ID,
		row.EmployeeID,
		row.ScheduleID,
		row.PaidPeriodID,
		row.EntryDate,
		row.Minutes,
		string(row.Reason),
		row.Description,
		string(row.Status),
		row.SubmittedAt,
		row.ApprovedAt,
		row.ApprovedByEmployeeID,
		row.RejectionReason,
		row.CreatedAt,
		row.UpdatedAt,
		"",
		"",
		nil,
		nil,
	)
}

func toDomainOvertimeEntryFromApproveRow(row db.ApproveOvertimeEntryRow) domain.OvertimeEntry {
	return buildDomainOvertimeEntry(
		row.ID,
		row.EmployeeID,
		row.ScheduleID,
		row.PaidPeriodID,
		row.EntryDate,
		row.Minutes,
		string(row.Reason),
		row.Description,
		string(row.Status),
		row.SubmittedAt,
		row.ApprovedAt,
		row.ApprovedByEmployeeID,
		row.RejectionReason,
		row.CreatedAt,
		row.UpdatedAt,
		row.EmployeeFirstName,
		row.EmployeeLastName,
		row.ApprovedByFirstName,
		row.ApprovedByLastName,
	)
}

func toDomainOvertimeEntryFromRejectRow(row db.RejectOvertimeEntryRow) domain.OvertimeEntry {
	return buildDomainOvertimeEntry(
		row.ID,
		row.EmployeeID,
		row.ScheduleID,
		row.PaidPeriodID,
		row.EntryDate,
		row.Minutes,
		string(row.Reason),
		row.Description,
		string(row.Status),
		row.SubmittedAt,
		row.ApprovedAt,
		row.ApprovedByEmployeeID,
		row.RejectionReason,
		row.CreatedAt,
		row.UpdatedAt,
		row.EmployeeFirstName,
		row.EmployeeLastName,
		row.ApprovedByFirstName,
		row.ApprovedByLastName,
	)
}

func toDomainOvertimeEntryFromUpdateByAdminRow(row db.UpdateOvertimeEntryByAdminRow) domain.OvertimeEntry {
	return buildDomainOvertimeEntry(
		row.ID,
		row.EmployeeID,
		row.ScheduleID,
		row.PaidPeriodID,
		row.EntryDate,
		row.Minutes,
		string(row.Reason),
		row.Description,
		string(row.Status),
		row.SubmittedAt,
		row.ApprovedAt,
		row.ApprovedByEmployeeID,
		row.RejectionReason,
		row.CreatedAt,
		row.UpdatedAt,
		row.EmployeeFirstName,
		row.EmployeeLastName,
		row.ApprovedByFirstName,
		row.ApprovedByLastName,
	)
}

func toDBOvertimeReason(value string) db.OvertimeReasonEnum {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case domain.OvertimeReasonClientCrisis:
		return db.OvertimeReasonEnumClientCrisis
	case domain.OvertimeReasonUnderstaffing:
		return db.OvertimeReasonEnumUnderstaffing
	case domain.OvertimeReasonMeetingConsultation:
		return db.OvertimeReasonEnumMeetingConsultation
	case domain.OvertimeReasonTrainingEducation:
		return db.OvertimeReasonEnumTrainingEducation
	case domain.OvertimeReasonCompletingAdministration:
		return db.OvertimeReasonEnumCompletingAdministration
	case domain.OvertimeReasonHandover:
		return db.OvertimeReasonEnumHandover
	case domain.OvertimeReasonEmergency:
		return db.OvertimeReasonEnumEmergency
	case domain.OvertimeReasonProjectWork:
		return db.OvertimeReasonEnumProjectWork
	case domain.OvertimeReasonEventActivity:
		return db.OvertimeReasonEnumEventActivity
	case domain.OvertimeReasonOther:
		return db.OvertimeReasonEnumOther
	default:
		return ""
	}
}

func toDBOvertimeReasonPtr(value *string) *db.OvertimeReasonEnum {
	if value == nil {
		return nil
	}
	return enumPtr(toDBOvertimeReason(*value))
}

func toDBOvertimeStatus(value string) db.OvertimeStatusEnum {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case domain.OvertimeStatusSubmitted:
		return db.OvertimeStatusEnumSubmitted
	case domain.OvertimeStatusApproved:
		return db.OvertimeStatusEnumApproved
	case domain.OvertimeStatusRejected:
		return db.OvertimeStatusEnumRejected
	default:
		return ""
	}
}

func toDBOvertimeStatusPtr(value *string) *db.OvertimeStatusEnum {
	if value == nil {
		return nil
	}
	return enumPtr(toDBOvertimeStatus(*value))
}
