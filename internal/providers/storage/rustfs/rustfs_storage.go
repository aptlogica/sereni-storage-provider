// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com
package rustfs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	app_errors "github.com/aptlogica/sereni-storage-provider/internal/app-errors"
	"github.com/aptlogica/sereni-storage-provider/internal/config"
	"github.com/aptlogica/sereni-storage-provider/internal/providers/storage/interfaces"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

type RustFSStorageProvider struct {
	Client  interfaces.S3Client
	Bucket  string
	BaseURL string
}

func NewRustFSStorageProvider(cfg *config.StorageRustFSConfig, serverCfg *config.ServerConfig) (interfaces.StorageProvider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	endpoint := cfg.Endpoint
	if !strings.Contains(endpoint, "://") {
		endpoint = "http://" + endpoint
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("invalid rustfs endpoint: %q", cfg.Endpoint)
	}

	awsCfg, err := awsConfig.LoadDefaultConfig(context.Background(),
		awsConfig.WithRegion("us-east-1"),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load rustfs aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
		o.BaseEndpoint = aws.String(endpoint)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(cfg.Bucket)})
	if err != nil {
		return nil, fmt.Errorf("failed to check rustfs bucket existence: %w", err)
	}

	port := u.Port()
	baseURL := fmt.Sprintf("%s://%s/%s/", serverCfg.Scheme, serverCfg.IP, cfg.Bucket)
	if port != "" {
		baseURL = fmt.Sprintf("%s://%s:%s/%s/", serverCfg.Scheme, serverCfg.IP, port, cfg.Bucket)
	}

	return &RustFSStorageProvider{
		Client:  client,
		Bucket:  cfg.Bucket,
		BaseURL: baseURL,
	}, nil
}

func (r *RustFSStorageProvider) Delete(ctx context.Context, objectName string) error {
	_, err := r.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.Bucket),
		Key:    aws.String(objectName),
	})
	return err
}

func (r *RustFSStorageProvider) Download(ctx context.Context, objectName string) (io.ReadCloser, error) {
	out, err := r.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.Bucket),
		Key:    aws.String(objectName),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (r *RustFSStorageProvider) Exists(ctx context.Context, objectName string) (bool, error) {
	_, err := r.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(r.Bucket),
		Key:    aws.String(objectName),
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && (apiErr.ErrorCode() == "NoSuchKey" || apiErr.ErrorCode() == "NotFound") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *RustFSStorageProvider) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	_, err := r.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.Bucket),
		Key:         aws.String(objectName),
		Body:        reader,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", err
	}
	return r.GetURL(ctx, objectName)
}

func (r *RustFSStorageProvider) GetURL(ctx context.Context, objectName string) (string, error) {
	cleanPath := strings.ReplaceAll(objectName, "\\", "/")
	return r.BaseURL + cleanPath, nil
}

func (r *RustFSStorageProvider) HealthCheck(ctx context.Context) error {
	_, err := r.Client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(r.Bucket)})
	if err != nil {
		return fmt.Errorf("rustfs health check failed: %w", err)
	}
	return nil
}

func (r *RustFSStorageProvider) GetSize(ctx context.Context, objectName string) (int64, bool, error) {
	out, err := r.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(r.Bucket),
		Key:    aws.String(objectName),
	})
	if err == nil && out.ContentLength != nil {
		return *out.ContentLength, false, nil
	}

	if err != nil && !isObjectNotFound(err) {
		return 0, false, fmt.Errorf("failed to get object metadata: %w", err)
	}

	size, hasObjects, listErr := r.getDirectorySize(ctx, objectName)
	if listErr != nil {
		return 0, false, listErr
	}
	if hasObjects || strings.HasSuffix(objectName, "/") {
		return size, true, nil
	}

	if err != nil {
		return 0, false, app_errors.FileNotFound
	}

	return 0, false, fmt.Errorf("failed to get object metadata: content length missing")
}

func (r *RustFSStorageProvider) getDirectorySize(ctx context.Context, prefix string) (int64, bool, error) {
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	var totalSize int64
	var objectCount int
	paginator := s3.NewListObjectsV2Paginator(r.Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(r.Bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return 0, false, fmt.Errorf("failed to list objects: %w", err)
		}
		for _, obj := range page.Contents {
			if obj.Size != nil {
				totalSize += *obj.Size
			}
			objectCount++
		}
	}

	return totalSize, objectCount > 0, nil
}

func isObjectNotFound(err error) bool {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code := strings.ToLower(apiErr.ErrorCode())
		return code == "nosuchkey" || code == "notfound" || code == "nosuchbucket"
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "nosuchkey") || strings.Contains(msg, "not found") || strings.Contains(msg, "status code: 404")
}
