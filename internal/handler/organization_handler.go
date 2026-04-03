package handler

import (
	"errors"
	"net/http"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RegisterOrganizationRoutes(
	rg *gin.RouterGroup,
	handler *OrganizationHandler,
	auth gin.HandlerFunc,
	requirePermission func(string) gin.HandlerFunc,
) {
	rg.POST(
		"/organizations",
		auth,
		requirePermission("LOCATION.CREATE"),
		handler.CreateOrganization,
	)
	rg.GET("/organizations", auth, requirePermission("LOCATION.VIEW"), handler.ListOrganizations)
	rg.GET(
		"/organizations/:id",
		auth,
		requirePermission("LOCATION.VIEW"),
		handler.GetOrganizationByID,
	)
	rg.GET(
		"/organizations/:id/counts",
		auth,
		requirePermission("LOCATION.VIEW"),
		handler.GetOrganizationCounts,
	)
	rg.POST(
		"/organizations/:id/locations",
		auth,
		requirePermission("LOCATION.CREATE"),
		handler.CreateOrganizationLocation,
	)
	rg.GET(
		"/organizations/:id/locations",
		auth,
		requirePermission("LOCATION.VIEW"),
		handler.ListOrganizationLocations,
	)
	rg.GET("/locations", auth, requirePermission("LOCATION.VIEW"), handler.ListAllLocations)
	rg.GET("/locations/:id", auth, requirePermission("LOCATION.VIEW"), handler.GetLocationByID)
	rg.PUT("/locations/:id", auth, requirePermission("LOCATION.UPDATE"), handler.UpdateLocation)
	rg.DELETE("/locations/:id", auth, requirePermission("LOCATION.DELETE"), handler.DeleteLocation)
	rg.POST("/locations/:id/shifts", auth, requirePermission("SHIFT.CREATE"), handler.CreateShift)
	rg.GET(
		"/locations/:id/shifts",
		auth,
		requirePermission("SHIFT.VIEW"),
		handler.ListShiftsByLocationID,
	)
	rg.PUT(
		"/locations/:id/shifts/:shift_id",
		auth,
		requirePermission("SHIFT.UPDATE"),
		handler.UpdateShift,
	)
	rg.DELETE(
		"/locations/:id/shifts/:shift_id",
		auth,
		requirePermission("SHIFT.DELETE"),
		handler.DeleteShift,
	)
	rg.PUT(
		"/organizations/:id",
		auth,
		requirePermission("LOCATION.UPDATE"),
		handler.UpdateOrganization,
	)
	rg.DELETE(
		"/organizations/:id",
		auth,
		requirePermission("LOCATION.DELETE"),
		handler.DeleteOrganization,
	)
	rg.GET(
		"/organizations/count",
		auth,
		requirePermission("LOCATION.VIEW"),
		handler.GetGlobalOrganizationCounts,
	)
}

type OrganizationHandler struct {
	service domain.OrganizationService
}

func NewOrganizationHandler(service domain.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{service: service}
}

func (h *OrganizationHandler) CreateOrganization(ctx *gin.Context) {
	var req createOrganizationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	organization, err := h.service.CreateOrganization(
		ctx.Request.Context(),
		toCreateOrganizationParams(req),
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to create organization", ""))
		return
	}

	ctx.JSON(
		http.StatusCreated,
		httpapi.OK(toCreateOrganizationResponse(organization), "Organization created successfully"),
	)
}

func (h *OrganizationHandler) GetOrganizationByID(ctx *gin.Context) {
	organizationID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid organization ID", ""))
		return
	}

	organization, err := h.service.GetOrganizationByID(ctx.Request.Context(), organizationID)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			httpapi.Fail("failed to get organization by id", ""),
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toGetOrganizationResponse(organization), "Organization retrieved successfully"),
	)
}

func (h *OrganizationHandler) UpdateOrganization(ctx *gin.Context) {
	organizationID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid organization ID", ""))
		return
	}

	var req updateOrganizationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	organization, err := h.service.UpdateOrganization(
		ctx.Request.Context(),
		organizationID,
		toUpdateOrganizationParams(req),
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to update organization", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toGetOrganizationResponse(organization), "Organization updated successfully"),
	)
}

func (h *OrganizationHandler) DeleteOrganization(ctx *gin.Context) {
	organizationID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid organization ID", ""))
		return
	}

	if err := h.service.DeleteOrganization(ctx.Request.Context(), organizationID); err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to delete organization", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			toDeleteOrganizationResponse(organizationID),
			"Organization deleted successfully",
		),
	)
}

func (h *OrganizationHandler) CreateOrganizationLocation(ctx *gin.Context) {
	organizationID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid organization ID", ""))
		return
	}

	var req createOrganizationLocationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	location, err := h.service.CreateOrganizationLocation(
		ctx.Request.Context(),
		organizationID,
		toCreateOrganizationLocationParams(req),
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to create location", ""))
		return
	}

	ctx.JSON(
		http.StatusCreated,
		httpapi.OK(toCreateOrganizationLocationResponse(location), "Location created successfully"),
	)
}

func (h *OrganizationHandler) ListAllLocations(ctx *gin.Context) {
	var req listLocationsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	page, err := h.service.ListAllLocations(ctx.Request.Context(), toListAllLocationsParams(req))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to list all locations", ""))
		return
	}

	results := make([]listLocationsResponse, len(page.Items))
	for i, item := range page.Items {
		results[i] = toListLocationsResponse(item)
	}

	response := httpapi.NewPageResponse(ctx, req.PageRequest, results, page.TotalCount)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "All locations retrieved successfully"))
}

func (h *OrganizationHandler) GetLocationByID(ctx *gin.Context) {
	locationID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid location ID", ""))
		return
	}

	location, err := h.service.GetLocationByID(ctx.Request.Context(), locationID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to get location by id", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toGetLocationResponse(location), "Location retrieved successfully"),
	)
}

func (h *OrganizationHandler) UpdateLocation(ctx *gin.Context) {
	locationID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid location ID", ""))
		return
	}

	var req updateLocationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	location, err := h.service.UpdateLocation(
		ctx.Request.Context(),
		locationID,
		toUpdateLocationParams(req),
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to update location", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toUpdateLocationResponse(location), "Location updated successfully"),
	)
}

func (h *OrganizationHandler) DeleteLocation(ctx *gin.Context) {
	locationID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid location ID", ""))
		return
	}

	if err := h.service.DeleteLocation(ctx.Request.Context(), locationID); err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to delete location", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toDeleteLocationResponse(locationID), "Location deleted successfully"),
	)
}

func (h *OrganizationHandler) CreateShift(ctx *gin.Context) {
	locationID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid location ID", ""))
		return
	}

	var req createShiftRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	shift, err := h.service.CreateShift(ctx.Request.Context(), toCreateShiftParams(locationID, req))
	if err != nil {
		if errors.Is(err, domain.ErrLocationShiftLimitReached) {
			ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to create shift", ""))
		return
	}

	ctx.JSON(
		http.StatusCreated,
		httpapi.OK(toCreateShiftResponse(shift), "Shift created successfully"),
	)
}

func (h *OrganizationHandler) ListShiftsByLocationID(ctx *gin.Context) {
	locationID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid location ID", ""))
		return
	}

	shifts, err := h.service.ListShiftsByLocationID(ctx.Request.Context(), locationID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to list shifts", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toListOrganizationLocationShiftDTOs(shifts), "Shifts retrieved successfully"),
	)
}

func (h *OrganizationHandler) UpdateShift(ctx *gin.Context) {
	locationID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid location ID", ""))
		return
	}
	_ = locationID

	shiftID, err := uuid.Parse(ctx.Param("shift_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid shift ID", ""))
		return
	}

	var req updateShiftRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	shift, err := h.service.UpdateShift(ctx.Request.Context(), shiftID, toUpdateShiftParams(req))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to update shift", ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toUpdateShiftResponse(shift), "Shift updated successfully"))
}

func (h *OrganizationHandler) DeleteShift(ctx *gin.Context) {
	locationID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid location ID", ""))
		return
	}
	_ = locationID

	shiftID, err := uuid.Parse(ctx.Param("shift_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid shift ID", ""))
		return
	}

	if err := h.service.DeleteShift(ctx.Request.Context(), shiftID); err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to delete shift", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toDeleteOrganizationResponse(shiftID), "Shift deleted successfully"),
	)
}

func (h *OrganizationHandler) GetOrganizationCounts(ctx *gin.Context) {
	organizationID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid organization ID", ""))
		return
	}

	counts, err := h.service.GetOrganizationCounts(ctx.Request.Context(), organizationID)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			httpapi.Fail("failed to get organization counts", ""),
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			toGetOrganizationCountsResponse(counts),
			"Organization counts retrieved successfully",
		),
	)
}

func (h *OrganizationHandler) GetGlobalOrganizationCounts(ctx *gin.Context) {
	counts, err := h.service.GetGlobalOrganizationCounts(ctx.Request.Context())
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			httpapi.Fail("failed to get global organization counts", ""),
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			toGetGlobalOrganizationCountsResponse(counts),
			"Global organization counts retrieved successfully",
		),
	)
}

func (h *OrganizationHandler) ListOrganizations(ctx *gin.Context) {
	var req listOrganizationsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	page, err := h.service.ListOrganizations(ctx.Request.Context(), toListOrganizationsParams(req))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to list organizations", ""))
		return
	}

	results := make([]listOrganizationsResponse, len(page.Items))
	for i, item := range page.Items {
		results[i] = toListOrganizationsResponse(item)
	}

	response := httpapi.NewPageResponse(ctx, req.PageRequest, results, page.TotalCount)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "Organizations retrieved successfully"))
}

func (h *OrganizationHandler) ListOrganizationLocations(ctx *gin.Context) {
	organizationID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid organization ID", ""))
		return
	}

	var req listOrganizationLocationsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	page, err := h.service.ListOrganizationLocations(
		ctx.Request.Context(),
		toListOrganizationLocationsParams(organizationID, req),
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to list locations", ""))
		return
	}

	results := make([]listOrganizationLocationsResponse, len(page.Items))
	for i, item := range page.Items {
		results[i] = toListOrganizationLocationsResponse(item)
	}

	response := httpapi.NewPageResponse(ctx, req.PageRequest, results, page.TotalCount)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "Locations retrieved successfully"))
}
