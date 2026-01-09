package s3

import (
	"context"
	"fmt"
	"io"
	"time"

	app_config "sereni-storage-provider/internal/config"
	"sereni-storage-provider/internal/providers/storage/interfaces"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3StorageProvider struct {
	client   *s3.Client
	bucket   string
	region   string
	uploader *manager.Uploader
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

	uploader := manager.NewUploader(client)

	return &S3StorageProvider{
		client:   client,
		bucket:   cfg.Bucket,
		region:   cfg.Region,
		uploader: uploader,
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
		return false, err
	}
	return true, nil
}

func (s *S3StorageProvider) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	upParams := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(objectName),
		Body:        reader,
		ContentType: aws.String(contentType),
	}

	_, err := s.uploader.Upload(ctx, upParams)
	if err != nil {
		return "", err
	}

	return s.GetURL(ctx, objectName)
}

func (s *S3StorageProvider) GetURL(ctx context.Context, objectName string) (string, error) {
	presigner := s3.NewPresignClient(s.client)
	out, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectName),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		// fallback to public URL
		return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, objectName), nil
	}
	return out.URL, nil
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
