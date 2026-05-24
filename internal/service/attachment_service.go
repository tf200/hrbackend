package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"hrbackend/internal/domain"
	"hrbackend/pkg/bucket"
)

type attachmentService struct {
	repo          domain.AttachmentRepository
	storageClient domain.Storage
	logger        domain.Logger
}

func NewAttachmentService(
	repo domain.AttachmentRepository,
	storageClient domain.Storage,
	logger domain.Logger,
) domain.AttachmentService {
	return &attachmentService{
		repo:          repo,
		storageClient: storageClient,
		logger:        logger,
	}
}

func (s *attachmentService) RequestUploadURL(
	ctx context.Context,
	filename string,
	size int64,
	tag *string,
) (*domain.UploadURLResponse, error) {
	if filename == "" {
		return nil, fmt.Errorf("filename cannot be empty")
	}
	if size <= 0 {
		return nil, fmt.Errorf("invalid file size")
	}

	// Generate a unique object key/filename
	objectKey := bucket.GenerateUniqueFilename(filename)

	// Store files under "attachments/" folder prefix in the bucket
	fileKey := fmt.Sprintf("attachments/%s", objectKey)

	// Save metadata in database with is_used = false
	attachment, err := s.repo.CreateAttachment(ctx, domain.CreateAttachmentParams{
		Name:   filename,
		File:   fileKey,
		Size:   int32(size),
		IsUsed: false,
		Tag:    tag,
	})
	if err != nil {
		s.logger.LogError(ctx, "AttachmentService.RequestUploadURL", "failed to create attachment record", err)
		return nil, fmt.Errorf("failed to create attachment record: %w", err)
	}

	// Generate the presigned upload URL (valid for 15 minutes)
	expiry := 15 * time.Minute
	uploadURL, err := s.storageClient.GeneratePresignedUploadURL(ctx, fileKey, expiry)
	if err != nil {
		s.logger.LogError(ctx, "AttachmentService.RequestUploadURL", "failed to generate presigned upload URL", err)
		// Clean up the created attachment record since we couldn't generate the upload URL
		_ = s.repo.DeleteAttachment(ctx, attachment.UUID)
		return nil, fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	return &domain.UploadURLResponse{
		AttachmentID: attachment.UUID,
		UploadURL:    uploadURL,
		FileKey:      fileKey,
	}, nil
}

func (s *attachmentService) GetAttachment(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Attachment, error) {
	return s.repo.GetAttachment(ctx, id)
}
