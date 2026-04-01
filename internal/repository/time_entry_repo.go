package repository

import (
	"context"
	"strings"

	"hrbackend/internal/domain"
	db "hrbackend/internal/repository/db"
	"hrbackend/pkg/conv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type TimeEntryRepository struct {
	store *db.Store
}

func NewTimeEntryRepository(store *db.Store) domain.TimeEntryRepository {
	return &TimeEntryRepository{store: store}
}

func (r *TimeEntryRepository) CreateTimeEntry(ctx context.Context, params domain.CreateTimeEntryParams) (*domain.TimeEntry, error) {
	row, err := r.store.CreateTimeEntry(ctx, db.CreateTimeEntryParams{
		EmployeeID:          params.EmployeeID,
		ScheduleID:          params.ScheduleID,
		EntryDate:           conv.PgDateFromTime(params.EntryDate),
		Hours:               params.Hours,
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

func (r *TimeEntryRepository) GetTimeEntryByID(ctx context.Context, id uuid.UUID) (*domain.TimeEntry, error) {
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

func (r *TimeEntryRepository) ListTimeEntries(ctx context.Context, params domain.ListTimeEntriesParams) (*domain.TimeEntryPage, error) {
	rows, err := r.store.ListTimeEntriesPaginated(ctx, db.ListTimeEntriesPaginatedParams{
		EmployeeID:     params.EmployeeID,
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
			row.EntryDate,
			row.Hours,
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

func (r *TimeEntryRepository) ListMyTimeEntries(ctx context.Context, params domain.ListMyTimeEntriesParams) (*domain.TimeEntryPage, error) {
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
			row.EntryDate,
			row.Hours,
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

func toDomainTimeEntryFromCreateRow(row db.CreateTimeEntryRow) domain.TimeEntry {
	return buildDomainTimeEntry(
		row.ID,
		row.EmployeeID,
		row.ScheduleID,
		row.EntryDate,
		row.Hours,
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
		row.EntryDate,
		row.Hours,
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
	entryDate pgtype.Date,
	hours float64,
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
		EntryDate:            conv.TimeFromPgDate(entryDate),
		Hours:                hours,
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
