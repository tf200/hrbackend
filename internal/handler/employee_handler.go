package handler

import (
	"errors"
	"net/http"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"
	"hrbackend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EmployeeHandler struct {
	service domain.EmployeeService
}

func NewEmployeeHandler(service domain.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{service: service}
}

func (h *EmployeeHandler) CreateEmployee(ctx *gin.Context) {
	var req createEmployeeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	if _, err := parseDatePtr(req.DateOfBirth); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employee, err := h.service.CreateEmployee(ctx.Request.Context(), toCreateEmployeeParams(req))
	if err != nil {
		if errors.Is(err, domain.ErrContractChangeInvalid) {
			ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to create employee", ""))
		return
	}

	ctx.JSON(
		http.StatusCreated,
		httpapi.OK(toEmployeeDetailResponse(employee), "Employee created successfully"),
	)
}

func (h *EmployeeHandler) ListEmployee(ctx *gin.Context) {
	var req listEmployeesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	page, err := h.service.ListEmployees(ctx.Request.Context(), toListEmployeesParams(req))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to list employees", ""))
		return
	}

	results := make([]employeeListItemResponse, len(page.Items))
	for i, item := range page.Items {
		results[i] = toEmployeeListItemResponse(item)
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(
			httpapi.NewPageResponse(ctx, req.PageRequest, results, page.TotalCount),
			"Employees retrieved successfully",
		),
	)
}

func (h *EmployeeHandler) GetEmployeeCounts(ctx *gin.Context) {
	counts, err := h.service.GetEmployeeCounts(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to get employee counts", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toEmployeeCountsResponse(counts), "Employee counts retrieved successfully"),
	)
}

func (h *EmployeeHandler) GetEmployeeByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}

	payload, ok := middleware.AuthPayloadFromContext(ctx.Request.Context())
	if !ok || payload == nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	employee, err := h.service.GetEmployeeByID(ctx.Request.Context(), id, payload.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to get employee by id", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toEmployeeDetailResponse(employee), "Employee retrieved successfully"),
	)
}

func (h *EmployeeHandler) UpdateEmployee(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}

	var req updateEmployeeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	if _, err := parseDatePtr(req.DateOfBirth); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	if req.SalaryAssignment != nil {
		if _, err := parseDate(req.SalaryAssignment.EffectiveFrom); err != nil {
			ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
			return
		}
		if _, err := parseDatePtr(req.SalaryAssignment.EffectiveTo); err != nil {
			ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
			return
		}
	}

	employee, err := h.service.UpdateEmployee(
		ctx.Request.Context(),
		id,
		toUpdateEmployeeParams(req),
	)
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		if errors.Is(err, domain.ErrContractChangeInvalid) {
			ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to update employee", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toEmployeeDetailResponse(employee), "Employee updated successfully"),
	)
}

func (h *EmployeeHandler) ListEmployeeContracts(ctx *gin.Context) {
	employeeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}

	contracts, err := h.service.ListEmployeeContracts(ctx.Request.Context(), employeeID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to list contracts", ""))
		return
	}

	items := make([]employeeContractDetailResponse, len(contracts))
	for i, c := range contracts {
		items[i] = *toEmployeeContractDetailResponse(&c)
	}

	ctx.JSON(http.StatusOK, httpapi.OK(items, ""))
}

func (h *EmployeeHandler) CreateContract(ctx *gin.Context) {
	employeeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}

	var req createContractRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employee, err := h.service.CreateNewContract(
		ctx.Request.Context(),
		employeeID,
		toCreateNewContractParams(req),
	)
	if err != nil {
		if errors.Is(err, domain.ErrContractChangeInvalid) {
			ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to create contract", ""))
		return
	}

	ctx.JSON(
		http.StatusCreated,
		httpapi.OK(toEmployeeDetailResponse(employee), "Contract created successfully"),
	)
}

func (h *EmployeeHandler) UpdateContract(ctx *gin.Context) {
	employeeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}

	contractID, err := uuid.Parse(ctx.Param("contract_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid contract ID", ""))
		return
	}

	var req updateContractRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	if req.StartDate != nil {
		if _, err := parseDatePtr(req.StartDate); err != nil {
			ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
			return
		}
	}
	if req.ContractEndDate != nil {
		if _, err := parseDatePtr(req.ContractEndDate); err != nil {
			ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
			return
		}
	}

	contract, err := h.service.UpdateEmployeeContract(
		ctx.Request.Context(),
		employeeID,
		contractID,
		toUpdateEmployeeContractParams(req),
	)
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		if errors.Is(err, domain.ErrContractChangeInvalid) {
			ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to update contract", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toEmployeeContractDetailResponse(contract), "Contract updated successfully"),
	)
}

func (h *EmployeeHandler) CreateContractAmendment(ctx *gin.Context) {
	employeeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}

	contractID, err := uuid.Parse(ctx.Param("contract_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid contract ID", ""))
		return
	}

	var req createContractAmendmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employee, err := h.service.CreateContractAmendment(
		ctx.Request.Context(),
		employeeID,
		contractID,
		toCreateContractAmendmentParams(req),
	)
	if err != nil {
		if errors.Is(err, domain.ErrContractChangeInvalid) {
			ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to create contract amendment", ""))
		return
	}

	ctx.JSON(
		http.StatusCreated,
		httpapi.OK(toEmployeeDetailResponse(employee), "Contract amendment created successfully"),
	)
}

func (h *EmployeeHandler) GetEmployeeProfile(ctx *gin.Context) {
	payload, ok := middleware.AuthPayloadFromContext(ctx.Request.Context())
	if !ok || payload == nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}

	profile, err := h.service.GetEmployeeProfile(ctx.Request.Context(), payload.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to get employee profile", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toEmployeeProfileResponse(profile), "Employee profile retrieved successfully"),
	)
}

func (h *EmployeeHandler) AddEducation(ctx *gin.Context) {
	employeeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}

	var req createEducationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	if _, err := parseDate(req.StartDate); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	if _, err := parseDate(req.EndDate); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	education, err := h.service.AddEducation(
		ctx.Request.Context(),
		employeeID,
		toCreateEducationParams(req),
	)
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to add education", ""))
		return
	}

	ctx.JSON(
		http.StatusCreated,
		httpapi.OK(toEducationResponse(education), "Education added successfully"),
	)
}

func (h *EmployeeHandler) ListEducation(ctx *gin.Context) {
	employeeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}

	educationList, err := h.service.ListEducation(ctx.Request.Context(), employeeID)
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to list education", ""))
		return
	}

	response := make([]educationResponse, len(educationList))
	for i := range educationList {
		response[i] = toEducationResponse(&educationList[i])
	}

	ctx.JSON(http.StatusOK, httpapi.OK(response, "Education retrieved successfully"))
}

func (h *EmployeeHandler) UpdateEducation(ctx *gin.Context) {
	employeeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}
	_ = employeeID

	educationID, err := uuid.Parse(ctx.Param("education_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid education ID", ""))
		return
	}

	var req updateEducationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	if _, err := parseDatePtr(req.StartDate); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	if _, err := parseDatePtr(req.EndDate); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	education, err := h.service.UpdateEducation(
		ctx.Request.Context(),
		educationID,
		toUpdateEducationParams(req),
	)
	if err != nil {
		if errors.Is(err, domain.ErrEducationNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to update education", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toEducationResponse(education), "Education updated successfully"),
	)
}

func (h *EmployeeHandler) DeleteEducation(ctx *gin.Context) {
	employeeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}
	_ = employeeID

	educationID, err := uuid.Parse(ctx.Param("education_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid education ID", ""))
		return
	}

	education, err := h.service.DeleteEducation(ctx.Request.Context(), educationID)
	if err != nil {
		if errors.Is(err, domain.ErrEducationNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to delete education", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toEducationResponse(education), "Education deleted successfully"),
	)
}

func (h *EmployeeHandler) AddExperience(ctx *gin.Context) {
	employeeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}

	var req createExperienceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	if _, err := parseDate(req.StartDate); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	if _, err := parseDate(req.EndDate); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	experience, err := h.service.AddExperience(
		ctx.Request.Context(),
		employeeID,
		toCreateExperienceParams(req),
	)
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to add experience", ""))
		return
	}

	ctx.JSON(
		http.StatusCreated,
		httpapi.OK(toExperienceResponse(experience), "Experience added successfully"),
	)
}

func (h *EmployeeHandler) ListExperience(ctx *gin.Context) {
	employeeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}

	experiences, err := h.service.ListExperience(ctx.Request.Context(), employeeID)
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to list experience", ""))
		return
	}

	response := make([]experienceResponse, len(experiences))
	for i := range experiences {
		response[i] = toExperienceResponse(&experiences[i])
	}

	ctx.JSON(http.StatusOK, httpapi.OK(response, "Experience retrieved successfully"))
}

func (h *EmployeeHandler) UpdateExperience(ctx *gin.Context) {
	employeeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}
	_ = employeeID

	experienceID, err := uuid.Parse(ctx.Param("experience_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid experience ID", ""))
		return
	}

	var req updateExperienceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	if _, err := parseDatePtr(req.StartDate); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	if _, err := parseDatePtr(req.EndDate); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	experience, err := h.service.UpdateExperience(
		ctx.Request.Context(),
		experienceID,
		toUpdateExperienceParams(req),
	)
	if err != nil {
		if errors.Is(err, domain.ErrExperienceNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to update experience", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toExperienceResponse(experience), "Experience updated successfully"),
	)
}

func (h *EmployeeHandler) DeleteExperience(ctx *gin.Context) {
	employeeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}
	_ = employeeID

	experienceID, err := uuid.Parse(ctx.Param("experience_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid experience ID", ""))
		return
	}

	experience, err := h.service.DeleteExperience(ctx.Request.Context(), experienceID)
	if err != nil {
		if errors.Is(err, domain.ErrExperienceNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to delete experience", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toExperienceResponse(experience), "Experience deleted successfully"),
	)
}

func (h *EmployeeHandler) AddQualification(ctx *gin.Context) {
	employeeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}

	var req []createQualificationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	if len(req) == 0 {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("at least one qualification is required", ""))
		return
	}

	params, err := toCreateQualificationsParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	count, err := h.service.AddQualifications(
		ctx.Request.Context(),
		employeeID,
		params,
	)
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to add qualifications", ""))
		return
	}

	ctx.JSON(
		http.StatusCreated,
		httpapi.OK(count, "Qualifications added successfully"),
	)
}

func (h *EmployeeHandler) ListQualification(ctx *gin.Context) {
	employeeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}

	qualifications, err := h.service.ListQualifications(ctx.Request.Context(), employeeID)
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to list qualifications", ""))
		return
	}

	response := make([]qualificationResponse, len(qualifications))
	for i := range qualifications {
		response[i] = toQualificationResponse(&qualifications[i])
	}

	ctx.JSON(http.StatusOK, httpapi.OK(response, "Qualifications retrieved successfully"))
}

func (h *EmployeeHandler) UpdateQualification(ctx *gin.Context) {
	employeeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}
	_ = employeeID

	qualificationID, err := uuid.Parse(ctx.Param("qualification_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid qualification ID", ""))
		return
	}

	var req updateQualificationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	if req.AchievedOn != nil {
		if _, err := parseDate(*req.AchievedOn); err != nil {
			ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
			return
		}
	}
	if req.ExpirationDate != nil {
		if _, err := parseDate(*req.ExpirationDate); err != nil {
			ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
			return
		}
	}

	qualification, err := h.service.UpdateQualification(
		ctx.Request.Context(),
		qualificationID,
		toUpdateQualificationParams(req),
	)
	if err != nil {
		if errors.Is(err, domain.ErrQualificationNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to update qualification", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toQualificationResponse(qualification), "Qualification updated successfully"),
	)
}

func (h *EmployeeHandler) DeleteQualification(ctx *gin.Context) {
	employeeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}
	_ = employeeID

	qualificationID, err := uuid.Parse(ctx.Param("qualification_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid qualification ID", ""))
		return
	}

	qualification, err := h.service.DeleteQualification(ctx.Request.Context(), qualificationID)
	if err != nil {
		if errors.Is(err, domain.ErrQualificationNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to delete qualification", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toQualificationResponse(qualification), "Qualification deleted successfully"),
	)
}

func (h *EmployeeHandler) ResetPassword(ctx *gin.Context) {
	employeeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}

	var req resetPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	result, err := h.service.ResetPassword(ctx.Request.Context(), employeeID, toResetPasswordParams(req))
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidPasswordResetRequest):
			ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		case errors.Is(err, domain.ErrEmployeeNotFound):
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
		case errors.Is(err, domain.ErrPasswordHashFailed),
			errors.Is(err, domain.ErrEmailDeliveryFailed):
			ctx.JSON(http.StatusInternalServerError, httpapi.Fail(err.Error(), ""))
		default:
			ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to reset password", ""))
		}
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toResetPasswordResponse(result), "Password reset successfully"),
	)
}

func (h *EmployeeHandler) SearchEmployeesByNameOrEmail(ctx *gin.Context) {
	var req searchEmployeesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	results, err := h.service.SearchEmployeesByNameOrEmail(ctx.Request.Context(), req.Search)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to search employees", ""))
		return
	}

	response := make([]employeeSearchResultResponse, len(results))
	for i, result := range results {
		response[i] = toEmployeeSearchResultResponse(result)
	}

	ctx.JSON(http.StatusOK, httpapi.OK(response, "Employees retrieved successfully"))
}

func (h *EmployeeHandler) AddEmployeeAuthorization(ctx *gin.Context) {
	employeeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}

	var req []createEmployeeAuthorizationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid request payload", err.Error()))
		return
	}

	if len(req) == 0 {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("at least one authorization is required", ""))
		return
	}

	params, err := toCreateEmployeeAuthorizationsParams(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	count, err := h.service.AddEmployeeAuthorizations(
		ctx.Request.Context(),
		employeeID,
		params,
	)
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to add employee authorizations", ""))
		return
	}

	ctx.JSON(http.StatusCreated,
		httpapi.OK(count, "Employee authorizations added successfully"),
	)
}

func (h *EmployeeHandler) ListEmployeeAuthorizations(ctx *gin.Context) {
	employeeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}

	authorizations, err := h.service.ListEmployeeAuthorizations(ctx.Request.Context(), employeeID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to list employee authorizations", ""))
		return
	}

	response := make([]employeeAuthorizationResponse, len(authorizations))
	for i := range authorizations {
		response[i] = toEmployeeAuthorizationResponse(&authorizations[i])
	}

	ctx.JSON(http.StatusOK, httpapi.OK(response, "Employee authorizations retrieved successfully"))
}

func (h *EmployeeHandler) UpdateEmployeeAuthorization(ctx *gin.Context) {
	authID, err := uuid.Parse(ctx.Param("authorization_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee authorization ID", ""))
		return
	}

	var req updateEmployeeAuthorizationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid request payload", err.Error()))
		return
	}

	authRecord, err := h.service.UpdateEmployeeAuthorization(
		ctx.Request.Context(),
		authID,
		toUpdateEmployeeAuthorizationParams(req),
	)
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeAuthorizationNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to update employee authorization", ""))
		return
	}

	ctx.JSON(http.StatusOK,
		httpapi.OK(toEmployeeAuthorizationResponse(authRecord), "Employee authorization updated successfully"),
	)
}

func (h *EmployeeHandler) DeleteEmployeeAuthorization(ctx *gin.Context) {
	authID, err := uuid.Parse(ctx.Param("authorization_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee authorization ID", ""))
		return
	}

	authRecord, err := h.service.DeleteEmployeeAuthorization(ctx.Request.Context(), authID)
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeAuthorizationNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to delete employee authorization", ""))
		return
	}

	ctx.JSON(
		http.StatusOK,
		httpapi.OK(toEmployeeAuthorizationResponse(authRecord), "Employee authorization deleted successfully"),
	)
}
