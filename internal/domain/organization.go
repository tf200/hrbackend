package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrLocationShiftLimitReached = errors.New(
	"location shift limit reached: max 4 shifts per location",
)

type Organization struct {
	ID                  uuid.UUID
	Name                string
	Street              string
	HouseNumber         string
	HouseNumberAddition *string
	PostalCode          string
	City                string
	Email               *string
	KvkNumber           *string
	BtwNumber           *string
	LocationCount       int64
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type CreateOrganizationParams struct {
	Name                string
	Street              string
	HouseNumber         string
	HouseNumberAddition *string
	PostalCode          string
	City                string
	Email               *string
	KvkNumber           *string
	BtwNumber           *string
}

type UpdateOrganizationParams struct {
	Name                *string
	Street              *string
	HouseNumber         *string
	HouseNumberAddition *string
	PostalCode          *string
	City                *string
	Email               *string
	KvkNumber           *string
	BtwNumber           *string
}

type UpdateOrganizationLocationParams struct {
	Name                *string
	Street              *string
	HouseNumber         *string
	HouseNumberAddition *string
	PostalCode          *string
	City                *string
	Capacity            *int32
}

type DeleteOrganizationLocationParams struct {
	LocationID uuid.UUID
}

type CreateOrganizationLocationParams struct {
	Name                string
	Street              string
	HouseNumber         string
	HouseNumberAddition *string
	PostalCode          string
	City                string
	Capacity            *int32
}

type ListOrganizationsParams struct {
	Limit  int32
	Offset int32
	Search string
}

type OrganizationPage struct {
	Items      []Organization
	TotalCount int64
}

type OrganizationLocation struct {
	ID                  uuid.UUID
	OrganizationID      uuid.UUID
	Name                string
	Street              string
	HouseNumber         string
	HouseNumberAddition *string
	PostalCode          string
	City                string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Shifts              []OrganizationLocationShift
}

type OrganizationLocationShift struct {
	ID         uuid.UUID
	LocationID uuid.UUID
	Slot       int16
	ShiftName  string
	StartTime  string
	EndTime    string
}

type ListOrganizationLocationsParams struct {
	OrganizationID uuid.UUID
	Limit          int32
	Offset         int32
	Search         string
}

type ListAllLocationsParams struct {
	Limit  int32
	Offset int32
	Search string
}

type GetLocationByIDParams struct {
	LocationID uuid.UUID
}

type CreateShiftParams struct {
	LocationID uuid.UUID
	Slot       int16
	ShiftName  string
	StartTime  string
	EndTime    string
}

type UpdateShiftParams struct {
	ShiftName string
	StartTime string
	EndTime   string
}

type OrganizationLocationPage struct {
	Items      []OrganizationLocation
	TotalCount int64
}

type OrganizationCounts struct {
	OrganizationID   uuid.UUID
	OrganizationName string
	LocationCount    int64
	EmployeeCount    int64
}

type GlobalOrganizationCounts struct {
	TotalLocations int64
	TotalEmployees int64
}

type OrganizationRepository interface {
	CreateOrganization(ctx context.Context, params CreateOrganizationParams) (*Organization, error)
	UpdateOrganization(
		ctx context.Context,
		organizationID uuid.UUID,
		params UpdateOrganizationParams,
	) (*Organization, error)
	DeleteOrganization(ctx context.Context, organizationID uuid.UUID) error
	CreateOrganizationLocation(
		ctx context.Context,
		organizationID uuid.UUID,
		params CreateOrganizationLocationParams,
	) (*OrganizationLocation, error)
	UpdateLocation(
		ctx context.Context,
		locationID uuid.UUID,
		params UpdateOrganizationLocationParams,
	) (*OrganizationLocation, error)
	DeleteLocation(ctx context.Context, locationID uuid.UUID) error
	CreateShift(ctx context.Context, params CreateShiftParams) (*OrganizationLocationShift, error)
	UpdateShift(
		ctx context.Context,
		shiftID uuid.UUID,
		params UpdateShiftParams,
	) (*OrganizationLocationShift, error)
	DeleteShift(ctx context.Context, shiftID uuid.UUID) error
	GetShiftsByLocationID(
		ctx context.Context,
		locationID uuid.UUID,
	) ([]OrganizationLocationShift, error)
	GetOrganizationCounts(
		ctx context.Context,
		organizationID uuid.UUID,
	) (*OrganizationCounts, error)
	GetGlobalOrganizationCounts(ctx context.Context) (*GlobalOrganizationCounts, error)
	GetOrganizationByID(ctx context.Context, organizationID uuid.UUID) (*Organization, error)
	GetLocationByID(ctx context.Context, locationID uuid.UUID) (*OrganizationLocation, error)
	ListOrganizations(
		ctx context.Context,
		params ListOrganizationsParams,
	) (*OrganizationPage, error)
	ListOrganizationLocations(
		ctx context.Context,
		params ListOrganizationLocationsParams,
	) (*OrganizationLocationPage, error)
	ListAllLocations(
		ctx context.Context,
		params ListAllLocationsParams,
	) (*OrganizationLocationPage, error)
}

type OrganizationService interface {
	CreateOrganization(ctx context.Context, params CreateOrganizationParams) (*Organization, error)
	UpdateOrganization(
		ctx context.Context,
		organizationID uuid.UUID,
		params UpdateOrganizationParams,
	) (*Organization, error)
	DeleteOrganization(ctx context.Context, organizationID uuid.UUID) error
	CreateOrganizationLocation(
		ctx context.Context,
		organizationID uuid.UUID,
		params CreateOrganizationLocationParams,
	) (*OrganizationLocation, error)
	UpdateLocation(
		ctx context.Context,
		locationID uuid.UUID,
		params UpdateOrganizationLocationParams,
	) (*OrganizationLocation, error)
	DeleteLocation(ctx context.Context, locationID uuid.UUID) error
	CreateShift(ctx context.Context, params CreateShiftParams) (*OrganizationLocationShift, error)
	UpdateShift(
		ctx context.Context,
		shiftID uuid.UUID,
		params UpdateShiftParams,
	) (*OrganizationLocationShift, error)
	DeleteShift(ctx context.Context, shiftID uuid.UUID) error
	ListShiftsByLocationID(
		ctx context.Context,
		locationID uuid.UUID,
	) ([]OrganizationLocationShift, error)
	GetOrganizationCounts(
		ctx context.Context,
		organizationID uuid.UUID,
	) (*OrganizationCounts, error)
	GetGlobalOrganizationCounts(ctx context.Context) (*GlobalOrganizationCounts, error)
	GetOrganizationByID(ctx context.Context, organizationID uuid.UUID) (*Organization, error)
	GetLocationByID(ctx context.Context, locationID uuid.UUID) (*OrganizationLocation, error)
	ListOrganizations(
		ctx context.Context,
		params ListOrganizationsParams,
	) (*OrganizationPage, error)
	ListOrganizationLocations(
		ctx context.Context,
		params ListOrganizationLocationsParams,
	) (*OrganizationLocationPage, error)
	ListAllLocations(
		ctx context.Context,
		params ListAllLocationsParams,
	) (*OrganizationLocationPage, error)
}
