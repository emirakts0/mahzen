package storage

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/emirakts0/mahzen/internal/config"
)

// NewS3Client creates an S3 client configured for the given S3/RustFS endpoint.
func NewS3Client(ctx context.Context, cfg config.S3Config) (*s3.Client, error) {
	slog.Info("s3 client connecting",
		"endpoint", cfg.Endpoint,
		"region", cfg.Region,
		"path_style", cfg.UsePathStyle,
	)

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		),
	)
	if err != nil {
		slog.Error("s3 client config failed", "error", err)
		return nil, fmt.Errorf("loading aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = cfg.UsePathStyle
	})

	slog.Info("s3 client initialized", "endpoint", cfg.Endpoint)
	return client, nil
}

// EnsureBucket creates the bucket if it does not already exist.
func EnsureBucket(ctx context.Context, client *s3.Client, bucket string) error {
	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucket)})
	if err == nil {
		slog.Info("s3 bucket exists", "bucket", bucket)
		return nil
	}

	slog.Info("s3 bucket not found, creating", "bucket", bucket)
	_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})
	if err != nil {
		slog.Error("s3 bucket creation failed", "bucket", bucket, "error", err)
		return fmt.Errorf("creating bucket %s: %w", bucket, err)
	}

	slog.Info("s3 bucket created", "bucket", bucket)
	return nil
}

// ObjectStore implements domain.ObjectStorage using S3-compatible storage.
type ObjectStore struct {
	client *s3.Client
	bucket string
}

// NewObjectStorage creates a new S3-backed object storage.
func NewObjectStorage(client *s3.Client, bucket string) *ObjectStore {
	return &ObjectStore{client: client, bucket: bucket}
}

func (s *ObjectStore) Upload(ctx context.Context, key string, reader io.Reader, contentType string, size int64) error {
	slog.Info("s3 uploading object",
		"bucket", s.bucket,
		"key", key,
		"content_type", contentType,
		"size_bytes", size,
	)

	input := &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          reader,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	}

	start := time.Now()
	if _, err := s.client.PutObject(ctx, input); err != nil {
		slog.Error("s3 upload failed",
			"key", key,
			"duration", time.Since(start),
			"error", err,
		)
		return fmt.Errorf("uploading object %s: %w", key, err)
	}

	slog.Info("s3 upload completed",
		"key", key,
		"size_bytes", size,
		"duration", time.Since(start),
	)
	return nil
}

func (s *ObjectStore) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	slog.Info("s3 downloading object", "bucket", s.bucket, "key", key)

	start := time.Now()
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		slog.Error("s3 download failed",
			"key", key,
			"duration", time.Since(start),
			"error", err,
		)
		return nil, fmt.Errorf("downloading object %s: %w", key, err)
	}

	slog.Info("s3 download completed",
		"key", key,
		"duration", time.Since(start),
	)
	return output.Body, nil
}

func (s *ObjectStore) Delete(ctx context.Context, key string) error {
	slog.Info("s3 deleting object", "bucket", s.bucket, "key", key)

	start := time.Now()
	if _, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}); err != nil {
		slog.Error("s3 delete failed",
			"key", key,
			"duration", time.Since(start),
			"error", err,
		)
		return fmt.Errorf("deleting object %s: %w", key, err)
	}

	slog.Info("s3 delete completed",
		"key", key,
		"duration", time.Since(start),
	)
	return nil
}

func (s *ObjectStore) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	slog.Debug("s3 generating presigned url", "key", key, "expiry", expiry)

	presigner := s3.NewPresignClient(s.client)
	req, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		slog.Error("s3 presign failed", "key", key, "error", err)
		return "", fmt.Errorf("presigning object %s: %w", key, err)
	}

	slog.Debug("s3 presigned url generated", "key", key)
	return req.URL, nil
}
