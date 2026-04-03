package domain

import (
	"context"
	"mime/multipart"
	"time"
)

type Storage interface {
	Upload(
		ctx context.Context,
		file multipart.File,
		filename string,
		contentType string,
	) (string, int64, error)
	GeneratePresignedURL(
		ctx context.Context,
		objectKey string,
		expiry time.Duration,
	) (string, error)
	GeneratePresignedUploadURL(
		ctx context.Context,
		objectKey string,
		expiry time.Duration,
	) (string, error)
	GetFileInfo(ctx context.Context, objectKey string) (int64, error)
	GetFileInfos(ctx context.Context, objectKeys []string) (map[string]int64, error)
	Delete(ctx context.Context, objectKey string) error
}
