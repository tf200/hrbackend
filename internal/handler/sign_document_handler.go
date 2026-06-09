package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"hrbackend/internal/domain"
	"hrbackend/internal/httpapi"
	"hrbackend/internal/middleware"
)

type SignDocumentHandler struct{ service domain.SignDocumentService }

func NewSignDocumentHandler(service domain.SignDocumentService) *SignDocumentHandler {
	return &SignDocumentHandler{service: service}
}

func (h *SignDocumentHandler) CreateDocument(ctx *gin.Context) {
	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}
	var req createSignDocumentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid request body", "INVALID_REQUEST"))
		return
	}
	doc, err := h.service.CreateDocument(
		ctx.Request.Context(),
		employeeID,
		toCreateSignDocumentParams(req),
	)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	ctx.JSON(http.StatusCreated, httpapi.OK(doc, "Sign document created"))
}

func (h *SignDocumentHandler) SetFields(ctx *gin.Context) {
	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}
	documentID, ok := parseUUIDParam(ctx, "document_id")
	if !ok {
		return
	}
	var req setSignDocumentFieldsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid request body", "INVALID_REQUEST"))
		return
	}
	fields, err := h.service.SetFields(
		ctx.Request.Context(),
		employeeID,
		documentID,
		toSignDocumentFieldParams(req.Fields),
	)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	ctx.JSON(http.StatusOK, httpapi.OK(fields, "Sign document fields saved"))
}

func (h *SignDocumentHandler) SendDocument(ctx *gin.Context) {
	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}
	documentID, ok := parseUUIDParam(ctx, "document_id")
	if !ok {
		return
	}
	doc, err := h.service.SendDocument(ctx.Request.Context(), employeeID, documentID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	ctx.JSON(http.StatusOK, httpapi.OK(doc, "Sign document sent"))
}
func (h *SignDocumentHandler) CancelDocument(ctx *gin.Context) {
	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}
	documentID, ok := parseUUIDParam(ctx, "document_id")
	if !ok {
		return
	}
	doc, err := h.service.CancelDocument(ctx.Request.Context(), employeeID, documentID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	ctx.JSON(http.StatusOK, httpapi.OK(doc, "Sign document cancelled"))
}
func (h *SignDocumentHandler) GetDocument(ctx *gin.Context) {
	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}
	documentID, ok := parseUUIDParam(ctx, "document_id")
	if !ok {
		return
	}
	doc, err := h.service.GetDocument(ctx.Request.Context(), employeeID, documentID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	ctx.JSON(http.StatusOK, httpapi.OK(doc, "Sign document retrieved"))
}
func (h *SignDocumentHandler) ListCreatedDocuments(ctx *gin.Context) {
	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}
	docs, err := h.service.ListMyCreatedDocuments(
		ctx.Request.Context(),
		employeeID,
		queryInt32(ctx, "limit", 50),
		queryInt32(ctx, "offset", 0),
	)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	ctx.JSON(http.StatusOK, httpapi.OK(docs, "Sign documents retrieved"))
}
func (h *SignDocumentHandler) ListMySigningDocuments(ctx *gin.Context) {
	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}
	docs, err := h.service.ListMySigningDocuments(
		ctx.Request.Context(),
		employeeID,
		queryInt32(ctx, "limit", 50),
		queryInt32(ctx, "offset", 0),
	)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	ctx.JSON(http.StatusOK, httpapi.OK(docs, "Signing documents retrieved"))
}
func (h *SignDocumentHandler) GetMySigningDocument(ctx *gin.Context) {
	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}
	documentID, ok := parseUUIDParam(ctx, "document_id")
	if !ok {
		return
	}
	doc, err := h.service.GetMySigningDocument(ctx.Request.Context(), employeeID, documentID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	ctx.JSON(http.StatusOK, httpapi.OK(doc, "Signing document retrieved"))
}
func (h *SignDocumentHandler) MarkViewed(ctx *gin.Context) {
	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}
	documentID, ok := parseUUIDParam(ctx, "document_id")
	if !ok {
		return
	}
	ip, ua := requestIPAndUserAgent(ctx)
	recipient, err := h.service.MarkViewed(ctx.Request.Context(), employeeID, documentID, ip, ua)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	ctx.JSON(http.StatusOK, httpapi.OK(recipient, "Signing document viewed"))
}
func (h *SignDocumentHandler) Sign(ctx *gin.Context) {
	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}
	documentID, ok := parseUUIDParam(ctx, "document_id")
	if !ok {
		return
	}
	var req signSignDocumentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid request body", "INVALID_REQUEST"))
		return
	}
	ip, ua := requestIPAndUserAgent(ctx)
	doc, err := h.service.Sign(
		ctx.Request.Context(),
		employeeID,
		toSignDocumentSignParams(documentID, req, ip, ua),
	)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	ctx.JSON(http.StatusOK, httpapi.OK(doc, "Document signed"))
}
func (h *SignDocumentHandler) GetSourceURL(ctx *gin.Context) { h.getURL(ctx, false) }
func (h *SignDocumentHandler) GetSignedURL(ctx *gin.Context) { h.getURL(ctx, true) }

func (h *SignDocumentHandler) getURL(ctx *gin.Context, signed bool) {
	employeeID := middleware.EmployeeIDFromContext(ctx.Request.Context())
	if employeeID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, httpapi.Fail("unauthorized", ""))
		return
	}
	documentID, ok := parseUUIDParam(ctx, "document_id")
	if !ok {
		return
	}
	var url string
	var err error
	if signed {
		url, err = h.service.GetSignedURL(ctx.Request.Context(), employeeID, documentID)
	} else {
		url, err = h.service.GetSourceURL(ctx.Request.Context(), employeeID, documentID)
	}
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail(err.Error(), ""))
		return
	}
	ctx.JSON(http.StatusOK, httpapi.OK(signDocumentURLResponse{URL: url}, "URL generated"))
}

func parseUUIDParam(ctx *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(ctx.Param(name))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpapi.Fail("invalid "+name, "INVALID_REQUEST"))
		return uuid.Nil, false
	}
	return id, true
}
func queryInt32(ctx *gin.Context, name string, fallback int32) int32 {
	value, err := strconv.ParseInt(ctx.Query(name), 10, 32)
	if err != nil {
		return fallback
	}
	return int32(value)
}
func requestIPAndUserAgent(ctx *gin.Context) (*string, *string) {
	ip := ctx.ClientIP()
	ua := ctx.Request.UserAgent()
	return &ip, &ua
}
