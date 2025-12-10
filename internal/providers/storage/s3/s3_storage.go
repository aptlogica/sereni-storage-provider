package s3

import (
	"context"
	"fmt"
	"io"

	app_config "sereni-storage-provider/internal/config"
	"sereni-storage-provider/internal/providers/storage/interfaces"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3StorageProvider struct {
	client *s3.Client
	bucket string
	region string
}

func NewS3StorageProvider(cfg *app_config.StorageAWSConfig) (interfaces.StorageProvider, error) {
	ctx := context.TODO()

	// Load AWS config
	// We use custom resolver if endpoint is provided (for S3 compatible services other than AWS if needed via this driver)
	// But usually we use standard AWS loading.

	opts := []func(*aws_config.LoadOptions) error{
		aws_config.WithRegion(cfg.Region),
		aws_config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
	}

	awsCfg, err := aws_config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.UsePathStyle {
			o.UsePathStyle = true
		}
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
	})

	return &S3StorageProvider{
		client: client,
		bucket: cfg.Bucket,
		region: cfg.Region,
	}, nil
}

func (s *S3StorageProvider) Delete(ctx context.Context, objectName string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectName),
	})
	return err
}

func (s *S3StorageProvider) Download(ctx context.Context, objectName string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectName),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (s *S3StorageProvider) Exists(ctx context.Context, objectName string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectName),
	})
	if err != nil {
		// Proper error checking for 404 in SDK v2 is verbose, generic check for now
		return false, nil
	}
	return true, nil
}

func (s *S3StorageProvider) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(objectName),
		Body:        reader,
		ContentType: aws.String(contentType),
	}

	// If size is known, set it (helps with performance/multipart decisions or strictly required for some readers)
	// v2 manager usually handles this, but here we use low level client for simplicity of one-file.
	// For production large file upload, use feature/s3/manager.NewUploader

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return "", err
	}

	return s.GetURL(ctx, objectName)
}

func (s *S3StorageProvider) GetURL(ctx context.Context, objectName string) (string, error) {
	// Public URL: https://bucket.s3.region.amazonaws.com/key
	// Or https://s3.region.amazonaws.com/bucket/key (path style)

	// Simple standard construction
	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, objectName)
	return url, nil
}

func (s *S3StorageProvider) HealthCheck(ctx context.Context) error {
	// Check bucket accessibility
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		return fmt.Errorf("s3 health check failed: %w", err)
	}
	return nil
}
