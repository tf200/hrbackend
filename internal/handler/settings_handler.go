package handler

import (
	"net/http"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"

	"github.com/gin-gonic/gin"
)

func RegisterSettingsRoutes(
	rg *gin.RouterGroup,
	handler *SettingsHandler,
	auth gin.HandlerFunc,
	requirePermission func(string) gin.HandlerFunc,
) {
	rg.GET(
		"/settings/organization-profile",
		auth,
		requirePermission("SETTINGS.VIEW"),
		handler.GetOrganizationProfile,
	)
	rg.PUT(
		"/settings/organization-profile",
		auth,
		requirePermission("SETTINGS.UPDATE"),
		handler.UpdateOrganizationProfile,
	)
}

type SettingsHandler struct {
	service domain.SettingsService
}

func NewSettingsHandler(service domain.SettingsService) *SettingsHandler {
	return &SettingsHandler{service: service}
}

func (h *SettingsHandler) GetOrganizationProfile(ctx *gin.Context) {
	profile, err := h.service.GetOrganizationProfile(ctx.Request.Context())
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			httpapi.Fail("failed to get organization profile", ""),
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			toGetOrganizationProfileResponse(profile),
			"Organization profile retrieved successfully",
		),
	)
}

func (h *SettingsHandler) UpdateOrganizationProfile(ctx *gin.Context) {
	var req updateOrganizationProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	profile, err := h.service.UpdateOrganizationProfile(
		ctx.Request.Context(),
		toUpdateOrganizationProfileParams(req),
	)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			httpapi.Fail("failed to update organization profile", ""),
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			toGetOrganizationProfileResponse(profile),
			"Organization profile updated successfully",
		),
	)
}
