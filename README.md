# Sereni Storage Provider - Multi-Backend File Storage Microservice

> A production-ready, scalable file storage microservice with support for Local, AWS S3, and MinIO backends. Deploy as a standalone storage service or integrate into any application ecosystem with language-agnostic REST APIs.

[![Version](https://img.shields.io/badge/Version-1.0.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24.4+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat-square&logo=docker)](https://www.docker.com/)
[![Quality Gate Status](https://sonar.aptlogica.com/api/project_badges/measure?project=aptlogica_sereni-storage-provider_268fdb46-e26c-4658-8e95-7ad63d65f666&metric=alert_status&token=sqb_6de6206ad8030012928d5d3ef806ce13462e6d4b)](https://sonar.aptlogica.com/dashboard?id=aptlogica_sereni-storage-provider_268fdb46-e26c-4658-8e95-7ad63d65f666)

## 📋 Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [API Documentation](#api-documentation)
- [Integration Guide](#integration-guide)
- [Architecture](#architecture)
- [Development](#development)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [FAQ](#faq)
- [License](#license)

## Overview

**Sereni Storage Provider** is a high-performance file storage microservice that abstracts away storage backend complexity. Upload, download, and manage files through a simple REST API while seamlessly switching between local filesystem, AWS S3, or MinIO storage without code changes. Built with Go and designed for cloud-native deployments, it provides automatic content-type detection, path normalization, rate limiting, and comprehensive error handling.

Sereni Storage Provider is designed with these key characteristics:

- **Multi-Backend Support**: Switch between Local, AWS S3, and MinIO storage with a single configuration change - no code modifications needed

- **Language-Agnostic**: Works with applications in any language (Node.js, Python, Java, .NET, PHP, Ruby, etc.) through standard HTTP REST APIs

- **Production-Ready**: Docker support, Swagger documentation, rate limiting, request tracking, health checks, and comprehensive testing

- **Developer-Friendly**: Zero-configuration local development with automatic directory creation, hot-reloading support, and clear error messages

### Why Choose Sereni Storage Provider?

- **Backend Flexibility**: Start with local storage for development, migrate to S3 or MinIO for production without changing application code

- **Cost Optimization**: Use local/MinIO for cost-sensitive workloads, S3 for global CDN distribution - switch anytime

- **S3-Compatible**: Works with AWS S3, DigitalOcean Spaces, Wasabi, Backblaze B2, or any S3-compatible storage

- **Smart Content-Type Detection**: Automatic MIME type detection from file extensions prevents browser rendering issues (SVG, PDF, etc.)

- **Path Normalization**: Automatic sanitization of file paths prevents directory traversal attacks and ensures cross-platform compatibility

- **Easy Integration**: Simple multipart/form-data uploads - works with standard HTML forms and any HTTP client

## Key Features

✅ **Multi-Backend Storage**
- Local Filesystem - Perfect for development and simple deployments
- AWS S3 - Global CDN distribution with high availability
- MinIO - Self-hosted S3-compatible storage for data sovereignty
- Switch backends with single environment variable change

✅ **File Operations**
- Upload files with multipart/form-data (standard HTML forms)
- Download files with automatic content-type headers
- Delete files with path-based API
- Get storage consumption statistics
- Automatic directory creation for local storage

✅ **Smart File Handling**
- Automatic content-type detection from file extensions
- Path normalization and sanitization (prevents directory traversal)
- SVG, PDF, and media file support with correct MIME types
- Configurable maximum upload size
- File size validation before storage

✅ **Production Features**
- Rate limiting to prevent API abuse
- Request ID tracking for debugging
- Health check endpoint for monitoring
- CORS support with configurable origins
- Comprehensive error handling with clear messages
- Docker and Docker Compose ready
- MinIO included in docker-compose for easy setup

✅ **Developer Experience**
- Swagger/OpenAPI documentation (auto-generated)
- Environment-based configuration
- Hot-reload support for development
- Comprehensive test coverage
- Clear logging with zerolog
- Single binary deployment

✅ **Security**
- Path traversal prevention
- Content-type validation (configurable whitelist)
- Request rate limiting
- Input sanitization
- Secure credential handling

✅ **Scalability**
- Stateless design for horizontal scaling
- Support for S3/MinIO distributed storage
- Concurrent request handling
- Optimized for high-throughput file operations

## Use Cases

### 1. **User-Generated Content (UGC) Storage**
Store user uploads (profile pictures, documents, media files) for web and mobile applications. Start with local storage for development, seamlessly migrate to S3 for production without changing application code.

```
User Upload → Storage Provider → Local/S3/MinIO → URL Returned → Display in App
```

### 2. **Multi-Tenant SaaS File Storage**
Provide isolated file storage for multiple tenants/customers. Use path-based organization (e.g., `/tenant-id/files/`) and switch to S3 for automatic replication and disaster recovery.

### 3. **Content Management Systems (CMS)**
Backend storage for CMS platforms managing images, videos, PDFs, and documents. Automatic content-type detection ensures files render correctly in browsers.

### 4. **E-Commerce Product Images**
Store product images, thumbnails, and media galleries. Use S3 with CloudFront CDN for fast global delivery, or MinIO for cost-effective self-hosted storage.

### 5. **Document Processing Pipelines**
Upload documents for processing (PDF generation, image conversion, etc.). Download processed results. Use local storage for temporary files, S3 for permanent archives.

### 6. **Mobile App Backend Storage**
Provide file upload/download APIs for iOS and Android apps. Works with standard multipart/form-data uploads from any mobile HTTP client.

### 7. **Microservices File Hub**
Centralize file storage for multiple microservices. All services use the same storage API, but storage backend can be optimized independently.

## Quick Start

### Prerequisites
- **Docker & Docker Compose** (v20.0+) - For containerized deployment
- **Go 1.24.4+** (optional) - For building from source
- **curl or Postman** (optional) - For API testing

### 30-Second Setup (Local Storage)

```bash
# 1. Clone the repository
git clone https://github.com/yourusername/sereni-storage-provider.git
cd sereni-storage-provider

# 2. Create environment file
cp example.env .env

# 3. Start the service with Docker Compose (includes MinIO)
docker-compose up -d

# 4. Verify installation
curl http://localhost:8083/health
```

**Service is now available at http://localhost:8083**

**MinIO Console:** http://localhost:9001 (credentials: minioadmin/minioadmin)

Visit **http://localhost:8083/swagger/index.html** for interactive API documentation.

**Next steps:** See [Installation](#installation) for more setup options, or [Usage](#usage) to upload your first file.

## Installation

### Option 1: Docker Compose (Recommended)

Includes MinIO for S3-compatible storage. Perfect for development and production.

```bash
# Step 1: Clone the repository
git clone https://github.com/yourusername/sereni-storage-provider.git
cd sereni-storage-provider

# Step 2: Create environment configuration
cp example.env .env

# Step 3: Configure storage backend (optional)
nano .env
# Set STORAGE_DRIVER to: local, s3, or minio
# For local storage (default), no additional config needed

# Step 4: Start services (storage provider + MinIO)
docker-compose up -d

# Step 5: Verify services are running
curl http://localhost:8083/health
# Expected: {"status":"healthy","storage":"local","timestamp":"..."}
```

**Result:** Storage service at http://localhost:8083, MinIO at http://localhost:9001

### Option 2: Docker (Storage Provider Only)

For custom deployments without MinIO.

```bash
# Step 1: Build the Docker image
docker build -t sereni-storage-provider:latest .

# Step 2: Run with local storage
docker run -d \
  -p 8083:8083 \
  -v $(pwd)/uploads:/app/uploads \
  -e STORAGE_DRIVER=local \
  -e STORAGE_DEV_PATH=/app/uploads \
  -e SERVER_PORT=8083 \
  -e SERVER_HOST=0.0.0.0 \
  -e SERVER_IP=localhost \
  -e SERVER_SCHEME=http \
  --name storage-provider \
  sereni-storage-provider:latest

# Step 3: Check logs
docker logs -f storage-provider

# Step 4: Verify service
curl http://localhost:8083/health
```

**Result:** Containerized storage service with local filesystem backend

### Option 3: Docker with AWS S3

Deploy with AWS S3 backend for production.

```bash
# Step 1: Build the Docker image
docker build -t sereni-storage-provider:latest .

# Step 2: Run with S3 configuration
docker run -d \
  -p 8083:8083 \
  -e STORAGE_DRIVER=s3 \
  -e AWS_REGION=us-east-1 \
  -e AWS_BUCKET=my-bucket-name \
  -e AWS_ACCESS_KEY=AKIAIOSFODNN7EXAMPLE \
  -e AWS_SECRET_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY \
  -e SERVER_PORT=8083 \
  -e SERVER_HOST=0.0.0.0 \
  -e SERVER_IP=yourdomain.com \
  -e SERVER_SCHEME=https \
  --name storage-provider \
  sereni-storage-provider:latest

# Step 3: Verify service
curl http://localhost:8083/health
```

**Result:** Storage service using AWS S3 as backend

### Option 4: From Source (Developers)

For development, testing, or modifications.

```bash
# Step 1: Ensure Go 1.24.4+ is installed
go version

# Step 2: Clone repository
git clone https://github.com/yourusername/sereni-storage-provider.git
cd sereni-storage-provider

# Step 3: Install dependencies
go mod download

# Step 4: Install Swagger CLI (for documentation)
go install github.com/swaggo/swag/cmd/swag@latest

# Step 5: Create environment file
cp example.env .env
nano .env  # Configure as needed

# Step 6: Generate Swagger documentation
swag init -g cmd/server/main.go

# Step 7: Build application
go build -o bin/storage-provider cmd/server/main.go

# Step 8: Run service
./bin/storage-provider
# or: go run cmd/server/main.go
```

**Result:** Service compiling from source and running on port 8083

## Configuration

### Environment Variables

Create `.env` file in your project root:

```dotenv
# === Server Configuration ===
SERVER_PORT=8083                                # HTTP server port
SERVER_HOST=0.0.0.0                             # Server bind address (0.0.0.0 for all interfaces)
SERVER_IP=localhost                             # Public IP/domain for constructing file URLs
SERVER_SCHEME=http                              # URL scheme: http or https
MAX_UPLOAD_SIZE_BYTES=10485760                  # Max upload size (default: 10MB)

# === Storage Backend ===
STORAGE_DRIVER=local                            # Storage backend: local, s3, minio

# === Local Storage Configuration ===
STORAGE_DEV_PATH=./uploads                      # Local storage directory path

# === AWS S3 Configuration ===
AWS_REGION=us-east-1                            # AWS region
AWS_BUCKET=my-bucket                            # S3 bucket name
AWS_ACCESS_KEY=AKIAIOSFODNN7EXAMPLE             # AWS access key ID
AWS_SECRET_KEY=wJalrXUtnFEMI/K7MDENG/bPx...     # AWS secret access key

# === MinIO Configuration ===
MINIO_ENDPOINT=localhost:9000                   # MinIO endpoint
MINIO_ACCESS_KEY=minioadmin                     # MinIO access key
MINIO_SECRET_KEY=minioadmin                     # MinIO secret key
MINIO_BUCKET=my-bucket                          # MinIO bucket name
MINIO_USE_SSL=false                             # Use SSL for MinIO connection

# === CORS Configuration ===
ALLOWED_ORIGINS=http://localhost:3000,http://example.com
                                                # Comma-separated allowed origins
```

### Default Values

If `.env` file values are not provided:
- `SERVER_PORT`: `8083`
- `SERVER_HOST`: `0.0.0.0`
- `SERVER_IP`: `localhost`
- `SERVER_SCHEME`: `http`
- `STORAGE_DRIVER`: `local`
- `STORAGE_DEV_PATH`: `./uploads`
- `MAX_UPLOAD_SIZE_BYTES`: `10485760` (10MB)

### Configuration Examples

**For Local Development:**
```dotenv
SERVER_PORT=8083
SERVER_HOST=0.0.0.0
SERVER_IP=localhost
SERVER_SCHEME=http
STORAGE_DRIVER=local
STORAGE_DEV_PATH=./uploads
ALLOWED_ORIGINS=*
```

**For Production with AWS S3:**
```dotenv
SERVER_PORT=8083
SERVER_HOST=0.0.0.0
SERVER_IP=api.yourdomain.com
SERVER_SCHEME=https
STORAGE_DRIVER=s3
AWS_REGION=us-east-1
AWS_BUCKET=production-files
AWS_ACCESS_KEY=${AWS_ACCESS_KEY_FROM_SECRETS}
AWS_SECRET_KEY=${AWS_SECRET_KEY_FROM_SECRETS}
ALLOWED_ORIGINS=https://yourdomain.com,https://app.yourdomain.com
MAX_UPLOAD_SIZE_BYTES=52428800  # 50MB
```

**For Self-Hosted with MinIO:**
```dotenv
SERVER_PORT=8083
SERVER_HOST=0.0.0.0
SERVER_IP=storage.yourdomain.com
SERVER_SCHEME=https
STORAGE_DRIVER=minio
MINIO_ENDPOINT=minio.yourdomain.com:9000
MINIO_ACCESS_KEY=your_minio_admin
MINIO_SECRET_KEY=your_secure_password
MINIO_BUCKET=production-files
MINIO_USE_SSL=true
ALLOWED_ORIGINS=https://yourdomain.com
```

**For Kubernetes (ConfigMap + Secret):**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: storage-provider-config
data:
  SERVER_PORT: "8083"
  SERVER_HOST: "0.0.0.0"
  SERVER_IP: "storage.cluster.local"
  SERVER_SCHEME: "http"
  STORAGE_DRIVER: "s3"
  AWS_REGION: "us-east-1"
  AWS_BUCKET: "production-files"
  ALLOWED_ORIGINS: "https://yourdomain.com"
---
apiVersion: v1
kind: Secret
metadata:
  name: storage-provider-secret
type: Opaque
data:
  AWS_ACCESS_KEY: <base64-encoded>
  AWS_SECRET_KEY: <base64-encoded>
```

### Storage Backend Comparison

| Feature | Local | MinIO | AWS S3 |
|---------|-------|-------|--------|
| **Cost** | Free | Self-hosted | Pay per GB |
| **Scalability** | Single node | Distributed | Unlimited |
| **Availability** | Single point of failure | High (clustered) | 99.99% SLA |
| **Setup** | Instant | Docker container | AWS account |
| **Best For** | Development, simple apps | On-premise, compliance | Production, global CDN |

## Usage

### Basic Usage

Upload a file using curl:

```bash
curl -X POST http://localhost:8083/api/v1/storage/upload \
  -F "file=@/path/to/your/file.jpg" \
  -F "path=images/profile.jpg"
```

### Example 1: Upload File

```bash
# Upload an image to specific path
curl -X POST http://localhost:8083/api/v1/storage/upload \
  -F "file=@./photo.jpg" \
  -F "path=users/123/profile.jpg"
```

**Output:**
```json
{
  "success": true,
  "message": "File uploaded successfully",
  "data": {
    "url": "http://localhost:8083/api/v1/storage/download?path=users/123/profile.jpg",
    "object_name": "users/123/profile.jpg"
  }
}
```

### Example 2: Upload to Auto-Generated Path

```bash
# Upload without specifying path (uses filename)
curl -X POST http://localhost:8083/api/v1/storage/upload \
  -F "file=@./document.pdf"
```

**Output:**
```json
{
  "success": true,
  "message": "File uploaded successfully",
  "data": {
    "url": "http://localhost:8083/api/v1/storage/download?path=document.pdf",
    "object_name": "document.pdf"
  }
}
```

### Example 3: Download File

```bash
# Download file by path
curl -X GET "http://localhost:8083/api/v1/storage/download?path=users/123/profile.jpg" \
  --output downloaded-file.jpg
```

**Result:** File downloaded with correct content-type header for browser rendering

### Example 4: Delete File

```bash
# Delete file by path
curl -X DELETE "http://localhost:8083/api/v1/storage/delete?path=users/123/profile.jpg"
```

**Output:**
```json
{
  "success": true,
  "message": "File deleted successfully"
}
```

### Example 5: Get Storage Consumption

```bash
# Get storage usage statistics (local storage only)
curl -X GET http://localhost:8083/api/v1/storage/consumption
```

**Output:**
```json
{
  "success": true,
  "message": "Storage consumption retrieved",
  "data": {
    "total_size_bytes": 15728640,
    "total_size_mb": 15.0,
    "file_count": 42
  }
}
```

### Example 6: Health Check

```bash
# Check service health
curl http://localhost:8083/health
```

**Output:**
```json
{
  "status": "healthy",
  "storage": "local",
  "timestamp": "2026-03-11T10:30:00Z"
}
```

## API Documentation

### Interactive API Docs

Once the service is running, visit: **http://localhost:8083/swagger/index.html**

Full interactive API documentation is available with the running service where you can test all endpoints.

### Endpoints

#### Upload File
```http
POST /api/v1/storage/upload
Content-Type: multipart/form-data
```

**Description:** Upload a file to storage. Supports any file type with automatic content-type detection.

**Request (multipart/form-data):**
```bash
curl -X POST http://localhost:8083/api/v1/storage/upload \
  -F "file=@./image.jpg" \
  -F "path=uploads/2026/03/image.jpg"
```

**Parameters:**
- `file` (required): File to upload (multipart file)
- `path` (optional): Destination path (if omitted, uses filename only)

**Response (Success - 200):**
```json
{
  "success": true,
  "message": "File uploaded successfully",
  "data": {
    "url": "http://localhost:8083/api/v1/storage/download?path=uploads/2026/03/image.jpg",
    "object_name": "uploads/2026/03/image.jpg"
  }
}
```

**Response (Error - 400):**
```json
{
  "success": false,
  "message": "File too large",
  "error": "file size exceeds maximum allowed (10MB)"
}
```

#### Download File
```http
GET /api/v1/storage/download?path={path}
```

**Description:** Download a file from storage. Returns file with appropriate content-type header.

**Request:**
```bash
curl -X GET "http://localhost:8083/api/v1/storage/download?path=uploads/2026/03/image.jpg" \
  --output downloaded-image.jpg
```

**Parameters:**
- `path` (required): Path to file in storage

**Response (Success - 200):**
- Binary file data with appropriate `Content-Type` header
- `Content-Disposition` header for filename

**Response (Error - 404):**
```json
{
  "success": false,
  "message": "File not found",
  "error": "file does not exist at path: uploads/2026/03/image.jpg"
}
```

#### Delete File
```http
DELETE /api/v1/storage/delete?path={path}
```

**Description:** Delete a file from storage.

**Request:**
```bash
curl -X DELETE "http://localhost:8083/api/v1/storage/delete?path=uploads/2026/03/image.jpg"
```

**Parameters:**
- `path` (required): Path to file in storage

**Response (Success - 200):**
```json
{
  "success": true,
  "message": "File deleted successfully"
}
```

**Response (Error - 404):**
```json
{
  "success": false,
  "message": "File not found",
  "error": "file does not exist"
}
```

#### Get Storage Consumption
```http
GET /api/v1/storage/consumption
```

**Description:** Get storage consumption statistics. Currently only works with local storage backend.

**Request:**
```bash
curl -X GET http://localhost:8083/api/v1/storage/consumption
```

**Response (Success - 200):**
```json
{
  "success": true,
  "message": "Storage consumption retrieved",
  "data": {
    "total_size_bytes": 15728640,
    "total_size_mb": 15.0,
    "file_count": 42
  }
}
```

**Response (S3/MinIO - 501):**
```json
{
  "success": false,
  "message": "Not implemented for this storage backend"
}
```

#### Health Check
```http
GET /health
```

**Description:** Check service health and storage backend status.

**Request:**
```bash
curl http://localhost:8083/health
```

**Response (Success - 200):**
```json
{
  "status": "healthy",
  "storage": "local",
  "timestamp": "2026-03-11T10:30:00Z"
}
```

### Error Codes

| Code | Name | Description |
|------|------|-------------|
| 200 | OK | Request successful |
| 400 | Bad Request | Invalid parameters (missing file, invalid path, file too large) |
| 404 | Not Found | File not found at specified path |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Server Error | Internal error or storage backend issue |
| 501 | Not Implemented | Feature not available for current storage backend |

## Integration Guide

### JavaScript / Node.js + Express

```javascript
const express = require('express');
const multer = require('multer');
const FormData = require('form-data');
const axios = require('axios');
const fs = require('fs');

const app = express();
const upload = multer({ dest: 'temp/' });
const STORAGE_PROVIDER_URL = 'http://localhost:8083';

// Upload file endpoint
app.post('/upload', upload.single('file'), async (req, res) => {
  try {
    const file = req.file;
    const destinationPath = req.body.path || `uploads/${Date.now()}_${file.originalname}`;
    
    // Create form data
    const formData = new FormData();
    formData.append('file', fs.createReadStream(file.path), file.originalname);
    formData.append('path', destinationPath);
    
    // Upload to storage provider
    const response = await axios.post(
      `${STORAGE_PROVIDER_URL}/api/v1/storage/upload`,
      formData,
      {
        headers: formData.getHeaders(),
        maxContentLength: Infinity,
        maxBodyLength: Infinity
      }
    );
    
    // Clean up temp file
    fs.unlinkSync(file.path);
    
    if (response.data.success) {
      res.json({
        message: 'File uploaded successfully',
        url: response.data.data.url,
        path: response.data.data.object_name
      });
    } else {
      res.status(500).json({ error: 'Upload failed' });
    }
    
  } catch (error) {
    console.error('Upload error:', error.message);
    res.status(500).json({ error: 'Upload failed' });
  }
});

// Download file endpoint (proxy)
app.get('/download/:filename', async (req, res) => {
  try {
    const filename = req.params.filename;
    const path = `uploads/${filename}`;
    
    const response = await axios.get(
      `${STORAGE_PROVIDER_URL}/api/v1/storage/download`,
      {
        params: { path: path },
        responseType: 'stream'
      }
    );
    
    // Forward headers
    res.setHeader('Content-Type', response.headers['content-type']);
    res.setHeader('Content-Disposition', response.headers['content-disposition']);
    
    // Pipe file stream to response
    response.data.pipe(res);
    
  } catch (error) {
    if (error.response && error.response.status === 404) {
      res.status(404).json({ error: 'File not found' });
    } else {
      res.status(500).json({ error: 'Download failed' });
    }
  }
});

// Delete file endpoint
app.delete('/delete/:filename', async (req, res) => {
  try {
    const filename = req.params.filename;
    const path = `uploads/${filename}`;
    
    const response = await axios.delete(
      `${STORAGE_PROVIDER_URL}/api/v1/storage/delete`,
      { params: { path: path } }
    );
    
    if (response.data.success) {
      res.json({ message: 'File deleted successfully' });
    } else {
      res.status(500).json({ error: 'Deletion failed' });
    }
    
  } catch (error) {
    res.status(500).json({ error: 'Deletion failed' });
  }
});

app.listen(3000, () => console.log('Server running on port 3000'));
```

### Python / Flask

```python
from flask import Flask, request, jsonify, send_file
import requests
from werkzeug.utils import secure_filename
import os
from datetime import datetime

app = Flask(__name__)
STORAGE_PROVIDER_URL = 'http://localhost:8083'
TEMP_FOLDER = 'temp'

os.makedirs(TEMP_FOLDER, exist_ok=True)

@app.route('/upload', methods=['POST'])
def upload_file():
    """Upload file to storage provider"""
    try:
        if 'file' not in request.files:
            return jsonify({'error': 'No file provided'}), 400
        
        file = request.files['file']
        if file.filename == '':
            return jsonify({'error': 'No file selected'}), 400
        
        # Get destination path from request or generate one
        destination_path = request.form.get('path')
        if not destination_path:
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
            destination_path = f'uploads/{timestamp}_{secure_filename(file.filename)}'
        
        # Prepare multipart form data
        files = {'file': (file.filename, file.stream, file.content_type)}
        data = {'path': destination_path}
        
        # Upload to storage provider
        response = requests.post(
            f'{STORAGE_PROVIDER_URL}/api/v1/storage/upload',
            files=files,
            data=data,
            timeout=30
        )
        
        if response.status_code == 200:
            result = response.json()
            if result.get('success'):
                return jsonify({
                    'message': 'File uploaded successfully',
                    'url': result['data']['url'],
                    'path': result['data']['object_name']
                }), 200
        
        return jsonify({'error': 'Upload failed'}), 500
        
    except Exception as e:
        return jsonify({'error': str(e)}), 500

@app.route('/download/<path:filepath>', methods=['GET'])
def download_file(filepath):
    """Download file from storage provider"""
    try:
        response = requests.get(
            f'{STORAGE_PROVIDER_URL}/api/v1/storage/download',
            params={'path': filepath},
            stream=True,
            timeout=30
        )
        
        if response.status_code == 200:
            # Save to temp file and send
            temp_path = os.path.join(TEMP_FOLDER, secure_filename(filepath.split('/')[-1]))
            with open(temp_path, 'wb') as f:
                for chunk in response.iter_content(chunk_size=8192):
                    f.write(chunk)
            
            return send_file(
                temp_path,
                mimetype=response.headers.get('Content-Type'),
                as_attachment=True,
                download_name=filepath.split('/')[-1]
            )
        
        return jsonify({'error': 'File not found'}), 404
        
    except Exception as e:
        return jsonify({'error': str(e)}), 500

@app.route('/delete/<path:filepath>', methods=['DELETE'])
def delete_file(filepath):
    """Delete file from storage provider"""
    try:
        response = requests.delete(
            f'{STORAGE_PROVIDER_URL}/api/v1/storage/delete',
            params={'path': filepath},
            timeout=10
        )
        
        if response.status_code == 200:
            result = response.json()
            if result.get('success'):
                return jsonify({'message': 'File deleted successfully'}), 200
        
        return jsonify({'error': 'Deletion failed'}), 500
        
    except Exception as e:
        return jsonify({'error': str(e)}), 500

@app.route('/storage-info', methods=['GET'])
def get_storage_info():
    """Get storage consumption"""
    try:
        response = requests.get(
            f'{STORAGE_PROVIDER_URL}/api/v1/storage/consumption',
            timeout=10
        )
        
        if response.status_code == 200:
            return jsonify(response.json()), 200
        
        return jsonify({'error': 'Failed to retrieve storage info'}), 500
        
    except Exception as e:
        return jsonify({'error': str(e)}), 500

if __name__ == '__main__':
    app.run(debug=True, port=5000)
```

### PHP / Laravel

```php
<?php
namespace App\Http\Controllers;

use Illuminate\Http\Request;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Storage;
use Illuminate\Support\Str;

class FileStorageController extends Controller
{
    private $storageProviderUrl = 'http://localhost:8083';
    
    /**
     * Upload file to storage provider
     */
    public function upload(Request $request)
    {
        try {
            $request->validate([
                'file' => 'required|file|max:10240', // Max 10MB
                'path' => 'nullable|string'
            ]);
            
            $file = $request->file('file');
            $destinationPath = $request->input('path', 
                'uploads/' . date('Y-m-d') . '/' . Str::uuid() . '_' . $file->getClientOriginalName()
            );
            
            // Upload to storage provider
            $response = Http::attach(
                'file',
                file_get_contents($file->getRealPath()),
                $file->getClientOriginalName()
            )->post($this->storageProviderUrl . '/api/v1/storage/upload', [
                'path' => $destinationPath
            ]);
            
            if ($response->successful() && $response->json('success')) {
                return response()->json([
                    'message' => 'File uploaded successfully',
                    'url' => $response->json('data.url'),
                    'path' => $response->json('data.object_name')
                ], 200);
            }
            
            return response()->json(['error' => 'Upload failed'], 500);
            
        } catch (\Exception $e) {
            return response()->json(['error' => $e->getMessage()], 500);
        }
    }
    
    /**
     * Download file from storage provider
     */
    public function download($filepath)
    {
        try {
            $response = Http::get($this->storageProviderUrl . '/api/v1/storage/download', [
                'path' => $filepath
            ]);
            
            if ($response->successful()) {
                $contentType = $response->header('Content-Type');
                $filename = basename($filepath);
                
                return response($response->body())
                    ->header('Content-Type', $contentType)
                    ->header('Content-Disposition', 'attachment; filename="' . $filename . '"');
            }
            
            return response()->json(['error' => 'File not found'], 404);
            
        } catch (\Exception $e) {
            return response()->json(['error' => $e->getMessage()], 500);
        }
    }
    
    /**
     * Delete file from storage provider
     */
    public function delete($filepath)
    {
        try {
            $response = Http::delete($this->storageProviderUrl . '/api/v1/storage/delete', [
                'path' => $filepath
            ]);
            
            if ($response->successful() && $response->json('success')) {
                return response()->json(['message' => 'File deleted successfully'], 200);
            }
            
            return response()->json(['error' => 'Deletion failed'], 500);
            
        } catch (\Exception $e) {
            return response()->json(['error' => $e->getMessage()], 500);
        }
    }
    
    /**
     * Get storage consumption
     */
    public function getStorageInfo()
    {
        try {
            $response = Http::get($this->storageProviderUrl . '/api/v1/storage/consumption');
            
            if ($response->successful()) {
                return response()->json($response->json(), 200);
            }
            
            return response()->json(['error' => 'Failed to retrieve storage info'], 500);
            
        } catch (\Exception $e) {
            return response()->json(['error' => $e->getMessage()], 500);
        }
    }
}
```

### Java / Spring Boot

```java
package com.example.storage;

import org.springframework.core.io.ByteArrayResource;
import org.springframework.core.io.Resource;
import org.springframework.http.*;
import org.springframework.stereotype.Controller;
import org.springframework.util.LinkedMultiValueMap;
import org.springframework.util.MultiValueMap;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.client.RestTemplate;
import org.springframework.web.multipart.MultipartFile;

import java.io.IOException;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.Map;

@RestController
@RequestMapping("/api/files")
public class FileStorageController {
    
    private final RestTemplate restTemplate = new RestTemplate();
    private static final String STORAGE_PROVIDER_URL = "http://localhost:8083";
    
    @PostMapping("/upload")
    public ResponseEntity<?> uploadFile(
            @RequestParam("file") MultipartFile file,
            @RequestParam(value = "path", required = false) String path
    ) {
        try {
            // Generate destination path if not provided
            if (path == null || path.isEmpty()) {
                String timestamp = LocalDateTime.now()
                    .format(DateTimeFormatter.ofPattern("yyyyMMdd_HHmmss"));
                path = "uploads/" + timestamp + "_" + file.getOriginalFilename();
            }
            
            // Prepare multipart request
            HttpHeaders headers = new HttpHeaders();
            headers.setContentType(MediaType.MULTIPART_FORM_DATA);
            
            MultiValueMap<String, Object> body = new LinkedMultiValueMap<>();
            body.add("file", new ByteArrayResource(file.getBytes()) {
                @Override
                public String getFilename() {
                    return file.getOriginalFilename();
                }
            });
            body.add("path", path);
            
            HttpEntity<MultiValueMap<String, Object>> requestEntity = 
                new HttpEntity<>(body, headers);
            
            // Upload to storage provider
            ResponseEntity<Map> response = restTemplate.exchange(
                STORAGE_PROVIDER_URL + "/api/v1/storage/upload",
                HttpMethod.POST,
                requestEntity,
                Map.class
            );
            
            if (response.getStatusCode() == HttpStatus.OK) {
                Map<String, Object> result = response.getBody();
                if (result != null && Boolean.TRUE.equals(result.get("success"))) {
                    Map<String, Object> data = (Map<String, Object>) result.get("data");
                    return ResponseEntity.ok(Map.of(
                        "message", "File uploaded successfully",
                        "url", data.get("url"),
                        "path", data.get("object_name")
                    ));
                }
            }
            
            return ResponseEntity.status(500).body(Map.of("error", "Upload failed"));
            
        } catch (IOException e) {
            return ResponseEntity.status(500).body(Map.of("error", e.getMessage()));
        }
    }
    
    @GetMapping("/download/{filepath}")
    public ResponseEntity<Resource> downloadFile(@PathVariable String filepath) {
        try {
            ResponseEntity<byte[]> response = restTemplate.exchange(
                STORAGE_PROVIDER_URL + "/api/v1/storage/download?path=" + filepath,
                HttpMethod.GET,
                null,
                byte[].class
            );
            
            if (response.getStatusCode() == HttpStatus.OK) {
                ByteArrayResource resource = new ByteArrayResource(response.getBody());
                
                return ResponseEntity.ok()
                    .contentType(MediaType.parseMediaType(
                        response.getHeaders().getContentType().toString()))
                    .header(HttpHeaders.CONTENT_DISPOSITION,
                        "attachment; filename=\"" + filepath.substring(filepath.lastIndexOf('/') + 1) + "\"")
                    .body(resource);
            }
            
            return ResponseEntity.notFound().build();
            
        } catch (Exception e) {
            return ResponseEntity.status(500).build();
        }
    }
    
    @DeleteMapping("/delete/{filepath}")
    public ResponseEntity<?> deleteFile(@PathVariable String filepath) {
        try {
            ResponseEntity<Map> response = restTemplate.exchange(
                STORAGE_PROVIDER_URL + "/api/v1/storage/delete?path=" + filepath,
                HttpMethod.DELETE,
                null,
                Map.class
            );
            
            if (response.getStatusCode() == HttpStatus.OK) {
                Map<String, Object> result = response.getBody();
                if (result != null && Boolean.TRUE.equals(result.get("success"))) {
                    return ResponseEntity.ok(Map.of("message", "File deleted successfully"));
                }
            }
            
            return ResponseEntity.status(500).body(Map.of("error", "Deletion failed"));
            
        } catch (Exception e) {
            return ResponseEntity.status(500).body(Map.of("error", e.getMessage()));
        }
    }
}
```

## Architecture

### System Architecture

```
┌─────────────────────────────────────────────────────────┐
│              Client Applications                        │
│     (Web, Mobile, Desktop, CLI, Other Services)        │
│   (Node.js, Python, Java, PHP, .NET, Ruby, Go, etc.)  │
└────────────────────┬────────────────────────────────────┘
                     │ HTTP/REST (multipart/form-data)
                     │
┌────────────────────▼────────────────────────────────────┐
│      Sereni Storage Provider (Port 8083)                │
│  ┌────────────────────────────────────────────────┐    │
│  │  HTTP Layer (Gin Router)                       │    │
│  │  • CORS Middleware                             │    │
│  │  • Rate Limiting                               │    │
│  │  • Request ID Tracking                         │    │
│  │  • Swagger Documentation                       │    │
│  └────────────┬───────────────────────────────────┘    │
│               │                                         │
│  ┌────────────▼──────────┐                              │
│  │  API Handlers         │                              │
│  │  • Upload Handler     │                              │
│  │  • Download Handler   │                              │
│  │  • Delete Handler     │                              │
│  │  • Health Handler     │                              │
│  └────────────┬──────────┘                              │
│               │                                         │
│  ┌────────────▼──────────┐                              │
│  │  Storage Service      │                              │
│  │  • Path Normalization │                              │
│  │  • Content-Type       │                              │
│  │    Detection          │                              │
│  │  • File Validation    │                              │
│  └────────────┬──────────┘                              │
│               │                                         │
│  ┌────────────▼──────────┐                              │
│  │  Storage Factory      │                              │
│  │  (Provider Selector)  │                              │
│  └────────────┬──────────┘                              │
└────────────────┼──────────────────────────────────────┘
                 │
    ┌────────────┼─────────────┐
    │            │             │
    ▼            ▼             ▼
┌─────────┐  ┌─────────┐  ┌─────────┐
│  Local  │  │   S3    │  │  MinIO  │
│Filesystem│  │ (AWS)   │  │ Storage │
└─────────┘  └─────────┘  └─────────┘
```

### Component Responsibilities

| Component | Responsibility |
|-----------|-----------------|
| **HTTP Layer** | Routes requests, handles CORS, rate limiting, request tracking, serves Swagger docs |
| **API Handlers** | Validates requests, enforces file size limits, formats responses, handles errors |
| **Storage Service** | Normalizes file paths, detects content types, manages file operations |
| **Storage Factory** | Selects appropriate storage provider based on configuration |
| **Storage Providers** | Implements backend-specific logic for Local/S3/MinIO storage |

### Design Patterns

**Strategy Pattern**: Storage providers implement a common interface, allowing runtime backend switching without code changes.

**Factory Pattern**: Storage factory creates appropriate provider instance based on configuration (Local, S3, or MinIO).

**Middleware Pattern**: Request processing pipeline with rate limiting, request ID tracking, and CORS handling.

**Service Layer Pattern**: Business logic separated from HTTP handlers for testability and reusability.

## Development

### Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go               # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/             # HTTP handlers
│   │   ├── middleware/           # Request middleware
│   │   └── routes/               # Route definitions
│   ├── config/
│   │   └── config.go             # Configuration loading
│   ├── services/
│   │   └── storage_service.go    # Business logic
│   ├── providers/
│   │   └── storage/
│   │       ├── factory.go        # Provider factory
│   │       ├── local/            # Local filesystem provider
│   │       ├── s3/               # AWS S3 provider
│   │       └── minio/            # MinIO provider
│   ├── utils/
│   │   └── file/                 # File utilities
│   └── app-errors/               # Custom errors
├── tests/                        # Test files
├── docs/                         # Swagger documentation
├── uploads/                      # Local storage directory
├── docker-compose.yml            # Multi-container setup
├── Dockerfile                    # Docker image definition
├── Makefile                      # Build automation
├── example.env                   # Environment template
├── go.mod                        # Go module definition
└── README.md                     # This file
```

### Development Setup

```bash
# 1. Clone repository
git clone https://github.com/yourusername/sereni-storage-provider.git
cd sereni-storage-provider

# 2. Install Go 1.24.4+ (if not installed)
# Download from: https://golang.org/dl/

# 3. Install dependencies
go mod download

# 4. Install Swagger CLI
go install github.com/swaggo/swag/cmd/swag@latest

# 5. Copy environment file
cp example.env .env
nano .env  # Configure as needed

# 6. Generate Swagger documentation
swag init -g cmd/server/main.go

# 7. Run tests
go test -v ./...

# 8. Run application
go run cmd/server/main.go
```

### Running Tests

```bash
# All tests
go test -v ./...

# With coverage
go test -v -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Specific package
go test -v ./internal/services/

# Race detection
go test -v -race ./...
```

### Building for Production

```bash
# Build binary
go build -o bin/storage-provider cmd/server/main.go

# Build optimized binary
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o bin/storage-provider cmd/server/main.go

# Build Docker image
docker build -t sereni-storage-provider:1.0.0 .

# Build multi-platform images
docker buildx build --platform linux/amd64,linux/arm64 -t sereni-storage-provider:1.0.0 .
```

## Troubleshooting

### Common Issues

#### 1. File Upload Fails - Size Too Large

**Error:**
```json
{"success":false,"message":"File too large","error":"file size exceeds maximum allowed (10MB)"}
```

**Solution:**
```bash
# Increase MAX_UPLOAD_SIZE_BYTES in .env (in bytes)
MAX_UPLOAD_SIZE_BYTES=52428800  # 50MB

# Restart service
docker-compose restart
```

#### 2. Local Storage - Directory Not Found

**Error:**
```
Failed to initialize storage provider: directory does not exist
```

**Solution:**
```bash
# Create uploads directory
mkdir -p uploads

# Or set different path in .env
STORAGE_DEV_PATH=./data/uploads

# Restart service
docker-compose restart
```

#### 3. S3/MinIO Connection Failed

**Error:**
```
Failed to initialize storage provider: connection refused
```

**Solution:**
```bash
# Verify credentials in .env
AWS_REGION=us-east-1
AWS_BUCKET=my-bucket
AWS_ACCESS_KEY=your_access_key
AWS_SECRET_KEY=your_secret_key

# Test S3 connection
aws s3 ls s3://my-bucket

# For MinIO, verify endpoint is accessible
curl http://localhost:9000/minio/health/live
```

#### 4. CORS Error in Browser

**Error:**
```
Access to XMLHttpRequest blocked by CORS policy
```

**Solution:**
```bash
# Add your frontend URL to ALLOWED_ORIGINS in .env
ALLOWED_ORIGINS=http://localhost:3000,https://yourapp.com

# Restart service
docker-compose restart
```

#### 5. Health Check Fails

**Error:**
```
curl: (7) Failed to connect to localhost port 8083
```

**Solution:**
```bash
# Check if service is running
docker ps

# Check logs
docker logs sereni-storage-provider

# Restart service
docker-compose up -d

# Verify
curl http://localhost:8083/health
```

## FAQ

**Q: Can I use this in production?**
A: Yes! Designed for production with Docker support, health checks, rate limiting, comprehensive error handling, and battle-tested storage backends (S3, MinIO).

**Q: Which storage backend should I use?**
A: **Local:** Development/simple apps. **MinIO:** Self-hosted, data sovereignty, cost-sensitive. **S3:** Production, global CDN, high availability. Switch anytime by changing STORAGE_DRIVER.

**Q: Does it work with languages other than Go?**
A: Absolutely! REST API works with any language that can make HTTP requests and send multipart/form-data (Node.js, Python, Java, PHP, .NET, Ruby, etc.). See [Integration Guide](#integration-guide).

**Q: Can I switch storage backends without downtime?**
A: Yes for S3↔MinIO (both S3-compatible). For Local↔S3/MinIO, you'll need to migrate existing files. Use tools like `aws s3 sync` or scripts to copy files between backends.

**Q: How do I migrate from local to S3?**
A: 1) Set up S3 bucket, 2) Use AWS CLI to sync: `aws s3 sync ./uploads s3://my-bucket/`, 3) Change STORAGE_DRIVER to s3, 4) Restart service.

**Q: What's the maximum file size?**
A: Configurable via MAX_UPLOAD_SIZE_BYTES (default: 10MB). For S3/MinIO, backend limits apply (S3: 5GB per PUT, use multipart for larger).

**Q: How do I handle large file uploads?**
A: Increase MAX_UPLOAD_SIZE_BYTES and consider chunked/multipart uploads for files >100MB. For very large files (>1GB), use direct S3 presigned URLs.

**Q: Can I restrict file types?**
A: Yes! Modify `allowedTypes` in main.go (currently allows all). Add validation: `allowedTypes := []string{"image/jpeg", "image/png", "application/pdf"}`.

**Q: How do I enable HTTPS?**
A: Use reverse proxy (Nginx, Traefik, Caddy) with SSL termination, or deploy behind cloud load balancer (AWS ALB, GCP Load Balancer) with TLS.

**Q: Does it support file versioning?**
A: Not built-in. S3/MinIO support versioning at bucket level - enable in your bucket settings.

**Q: How do I report issues?**
A: Open an issue on GitHub with description, steps to reproduce, logs, and configuration (redact credentials).

## License

This project is licensed under the **Apache License 2.0**.

Full license text: See [LICENSE](LICENSE) file in repository.

---

**Made with ❤️ by the Sereni Team**
