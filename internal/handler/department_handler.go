package handler

import (
	"net/http"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DepartmentHandler struct {
	service domain.DepartmentService
}

func NewDepartmentHandler(service domain.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{service: service}
}

func (h *DepartmentHandler) CreateDepartment(ctx *gin.Context) {
	var req createDepartmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	department, err := h.service.CreateDepartment(
		ctx.Request.Context(),
		toCreateDepartmentParams(req),
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to create department", ""))
		return
	}

	ctx.JSON(
		http.StatusCreated,
		httpapi.OK(toDepartmentResponse(department), "Department created successfully"),
	)
}

func (h *DepartmentHandler) ListDepartments(ctx *gin.Context) {
	var req listDepartmentsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	page, err := h.service.ListDepartments(ctx.Request.Context(), toListDepartmentsParams(req))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to list departments", ""))
		return
	}

	results := make([]departmentResponse, len(page.Items))
	for i, item := range page.Items {
		results[i] = toDepartmentItemResponse(item)
	}

	response := httpapi.NewPageResponse(ctx, req.PageRequest, results, page.TotalCount)
	ctx.JSON(http.StatusOK, httpapi.OK(response, "Departments retrieved successfully"))
}

func (h *DepartmentHandler) GetDepartmentByID(ctx *gin.Context) {
	departmentID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid department ID", ""))
		return
	}

	department, err := h.service.GetDepartmentByID(ctx.Request.Context(), departmentID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to get department by id", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toDepartmentResponse(department), "Department retrieved successfully"),
	)
}

func (h *DepartmentHandler) UpdateDepartment(ctx *gin.Context) {
	departmentID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid department ID", ""))
		return
	}

	var req updateDepartmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	department, err := h.service.UpdateDepartment(
		ctx.Request.Context(),
		departmentID,
		toUpdateDepartmentParams(req),
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to update department", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toDepartmentResponse(department), "Department updated successfully"),
	)
}

func (h *DepartmentHandler) DeleteDepartment(ctx *gin.Context) {
	departmentID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid department ID", ""))
		return
	}

	if err := h.service.DeleteDepartment(ctx.Request.Context(), departmentID); err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to delete department", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(deleteDepartmentResponse{ID: departmentID}, "Department deleted successfully"),
	)
}
