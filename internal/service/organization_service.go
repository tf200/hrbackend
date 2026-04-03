package service

import (
	"context"
	"time"

	"hrbackend/internal/domain"
	"hrbackend/pkg/ptr"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type OrganizationService struct {
	repository domain.OrganizationRepository
	logger     domain.Logger
}

func NewOrganizationService(
	repository domain.OrganizationRepository,
	logger domain.Logger,
) domain.OrganizationService {
	return &OrganizationService{
		repository: repository,
		logger:     logger,
	}
}

func (s *OrganizationService) CreateOrganization(
	ctx context.Context,
	params domain.CreateOrganizationParams,
) (*domain.Organization, error) {
	organization, err := s.repository.CreateOrganization(ctx, params)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"OrganizationService.CreateOrganization",
				"failed to create organization",
				err,
				zap.String("name", params.Name),
				zap.String("city", params.City),
				zap.String("postal_code", params.PostalCode),
			)
		}
		return nil, err
	}

	return organization, nil
}

func (s *OrganizationService) UpdateOrganization(
	ctx context.Context,
	organizationID uuid.UUID,
	params domain.UpdateOrganizationParams,
) (*domain.Organization, error) {
	organization, err := s.repository.UpdateOrganization(ctx, organizationID, params)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"OrganizationService.UpdateOrganization",
				"failed to update organization",
				err,
				zap.String("organization_id", organizationID.String()),
				zap.String("name", ptr.OrEmpty(params.Name)),
			)
		}
		return nil, err
	}

	return organization, nil
}

func (s *OrganizationService) DeleteOrganization(
	ctx context.Context,
	organizationID uuid.UUID,
) error {
	if err := s.repository.DeleteOrganization(ctx, organizationID); err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"OrganizationService.DeleteOrganization",
				"failed to delete organization",
				err,
				zap.String("organization_id", organizationID.String()),
			)
		}
		return err
	}

	return nil
}

func (s *OrganizationService) CreateOrganizationLocation(
	ctx context.Context,
	organizationID uuid.UUID,
	params domain.CreateOrganizationLocationParams,
) (*domain.OrganizationLocation, error) {
	location, err := s.repository.CreateOrganizationLocation(ctx, organizationID, params)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"OrganizationService.CreateOrganizationLocation",
				"failed to create organization location",
				err,
				zap.String("organization_id", organizationID.String()),
				zap.String("name", params.Name),
				zap.String("city", params.City),
			)
		}
		return nil, err
	}

	return location, nil
}

func (s *OrganizationService) UpdateLocation(
	ctx context.Context,
	locationID uuid.UUID,
	params domain.UpdateOrganizationLocationParams,
) (*domain.OrganizationLocation, error) {
	location, err := s.repository.UpdateLocation(ctx, locationID, params)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"OrganizationService.UpdateLocation",
				"failed to update location",
				err,
				zap.String("location_id", locationID.String()),
			)
		}
		return nil, err
	}

	return location, nil
}

func (s *OrganizationService) DeleteLocation(ctx context.Context, locationID uuid.UUID) error {
	if err := s.repository.DeleteLocation(ctx, locationID); err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"OrganizationService.DeleteLocation",
				"failed to delete location",
				err,
				zap.String("location_id", locationID.String()),
			)
		}
		return err
	}

	return nil
}

func (s *OrganizationService) CreateShift(
	ctx context.Context,
	params domain.CreateShiftParams,
) (*domain.OrganizationLocationShift, error) {
	existingShifts, err := s.repository.GetShiftsByLocationID(ctx, params.LocationID)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"OrganizationService.CreateShift",
				"failed to check existing shifts",
				err,
				zap.String("location_id", params.LocationID.String()),
			)
		}
		return nil, err
	}

	if len(existingShifts) >= 4 {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"OrganizationService.CreateShift",
				"location shift limit reached",
				domain.ErrLocationShiftLimitReached,
				zap.String("location_id", params.LocationID.String()),
			)
		}
		return nil, domain.ErrLocationShiftLimitReached
	}

	usedSlots := make(map[int16]struct{}, len(existingShifts))
	for _, existingShift := range existingShifts {
		usedSlots[existingShift.Slot] = struct{}{}
	}

	var selectedSlot int16
	for slot := int16(1); slot <= 4; slot++ {
		if _, exists := usedSlots[slot]; !exists {
			selectedSlot = slot
			break
		}
	}

	if selectedSlot == 0 {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"OrganizationService.CreateShift",
				"no available shift slot found",
				domain.ErrLocationShiftLimitReached,
				zap.String("location_id", params.LocationID.String()),
			)
		}
		return nil, domain.ErrLocationShiftLimitReached
	}

	startTime, err := parseShiftTime(params.StartTime)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"OrganizationService.CreateShift",
				"invalid start time format",
				err,
				zap.String("location_id", params.LocationID.String()),
			)
		}
		return nil, err
	}

	endTime, err := parseShiftTime(params.EndTime)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"OrganizationService.CreateShift",
				"invalid end time format",
				err,
				zap.String("location_id", params.LocationID.String()),
			)
		}
		return nil, err
	}

	shift, err := s.repository.CreateShift(ctx, domain.CreateShiftParams{
		LocationID: params.LocationID,
		Slot:       selectedSlot,
		ShiftName:  params.ShiftName,
		StartTime:  startTime,
		EndTime:    endTime,
	})
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(ctx, "OrganizationService.CreateShift", "failed to create shift", err,
				zap.String("location_id", params.LocationID.String()),
				zap.String("shift_name", params.ShiftName),
			)
		}
		return nil, err
	}

	if s.logger != nil {
		s.logger.LogInfo(ctx, "OrganizationService.CreateShift", "shift created successfully",
			zap.String("shift_id", shift.ID.String()),
		)
	}

	return shift, nil
}

func (s *OrganizationService) UpdateShift(
	ctx context.Context,
	shiftID uuid.UUID,
	params domain.UpdateShiftParams,
) (*domain.OrganizationLocationShift, error) {
	startTime, err := parseShiftTime(params.StartTime)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"OrganizationService.UpdateShift",
				"invalid start time format",
				err,
				zap.String("shift_id", shiftID.String()),
			)
		}
		return nil, err
	}

	endTime, err := parseShiftTime(params.EndTime)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"OrganizationService.UpdateShift",
				"invalid end time format",
				err,
				zap.String("shift_id", shiftID.String()),
			)
		}
		return nil, err
	}

	shift, err := s.repository.UpdateShift(ctx, shiftID, domain.UpdateShiftParams{
		ShiftName: params.ShiftName,
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(ctx, "OrganizationService.UpdateShift", "failed to update shift", err,
				zap.String("shift_id", shiftID.String()),
				zap.String("shift_name", params.ShiftName),
			)
		}
		return nil, err
	}

	if s.logger != nil {
		s.logger.LogInfo(ctx, "OrganizationService.UpdateShift", "shift updated successfully",
			zap.String("shift_id", shift.ID.String()),
		)
	}

	return shift, nil
}

func (s *OrganizationService) DeleteShift(ctx context.Context, shiftID uuid.UUID) error {
	if err := s.repository.DeleteShift(ctx, shiftID); err != nil {
		if s.logger != nil {
			s.logger.LogError(ctx, "OrganizationService.DeleteShift", "failed to delete shift", err,
				zap.String("shift_id", shiftID.String()),
			)
		}
		return err
	}

	if s.logger != nil {
		s.logger.LogInfo(ctx, "OrganizationService.DeleteShift", "shift deleted successfully",
			zap.String("shift_id", shiftID.String()),
		)
	}

	return nil
}

func (s *OrganizationService) ListShiftsByLocationID(
	ctx context.Context,
	locationID uuid.UUID,
) ([]domain.OrganizationLocationShift, error) {
	shifts, err := s.repository.GetShiftsByLocationID(ctx, locationID)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"OrganizationService.ListShiftsByLocationID",
				"failed to list shifts",
				err,
				zap.String("location_id", locationID.String()),
			)
		}
		return nil, err
	}

	return shifts, nil
}

func parseShiftTime(value string) (string, error) {
	parsed, err := time.Parse("15:04:05", value)
	if err != nil {
		parsed, err = time.Parse("15:04", value)
		if err != nil {
			return "", err
		}
	}

	return parsed.Format("15:04:05"), nil
}

func (s *OrganizationService) GetLocationByID(
	ctx context.Context,
	locationID uuid.UUID,
) (*domain.OrganizationLocation, error) {
	location, err := s.repository.GetLocationByID(ctx, locationID)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"OrganizationService.GetLocationByID",
				"failed to get location by id",
				err,
				zap.String("location_id", locationID.String()),
			)
		}
		return nil, err
	}

	return location, nil
}

func (s *OrganizationService) ListAllLocations(
	ctx context.Context,
	params domain.ListAllLocationsParams,
) (*domain.OrganizationLocationPage, error) {
	page, err := s.repository.ListAllLocations(ctx, params)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"OrganizationService.ListAllLocations",
				"failed to list all locations",
				err,
				zap.String("search", params.Search),
				zap.Int32("limit", params.Limit),
				zap.Int32("offset", params.Offset),
			)
		}
		return nil, err
	}

	return page, nil
}

func (s *OrganizationService) GetOrganizationCounts(
	ctx context.Context,
	organizationID uuid.UUID,
) (*domain.OrganizationCounts, error) {
	counts, err := s.repository.GetOrganizationCounts(ctx, organizationID)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"OrganizationService.GetOrganizationCounts",
				"failed to get organization counts",
				err,
				zap.String("organization_id", organizationID.String()),
			)
		}
		return nil, err
	}

	return counts, nil
}

func (s *OrganizationService) GetGlobalOrganizationCounts(
	ctx context.Context,
) (*domain.GlobalOrganizationCounts, error) {
	counts, err := s.repository.GetGlobalOrganizationCounts(ctx)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"OrganizationService.GetGlobalOrganizationCounts",
				"failed to get global organization counts",
				err,
			)
		}
		return nil, err
	}

	return counts, nil
}

func (s *OrganizationService) GetOrganizationByID(
	ctx context.Context,
	organizationID uuid.UUID,
) (*domain.Organization, error) {
	organization, err := s.repository.GetOrganizationByID(ctx, organizationID)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"OrganizationService.GetOrganizationByID",
				"failed to get organization by id",
				err,
				zap.String("organization_id", organizationID.String()),
			)
		}
		return nil, err
	}

	return organization, nil
}

func (s *OrganizationService) ListOrganizations(
	ctx context.Context,
	params domain.ListOrganizationsParams,
) (*domain.OrganizationPage, error) {
	page, err := s.repository.ListOrganizations(ctx, params)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"OrganizationService.ListOrganizations",
				"failed to list organizations",
				err,
				zap.String("search", params.Search),
				zap.Int32("limit", params.Limit),
				zap.Int32("offset", params.Offset),
			)
		}
		return nil, err
	}

	return page, nil
}

func (s *OrganizationService) ListOrganizationLocations(
	ctx context.Context,
	params domain.ListOrganizationLocationsParams,
) (*domain.OrganizationLocationPage, error) {
	page, err := s.repository.ListOrganizationLocations(ctx, params)
	if err != nil {
		if s.logger != nil {
			s.logger.LogError(
				ctx,
				"OrganizationService.ListOrganizationLocations",
				"failed to list organization locations",
				err,
				zap.String("organization_id", params.OrganizationID.String()),
				zap.String("search", params.Search),
				zap.Int32("limit", params.Limit),
				zap.Int32("offset", params.Offset),
			)
		}
		return nil, err
	}

	return page, nil
}

var _ domain.OrganizationService = (*OrganizationService)(nil)
