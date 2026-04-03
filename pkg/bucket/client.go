package bucket

import (
	"context"
	"fmt"
	"mime/multipart"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"golang.org/x/sync/errgroup"
)

type Config struct {
	Endpoint string
	KeyID    string
	Key      string
	Bucket   string
	Region   string
	Secure   bool
}

type ObjectStorageClient struct {
	client *minio.Client
	bucket string
}

func New(ctx context.Context, cfg Config) (*ObjectStorageClient, error) {
	region := cfg.Region
	if region == "" {
		region = "eu-central-003"
	}

	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.KeyID, cfg.Key, ""),
		Secure: cfg.Secure,
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	exists, err := minioClient.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check if bucket exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("bucket %s does not exist", cfg.Bucket)
	}

	return &ObjectStorageClient{
		client: minioClient,
		bucket: cfg.Bucket,
	}, nil
}

func (o *ObjectStorageClient) Upload(
	ctx context.Context,
	file multipart.File,
	filename string,
	contentType string,
) (string, int64, error) {
	uploadInfo, err := o.client.PutObject(ctx, o.bucket, filename, file, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", 0, fmt.Errorf("failed to upload file: %w", err)
	}

	return uploadInfo.Key, uploadInfo.Size, nil
}

func (o *ObjectStorageClient) GeneratePresignedURL(
	ctx context.Context,
	objectKey string,
	expiry time.Duration,
) (string, error) {
	presignedURL, err := o.client.PresignedGetObject(ctx, o.bucket, objectKey, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.String(), nil
}

func (o *ObjectStorageClient) GeneratePresignedUploadURL(
	ctx context.Context,
	objectKey string,
	expiry time.Duration,
) (string, error) {
	presignedURL, err := o.client.PresignedPutObject(ctx, o.bucket, objectKey, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	return presignedURL.String(), nil
}

func (o *ObjectStorageClient) GetFileInfo(ctx context.Context, objectKey string) (int64, error) {
	objInfo, err := o.client.StatObject(ctx, o.bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to get file info: %w", err)
	}

	return objInfo.Size, nil
}

func (o *ObjectStorageClient) GetFileInfos(
	ctx context.Context,
	objectKeys []string,
) (map[string]int64, error) {
	results := make(map[string]int64, len(objectKeys))
	if len(objectKeys) == 0 {
		return results, nil
	}

	var mu sync.Mutex
	grp, groupCtx := errgroup.WithContext(ctx)
	grp.SetLimit(8)

	for _, key := range objectKeys {
		key := key
		grp.Go(func() error {
			size, err := o.GetFileInfo(groupCtx, key)
			if err != nil {
				return err
			}

			mu.Lock()
			results[key] = size
			mu.Unlock()
			return nil
		})
	}

	if err := grp.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

func (o *ObjectStorageClient) Delete(ctx context.Context, objectKey string) error {
	if err := o.client.RemoveObject(
		ctx,
		o.bucket,
		objectKey,
		minio.RemoveObjectOptions{},
	); err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}
