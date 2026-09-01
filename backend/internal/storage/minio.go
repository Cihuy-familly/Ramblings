package storage

import (
	"context"
	"io"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOStorage wraps a MinIO client with a predefined bucket.
type MinIOStorage struct {
	client *minio.Client
	bucket string
}

// NewMinIOClient creates a new MinIO client and ensures the bucket exists.
func NewMinIOClient(endpoint, accessKey, secretKey, bucket string, useSSL bool) *MinIOStorage {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalf("Failed to create MinIO client: %v", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		log.Fatalf("Failed to check MinIO bucket: %v", err)
	}
	if !exists {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatalf("Failed to create MinIO bucket: %v", err)
		}
		log.Printf("MinIO bucket %q created", bucket)
	}

	log.Println("MinIO connection established")
	return &MinIOStorage{client: client, bucket: bucket}
}

// Upload stores a file in MinIO.
func (s *MinIOStorage) Upload(ctx context.Context, file io.Reader, filename, contentType string, size int64) error {
	_, err := s.client.PutObject(ctx, s.bucket, filename, file, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

// Get retrieves a file from MinIO.
func (s *MinIOStorage) Get(ctx context.Context, filename string) (*minio.Object, error) {
	return s.client.GetObject(ctx, s.bucket, filename, minio.GetObjectOptions{})
}