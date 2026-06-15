package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"hrbackend/internal/domain"
	"hrbackend/pkg/bucket"
)

const (
	signatureUploadMaxSize = 2 << 20
	signatureURLTTL        = 15 * time.Minute
)

type employeeSignatureProfileService struct {
	repo    domain.EmployeeSignatureProfileRepository
	storage domain.Storage
	logger  domain.Logger
}

func NewEmployeeSignatureProfileService(
	repo domain.EmployeeSignatureProfileRepository,
	storage domain.Storage,
	logger domain.Logger,
) domain.EmployeeSignatureProfileService {
	return &employeeSignatureProfileService{repo: repo, storage: storage, logger: logger}
}

func (s *employeeSignatureProfileService) GetMySignatureProfile(
	ctx context.Context,
	employeeID uuid.UUID,
) (*domain.EmployeeSignatureProfile, error) {
	if employeeID == uuid.Nil {
		return nil, domain.ErrEmployeeSignatureProfileInvalid
	}
	return s.repo.GetByEmployeeID(ctx, employeeID)
}

func (s *employeeSignatureProfileService) UpsertMySignatureProfile(
	ctx context.Context,
	employeeID uuid.UUID,
	params domain.UpsertEmployeeSignatureProfileParams,
) (*domain.EmployeeSignatureProfile, error) {
	if employeeID == uuid.Nil {
		return nil, domain.ErrEmployeeSignatureProfileInvalid
	}
	params.Type = strings.ToLower(strings.TrimSpace(params.Type))
	if params.TypedName != nil {
		trimmed := strings.TrimSpace(*params.TypedName)
		params.TypedName = &trimmed
	}
	if params.ImageFileKey != nil {
		trimmed := strings.TrimSpace(*params.ImageFileKey)
		params.ImageFileKey = &trimmed
	}
	if !validSignatureProfile(employeeID, params) {
		return nil, domain.ErrEmployeeSignatureProfileInvalid
	}
	return s.repo.Upsert(ctx, employeeID, params)
}

func (s *employeeSignatureProfileService) DeleteMySignatureProfile(
	ctx context.Context,
	employeeID uuid.UUID,
) error {
	if employeeID == uuid.Nil {
		return domain.ErrEmployeeSignatureProfileInvalid
	}
	return s.repo.DeleteByEmployeeID(ctx, employeeID)
}

func (s *employeeSignatureProfileService) RequestUploadURL(
	ctx context.Context,
	employeeID uuid.UUID,
	params domain.RequestSignatureUploadURLParams,
) (*domain.SignatureUploadURLResponse, error) {
	if employeeID == uuid.Nil || strings.TrimSpace(params.Filename) == "" || params.Size <= 0 ||
		params.Size > signatureUploadMaxSize || !allowedSignatureContentType(params.ContentType) {
		return nil, domain.ErrEmployeeSignatureProfileInvalid
	}
	fileKey := fmt.Sprintf(
		"employee-signatures/%s/%s",
		employeeID.String(),
		bucket.GenerateUniqueFilename(params.Filename),
	)
	url, err := s.storage.GeneratePresignedUploadURL(ctx, fileKey, signatureURLTTL)
	if err != nil {
		s.logger.LogError(ctx, "EmployeeSignatureProfileService.RequestUploadURL", "failed to generate upload URL", err)
		return nil, err
	}
	return &domain.SignatureUploadURLResponse{UploadURL: url, FileKey: fileKey}, nil
}

func (s *employeeSignatureProfileService) GetSignatureImageURL(
	ctx context.Context,
	employeeID uuid.UUID,
) (string, error) {
	profile, err := s.GetMySignatureProfile(ctx, employeeID)
	if err != nil {
		return "", err
	}
	if profile.ImageFileKey == nil || strings.TrimSpace(*profile.ImageFileKey) == "" {
		return "", domain.ErrEmployeeSignatureProfileInvalid
	}
	return s.storage.GeneratePresignedURL(ctx, *profile.ImageFileKey, signatureURLTTL)
}

func validSignatureProfile(
	employeeID uuid.UUID,
	params domain.UpsertEmployeeSignatureProfileParams,
) bool {
	switch params.Type {
	case "typed":
		return params.TypedName != nil && *params.TypedName != ""
	case "drawn", "uploaded":
		return params.ImageFileKey != nil && validSignatureFileKey(employeeID, *params.ImageFileKey)
	default:
		return false
	}
}

func validSignatureFileKey(employeeID uuid.UUID, key string) bool {
	key = strings.TrimSpace(key)
	if !strings.HasPrefix(key, "employee-signatures/"+employeeID.String()+"/") {
		return false
	}
	switch strings.ToLower(filepath.Ext(key)) {
	case ".png", ".jpg", ".jpeg":
		return true
	default:
		return false
	}
}

func allowedSignatureContentType(contentType string) bool {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/png", "image/jpeg":
		return true
	default:
		return false
	}
}
