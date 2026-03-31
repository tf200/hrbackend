package bucket

import (
	"context"
	"errors"
	"mime/multipart"
	"time"
)

var ErrBucketDisabled = errors.New("object storage disabled")

type NoopClient struct{}

func NewNoop() *NoopClient {
	return &NoopClient{}
}

func (n *NoopClient) Upload(ctx context.Context, file multipart.File, filename string, contentType string) (string, int64, error) {
	return "", 0, ErrBucketDisabled
}

func (n *NoopClient) GeneratePresignedURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	return "", ErrBucketDisabled
}

func (n *NoopClient) GeneratePresignedUploadURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	return "", ErrBucketDisabled
}

func (n *NoopClient) GetFileInfo(ctx context.Context, objectKey string) (int64, error) {
	return 0, ErrBucketDisabled
}

func (n *NoopClient) GetFileInfos(ctx context.Context, objectKeys []string) (map[string]int64, error) {
	return nil, ErrBucketDisabled
}

func (n *NoopClient) Delete(ctx context.Context, objectKey string) error {
	return ErrBucketDisabled
}
