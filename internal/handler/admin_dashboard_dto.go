package handler

import (
	"time"

	"hrbackend/internal/domain"

	"github.com/google/uuid"
)

const adminDashboardDateLayout = "2006-01-02"

type listOpenShiftCoverageRequest struct {
	Days *int32 `form:"days" binding:"omitempty,min=1,max=31"`
}

type adminAvailabilityStatsResponse struct {
	Active       int64 `json:"active"`
	OnLeave      int64 `json:"on_leave"`
	Sick         int64 `json:"sick"`
	OutOfService int64 `json:"out_of_service"`
	Unscheduled  int64 `json:"unscheduled"`
}

func toAdminAvailabilityStatsResponse(
	stats *domain.AvailabilityStats,
) adminAvailabilityStatsResponse {
	return adminAvailabilityStatsResponse{
		Active:       stats.Active,
		OnLeave:      stats.OnLeave,
		Sick:         stats.Sick,
		OutOfService: stats.OutOfService,
		Unscheduled:  stats.Unscheduled,
	}
}

type adminOpenShiftCoverageResponse struct {
	Days      int32                                    `json:"days"`
	Locations []adminOpenShiftCoverageLocationResponse `json:"locations"`
}

type adminOpenShiftCoverageLocationResponse struct {
	LocationID          uuid.UUID                             `json:"location_id"`
	LocationName        string                                `json:"location_name"`
	Street              string                                `json:"street"`
	HouseNumber         string                                `json:"house_number"`
	HouseNumberAddition *string                               `json:"house_number_addition,omitempty"`
	PostalCode          string                                `json:"postal_code"`
	City                string                                `json:"city"`
	OpenSlots           int64                                 `json:"open_slots"`
	OpenDays            int                                   `json:"open_days"`
	Shifts              []adminOpenShiftCoverageShiftResponse `json:"shifts"`
	openDates           map[string]struct{}
}

type adminOpenShiftCoverageShiftResponse struct {
	Date      string    `json:"date"`
	ShiftID   uuid.UUID `json:"shift_id"`
	ShiftName string    `json:"shift_name"`
	OpenSlots int64     `json:"open_slots"`
}

func toAdminOpenShiftCoverageResponse(
	days int32,
	items []domain.OpenShiftCoverageItem,
) adminOpenShiftCoverageResponse {
	locations := make([]adminOpenShiftCoverageLocationResponse, 0)
	locationIndexes := make(map[uuid.UUID]int)

	for _, item := range items {
		idx, ok := locationIndexes[item.LocationID]
		if !ok {
			idx = len(locations)
			locationIndexes[item.LocationID] = idx
			locations = append(locations, adminOpenShiftCoverageLocationResponse{
				LocationID:          item.LocationID,
				LocationName:        item.LocationName,
				Street:              item.Street,
				HouseNumber:         item.HouseNumber,
				HouseNumberAddition: item.HouseNumberAddition,
				PostalCode:          item.PostalCode,
				City:                item.City,
				Shifts:              []adminOpenShiftCoverageShiftResponse{},
				openDates:           make(map[string]struct{}),
			})
		}

		date := formatAdminDashboardDate(item.ShiftDate)
		locations[idx].OpenSlots += item.OpenSlots
		locations[idx].openDates[date] = struct{}{}
		locations[idx].Shifts = append(locations[idx].Shifts, adminOpenShiftCoverageShiftResponse{
			Date:      date,
			ShiftID:   item.ShiftID,
			ShiftName: item.ShiftName,
			OpenSlots: item.OpenSlots,
		})
	}

	for i := range locations {
		locations[i].OpenDays = len(locations[i].openDates)
		locations[i].openDates = nil
	}

	return adminOpenShiftCoverageResponse{
		Days:      days,
		Locations: locations,
	}
}

type adminCriticalActionStatsResponse struct {
	PendingLeaveRequests int64 `json:"pending_leave_requests"`
	PendingShiftSwaps    int64 `json:"pending_shift_swaps"`
	PendingExpenseClaims int64 `json:"pending_expense_claims"`
	PendingTimeEntries   int64 `json:"pending_time_entries"`
	Total                int64 `json:"total"`
}

func toAdminCriticalActionStatsResponse(
	stats *domain.CriticalActionStats,
) adminCriticalActionStatsResponse {
	return adminCriticalActionStatsResponse{
		PendingLeaveRequests: stats.PendingLeaveRequests,
		PendingShiftSwaps:    stats.PendingShiftSwaps,
		PendingExpenseClaims: stats.PendingExpenseClaims,
		PendingTimeEntries:   stats.PendingTimeEntries,
		Total: stats.PendingLeaveRequests +
			stats.PendingShiftSwaps +
			stats.PendingExpenseClaims +
			stats.PendingTimeEntries,
	}
}

type adminPayrollTotalStatsResponse struct {
	Month       string  `json:"month"`
	SalaryTotal float64 `json:"salary_total"`
	ZZPTotal    float64 `json:"zzp_total"`
	ORTTotal    float64 `json:"ort_total"`
	GrossTotal  float64 `json:"gross_total"`
}

func toAdminPayrollTotalStatsResponse(
	stats *domain.PayrollTotalStats,
) adminPayrollTotalStatsResponse {
	return adminPayrollTotalStatsResponse{
		Month:       stats.MonthStart.Format("2006-01"),
		SalaryTotal: stats.SalaryTotal,
		ZZPTotal:    stats.ZZPTotal,
		ORTTotal:    stats.ORTTotal,
		GrossTotal:  stats.SalaryTotal + stats.ZZPTotal + stats.ORTTotal,
	}
}

type adminRiskRadarStatsResponse struct {
	Month                    string `json:"month"`
	ContractsEndingThisMonth int64  `json:"contracts_ending_this_month"`
	OverdueTraining          int64  `json:"overdue_training"`
	LateArrivalsThisMonth    int64  `json:"late_arrivals_this_month"`
	Total                    int64  `json:"total"`
}

func toAdminRiskRadarStatsResponse(
	stats *domain.RiskRadarStats,
) adminRiskRadarStatsResponse {
	return adminRiskRadarStatsResponse{
		Month:                    stats.MonthStart.Format("2006-01"),
		ContractsEndingThisMonth: stats.ContractsEndingThisMonth,
		OverdueTraining:          stats.OverdueTraining,
		LateArrivalsThisMonth:    stats.LateArrivalsThisMonth,
		Total: stats.ContractsEndingThisMonth +
			stats.OverdueTraining +
			stats.LateArrivalsThisMonth,
	}
}

type adminTeamHealthByDepartmentResponse struct {
	Departments []adminTeamHealthDepartmentResponse `json:"departments"`
}

type adminTeamHealthDepartmentResponse struct {
	DepartmentID    uuid.UUID `json:"department_id"`
	DepartmentName  string    `json:"department_name"`
	StaffingPercent float64   `json:"staffing_percent"`
	AbsencePercent  float64   `json:"absence_percent"`
	TrainingPercent float64   `json:"training_percent"`
	Score           float64   `json:"score"`
	Risk            string    `json:"risk"`
}

func toAdminTeamHealthByDepartmentResponse(
	items []domain.TeamHealthByDepartmentItem,
) adminTeamHealthByDepartmentResponse {
	departments := make([]adminTeamHealthDepartmentResponse, 0, len(items))
	for _, item := range items {
		departments = append(departments, adminTeamHealthDepartmentResponse{
			DepartmentID:    item.DepartmentID,
			DepartmentName:  item.DepartmentName,
			StaffingPercent: item.StaffingPercent,
			AbsencePercent:  item.AbsencePercent,
			TrainingPercent: item.TrainingPercent,
			Score:           item.Score,
			Risk:            item.Risk,
		})
	}

	return adminTeamHealthByDepartmentResponse{Departments: departments}
}

func formatAdminDashboardDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(adminDashboardDateLayout)
}
