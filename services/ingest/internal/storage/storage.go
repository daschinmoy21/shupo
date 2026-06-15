package storage

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStore struct {
	client *minio.Client
	bucket string
}

func NewMinioStore(endpoint, accessKey, secretKey, bucket string) (*MinioStore, error) {
	cli, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}
	return &MinioStore{client: cli, bucket: bucket}, nil
}

func (s *MinioStore) Put(ctx context.Context, key string, src io.Reader, size int64, contentType string) (int64, error) {
	info, err := s.client.PutObject(ctx, s.bucket, key, src, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return 0, err
	}
	return info.Size, nil
}
