package handler

import (
	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"

	"github.com/google/uuid"
)

type fullTimeEmployeeDeptBreakdownItemResponse struct {
	DepartmentID   uuid.UUID `json:"department_id"`
	DepartmentName string    `json:"department_name"`
	TotalEmployees int64     `json:"total_employees"`
}

type fullTimeEmployeeLocBreakdownItemResponse struct {
	LocationID     uuid.UUID `json:"location_id"`
	LocationName   string    `json:"location_name"`
	TotalEmployees int64     `json:"total_employees"`
}

type fullTimeEmployeeBreakdownsResponse struct {
	ByDepartment []fullTimeEmployeeDeptBreakdownItemResponse `json:"by_department"`
	ByLocation   []fullTimeEmployeeLocBreakdownItemResponse  `json:"by_location"`
}

type getLeaveAbsenceTrendsRequest struct {
	View string `form:"view" binding:"omitempty,oneof=yearly last_6_months"`
	Year *int32 `form:"year"`
}

type leaveAbsenceTrendPointResponse struct {
	Month        string `json:"month"`
	EmployeesOut int64  `json:"employees_out"`
}

type leaveAbsenceTrendsResponse struct {
	View   string                           `json:"view"`
	Year   *int32                           `json:"year,omitempty"`
	Points []leaveAbsenceTrendPointResponse `json:"points"`
}

type getUpcomingDashboardAlertsRequest struct {
	Days  int32 `form:"days"`
	Limit int32 `form:"limit"`
}

type endingContractAlertResponse struct {
	EmployeeID      uuid.UUID `json:"employee_id"`
	EmployeeName    string    `json:"employee_name"`
	ContractID      uuid.UUID `json:"contract_id"`
	ContractType    string    `json:"contract_type"`
	ContractEndDate string    `json:"contract_end_date"`
	DaysRemaining   int32     `json:"days_remaining"`
	Department      string    `json:"department"`
	Location        string    `json:"location"`
}

type expiringCredentialAlertResponse struct {
	EmployeeID     uuid.UUID `json:"employee_id"`
	EmployeeName   string    `json:"employee_name"`
	CredentialID   uuid.UUID `json:"credential_id"`
	CredentialType string    `json:"credential_type"`
	Name           string    `json:"name"`
	ExpiryDate     string    `json:"expiry_date"`
	DaysRemaining  int32     `json:"days_remaining"`
}

type returningFromLeaveAlertResponse struct {
	EmployeeID      uuid.UUID `json:"employee_id"`
	EmployeeName    string    `json:"employee_name"`
	LeaveRequestID  uuid.UUID `json:"leave_request_id"`
	LeaveType       string    `json:"leave_type"`
	LeaveEndDate    string    `json:"leave_end_date"`
	ReturnDate      string    `json:"return_date"`
	DaysUntilReturn int32     `json:"days_until_return"`
}

type upcomingDashboardAlertsResponse struct {
	EndingContracts     []endingContractAlertResponse     `json:"ending_contracts"`
	ExpiringCredentials []expiringCredentialAlertResponse `json:"expiring_credentials"`
	ReturningFromLeave  []returningFromLeaveAlertResponse `json:"returning_from_leave"`
}

func toFullTimeEmployeeBreakdownsResponse(
	b *domain.FullTimeEmployeeBreakdowns,
) fullTimeEmployeeBreakdownsResponse {
	deptItems := make([]fullTimeEmployeeDeptBreakdownItemResponse, len(b.ByDepartment))
	for i, item := range b.ByDepartment {
		deptItems[i] = fullTimeEmployeeDeptBreakdownItemResponse{
			DepartmentID:   item.DepartmentID,
			DepartmentName: item.DepartmentName,
			TotalEmployees: item.TotalEmployees,
		}
	}

	locItems := make([]fullTimeEmployeeLocBreakdownItemResponse, len(b.ByLocation))
	for i, item := range b.ByLocation {
		locItems[i] = fullTimeEmployeeLocBreakdownItemResponse{
			LocationID:     item.LocationID,
			LocationName:   item.LocationName,
			TotalEmployees: item.TotalEmployees,
		}
	}

	return fullTimeEmployeeBreakdownsResponse{
		ByDepartment: deptItems,
		ByLocation:   locItems,
	}
}

func toLeaveAbsenceTrendsResponse(trends *domain.LeaveAbsenceTrends) leaveAbsenceTrendsResponse {
	points := make([]leaveAbsenceTrendPointResponse, len(trends.Points))
	for i, point := range trends.Points {
		points[i] = leaveAbsenceTrendPointResponse{
			Month:        point.Month.Format("2006-01"),
			EmployeesOut: point.EmployeesOut,
		}
	}

	return leaveAbsenceTrendsResponse{
		View:   trends.View,
		Year:   trends.Year,
		Points: points,
	}
}

func toUpcomingDashboardAlertsResponse(
	alerts *domain.UpcomingDashboardAlerts,
) upcomingDashboardAlertsResponse {
	endingContracts := make([]endingContractAlertResponse, len(alerts.EndingContracts))
	for i, item := range alerts.EndingContracts {
		endingContracts[i] = endingContractAlertResponse{
			EmployeeID:      item.EmployeeID,
			EmployeeName:    item.EmployeeName,
			ContractID:      item.ContractID,
			ContractType:    item.ContractType,
			ContractEndDate: item.ContractEndDate.Format("2006-01-02"),
			DaysRemaining:   item.DaysRemaining,
			Department:      item.Department,
			Location:        item.Location,
		}
	}

	expiringCredentials := make([]expiringCredentialAlertResponse, len(alerts.ExpiringCredentials))
	for i, item := range alerts.ExpiringCredentials {
		expiringCredentials[i] = expiringCredentialAlertResponse{
			EmployeeID:     item.EmployeeID,
			EmployeeName:   item.EmployeeName,
			CredentialID:   item.CredentialID,
			CredentialType: item.CredentialType,
			Name:           item.Name,
			ExpiryDate:     item.ExpiryDate.Format("2006-01-02"),
			DaysRemaining:  item.DaysRemaining,
		}
	}

	returningFromLeave := make([]returningFromLeaveAlertResponse, len(alerts.ReturningFromLeave))
	for i, item := range alerts.ReturningFromLeave {
		returningFromLeave[i] = returningFromLeaveAlertResponse{
			EmployeeID:      item.EmployeeID,
			EmployeeName:    item.EmployeeName,
			LeaveRequestID:  item.LeaveRequestID,
			LeaveType:       item.LeaveType,
			LeaveEndDate:    item.LeaveEndDate.Format("2006-01-02"),
			ReturnDate:      item.ReturnDate.Format("2006-01-02"),
			DaysUntilReturn: item.DaysUntilReturn,
		}
	}

	return upcomingDashboardAlertsResponse{
		EndingContracts:     endingContracts,
		ExpiringCredentials: expiringCredentials,
		ReturningFromLeave:  returningFromLeave,
	}
}

type adminDashboardKPIsResponse struct {
	TotalEmployees   int64 `json:"total_employees"`
	EmployeesPresent int64 `json:"employees_present"`
	TotalDocuments   int64 `json:"total_documents"`
	ProcessingDocs   int64 `json:"processing_docs"`
}

func toAdminDashboardKPIsResponse(kpi *domain.AdminDashboardKPI) adminDashboardKPIsResponse {
	return adminDashboardKPIsResponse{
		TotalEmployees:   kpi.TotalEmployees,
		EmployeesPresent: kpi.EmployeesPresent,
		TotalDocuments:   kpi.TotalDocuments,
		ProcessingDocs:   kpi.ProcessingDocs,
	}
}

type listRecentEmployeesRequest struct {
	httpapi.PageRequest
}

type recentEmployeeItemResponse struct {
	ID                 uuid.UUID `json:"id"`
	Name               string    `json:"name"`
	Department         *string   `json:"department"`
	Location           *string   `json:"location"`
	CreatedAt          string    `json:"created_at"`
}

func toRecentEmployeeItemResponse(emp domain.RecentDashboardEmployee) recentEmployeeItemResponse {
	return recentEmployeeItemResponse{
		ID:                 emp.ID,
		Name:               emp.FirstName + " " + emp.LastName,
		Department:         emp.DepartmentName,
		Location:           emp.LocationName,
		CreatedAt:          emp.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
