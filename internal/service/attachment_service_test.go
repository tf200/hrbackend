package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"hrbackend/internal/domain"
)

type fakeAttachmentRepository struct {
	attachments map[uuid.UUID]*domain.Attachment
	createErr   error
	getErr      error
	updateErr   error
	deleteErr   error
}

func (f *fakeAttachmentRepository) CreateAttachment(
	ctx context.Context,
	params domain.CreateAttachmentParams,
) (*domain.Attachment, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	id := uuid.New()
	tag := params.Tag
	att := &domain.Attachment{
		UUID:    id,
		Name:    params.Name,
		File:    params.File,
		Size:    params.Size,
		IsUsed:  params.IsUsed,
		Tag:     tag,
		Updated: time.Now(),
		Created: time.Now(),
	}
	if f.attachments == nil {
		f.attachments = make(map[uuid.UUID]*domain.Attachment)
	}
	f.attachments[id] = att
	return att, nil
}

func (f *fakeAttachmentRepository) GetAttachment(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Attachment, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	att, ok := f.attachments[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return att, nil
}

func (f *fakeAttachmentRepository) UpdateAttachmentUsed(
	ctx context.Context,
	params domain.UpdateAttachmentUsedParams,
) (*domain.Attachment, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	att, ok := f.attachments[params.UUID]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	att.IsUsed = params.IsUsed
	att.Updated = time.Now()
	return att, nil
}

func (f *fakeAttachmentRepository) DeleteAttachment(ctx context.Context, id uuid.UUID) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	delete(f.attachments, id)
	return nil
}

type fakeStorage struct {
	uploadErr    error
	urlErr       error
	uploadUrlErr error
	infoErr      error
	deleteErr    error
}

func (f *fakeStorage) Upload(
	ctx context.Context,
	file multipart.File,
	filename string,
	contentType string,
) (string, int64, error) {
	return "", 0, f.uploadErr
}

func (f *fakeStorage) GeneratePresignedURL(
	ctx context.Context,
	objectKey string,
	expiry time.Duration,
) (string, error) {
	return "", f.urlErr
}

func (f *fakeStorage) GeneratePresignedUploadURL(
	ctx context.Context,
	objectKey string,
	expiry time.Duration,
) (string, error) {
	if f.uploadUrlErr != nil {
		return "", f.uploadUrlErr
	}
	return "https://example.com/upload/" + objectKey, nil
}

func (f *fakeStorage) GetFileInfo(ctx context.Context, objectKey string) (int64, error) {
	return 0, f.infoErr
}

func (f *fakeStorage) GetFileInfos(
	ctx context.Context,
	objectKeys []string,
) (map[string]int64, error) {
	return nil, nil
}

func (f *fakeStorage) Download(ctx context.Context, objectKey string) ([]byte, error) {
	return nil, nil
}

func (f *fakeStorage) Delete(ctx context.Context, objectKey string) error {
	return f.deleteErr
}

type fakeLogger struct{}

func (f *fakeLogger) LogError(
	ctx context.Context,
	operation, message string,
	err error,
	fields ...zap.Field,
) {
}
func (f *fakeLogger) LogWarn(ctx context.Context, operation, message string, fields ...zap.Field) {}
func (f *fakeLogger) LogInfo(ctx context.Context, operation, message string, fields ...zap.Field) {}

func TestAttachmentService_RequestUploadURL_Success(t *testing.T) {
	repo := &fakeAttachmentRepository{attachments: make(map[uuid.UUID]*domain.Attachment)}
	storage := &fakeStorage{}
	logger := &fakeLogger{}

	svc := NewAttachmentService(repo, storage, logger)

	tag := "resume"
	resp, err := svc.RequestUploadURL(context.Background(), "resume.pdf", 5000, &tag)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.AttachmentID == uuid.Nil {
		t.Errorf("expected non-nil attachment ID")
	}

	if !strings.Contains(resp.UploadURL, resp.FileKey) {
		t.Errorf(
			"expected upload URL to contain file key %s, got: %s",
			resp.FileKey,
			resp.UploadURL,
		)
	}

	if !strings.HasPrefix(resp.FileKey, "attachments/") {
		t.Errorf("expected file key to start with 'attachments/', got: %s", resp.FileKey)
	}

	// Verify that the record was created in the fake repo
	createdAtt, exists := repo.attachments[resp.AttachmentID]
	if !exists {
		t.Fatalf("attachment record was not saved in repo")
	}

	if createdAtt.Name != "resume.pdf" {
		t.Errorf("expected attachment name 'resume.pdf', got '%s'", createdAtt.Name)
	}

	if createdAtt.IsUsed {
		t.Errorf("expected attachment IsUsed to be false, got true")
	}

	if createdAtt.Tag == nil || *createdAtt.Tag != "resume" {
		t.Errorf("expected tag 'resume', got '%v'", createdAtt.Tag)
	}
}

func TestAttachmentService_RequestUploadURL_ValidationErrors(t *testing.T) {
	repo := &fakeAttachmentRepository{}
	storage := &fakeStorage{}
	logger := &fakeLogger{}

	svc := NewAttachmentService(repo, storage, logger)

	t.Run("empty filename", func(t *testing.T) {
		_, err := svc.RequestUploadURL(context.Background(), "", 100, nil)
		if err == nil {
			t.Errorf("expected error for empty filename")
		}
	})

	t.Run("zero size", func(t *testing.T) {
		_, err := svc.RequestUploadURL(context.Background(), "test.pdf", 0, nil)
		if err == nil {
			t.Errorf("expected error for zero size")
		}
	})

	t.Run("negative size", func(t *testing.T) {
		_, err := svc.RequestUploadURL(context.Background(), "test.pdf", -50, nil)
		if err == nil {
			t.Errorf("expected error for negative size")
		}
	})
}

func TestAttachmentService_RequestUploadURL_CreateAttachmentError(t *testing.T) {
	expectedErr := errors.New("database connection down")
	repo := &fakeAttachmentRepository{createErr: expectedErr}
	storage := &fakeStorage{}
	logger := &fakeLogger{}

	svc := NewAttachmentService(repo, storage, logger)

	_, err := svc.RequestUploadURL(context.Background(), "test.pdf", 100, nil)
	if err == nil {
		t.Fatalf("expected error from repository create, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}
}

func TestAttachmentService_RequestUploadURL_PresignedURLError(t *testing.T) {
	expectedErr := errors.New("S3 region configuration invalid")
	repo := &fakeAttachmentRepository{attachments: make(map[uuid.UUID]*domain.Attachment)}
	storage := &fakeStorage{uploadUrlErr: expectedErr}
	logger := &fakeLogger{}

	svc := NewAttachmentService(repo, storage, logger)

	_, err := svc.RequestUploadURL(context.Background(), "test.pdf", 100, nil)
	if err == nil {
		t.Fatalf("expected error from S3 upload url generation, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}

	// Verify that the record was cleaned up from the repo
	if len(repo.attachments) != 0 {
		t.Errorf("expected attachment record to be cleaned up on S3 failure, but it remained")
	}
}
