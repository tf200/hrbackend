package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/goccy/go-json"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"
	"hrbackend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	errCodeHandbookDraftAlreadyExists  = "HANDBOOK_DRAFT_ALREADY_EXISTS"
	errCodeHandbookTemplateNotFound    = "HANDBOOK_TEMPLATE_NOT_FOUND"
	errCodeHandbookTemplateNotDraft    = "HANDBOOK_TEMPLATE_NOT_DRAFT"
	errCodeHandbookTemplateNoSteps     = "HANDBOOK_TEMPLATE_NO_STEPS"
	errCodeHandbookStepNotFound        = "HANDBOOK_STEP_NOT_FOUND"
	errCodeEmployeeHandbookNotFound    = "EMPLOYEE_HANDBOOK_NOT_FOUND"
	errCodeEmployeeHandbookNotActive   = "EMPLOYEE_HANDBOOK_NOT_ACTIVE"
	errCodeHandbookLinkURLInvalid      = "HANDBOOK_LINK_URL_INVALID"
	errCodeHandbookQuizContentInvalid  = "HANDBOOK_QUIZ_CONTENT_INVALID"
	errCodeHandbookStepReorderMismatch = "HANDBOOK_STEP_REORDER_SET_MISMATCH"
	errCodeInvalidRequest              = "INVALID_REQUEST"
)

func RegisterHandbookRoutes(
	rg *gin.RouterGroup,
	handler *HandbookHandler,
	auth gin.HandlerFunc,
	requirePermission func(string) gin.HandlerFunc,
) {
	handbook := rg.Group("/handbook")
	handbook.Use(auth)

	handbook.GET("/me", requirePermission("HANDBOOK.SELF.VIEW"), handler.GetMyActiveHandbook)
	handbook.POST("/me/start", requirePermission("HANDBOOK.SELF.UPDATE"), handler.StartMyHandbook)
	handbook.POST(
		"/me/steps/:step_id/complete",
		requirePermission("HANDBOOK.SELF.UPDATE"),
		handler.CompleteMyHandbookStep,
	)

	handbook.POST(
		"/templates",
		requirePermission("HANDBOOK.TEMPLATE.CREATE"),
		handler.CreateHandbookTemplate,
	)
	handbook.POST(
		"/templates/clone",
		requirePermission("HANDBOOK.TEMPLATE.CREATE"),
		handler.CloneHandbookTemplate,
	)
	handbook.PATCH(
		"/templates/:template_id",
		requirePermission("HANDBOOK.TEMPLATE.UPDATE"),
		handler.UpdateHandbookTemplate,
	)
	handbook.POST(
		"/templates/publish",
		requirePermission("HANDBOOK.TEMPLATE.PUBLISH"),
		handler.PublishHandbookTemplate,
	)
	handbook.GET(
		"/departments/:department_id/templates",
		requirePermission("HANDBOOK.TEMPLATE.VIEW"),
		handler.ListHandbookTemplatesByDepartment,
	)

	handbook.POST("/steps", requirePermission("HANDBOOK.STEP.CREATE"), handler.CreateHandbookStep)
	handbook.PATCH(
		"/steps/:step_id",
		requirePermission("HANDBOOK.STEP.UPDATE"),
		handler.UpdateHandbookStep,
	)
	handbook.DELETE(
		"/steps/:step_id",
		requirePermission("HANDBOOK.STEP.DELETE"),
		handler.DeleteHandbookStep,
	)
	handbook.GET(
		"/templates/:template_id/steps",
		requirePermission("HANDBOOK.STEP.VIEW"),
		handler.ListHandbookStepsByTemplate,
	)
	handbook.POST(
		"/templates/:template_id/steps/reorder",
		requirePermission("HANDBOOK.STEP.UPDATE"),
		handler.ReorderHandbookSteps,
	)
	handbook.GET(
		"/employees/:employee_id/history",
		requirePermission("HANDBOOK.ASSIGN"),
		handler.ListEmployeeHandbookHistory,
	)

	handbook.GET(
		"/assignments",
		requirePermission("HANDBOOK.ASSIGN"),
		handler.ListEmployeeHandbookAssignments,
	)
	handbook.GET(
		"/assignments/eligible-employees",
		requirePermission("HANDBOOK.ASSIGN"),
		handler.ListEligibleEmployees,
	)
	handbook.GET(
		"/assignments/:handbook_id",
		requirePermission("HANDBOOK.ASSIGN"),
		handler.GetEmployeeHandbookDetails,
	)
	handbook.POST(
		"/assignments",
		requirePermission("HANDBOOK.ASSIGN"),
		handler.AssignHandbookTemplateToEmployee,
	)
	handbook.POST(
		"/assignments/:handbook_id/waive",
		requirePermission("HANDBOOK.ASSIGN"),
		handler.WaiveEmployeeHandbook,
	)
}

type HandbookHandler struct {
	service domain.HandbookService
}

func NewHandbookHandler(service domain.HandbookService) *HandbookHandler {
	return &HandbookHandler{service: service}
}

func (h *HandbookHandler) GetMyActiveHandbook(ctx *gin.Context) {
	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	res, err := h.service.GetMyActiveHandbook(ctx.Request.Context(), employeeID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
		return
	}
	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toMyActiveHandbookResponse(res), "Active handbook retrieved"),
	)
}

func (h *HandbookHandler) StartMyHandbook(ctx *gin.Context) {
	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}
	res, err := h.service.StartMyHandbook(ctx.Request.Context(), employeeID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	ctx.JSON(http.StatusOK, httpapi.OK(toStartedHandbookResponse(res), "Handbook started"))
}

func (h *HandbookHandler) CompleteMyHandbookStep(ctx *gin.Context) {
	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	stepID, err := uuid.Parse(ctx.Param("step_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid step_id", errCodeInvalidRequest))
		return
	}

	var req completeMyHandbookStepRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid request body", errCodeInvalidRequest))
		return
	}

	res, err := h.service.CompleteMyHandbookStep(
		ctx.Request.Context(),
		employeeID,
		stepID,
		req.Response,
	)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	ctx.JSON(http.StatusOK, httpapi.OK(toCompletedHandbookStepResponse(res), "Step completed"))
}

func (h *HandbookHandler) CreateHandbookTemplate(ctx *gin.Context) {
	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	var req createHandbookTemplateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid request body", errCodeInvalidRequest))
		return
	}

	res, err := h.service.CreateTemplateForDepartment(
		ctx.Request.Context(),
		employeeID,
		toCreateTemplateForDepartmentParams(req),
	)
	if err != nil {
		if errors.Is(err, domain.ErrDraftTemplateAlreadyExists) {
			ctx.JSON(
				http.StatusConflict,
				httpapi.Fail(err.Error(), errCodeHandbookDraftAlreadyExists),
			)
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail(err.Error(), ""))
		return
	}
	ctx.JSON(http.StatusCreated, httpapi.OK(toHandbookTemplateResponse(res), "Template created"))
}

func (h *HandbookHandler) CloneHandbookTemplate(ctx *gin.Context) {
	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	var req cloneHandbookTemplateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid request body", errCodeInvalidRequest))
		return
	}

	res, err := h.service.CloneTemplateToDraft(
		ctx.Request.Context(),
		employeeID,
		toCloneTemplateToDraftParams(req),
	)
	if err != nil {
		if errors.Is(err, domain.ErrDraftTemplateAlreadyExists) {
			ctx.JSON(
				http.StatusConflict,
				httpapi.Fail(err.Error(), errCodeHandbookDraftAlreadyExists),
			)
			return
		}
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	ctx.JSON(http.StatusCreated, httpapi.OK(toHandbookTemplateResponse(res), "Template cloned"))
}

func (h *HandbookHandler) UpdateHandbookTemplate(ctx *gin.Context) {
	templateID, err := uuid.Parse(ctx.Param("template_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid template_id", errCodeInvalidRequest))
		return
	}

	var req updateHandbookTemplateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid request body", errCodeInvalidRequest))
		return
	}
	if req.Title == nil && req.Description == nil {
		ctx.JSON(
			http.StatusBadRequest,
			httpapi.Fail(
				"at least one field is required: title or description",
				errCodeInvalidRequest,
			),
		)
		return
	}

	var (
		titleVal *string
		setTitle bool
		descVal  *string
		setDesc  bool
	)
	if req.Title != nil {
		setTitle = true
		if string(*req.Title) != "null" {
			var title string
			if err := json.Unmarshal(*req.Title, &title); err != nil {
				ctx.JSON(
					http.StatusBadRequest,
					httpapi.Fail("title must be a string", errCodeInvalidRequest),
				)
				return
			}
			if strings.TrimSpace(title) == "" {
				ctx.JSON(
					http.StatusBadRequest,
					httpapi.Fail("title cannot be empty", errCodeInvalidRequest),
				)
				return
			}
			titleVal = &title
		}
	}
	if setTitle && titleVal == nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("title cannot be null", errCodeInvalidRequest))
		return
	}
	if req.Description != nil {
		setDesc = true
		if string(*req.Description) != "null" {
			var description string
			if err := json.Unmarshal(*req.Description, &description); err != nil {
				ctx.JSON(
					http.StatusBadRequest,
					httpapi.Fail("description must be a string or null", errCodeInvalidRequest),
				)
				return
			}
			descVal = &description
		}
	}

	res, err := h.service.UpdateTemplate(ctx.Request.Context(), domain.UpdateTemplateParams{
		TemplateID:     templateID,
		Title:          titleVal,
		SetTitle:       setTitle,
		Description:    descVal,
		SetDescription: setDesc,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTemplateNotFound):
			ctx.JSON(
				http.StatusNotFound,
				httpapi.Fail(err.Error(), errCodeHandbookTemplateNotFound),
			)
		case errors.Is(err, domain.ErrTemplateNotDraft):
			ctx.JSON(
				http.StatusBadRequest,
				httpapi.Fail(err.Error(), errCodeHandbookTemplateNotDraft),
			)
		default:
			ctx.JSON(http.StatusInternalServerError, httpapi.Fail(err.Error(), ""))
		}
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toHandbookTemplateResponse(res), "Template updated"))
}

func (h *HandbookHandler) PublishHandbookTemplate(ctx *gin.Context) {
	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	var req publishHandbookTemplateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid request body", errCodeInvalidRequest))
		return
	}

	res, err := h.service.PublishTemplate(
		ctx.Request.Context(),
		employeeID,
		toPublishTemplateParams(req),
	)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTemplateNotFound):
			ctx.JSON(
				http.StatusNotFound,
				httpapi.Fail(err.Error(), errCodeHandbookTemplateNotFound),
			)
		case errors.Is(err, domain.ErrTemplateNotDraft):
			ctx.JSON(
				http.StatusBadRequest,
				httpapi.Fail(err.Error(), errCodeHandbookTemplateNotDraft),
			)
		case errors.Is(err, domain.ErrTemplateHasNoSteps):
			ctx.JSON(
				http.StatusBadRequest,
				httpapi.Fail(err.Error(), errCodeHandbookTemplateNoSteps),
			)
		default:
			ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		}
		return
	}
	ctx.JSON(http.StatusOK, httpapi.OK(toHandbookTemplateResponse(res), "Template published"))
}

func (h *HandbookHandler) ListHandbookTemplatesByDepartment(ctx *gin.Context) {
	deptID, err := uuid.Parse(ctx.Param("department_id"))
	if err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			httpapi.Fail("invalid department_id", errCodeInvalidRequest),
		)
		return
	}
	res, err := h.service.ListTemplatesByDepartment(ctx.Request.Context(), deptID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail(err.Error(), ""))
		return
	}
	results := make([]handbookTemplateResponse, len(res))
	for i := range res {
		item := res[i]
		results[i] = toHandbookTemplateResponse(&item)
	}
	ctx.JSON(http.StatusOK, httpapi.OK(results, "Templates retrieved"))
}

func (h *HandbookHandler) CreateHandbookStep(ctx *gin.Context) {
	var req createHandbookStepRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid request body", errCodeInvalidRequest))
		return
	}
	res, err := h.service.CreateStep(ctx.Request.Context(), toCreateHandbookStepParams(req))
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTemplateNotDraft):
			ctx.JSON(
				http.StatusBadRequest,
				httpapi.Fail(err.Error(), errCodeHandbookTemplateNotDraft),
			)
		case strings.Contains(strings.ToLower(err.Error()), "invalid url"):
			ctx.JSON(
				http.StatusBadRequest,
				httpapi.Fail(err.Error(), errCodeHandbookLinkURLInvalid),
			)
		case strings.Contains(strings.ToLower(err.Error()), "invalid step content"):
			ctx.JSON(
				http.StatusBadRequest,
				httpapi.Fail(err.Error(), errCodeHandbookQuizContentInvalid),
			)
		default:
			ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		}
		return
	}
	ctx.JSON(http.StatusCreated, httpapi.OK(toHandbookStepResponse(res), "Step created"))
}

func (h *HandbookHandler) UpdateHandbookStep(ctx *gin.Context) {
	stepID, err := uuid.Parse(ctx.Param("step_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid step_id", errCodeInvalidRequest))
		return
	}

	var req updateHandbookStepRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid request body", errCodeInvalidRequest))
		return
	}

	updateReq := domain.UpdateStepParams{StepID: stepID}
	if req.Title != nil {
		updateReq.SetTitle = true
		if string(*req.Title) != "null" {
			var v string
			if err := json.Unmarshal(*req.Title, &v); err != nil {
				ctx.JSON(
					http.StatusBadRequest,
					httpapi.Fail("title must be a string", errCodeInvalidRequest),
				)
				return
			}
			updateReq.Title = &v
		}
	}
	if req.Body != nil {
		updateReq.SetBody = true
		if string(*req.Body) != "null" {
			var v string
			if err := json.Unmarshal(*req.Body, &v); err != nil {
				ctx.JSON(
					http.StatusBadRequest,
					httpapi.Fail("body must be a string or null", errCodeInvalidRequest),
				)
				return
			}
			updateReq.Body = &v
		}
	}
	if req.Content != nil {
		updateReq.ContentProvided = true
		if string(*req.Content) != "null" {
			updateReq.Content = *req.Content
		}
	}
	if req.IsRequired != nil {
		updateReq.SetIsRequired = true
		var v bool
		if err := json.Unmarshal(*req.IsRequired, &v); err != nil {
			ctx.JSON(
				http.StatusBadRequest,
				httpapi.Fail("is_required must be a boolean", errCodeInvalidRequest),
			)
			return
		}
		updateReq.IsRequired = &v
	}

	res, err := h.service.UpdateStep(ctx.Request.Context(), updateReq)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrStepNotFound):
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), errCodeHandbookStepNotFound))
		case errors.Is(err, domain.ErrTemplateNotDraft):
			ctx.JSON(
				http.StatusBadRequest,
				httpapi.Fail(err.Error(), errCodeHandbookTemplateNotDraft),
			)
		default:
			ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		}
		return
	}
	ctx.JSON(http.StatusOK, httpapi.OK(toHandbookStepResponse(res), "Step updated"))
}

func (h *HandbookHandler) DeleteHandbookStep(ctx *gin.Context) {
	stepID, err := uuid.Parse(ctx.Param("step_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid step_id", errCodeInvalidRequest))
		return
	}
	err = h.service.DeleteStep(ctx.Request.Context(), domain.DeleteStepParams{StepID: stepID})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrStepNotFound):
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), errCodeHandbookStepNotFound))
		case errors.Is(err, domain.ErrTemplateNotDraft):
			ctx.JSON(
				http.StatusBadRequest,
				httpapi.Fail(err.Error(), errCodeHandbookTemplateNotDraft),
			)
		default:
			ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		}
		return
	}
	ctx.JSON(http.StatusOK, httpapi.OK(struct {
		StepID  uuid.UUID `json:"step_id"`
		Deleted bool      `json:"deleted"`
	}{StepID: stepID, Deleted: true}, "Step deleted"))
}

func (h *HandbookHandler) ReorderHandbookSteps(ctx *gin.Context) {
	templateID, err := uuid.Parse(ctx.Param("template_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid template_id", errCodeInvalidRequest))
		return
	}
	var req reorderHandbookStepsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid request body", errCodeInvalidRequest))
		return
	}
	res, err := h.service.ReorderTemplateSteps(ctx.Request.Context(), domain.ReorderStepsParams{
		TemplateID:     templateID,
		OrderedStepIDs: req.OrderedStepIDs,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidStepReorder) {
			ctx.JSON(
				http.StatusBadRequest,
				httpapi.Fail(err.Error(), errCodeHandbookStepReorderMismatch),
			)
			return
		}
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	results := make([]handbookStepResponse, len(res))
	for i := range res {
		item := res[i]
		results[i] = toHandbookStepResponse(&item)
	}
	ctx.JSON(http.StatusOK, httpapi.OK(results, "Steps reordered"))
}

func (h *HandbookHandler) ListHandbookStepsByTemplate(ctx *gin.Context) {
	templateID, err := uuid.Parse(ctx.Param("template_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid template_id", errCodeInvalidRequest))
		return
	}
	res, err := h.service.ListStepsByTemplate(ctx.Request.Context(), templateID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail(err.Error(), ""))
		return
	}
	results := make([]handbookStepResponse, len(res))
	for i := range res {
		item := res[i]
		results[i] = toHandbookStepResponse(&item)
	}
	ctx.JSON(http.StatusOK, httpapi.OK(results, "Steps retrieved"))
}

func (h *HandbookHandler) AssignHandbookTemplateToEmployee(ctx *gin.Context) {
	actor := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if actor == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}
	var req assignHandbookTemplateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid request body", errCodeInvalidRequest))
		return
	}
	res, err := h.service.AssignTemplateToEmployee(
		ctx.Request.Context(),
		actor,
		toAssignHandbookTemplateParams(req),
	)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	ctx.JSON(
		http.StatusCreated,
		httpapi.OK(toEmployeeHandbookAssignmentResponse(res), "Handbook assigned"),
	)
}

func (h *HandbookHandler) ListEmployeeHandbookHistory(ctx *gin.Context) {
	employeeID, err := uuid.Parse(ctx.Param("employee_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee_id", errCodeInvalidRequest))
		return
	}
	res, err := h.service.ListEmployeeHandbookHistory(ctx.Request.Context(), employeeID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail(err.Error(), ""))
		return
	}
	results := make([]handbookAssignmentHistoryEntryResponse, len(res))
	for i := range res {
		results[i] = toHandbookAssignmentHistoryEntryResponse(res[i])
	}
	ctx.JSON(http.StatusOK, httpapi.OK(results, "History retrieved"))
}

func (h *HandbookHandler) ListEmployeeHandbookAssignments(ctx *gin.Context) {
	var req listEmployeeHandbookAssignmentsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), errCodeInvalidRequest))
		return
	}
	res, err := h.service.ListEmployeeHandbookAssignments(
		ctx.Request.Context(),
		toListEmployeeHandbookAssignmentsParams(req),
	)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	results := make([]employeeHandbookAssignmentSummaryResponse, len(res.Items))
	for i, item := range res.Items {
		results[i] = toEmployeeHandbookAssignmentSummaryResponse(item)
	}
	page := httpapi.NewPageResponse(ctx, req.PageRequest, results, res.TotalCount)
	ctx.JSON(http.StatusOK, httpapi.OK(page, "Assignments retrieved"))
}

func (h *HandbookHandler) ListEligibleEmployees(ctx *gin.Context) {
	actor := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if actor == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}
	var req listEligibleEmployeesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), errCodeInvalidRequest))
		return
	}
	res, err := h.service.ListEligibleEmployees(
		ctx.Request.Context(),
		actor,
		toListEligibleEmployeesParams(req),
	)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	results := make([]eligibleEmployeeResponse, len(res.Items))
	for i, item := range res.Items {
		results[i] = toEligibleEmployeeResponse(item)
	}
	page := httpapi.NewPageResponse(ctx, req.PageRequest, results, res.TotalCount)
	ctx.JSON(http.StatusOK, httpapi.OK(page, "Eligible employees retrieved"))
}

func (h *HandbookHandler) GetEmployeeHandbookDetails(ctx *gin.Context) {
	handbookID, err := uuid.Parse(ctx.Param("handbook_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid handbook_id", errCodeInvalidRequest))
		return
	}
	res, err := h.service.GetEmployeeHandbookDetails(ctx.Request.Context(), handbookID)
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeHandbookNotFound) {
			ctx.JSON(
				http.StatusNotFound,
				httpapi.Fail(err.Error(), errCodeEmployeeHandbookNotFound),
			)
			return
		}
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toEmployeeHandbookDetailsResponse(res), "Employee handbook details retrieved"),
	)
}

func (h *HandbookHandler) WaiveEmployeeHandbook(ctx *gin.Context) {
	actor := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if actor == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}
	handbookID, err := uuid.Parse(ctx.Param("handbook_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid handbook_id", errCodeInvalidRequest))
		return
	}
	var req waiveEmployeeHandbookRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid request body", errCodeInvalidRequest))
		return
	}
	res, err := h.service.WaiveEmployeeHandbook(
		ctx.Request.Context(),
		actor,
		domain.WaiveEmployeeHandbookParams{
			EmployeeHandbookID: handbookID,
			Reason:             req.Reason,
		},
	)
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeHandbookNotActive) {
			ctx.JSON(
				http.StatusBadRequest,
				httpapi.Fail(err.Error(), errCodeEmployeeHandbookNotActive),
			)
			return
		}
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toWaivedEmployeeHandbookResponse(res), "Employee handbook waived"),
	)
}
