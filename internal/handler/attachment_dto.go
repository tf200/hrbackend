package handler

import (
	"github.com/google/uuid"
	"hrbackend/internal/domain"
)

type requestUploadURLRequest struct {
	Filename string  `json:"filename" binding:"required"`
	Size     int64   `json:"size"     binding:"required,gt=0"`
	Tag      *string `json:"tag"      binding:"omitempty"`
}

type uploadURLResponse struct {
	AttachmentID uuid.UUID `json:"attachment_id"`
	UploadURL    string    `json:"upload_url"`
	FileKey      string    `json:"file_key"`
}

func toUploadURLResponse(resp *domain.UploadURLResponse) uploadURLResponse {
	return uploadURLResponse{
		AttachmentID: resp.AttachmentID,
		UploadURL:    resp.UploadURL,
		FileKey:      resp.FileKey,
	}
}
