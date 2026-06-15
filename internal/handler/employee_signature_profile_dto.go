package handler

import (
	"time"

	"github.com/google/uuid"

	"hrbackend/internal/domain"
)

type employeeSignatureProfileResponse struct {
	Exists    bool                      `json:"exists"`
	Signature *employeeSignatureProfile `json:"signature"`
}

type employeeSignatureProfile struct {
	ID           uuid.UUID `json:"id"`
	Type         string    `json:"type"`
	TypedName    *string   `json:"typed_name"`
	ImageFileKey *string   `json:"image_file_key"`
	ImageURL     *string   `json:"image_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type upsertEmployeeSignatureProfileRequest struct {
	Type         string  `json:"type" binding:"required"`
	TypedName    *string `json:"typed_name"`
	ImageFileKey *string `json:"image_file_key"`
}

type requestSignatureUploadURLRequest struct {
	Filename    string `json:"filename"     binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
	Size        int64  `json:"size"         binding:"required,gt=0"`
}

type signatureUploadURLResponse struct {
	UploadURL string `json:"upload_url"`
	FileKey   string `json:"file_key"`
}

type signatureImageURLResponse struct {
	URL string `json:"url"`
}

func toUpsertEmployeeSignatureProfileParams(
	req upsertEmployeeSignatureProfileRequest,
) domain.UpsertEmployeeSignatureProfileParams {
	return domain.UpsertEmployeeSignatureProfileParams{
		Type:         req.Type,
		TypedName:    req.TypedName,
		ImageFileKey: req.ImageFileKey,
	}
}

func toRequestSignatureUploadURLParams(
	req requestSignatureUploadURLRequest,
) domain.RequestSignatureUploadURLParams {
	return domain.RequestSignatureUploadURLParams{
		Filename:    req.Filename,
		ContentType: req.ContentType,
		Size:        req.Size,
	}
}

func toEmployeeSignatureProfileResponse(
	profile *domain.EmployeeSignatureProfile,
	imageURL *string,
) employeeSignatureProfileResponse {
	if profile == nil {
		return employeeSignatureProfileResponse{Exists: false}
	}
	return employeeSignatureProfileResponse{
		Exists: true,
		Signature: &employeeSignatureProfile{
			ID:           profile.ID,
			Type:         profile.Type,
			TypedName:    profile.TypedName,
			ImageFileKey: profile.ImageFileKey,
			ImageURL:     imageURL,
			CreatedAt:    profile.CreatedAt,
			UpdatedAt:    profile.UpdatedAt,
		},
	}
}

func toSignatureUploadURLResponse(
	resp *domain.SignatureUploadURLResponse,
) signatureUploadURLResponse {
	return signatureUploadURLResponse{UploadURL: resp.UploadURL, FileKey: resp.FileKey}
}
