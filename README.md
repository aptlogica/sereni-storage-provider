<h1 align="center">sereni-storage-provider - Cloud-Native Object Storage Service</h1>

<p align="center">Enterprise-grade object storage service and open source storage service with S3-compatible APIs. A comprehensive file storage API and cloud storage backend providing advanced access control, multi-tenant architecture, and seamless integration with modern cloud infrastructure.</p>

<p align="center">
<a href="LICENSE"><img src="https://img.shields.io/badge/Version-1.0.0-blue?style=for-the-badge" alt="Version"></a>
<a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.26.2-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version"></a>
<a href="https://aws.amazon.com/s3/"><img src="https://img.shields.io/badge/AWS_S3-Compatible-FF9900?style=for-the-badge&logo=amazons3&logoColor=white" alt="AWS S3"></a>
<a href="https://min.io/"><img src="https://img.shields.io/badge/MinIO-Supported-C72E49?style=for-the-badge&logo=minio&logoColor=white" alt="MinIO"></a>
<a href="https://gin-gonic.com/"><img src="https://img.shields.io/badge/Gin-Framework-008ECF?style=for-the-badge&logo=gin&logoColor=white" alt="Gin"></a>
<a href="https://www.docker.com/"><img src="https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker"></a>
<a href="https://swagger.io/"><img src="https://img.shields.io/badge/Swagger-Documented-85EA2D?style=for-the-badge&logo=swagger&logoColor=black" alt="Swagger"></a>
</p>

<p align="center">
<a href="https://github.com/aptlogica/sereni-storage-provider/actions/workflows/ci.yml"><img src="https://github.com/aptlogica/sereni-storage-provider/actions/workflows/ci.yml/badge.svg" alt="Build Status"></a>
<a href="https://github.com/aptlogica/sereni-storage-provider/actions/workflows/codeql.yml"><img src="https://github.com/aptlogica/sereni-storage-provider/actions/workflows/codeql.yml/badge.svg" alt="CodeQL"></a>
<a href="https://sonarcloud.io/dashboard?id=aptlogica_sereni-storage-provider"><img src="https://sonarcloud.io/api/project_badges/measure?project=aptlogica_sereni-storage-provider&metric=alert_status" alt="Quality Gate Status"></a>
<a href="./coverage.html"><img src="https://img.shields.io/badge/coverage-91%25-brightgreen" alt="Coverage Status"></a>
</p>

<p align="center">
<a href="LICENSE"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License"></a>
</p>
## API Endpoints

All endpoints are prefixed with `/api/v1`.

| Method | Path                  | Description                        |
|--------|-----------------------|------------------------------------|
| GET    | /health               | Health check                       |
| POST   | /storage/upload       | Upload a file                      |
| GET    | /storage/download     | Download a file                    |
| DELETE | /storage/delete       | Delete a file                      |
| GET    | /storage/consumption  | Get file/directory size            |

See [docs/swagger.yaml](docs/swagger.yaml) for full OpenAPI spec and request/response details.

## Architecture Diagram

```mermaid
flowchart TD
    Client-->|REST/HTTP|API[API Layer]
    API-->|Service Calls|Service[Service Layer]
    Service-->|Provider Interface|Provider[Storage Provider]
    Provider-->|Local/S3/MinIO|Backend[Storage Backend]
```

## How to Contribute

We welcome contributions! Please:
- Read [CONTRIBUTING.md](CONTRIBUTING.md)
- Open issues for bugs/feature requests
- Fork and submit pull requests
- Follow the code style and add tests


## Overview

**Sereni Storage Provider** is an open-source, enterprise-grade storage platform that provides S3-compatible storage, MinIO support, and local file system integration. Built for scalability and flexibility, it enables developers to manage and store data securely across cloud-native and on-premise infrastructures. With API compatibility, robust access control mechanisms, and a scalable backend architecture, Sereni simplifies storage management for modern, distributed applications.

**MinIO vs Amazon S3:** MinIO is an S3-compatible open-source object store — the same client code works against both. In production, point S3_ENDPOINT at your Amazon S3 region endpoint. In development, MinIO runs locally via Docker as a drop-in replacement with no code changes.

- **S3-Compatible APIs**: Full compatibility with Amazon S3 client libraries and tools
- **Advanced Access Control**: Role-based permissions with fine-grained access policies
- **Multi-Tenant Architecture**: Isolated storage contexts for different organizations
- **High Performance**: Optimized for large file uploads with high throughput
- **Data Security**: Encryption at rest and in transit with comprehensive audit logging
- **Storage Microservice**: Complete object storage API with file upload service capabilities
- **Cloud-Native Design**: Kubernetes deployment with auto-scaling capabilities

### Antivirus Integration
When used with [sereni-antivirus-clamav](https://github.com/aptlogica/sereni-antivirus-clamav), all file uploads are scanned for malware before being committed to storage. The scanning step is handled by the sereni-base orchestration layer — neither service calls the other directly. Configure `ANTIVIRUS_SERVICE_URL` in sereni-base to enable this behaviour.

## Architecture
- Go 1.26.2+, idiomatic design
- Modular, testable codebase

## Installation
```sh
go get github.com/aptlogica/sereni-storage-provider
```

## Configuration
See `.env.example` for environment variables and configuration options.

## Quick Start

```go
package main

import (
    "context"
    "log"
    "os"
    
    "github.com/aptlogica/sereni-storage-provider/pkg/client"
    "github.com/aptlogica/sereni-storage-provider/pkg/config"
)

func main() {
    // Initialize configuration
    cfg := config.New()
    cfg.StorageBackend = "s3"
    cfg.S3Endpoint = "https://s3.amazonaws.com"
    cfg.S3Region = "us-east-1"
    cfg.S3AccessKey = "your-access-key"
    cfg.S3SecretKey = "your-secret-key"
    
    // Create storage client
    client, err := client.New(cfg)
    if err != nil {
        log.Fatal("Failed to create client:", err)
    }
    defer client.Close()
    
    // Upload a file
    file, err := os.Open("example.txt")
    if err != nil {
        log.Fatal("Failed to open file:", err)
    }
    defer file.Close()
    
    ctx := context.Background()
    result, err := client.Upload(ctx, "my-bucket", "example.txt", file)
    if err != nil {
        log.Fatal("Failed to upload file:", err)
    }
    
    log.Printf("File uploaded: %s", result.URL)
}
```

## Development

### Local Setup
```bash
# Clone the repository
git clone https://github.com/aptlogica/sereni-storage-provider.git
cd sereni-storage-provider

# Install dependencies
go mod download

# Set up environment
cp .env.example .env
# Configure your storage backend in .env

# Start MinIO for local development (optional)
docker run -d \
  -p 9000:9000 -p 9001:9001 \
  --name minio \
  minio/minio server /data --console-address ":9001"

# Start development server
go run ./cmd/server
```

### Environment Configuration
```bash
STORAGE_BACKEND=s3
S3_ENDPOINT=http://localhost:9000
S3_REGION=us-east-1
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
MAX_FILE_SIZE=100MB
PORT=8080
LOG_LEVEL=debug
```

### Storage Backends
```bash
# Local filesystem
STORAGE_BACKEND=local
LOCAL_STORAGE_PATH=./uploads

# Amazon S3
STORAGE_BACKEND=s3
S3_ENDPOINT=https://s3.amazonaws.com

# MinIO (S3-compatible)
STORAGE_BACKEND=s3
S3_ENDPOINT=http://localhost:9000
```

## Testing
- Run `go test ./...` to execute unit tests

## Security
See [SECURITY.md](SECURITY.md) for reporting vulnerabilities.

## SereniBase Ecosystem

Sereni Storage Provider is part of the [SereniBase](https://github.com/aptlogica/sereni-base) ecosystem, a collection of microservices designed to build robust, scalable applications.

- [sereni-base](https://github.com/aptlogica/sereni-base) - Core orchestration layer
- [sereni-antivirus-clamav](https://github.com/aptlogica/sereni-antivirus-clamav) - Antivirus scanning service
- [sereni-auth-provider](https://github.com/aptlogica/sereni-auth-provider) - Authentication and authorization service

## License
Apache License 2.0. Copyright (c) 2026 Aptlogica Technologies.
