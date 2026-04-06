package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type TimeEntryRepository struct {
	store *db.Store
}

func NewTimeEntryRepository(store *db.Store) domain.TimeEntryRepository {
	return &TimeEntryRepository{store: store}
}

func (r *TimeEntryRepository) WithTx(
	ctx context.Context,
	fn func(tx domain.TimeEntryTxRepository) error,
) error {
	return r.store.ExecTx(ctx, func(q *db.Queries) error {
		return fn(&timeEntryTxRepo{queries: q})
	})
}

func (r *TimeEntryRepository) CreateTimeEntry(
	ctx context.Context,
	params domain.CreateTimeEntryParams,
) (*domain.TimeEntry, error) {
	if params.ScheduleID != nil {
		schedule, err := r.store.GetScheduleById(ctx, *params.ScheduleID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, domain.ErrTimeEntryInvalidRequest
			}
			return nil, err
		}
		if schedule.EmployeeID != params.EmployeeID {
			return nil, domain.ErrTimeEntryInvalidRequest
		}
	}
	startTime, err := conv.PgTimeFromString(params.StartTime)
	if err != nil {
		return nil, domain.ErrTimeEntryInvalidRequest
	}
	endTime, err := conv.PgTimeFromString(params.EndTime)
	if err != nil {
		return nil, domain.ErrTimeEntryInvalidRequest
	}

	row, err := r.store.CreateTimeEntry(ctx, db.CreateTimeEntryParams{
		EmployeeID:          params.EmployeeID,
		ScheduleID:          params.ScheduleID,
		EntryDate:           conv.PgDateFromTime(params.EntryDate),
		StartTime:           startTime,
		EndTime:             endTime,
		BreakMinutes:        params.BreakMinutes,
		HourType:            toDBTimeEntryHourType(params.HourType),
		ProjectName:         params.ProjectName,
		ProjectNumber:       params.ProjectNumber,
		ClientName:          params.ClientName,
		ActivityCategory:    params.ActivityCategory,
		ActivityDescription: params.ActivityDescription,
		Notes:               params.Notes,
	})
	if err != nil {
		return nil, err
	}

	result := toDomainTimeEntryFromCreateRow(row)
	return &result, nil
}

func (r *TimeEntryRepository) GetTimeEntryByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.TimeEntry, error) {
	row, err := r.store.GetTimeEntryByID(ctx, id)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrTimeEntryNotFound
		}
		return nil, err
	}

	result := toDomainTimeEntryFromGetRow(row)
	return &result, nil
}

func (r *TimeEntryRepository) ListTimeEntries(
	ctx context.Context,
	params domain.ListTimeEntriesParams,
) (*domain.TimeEntryPage, error) {
	rows, err := r.store.ListTimeEntriesPaginated(ctx, db.ListTimeEntriesPaginatedParams{
		Status:         toDBNullTimeEntryStatus(params.Status),
		EmployeeSearch: trimStringPtr(params.EmployeeSearch),
		Limit:          params.Limit,
		Offset:         params.Offset,
	})
	if err != nil {
		return nil, err
	}

	page := &domain.TimeEntryPage{
		Items: make([]domain.TimeEntry, 0, len(rows)),
	}
	if len(rows) > 0 {
		page.TotalCount = rows[0].TotalCount
	}

	for _, row := range rows {
		page.Items = append(page.Items, buildDomainTimeEntry(
			row.ID,
			row.EmployeeID,
			row.ScheduleID,
			row.PaidPeriodID,
			row.EntryDate,
			row.StartTime,
			row.EndTime,
			row.BreakMinutes,
			string(row.HourType),
			row.ProjectName,
			row.ProjectNumber,
			row.ClientName,
			row.ActivityCategory,
			row.ActivityDescription,
			string(row.Status),
			row.SubmittedAt,
			row.ApprovedAt,
			row.ApprovedByEmployeeID,
			row.RejectionReason,
			row.Notes,
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

func (r *TimeEntryRepository) ListMyTimeEntries(
	ctx context.Context,
	params domain.ListMyTimeEntriesParams,
) (*domain.TimeEntryPage, error) {
	rows, err := r.store.ListMyTimeEntriesPaginated(ctx, db.ListMyTimeEntriesPaginatedParams{
		EmployeeID: params.EmployeeID,
		Status:     toDBNullTimeEntryStatus(params.Status),
		Limit:      params.Limit,
		Offset:     params.Offset,
	})
	if err != nil {
		return nil, err
	}

	page := &domain.TimeEntryPage{
		Items: make([]domain.TimeEntry, 0, len(rows)),
	}
	if len(rows) > 0 {
		page.TotalCount = rows[0].TotalCount
	}

	for _, row := range rows {
		page.Items = append(page.Items, buildDomainTimeEntry(
			row.ID,
			row.EmployeeID,
			row.ScheduleID,
			row.PaidPeriodID,
			row.EntryDate,
			row.StartTime,
			row.EndTime,
			row.BreakMinutes,
			string(row.HourType),
			row.ProjectName,
			row.ProjectNumber,
			row.ClientName,
			row.ActivityCategory,
			row.ActivityDescription,
			string(row.Status),
			row.SubmittedAt,
			row.ApprovedAt,
			row.ApprovedByEmployeeID,
			row.RejectionReason,
			row.Notes,
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

func (r *TimeEntryRepository) GetCurrentMonthTimeEntryStats(
	ctx context.Context,
) (*domain.TimeEntryStats, error) {
	row, err := r.store.GetCurrentMonthTimeEntryStats(ctx)
	if err != nil {
		return nil, err
	}

	return &domain.TimeEntryStats{
		TotalHours:            float64(row.TotalWorkedMinutes) / 60.0,
		TotalAwaitingApproval: row.TotalAwaitingApproval,
		TotalApproved:         row.TotalApproved,
		TotalConcepts:         row.TotalConcepts,
	}, nil
}

func toDomainTimeEntryFromCreateRow(row db.CreateTimeEntryRow) domain.TimeEntry {
	return buildDomainTimeEntry(
		row.ID,
		row.EmployeeID,
		row.ScheduleID,
		row.PaidPeriodID,
		row.EntryDate,
		row.StartTime,
		row.EndTime,
		row.BreakMinutes,
		string(row.HourType),
		row.ProjectName,
		row.ProjectNumber,
		row.ClientName,
		row.ActivityCategory,
		row.ActivityDescription,
		string(row.Status),
		row.SubmittedAt,
		row.ApprovedAt,
		row.ApprovedByEmployeeID,
		row.RejectionReason,
		row.Notes,
		row.CreatedAt,
		row.UpdatedAt,
		row.EmployeeFirstName,
		row.EmployeeLastName,
		row.ApprovedByFirstName,
		row.ApprovedByLastName,
	)
}

func toDomainTimeEntryFromGetRow(row db.GetTimeEntryByIDRow) domain.TimeEntry {
	return buildDomainTimeEntry(
		row.ID,
		row.EmployeeID,
		row.ScheduleID,
		row.PaidPeriodID,
		row.EntryDate,
		row.StartTime,
		row.EndTime,
		row.BreakMinutes,
		string(row.HourType),
		row.ProjectName,
		row.ProjectNumber,
		row.ClientName,
		row.ActivityCategory,
		row.ActivityDescription,
		string(row.Status),
		row.SubmittedAt,
		row.ApprovedAt,
		row.ApprovedByEmployeeID,
		row.RejectionReason,
		row.Notes,
		row.CreatedAt,
		row.UpdatedAt,
		row.EmployeeFirstName,
		row.EmployeeLastName,
		row.ApprovedByFirstName,
		row.ApprovedByLastName,
	)
}

func buildDomainTimeEntry(
	id uuid.UUID,
	employeeID uuid.UUID,
	scheduleID *uuid.UUID,
	paidPeriodID *uuid.UUID,
	entryDate pgtype.Date,
	startTime pgtype.Time,
	endTime pgtype.Time,
	breakMinutes int32,
	hourType string,
	projectName *string,
	projectNumber *string,
	clientName *string,
	activityCategory *string,
	activityDescription *string,
	status string,
	submittedAt pgtype.Timestamptz,
	approvedAt pgtype.Timestamptz,
	approvedByEmployeeID *uuid.UUID,
	rejectionReason *string,
	notes *string,
	createdAt pgtype.Timestamptz,
	updatedAt pgtype.Timestamptz,
	employeeFirstName string,
	employeeLastName string,
	approvedByFirstName *string,
	approvedByLastName *string,
) domain.TimeEntry {
	return domain.TimeEntry{
		ID:                   id,
		EmployeeID:           employeeID,
		EmployeeName:         fullName(employeeFirstName, employeeLastName),
		ScheduleID:           scheduleID,
		PaidPeriodID:         paidPeriodID,
		EntryDate:            conv.TimeFromPgDate(entryDate),
		StartTime:            conv.StringFromPgTime(startTime),
		EndTime:              conv.StringFromPgTime(endTime),
		BreakMinutes:         breakMinutes,
		HourType:             hourType,
		ProjectName:          projectName,
		ProjectNumber:        projectNumber,
		ClientName:           clientName,
		ActivityCategory:     activityCategory,
		ActivityDescription:  activityDescription,
		Status:               status,
		SubmittedAt:          timePtrFromPgTimestamptz(submittedAt),
		ApprovedAt:           timePtrFromPgTimestamptz(approvedAt),
		ApprovedByEmployeeID: approvedByEmployeeID,
		ApprovedByName:       nullableFullName(approvedByFirstName, approvedByLastName),
		RejectionReason:      rejectionReason,
		Notes:                notes,
		CreatedAt:            conv.TimeFromPgTimestamptz(createdAt),
		UpdatedAt:            conv.TimeFromPgTimestamptz(updatedAt),
	}
}

func fullName(firstName, lastName string) string {
	return strings.TrimSpace(firstName + " " + lastName)
}

func nullableFullName(firstName, lastName *string) *string {
	if firstName == nil && lastName == nil {
		return nil
	}

	name := fullName(valueOrEmpty(firstName), valueOrEmpty(lastName))
	if name == "" {
		return nil
	}
	return &name
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func toDBTimeEntryHourType(value string) db.TimeEntryHourTypeEnum {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case domain.TimeEntryHourTypeNormal:
		return db.TimeEntryHourTypeEnumNormal
	case domain.TimeEntryHourTypeOvertime:
		return db.TimeEntryHourTypeEnumOvertime
	case domain.TimeEntryHourTypeTravel:
		return db.TimeEntryHourTypeEnumTravel
	case domain.TimeEntryHourTypeLeave:
		return db.TimeEntryHourTypeEnumLeave
	case domain.TimeEntryHourTypeSick:
		return db.TimeEntryHourTypeEnumSick
	case domain.TimeEntryHourTypeTraining:
		return db.TimeEntryHourTypeEnumTraining
	default:
		return ""
	}
}

func toDBTimeEntryStatus(value string) db.TimeEntryStatusEnum {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case domain.TimeEntryStatusDraft:
		return db.TimeEntryStatusEnumDraft
	case domain.TimeEntryStatusSubmitted:
		return db.TimeEntryStatusEnumSubmitted
	case domain.TimeEntryStatusApproved:
		return db.TimeEntryStatusEnumApproved
	case domain.TimeEntryStatusRejected:
		return db.TimeEntryStatusEnumRejected
	default:
		return ""
	}
}

func toDBNullTimeEntryStatus(value *string) db.NullTimeEntryStatusEnum {
	if value == nil {
		return db.NullTimeEntryStatusEnum{}
	}

	return db.NullTimeEntryStatusEnum{
		TimeEntryStatusEnum: toDBTimeEntryStatus(*value),
		Valid:               true,
	}
}

type timeEntryTxRepo struct {
	queries *db.Queries
}

func (r *timeEntryTxRepo) GetTimeEntryForUpdate(
	ctx context.Context,
	timeEntryID uuid.UUID,
) (*domain.TimeEntry, error) {
	row, err := r.queries.LockTimeEntryByID(ctx, timeEntryID)
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrTimeEntryNotFound
		}
		return nil, err
	}

	model := toDomainTimeEntryFromDBTimeEntry(row)
	return &model, nil
}

func (r *timeEntryTxRepo) ApproveTimeEntry(
	ctx context.Context,
	timeEntryID, approvedByEmployeeID uuid.UUID,
) (*domain.TimeEntry, error) {
	row, err := r.queries.ApproveTimeEntry(ctx, db.ApproveTimeEntryParams{
		ID:                   timeEntryID,
		ApprovedByEmployeeID: &approvedByEmployeeID,
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrTimeEntryNotFound
		}
		return nil, err
	}

	model := toDomainTimeEntryFromApproveRow(row)
	return &model, nil
}

func (r *timeEntryTxRepo) RejectTimeEntry(
	ctx context.Context,
	timeEntryID uuid.UUID,
	rejectionReason *string,
) (*domain.TimeEntry, error) {
	row, err := r.queries.RejectTimeEntry(ctx, db.RejectTimeEntryParams{
		ID:              timeEntryID,
		RejectionReason: rejectionReason,
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrTimeEntryNotFound
		}
		return nil, err
	}

	model := toDomainTimeEntryFromRejectRow(row)
	return &model, nil
}

func (r *timeEntryTxRepo) UpdateTimeEntryByAdmin(
	ctx context.Context,
	timeEntryID uuid.UUID,
	params domain.UpdateTimeEntryByAdminParams,
) (*domain.TimeEntry, error) {
	if params.ScheduleID != nil {
		schedule, err := r.queries.GetScheduleById(ctx, *params.ScheduleID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, domain.ErrTimeEntryInvalidRequest
			}
			return nil, err
		}
		if schedule.EmployeeID != params.EmployeeID {
			return nil, domain.ErrTimeEntryInvalidRequest
		}
	}

	startTime, err := toNullablePgTime(params.StartTime)
	if err != nil {
		return nil, domain.ErrTimeEntryInvalidRequest
	}
	endTime, err := toNullablePgTime(params.EndTime)
	if err != nil {
		return nil, domain.ErrTimeEntryInvalidRequest
	}

	row, err := r.queries.UpdateTimeEntryByAdmin(ctx, db.UpdateTimeEntryByAdminParams{
		ScheduleID:          params.ScheduleID,
		EntryDate:           toNullablePgDate(params.EntryDate),
		StartTime:           startTime,
		EndTime:             endTime,
		BreakMinutes:        params.BreakMinutes,
		HourType:            toDBNullTimeEntryHourType(params.HourType),
		ProjectName:         params.ProjectName,
		ProjectNumber:       params.ProjectNumber,
		ClientName:          params.ClientName,
		ActivityCategory:    params.ActivityCategory,
		ActivityDescription: params.ActivityDescription,
		Notes:               params.Notes,
		SetSubmitted:        shouldSetSubmitted(params.Status),
		ID:                  timeEntryID,
	})
	if err != nil {
		if isDBNotFound(err) {
			return nil, domain.ErrTimeEntryNotFound
		}
		return nil, err
	}

	model := toDomainTimeEntryFromUpdateRow(row)
	return &model, nil
}

func (r *timeEntryTxRepo) CreateTimeEntryUpdateAudit(
	ctx context.Context,
	params domain.CreateTimeEntryUpdateAuditParams,
) error {
	return r.queries.CreateTimeEntryUpdateAudit(ctx, db.CreateTimeEntryUpdateAuditParams{
		TimeEntryID:     params.TimeEntryID,
		AdminEmployeeID: params.AdminEmployeeID,
		AdminUpdateNote: params.AdminUpdateNote,
		BeforeSnapshot:  params.BeforeSnapshot,
		AfterSnapshot:   params.AfterSnapshot,
	})
}

func toDomainTimeEntryFromDBTimeEntry(row db.TimeEntry) domain.TimeEntry {
	return buildDomainTimeEntry(
		row.ID,
		row.EmployeeID,
		row.ScheduleID,
		row.PaidPeriodID,
		row.EntryDate,
		row.StartTime,
		row.EndTime,
		row.BreakMinutes,
		string(row.HourType),
		row.ProjectName,
		row.ProjectNumber,
		row.ClientName,
		row.ActivityCategory,
		row.ActivityDescription,
		string(row.Status),
		row.SubmittedAt,
		row.ApprovedAt,
		row.ApprovedByEmployeeID,
		row.RejectionReason,
		row.Notes,
		row.CreatedAt,
		row.UpdatedAt,
		"",
		"",
		nil,
		nil,
	)
}

func toDomainTimeEntryFromApproveRow(row db.ApproveTimeEntryRow) domain.TimeEntry {
	return buildDomainTimeEntry(
		row.ID,
		row.EmployeeID,
		row.ScheduleID,
		row.PaidPeriodID,
		row.EntryDate,
		row.StartTime,
		row.EndTime,
		row.BreakMinutes,
		string(row.HourType),
		row.ProjectName,
		row.ProjectNumber,
		row.ClientName,
		row.ActivityCategory,
		row.ActivityDescription,
		string(row.Status),
		row.SubmittedAt,
		row.ApprovedAt,
		row.ApprovedByEmployeeID,
		row.RejectionReason,
		row.Notes,
		row.CreatedAt,
		row.UpdatedAt,
		row.EmployeeFirstName,
		row.EmployeeLastName,
		row.ApprovedByFirstName,
		row.ApprovedByLastName,
	)
}

func toDomainTimeEntryFromRejectRow(row db.RejectTimeEntryRow) domain.TimeEntry {
	return buildDomainTimeEntry(
		row.ID,
		row.EmployeeID,
		row.ScheduleID,
		row.PaidPeriodID,
		row.EntryDate,
		row.StartTime,
		row.EndTime,
		row.BreakMinutes,
		string(row.HourType),
		row.ProjectName,
		row.ProjectNumber,
		row.ClientName,
		row.ActivityCategory,
		row.ActivityDescription,
		string(row.Status),
		row.SubmittedAt,
		row.ApprovedAt,
		row.ApprovedByEmployeeID,
		row.RejectionReason,
		row.Notes,
		row.CreatedAt,
		row.UpdatedAt,
		row.EmployeeFirstName,
		row.EmployeeLastName,
		row.ApprovedByFirstName,
		row.ApprovedByLastName,
	)
}

func toDomainTimeEntryFromUpdateRow(row db.UpdateTimeEntryByAdminRow) domain.TimeEntry {
	return buildDomainTimeEntry(
		row.ID,
		row.EmployeeID,
		row.ScheduleID,
		row.PaidPeriodID,
		row.EntryDate,
		row.StartTime,
		row.EndTime,
		row.BreakMinutes,
		string(row.HourType),
		row.ProjectName,
		row.ProjectNumber,
		row.ClientName,
		row.ActivityCategory,
		row.ActivityDescription,
		string(row.Status),
		row.SubmittedAt,
		row.ApprovedAt,
		row.ApprovedByEmployeeID,
		row.RejectionReason,
		row.Notes,
		row.CreatedAt,
		row.UpdatedAt,
		row.EmployeeFirstName,
		row.EmployeeLastName,
		row.ApprovedByFirstName,
		row.ApprovedByLastName,
	)
}

func toNullablePgDate(value *time.Time) pgtype.Date {
	if value == nil {
		return pgtype.Date{}
	}
	return conv.PgDateFromTime(*value)
}

func toNullablePgTime(value *string) (pgtype.Time, error) {
	if value == nil {
		return pgtype.Time{}, nil
	}
	parsed, err := conv.PgTimeFromString(*value)
	if err != nil {
		return pgtype.Time{}, err
	}
	return parsed, nil
}

func toDBNullTimeEntryHourType(value *string) db.NullTimeEntryHourTypeEnum {
	if value == nil {
		return db.NullTimeEntryHourTypeEnum{}
	}

	return db.NullTimeEntryHourTypeEnum{
		TimeEntryHourTypeEnum: toDBTimeEntryHourType(*value),
		Valid:                 true,
	}
}

func shouldSetSubmitted(status *string) bool {
	return status != nil && strings.TrimSpace(strings.ToLower(*status)) == domain.TimeEntryStatusSubmitted
}
