# MinIO to RustFS Migration Guide (Simple)

This file explains how to remove MinIO completely and switch to RustFS in a Go storage service.

Use this when you want one S3-compatible backend (RustFS) and no MinIO code left.

## Goal

- Remove MinIO from code, config, tests, and Docker
- Use RustFS as S3-compatible storage
- Keep app behavior same for upload, download, delete, and size/consumption

## Quick Idea (for juniors)

Think of MinIO and RustFS as storage engines behind the same API.
If your app already uses an interface (StorageProvider), you can swap the engine by:

1. removing MinIO-specific code
2. wiring RustFS provider in factory/config
3. updating Docker/env values
4. fixing tests

## Step 1: Remove MinIO code and references

Delete MinIO provider files and imports.

What to remove:
- MinIO provider implementation file(s)
- MinIO branch from storage factory switch-case
- MinIO config struct/fields
- MinIO environment variables in env files

In this project, examples were:
- remove MinIO provider under internal/providers/storage/minio/
- remove minio case from storage factory
- remove MINIO_* config fields from internal/config/config.go

## Step 2: Use one S3-compatible client for RustFS

RustFS is S3-compatible, so use AWS SDK v2 S3 client with custom endpoint.

Important settings:
- BaseEndpoint = RustFS endpoint
- UsePathStyle = true
- credentials from env

Why this works:
- RustFS speaks S3 API
- app code can use same S3 operations (PutObject, GetObject, HeadObject, ListObjectsV2)

## Step 3: Update config and env variables

Keep only rustfs + s3 + local config.

Required RustFS env values (example):
- STORAGE_DRIVER=rustfs
- RUSTFS_ENDPOINT=http://rustfs-server:9000
- RUSTFS_ACCESS_KEY=rustfsadmin
- RUSTFS_SECRET_KEY=rustfsadmin
- RUSTFS_BUCKET=serenibase
- RUSTFS_USE_SSL=false

Important:
- endpoint must be valid URI (include http:// or https://)
- inside Docker network, use service name (rustfs-server), not localhost

## Step 4: Update docker-compose

Expose RustFS ports:
- 9000:9000 for S3 API
- 9001:9001 for RustFS console

For local dev on one disk, RustFS may fail disk safety check for erasure mode.
For local testing only, set:
- RUSTFS_UNSAFE_BYPASS_DISK_CHECK=true

Do not use that bypass in production.

## Step 5: Remove MinIO dependencies

Clean module dependencies:
- remove github.com/minio/minio-go/v7
- run go mod tidy

Confirm no MinIO package remains in go.mod/go.sum.

## Step 6: Rewrite tests

Replace MinIO mocks with AWS SDK v2 style mocks.

Test these methods at minimum:
- Upload
- Download
- Delete
- Exists
- GetSize (file and directory)
- HealthCheck

Also test not-found behavior and path-without-trailing-slash directory behavior.

## Step 7: Verify everything

Run checks:

1. go test ./...
2. docker compose up -d
3. open API docs: http://localhost:8083/swagger/index.html
4. open RustFS console: http://localhost:9001
5. test upload/download/consumption endpoints

## Common mistakes and fixes

1. Error: custom endpoint was not a valid URI
- Fix: set RUSTFS_ENDPOINT with scheme, for example http://rustfs-server:9000

2. Swagger not reachable from host
- Fix: set SERVER_HOST=0.0.0.0 in .env, then recreate app container

3. RustFS fatal disk check in local setup
- Fix (local only): set RUSTFS_UNSAFE_BYPASS_DISK_CHECK=true

4. Consumption endpoint returns failed to get size information
- Fix: handle folder paths with and without trailing slash in provider logic

## Done checklist

- [ ] No MinIO code files left
- [ ] No MINIO_* env variables left
- [ ] No MinIO dependency in go.mod
- [ ] Factory has no minio branch
- [ ] RustFS starts on 9000 and 9001
- [ ] App points to RustFS endpoint
- [ ] Tests pass

This is the clean migration path from MinIO to RustFS for this repository.
