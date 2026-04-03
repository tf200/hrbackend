package handler

import (
	"time"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"

	"github.com/google/uuid"
)

type listOrganizationsRequest struct {
	httpapi.PageRequest
	Search *string `form:"search"`
}

type createOrganizationRequest struct {
	Name                string  `json:"name" binding:"required"`
	Street              string  `json:"street" binding:"required"`
	HouseNumber         string  `json:"house_number" binding:"required"`
	HouseNumberAddition *string `json:"house_number_addition"`
	PostalCode          string  `json:"postal_code" binding:"required"`
	City                string  `json:"city" binding:"required"`
	Email               *string `json:"email"`
	KvkNumber           *string `json:"kvk_number"`
	BtwNumber           *string `json:"btw_number"`
}

type updateOrganizationRequest struct {
	Name                *string `json:"name"`
	Street              *string `json:"street"`
	HouseNumber         *string `json:"house_number"`
	HouseNumberAddition *string `json:"house_number_addition"`
	PostalCode          *string `json:"postal_code"`
	City                *string `json:"city"`
	Email               *string `json:"email"`
	KvkNumber           *string `json:"kvk_number"`
	BtwNumber           *string `json:"btw_number"`
}

type createOrganizationResponse struct {
	ID                  uuid.UUID `json:"id"`
	Name                string    `json:"name"`
	Street              string    `json:"street"`
	HouseNumber         string    `json:"house_number"`
	HouseNumberAddition *string   `json:"house_number_addition"`
	PostalCode          string    `json:"postal_code"`
	City                string    `json:"city"`
	Email               *string   `json:"email"`
	KvkNumber           *string   `json:"kvk_number"`
	BtwNumber           *string   `json:"btw_number"`
}

type updateOrganizationResponse = getOrganizationResponse

type deleteOrganizationResponse struct {
	ID uuid.UUID `json:"id"`
}

type createOrganizationLocationRequest struct {
	Name                string  `json:"name" binding:"required"`
	Street              string  `json:"street" binding:"required"`
	HouseNumber         string  `json:"house_number" binding:"required"`
	HouseNumberAddition *string `json:"house_number_addition"`
	PostalCode          string  `json:"postal_code" binding:"required"`
	City                string  `json:"city" binding:"required"`
}

type createOrganizationLocationResponse struct {
	ID                  uuid.UUID `json:"id"`
	Name                string    `json:"name"`
	Street              string    `json:"street"`
	HouseNumber         string    `json:"house_number"`
	HouseNumberAddition *string   `json:"house_number_addition"`
	PostalCode          string    `json:"postal_code"`
	City                string    `json:"city"`
}

type getLocationResponse struct {
	ID                  uuid.UUID `json:"id"`
	Name                string    `json:"name"`
	Street              string    `json:"street"`
	HouseNumber         string    `json:"house_number"`
	HouseNumberAddition *string   `json:"house_number_addition"`
	PostalCode          string    `json:"postal_code"`
	City                string    `json:"city"`
}

type updateLocationRequest struct {
	Name                *string `json:"name"`
	Street              *string `json:"street"`
	HouseNumber         *string `json:"house_number"`
	HouseNumberAddition *string `json:"house_number_addition"`
	PostalCode          *string `json:"postal_code"`
	City                *string `json:"city"`
}

type updateLocationResponse = getLocationResponse

type deleteLocationResponse struct {
	ID uuid.UUID `json:"id"`
}

type createShiftRequest struct {
	ShiftName string `json:"shift" binding:"required"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
}

type updateShiftRequest = createShiftRequest

type createShiftResponse struct {
	ID         uuid.UUID `json:"id"`
	LocationID uuid.UUID `json:"location_id"`
	Slot       int16     `json:"slot"`
	ShiftName  string    `json:"shift"`
	StartTime  string    `json:"start_time"`
	EndTime    string    `json:"end_time"`
}

type updateShiftResponse = createShiftResponse

type getOrganizationResponse struct {
	ID                  uuid.UUID `json:"id"`
	Name                string    `json:"name"`
	Street              string    `json:"street"`
	HouseNumber         string    `json:"house_number"`
	HouseNumberAddition *string   `json:"house_number_addition"`
	PostalCode          string    `json:"postal_code"`
	City                string    `json:"city"`
	Email               *string   `json:"email"`
	KvkNumber           *string   `json:"kvk_number"`
	BtwNumber           *string   `json:"btw_number"`
	LocationCount       int64     `json:"location_count"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type getOrganizationCountsResponse struct {
	OrganizationID   uuid.UUID `json:"organization_id"`
	OrganizationName string    `json:"organization_name"`
	LocationCount    int64     `json:"location_count"`
	EmployeeCount    int64     `json:"employee_count"`
}

type getGlobalOrganizationCountsResponse struct {
	TotalLocations int64 `json:"total_locations"`
	TotalEmployees int64 `json:"total_employees"`
}

type listOrganizationsResponse struct {
	ID                  uuid.UUID `json:"id"`
	Name                string    `json:"name"`
	Street              string    `json:"street"`
	HouseNumber         string    `json:"house_number"`
	HouseNumberAddition *string   `json:"house_number_addition"`
	PostalCode          string    `json:"postal_code"`
	City                string    `json:"city"`
	Email               *string   `json:"email"`
	KvkNumber           *string   `json:"kvk_number"`
	BtwNumber           *string   `json:"btw_number"`
	LocationCount       int64     `json:"location_count"`
}

type listOrganizationLocationsRequest struct {
	httpapi.PageRequest
	Search *string `form:"search"`
}

type listLocationsRequest = listOrganizationLocationsRequest

type listLocationsResponse = listOrganizationLocationsResponse

type listOrganizationLocationsResponse struct {
	ID                  uuid.UUID                          `json:"id"`
	Name                string                             `json:"name"`
	Street              string                             `json:"street"`
	HouseNumber         string                             `json:"house_number"`
	HouseNumberAddition *string                            `json:"house_number_addition"`
	PostalCode          string                             `json:"postal_code"`
	City                string                             `json:"city"`
	CreatedAt           time.Time                          `json:"created_at"`
	UpdatedAt           time.Time                          `json:"updated_at"`
	Shifts              []listOrganizationLocationShiftDTO `json:"shifts"`
}

type listOrganizationLocationShiftDTO struct {
	ID         uuid.UUID `json:"id"`
	LocationID uuid.UUID `json:"location_id"`
	Slot       int16     `json:"slot"`
	ShiftName  string    `json:"shift"`
	StartTime  string    `json:"start_time"`
	EndTime    string    `json:"end_time"`
}

func toListOrganizationsParams(req listOrganizationsRequest) domain.ListOrganizationsParams {
	search := ""
	if req.Search != nil {
		search = *req.Search
	}

	return domain.ListOrganizationsParams{
		Limit:  req.PageSize,
		Offset: (req.Page - 1) * req.PageSize,
		Search: search,
	}
}

func toCreateOrganizationParams(req createOrganizationRequest) domain.CreateOrganizationParams {
	return domain.CreateOrganizationParams{
		Name:                req.Name,
		Street:              req.Street,
		HouseNumber:         req.HouseNumber,
		HouseNumberAddition: req.HouseNumberAddition,
		PostalCode:          req.PostalCode,
		City:                req.City,
		Email:               req.Email,
		KvkNumber:           req.KvkNumber,
		BtwNumber:           req.BtwNumber,
	}
}

func toUpdateOrganizationParams(req updateOrganizationRequest) domain.UpdateOrganizationParams {
	return domain.UpdateOrganizationParams{
		Name:                req.Name,
		Street:              req.Street,
		HouseNumber:         req.HouseNumber,
		HouseNumberAddition: req.HouseNumberAddition,
		PostalCode:          req.PostalCode,
		City:                req.City,
		Email:               req.Email,
		KvkNumber:           req.KvkNumber,
		BtwNumber:           req.BtwNumber,
	}
}

func toCreateOrganizationResponse(organization *domain.Organization) createOrganizationResponse {
	return createOrganizationResponse{
		ID:                  organization.ID,
		Name:                organization.Name,
		Street:              organization.Street,
		HouseNumber:         organization.HouseNumber,
		HouseNumberAddition: organization.HouseNumberAddition,
		PostalCode:          organization.PostalCode,
		City:                organization.City,
		Email:               organization.Email,
		KvkNumber:           organization.KvkNumber,
		BtwNumber:           organization.BtwNumber,
	}
}

func toGetOrganizationResponse(organization *domain.Organization) getOrganizationResponse {
	return getOrganizationResponse{
		ID:                  organization.ID,
		Name:                organization.Name,
		Street:              organization.Street,
		HouseNumber:         organization.HouseNumber,
		HouseNumberAddition: organization.HouseNumberAddition,
		PostalCode:          organization.PostalCode,
		City:                organization.City,
		Email:               organization.Email,
		KvkNumber:           organization.KvkNumber,
		BtwNumber:           organization.BtwNumber,
		LocationCount:       organization.LocationCount,
		CreatedAt:           organization.CreatedAt,
		UpdatedAt:           organization.UpdatedAt,
	}
}

func toDeleteOrganizationResponse(id uuid.UUID) deleteOrganizationResponse {
	return deleteOrganizationResponse{ID: id}
}

func toCreateOrganizationLocationParams(req createOrganizationLocationRequest) domain.CreateOrganizationLocationParams {
	return domain.CreateOrganizationLocationParams{
		Name:                req.Name,
		Street:              req.Street,
		HouseNumber:         req.HouseNumber,
		HouseNumberAddition: req.HouseNumberAddition,
		PostalCode:          req.PostalCode,
		City:                req.City,
	}
}

func toCreateOrganizationLocationResponse(location *domain.OrganizationLocation) createOrganizationLocationResponse {
	return createOrganizationLocationResponse{
		ID:                  location.ID,
		Name:                location.Name,
		Street:              location.Street,
		HouseNumber:         location.HouseNumber,
		HouseNumberAddition: location.HouseNumberAddition,
		PostalCode:          location.PostalCode,
		City:                location.City,
	}
}

func toGetLocationResponse(location *domain.OrganizationLocation) getLocationResponse {
	return getLocationResponse{
		ID:                  location.ID,
		Name:                location.Name,
		Street:              location.Street,
		HouseNumber:         location.HouseNumber,
		HouseNumberAddition: location.HouseNumberAddition,
		PostalCode:          location.PostalCode,
		City:                location.City,
	}
}

func toUpdateLocationParams(req updateLocationRequest) domain.UpdateOrganizationLocationParams {
	return domain.UpdateOrganizationLocationParams{
		Name:                req.Name,
		Street:              req.Street,
		HouseNumber:         req.HouseNumber,
		HouseNumberAddition: req.HouseNumberAddition,
		PostalCode:          req.PostalCode,
		City:                req.City,
	}
}

func toUpdateLocationResponse(location *domain.OrganizationLocation) updateLocationResponse {
	return toGetLocationResponse(location)
}

func toDeleteLocationResponse(id uuid.UUID) deleteLocationResponse {
	return deleteLocationResponse{ID: id}
}

func toCreateShiftParams(locationID uuid.UUID, req createShiftRequest) domain.CreateShiftParams {
	return domain.CreateShiftParams{
		LocationID: locationID,
		ShiftName:  req.ShiftName,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
	}
}

func toUpdateShiftParams(req updateShiftRequest) domain.UpdateShiftParams {
	return domain.UpdateShiftParams{
		ShiftName: req.ShiftName,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}
}

func toCreateShiftResponse(shift *domain.OrganizationLocationShift) createShiftResponse {
	return createShiftResponse{
		ID:         shift.ID,
		LocationID: shift.LocationID,
		Slot:       shift.Slot,
		ShiftName:  shift.ShiftName,
		StartTime:  shift.StartTime,
		EndTime:    shift.EndTime,
	}
}

func toUpdateShiftResponse(shift *domain.OrganizationLocationShift) updateShiftResponse {
	return toCreateShiftResponse(shift)
}

func toGetOrganizationCountsResponse(counts *domain.OrganizationCounts) getOrganizationCountsResponse {
	return getOrganizationCountsResponse{
		OrganizationID:   counts.OrganizationID,
		OrganizationName: counts.OrganizationName,
		LocationCount:    counts.LocationCount,
		EmployeeCount:    counts.EmployeeCount,
	}
}

func toGetGlobalOrganizationCountsResponse(counts *domain.GlobalOrganizationCounts) getGlobalOrganizationCountsResponse {
	return getGlobalOrganizationCountsResponse{
		TotalLocations: counts.TotalLocations,
		TotalEmployees: counts.TotalEmployees,
	}
}

func toListOrganizationsResponse(organization domain.Organization) listOrganizationsResponse {
	return listOrganizationsResponse{
		ID:                  organization.ID,
		Name:                organization.Name,
		Street:              organization.Street,
		HouseNumber:         organization.HouseNumber,
		HouseNumberAddition: organization.HouseNumberAddition,
		PostalCode:          organization.PostalCode,
		City:                organization.City,
		Email:               organization.Email,
		KvkNumber:           organization.KvkNumber,
		BtwNumber:           organization.BtwNumber,
		LocationCount:       organization.LocationCount,
	}
}

func toListOrganizationLocationsParams(organizationID uuid.UUID, req listOrganizationLocationsRequest) domain.ListOrganizationLocationsParams {
	search := ""
	if req.Search != nil {
		search = *req.Search
	}

	return domain.ListOrganizationLocationsParams{
		OrganizationID: organizationID,
		Limit:          req.PageSize,
		Offset:         (req.Page - 1) * req.PageSize,
		Search:         search,
	}
}

func toListAllLocationsParams(req listLocationsRequest) domain.ListAllLocationsParams {
	search := ""
	if req.Search != nil {
		search = *req.Search
	}

	return domain.ListAllLocationsParams{
		Limit:  req.PageSize,
		Offset: (req.Page - 1) * req.PageSize,
		Search: search,
	}
}

func toListOrganizationLocationsResponse(location domain.OrganizationLocation) listOrganizationLocationsResponse {
	return listOrganizationLocationsResponse{
		ID:                  location.ID,
		Name:                location.Name,
		Street:              location.Street,
		HouseNumber:         location.HouseNumber,
		HouseNumberAddition: location.HouseNumberAddition,
		PostalCode:          location.PostalCode,
		City:                location.City,
		CreatedAt:           location.CreatedAt,
		UpdatedAt:           location.UpdatedAt,
		Shifts:              toListOrganizationLocationShiftDTOs(location.Shifts),
	}
}

func toListLocationsResponse(location domain.OrganizationLocation) listLocationsResponse {
	return toListOrganizationLocationsResponse(location)
}

func toListOrganizationLocationShiftDTOs(shifts []domain.OrganizationLocationShift) []listOrganizationLocationShiftDTO {
	result := make([]listOrganizationLocationShiftDTO, len(shifts))
	for i, shift := range shifts {
		result[i] = listOrganizationLocationShiftDTO{
			ID:         shift.ID,
			LocationID: shift.LocationID,
			Slot:       shift.Slot,
			ShiftName:  shift.ShiftName,
			StartTime:  shift.StartTime,
			EndTime:    shift.EndTime,
		}
	}
	return result
}
