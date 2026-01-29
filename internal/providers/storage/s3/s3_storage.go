package s3

import (
	"context"
	"fmt"
	"io"
	"strings"
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
	Client   interfaces.S3Client
	Bucket   string
	Region   string
	Uploader *manager.Uploader
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
		Client:   client,
		Bucket:   cfg.Bucket,
		Region:   cfg.Region,
		Uploader: uploader,
	}, nil
}

func (s *S3StorageProvider) Delete(ctx context.Context, objectName string) error {
	_, err := s.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(objectName),
	})
	return err
}

func (s *S3StorageProvider) Download(ctx context.Context, objectName string) (io.ReadCloser, error) {
	out, err := s.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(objectName),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (s *S3StorageProvider) Exists(ctx context.Context, objectName string) (bool, error) {
	_, err := s.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(objectName),
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *S3StorageProvider) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	upParams := &s3.PutObjectInput{
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(objectName),
		Body:        reader,
		ContentType: aws.String(contentType),
	}

	_, err := s.Client.PutObject(ctx, upParams)
	if err != nil {
		return "", err
	}

	return s.GetURL(ctx, objectName)
}

func (s *S3StorageProvider) GetURL(ctx context.Context, objectName string) (string, error) {
	if client, ok := s.Client.(*s3.Client); ok {
		presigner := s3.NewPresignClient(client)
		out, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(s.Bucket),
			Key:    aws.String(objectName),
		}, s3.WithPresignExpires(15*time.Minute))
		if err != nil {
			// fallback to public URL
			return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.Bucket, s.Region, objectName), nil
		}
		return out.URL, nil
	}
	// fallback to public URL for mocks
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.Bucket, s.Region, objectName), nil
}

func (s *S3StorageProvider) HealthCheck(ctx context.Context) error {
	// Check bucket accessibility
	_, err := s.Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.Bucket),
	})
	if err != nil {
		return fmt.Errorf("s3 health check failed: %w", err)
	}
	return nil
}

// GetSize returns the size in bytes of an object or directory in S3
// For S3, directories are virtual - they are prefixes for object keys
// Returns (size, isDirectory, error)
func (s *S3StorageProvider) GetSize(ctx context.Context, objectName string) (int64, bool, error) {
	// First try to get as a single object
	out, err := s.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(objectName),
	})
	if err == nil && out.ContentLength != nil {
		// Object exists, return its size
		return *out.ContentLength, false, nil
	}

	// If the path ends with "/", treat it as a directory
	if strings.HasSuffix(objectName, "/") {
		size, err := s.getDirectorySize(ctx, objectName)
		return size, true, err
	}

	// Otherwise, return the HeadObject error
	return 0, false, fmt.Errorf("failed to get object metadata: %w", err)
}

// getDirectorySize calculates the total size of all objects with the given prefix
func (s *S3StorageProvider) getDirectorySize(ctx context.Context, prefix string) (int64, error) {
	// Ensure prefix ends with "/" for directory-like behavior
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	var totalSize int64
	paginator := s3.NewListObjectsV2Paginator(s.Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.Bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return 0, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range page.Contents {
			if obj.Size != nil {
				totalSize += *obj.Size
			}
		}
	}

	return totalSize, nil
}
