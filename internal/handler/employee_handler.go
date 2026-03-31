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
	if _, err := parseDatePtr(req.ContractStartDate); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	if _, err := parseDatePtr(req.ContractEndDate); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employee, err := h.service.CreateEmployee(ctx.Request.Context(), toCreateEmployeeParams(req))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to create employee", ""))
		return
	}

	ctx.JSON(http.StatusCreated, httpapi.OK(toEmployeeDetailResponse(employee), "Employee created successfully"))
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

	ctx.JSON(http.StatusOK, httpapi.OK(httpapi.NewPageResponse(ctx, req.PageRequest, results, page.TotalCount), "Employees retrieved successfully"))
}

func (h *EmployeeHandler) GetEmployeeCounts(ctx *gin.Context) {
	counts, err := h.service.GetEmployeeCounts(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to get employee counts", ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toEmployeeCountsResponse(counts), "Employee counts retrieved successfully"))
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

	ctx.JSON(http.StatusOK, httpapi.OK(toEmployeeDetailResponse(employee), "Employee retrieved successfully"))
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

	employee, err := h.service.UpdateEmployee(ctx.Request.Context(), id, toUpdateEmployeeParams(req))
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to update employee", ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toEmployeeDetailResponse(employee), "Employee updated successfully"))
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

	ctx.JSON(http.StatusOK, httpapi.OK(toEmployeeProfileResponse(profile), "Employee profile retrieved successfully"))
}

func (h *EmployeeHandler) UpdateIsSubcontractor(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}

	var req updateIsSubcontractorRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employee, err := h.service.UpdateIsSubcontractor(ctx.Request.Context(), id, domain.UpdateIsSubcontractorParams{IsSubcontractor: *req.IsSubcontractor})
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to update is subcontractor", ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toEmployeeDetailResponse(employee), "Employee subcontractor status updated successfully"))
}

func (h *EmployeeHandler) AddContractDetails(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}

	var req addContractDetailsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	if _, err := parseDate(*req.ContractStartDate); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	if _, err := parseDate(*req.ContractEndDate); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	employee, err := h.service.AddContractDetails(ctx.Request.Context(), id, toAddContractDetailsParams(req))
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to add contract details", ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toEmployeeDetailResponse(employee), "Contract details updated successfully"))
}

func (h *EmployeeHandler) GetContractDetails(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}

	details, err := h.service.GetContractDetails(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to get contract details", ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toContractDetailsResponse(details), "Contract details retrieved successfully"))
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

	education, err := h.service.AddEducation(ctx.Request.Context(), employeeID, toCreateEducationParams(req))
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to add education", ""))
		return
	}

	ctx.JSON(http.StatusCreated, httpapi.OK(toEducationResponse(education), "Education added successfully"))
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

	education, err := h.service.UpdateEducation(ctx.Request.Context(), educationID, toUpdateEducationParams(req))
	if err != nil {
		if errors.Is(err, domain.ErrEducationNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to update education", ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toEducationResponse(education), "Education updated successfully"))
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

	ctx.JSON(http.StatusOK, httpapi.OK(toEducationResponse(education), "Education deleted successfully"))
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

	experience, err := h.service.AddExperience(ctx.Request.Context(), employeeID, toCreateExperienceParams(req))
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to add experience", ""))
		return
	}

	ctx.JSON(http.StatusCreated, httpapi.OK(toExperienceResponse(experience), "Experience added successfully"))
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

	experience, err := h.service.UpdateExperience(ctx.Request.Context(), experienceID, toUpdateExperienceParams(req))
	if err != nil {
		if errors.Is(err, domain.ErrExperienceNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to update experience", ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toExperienceResponse(experience), "Experience updated successfully"))
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

	ctx.JSON(http.StatusOK, httpapi.OK(toExperienceResponse(experience), "Experience deleted successfully"))
}

func (h *EmployeeHandler) AddCertification(ctx *gin.Context) {
	employeeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}

	var req createCertificationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	if _, err := parseDate(req.DateIssued); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	certification, err := h.service.AddCertification(ctx.Request.Context(), employeeID, toCreateCertificationParams(req))
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to add certification", ""))
		return
	}

	ctx.JSON(http.StatusCreated, httpapi.OK(toCertificationResponse(certification), "Certification added successfully"))
}

func (h *EmployeeHandler) ListCertification(ctx *gin.Context) {
	employeeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}

	certifications, err := h.service.ListCertification(ctx.Request.Context(), employeeID)
	if err != nil {
		if errors.Is(err, domain.ErrEmployeeNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to list certification", ""))
		return
	}

	response := make([]certificationResponse, len(certifications))
	for i := range certifications {
		response[i] = toCertificationResponse(&certifications[i])
	}

	ctx.JSON(http.StatusOK, httpapi.OK(response, "Certification retrieved successfully"))
}

func (h *EmployeeHandler) UpdateCertification(ctx *gin.Context) {
	employeeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}
	_ = employeeID

	certificationID, err := uuid.Parse(ctx.Param("certification_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid certification ID", ""))
		return
	}

	var req updateCertificationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	if _, err := parseDatePtr(req.DateIssued); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}

	certification, err := h.service.UpdateCertification(ctx.Request.Context(), certificationID, toUpdateCertificationParams(req))
	if err != nil {
		if errors.Is(err, domain.ErrCertificationNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to update certification", ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toCertificationResponse(certification), "Certification updated successfully"))
}

func (h *EmployeeHandler) DeleteCertification(ctx *gin.Context) {
	employeeID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid employee ID", ""))
		return
	}
	_ = employeeID

	certificationID, err := uuid.Parse(ctx.Param("certification_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid certification ID", ""))
		return
	}

	certification, err := h.service.DeleteCertification(ctx.Request.Context(), certificationID)
	if err != nil {
		if errors.Is(err, domain.ErrCertificationNotFound) {
			ctx.JSON(http.StatusNotFound, httpapi.Fail(err.Error(), ""))
			return
		}
		ctx.JSON(http.StatusInternalServerError, httpapi.Fail("failed to delete certification", ""))
		return
	}

	ctx.JSON(http.StatusOK, httpapi.OK(toCertificationResponse(certification), "Certification deleted successfully"))
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
