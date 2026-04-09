# Afrisinc Assets API Documentation

## Overview

The Afrisinc Assets API is a RESTful service for managing files, images, videos, documents, and fonts with folder organization and tagging capabilities.

**Base URL:** `https://assets.afrisinc.com`

## Authentication

All API endpoints (except health checks) require an API key passed via the `X-API-Key` header:

```bash
curl -H "X-API-Key: your-api-key" \
  https://assets.afrisinc.com/api/v1/assets
```

## OpenAPI Specification

The complete API specification is available in [openapi.yaml](./openapi.yaml) in OpenAPI 3.0 format.

### View in Swagger UI

You can view the API documentation interactively using Swagger UI:

**Online (using Rapidoc/Swagger UI CDN):**
```
https://swagger.io/tools/swagger-ui/
```
Upload the `openapi.yaml` file or paste the URL.

**Local (self-hosted):**
Add to your project's server startup:

```go
import "github.com/swaggo/http-swagger"

// In your router setup:
r.Get("/api/docs/*", httpSwagger.WrapHandler)
```

## Supported File Types

### Images
- `image/jpeg` - JPEG/JPG
- `image/png` - PNG
- `image/webp` - WebP
- `image/gif` - GIF
- `image/svg+xml` - SVG

### Videos
- `video/mp4` - MP4
- `video/webm` - WebM

### Documents
- `application/pdf` - PDF

### Fonts
- `font/ttf` - TrueType
- `font/otf` - OpenType
- `font/woff` - WOFF
- `font/woff2` - WOFF2

**Max file size:** 50 MB (configurable via `UPLOAD_MAX_MB` env var)

## API Endpoints

### Health Checks (No Authentication)

```bash
# Liveness probe
GET /health/live

# Readiness probe
GET /health/ready
```

### Assets

#### List Assets
```bash
GET /api/v1/assets?folder_id=&search=&tags=tag1,tag2&page=1&page_size=50&sort_by=created_at&sort_dir=desc
```

**Query Parameters:**
- `folder_id` (UUID) - Filter by folder
- `type` (string) - Filter by asset type: `image`, `video`, `document`, `font`
- `search` (string) - Search filename
- `tags` (string) - Comma-separated tags
- `page` (int, default: 1) - Pagination page
- `page_size` (int, default: 50) - Items per page
- `sort_by` (string) - Sort field: `name`, `created_at`, `size`
- `sort_dir` (string) - Sort direction: `asc`, `desc`

#### Upload Asset
```bash
curl -X POST https://assets.afrisinc.com/api/v1/assets \
  -H "X-API-Key: your-api-key" \
  -F "file=@/path/to/image.jpg" \
  -F "folder_id=123e4567-e89b-12d3-a456-426614174000" \
  -F "tags=profile,important,2024"
```

**Form Fields:**
- `file` (required, multipart/form-data) - The file to upload
- `folder_id` (optional, UUID) - Folder to organize into
- `tags` (optional, comma-separated) - Categorization tags

**Response (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "550e8400-e29b-41d4-a716-446655440000.jpg",
  "original_name": "image.jpg",
  "mime_type": "image/jpeg",
  "size_bytes": 245678,
  "folder_id": "123e4567-e89b-12d3-a456-426614174000",
  "tags": ["profile", "important"],
  "created_at": "2026-04-09T10:30:00Z",
  "url": "https://assets.afrisinc.com/550e8400-e29b-41d4-a716-446655440000.jpg"
}
```

#### Get Asset Details
```bash
GET /api/v1/assets/{id}
```

#### Download Asset
```bash
GET /api/v1/assets/{id}/download
```
Returns the raw file with `Content-Disposition: attachment` header.

#### Delete Asset
```bash
DELETE /api/v1/assets/{id}
```

#### Bulk Delete Assets
```bash
curl -X DELETE https://assets.afrisinc.com/api/v1/assets \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"ids": ["id1", "id2", "id3"]}'
```

**Constraints:**
- Minimum 1 asset
- Maximum 100 assets per request

#### Asset Statistics
```bash
GET /api/v1/assets/stats
```

**Response:**
```json
{
  "total_count": 150,
  "total_size_bytes": 524288000,
  "by_type": {
    "image/jpeg": 50,
    "image/png": 30,
    "video/mp4": 20,
    "application/pdf": 50
  }
}
```

### Folders

#### List Folders
```bash
GET /api/v1/folders
```

**Response:**
```json
{
  "folders": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "name": "Client Projects",
      "description": "All assets for client ABC",
      "created_at": "2026-04-09T10:30:00Z"
    }
  ]
}
```

#### Create Folder
```bash
curl -X POST https://assets.afrisinc.com/api/v1/folders \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Client ABC",
    "description": "All assets for client ABC"
  }'
```

**Constraints:**
- Name: required, max 100 characters
- Description: optional, max 500 characters

#### Get Folder
```bash
GET /api/v1/folders/{id}
```

#### Delete Folder
```bash
DELETE /api/v1/folders/{id}
```

## Error Responses

### 400 Bad Request
Invalid input or validation error.

```json
{
  "ok": false,
  "error": "file type \"text/plain\" is not allowed"
}
```

### 401 Unauthorized
Missing or invalid API key.

```json
{
  "ok": false,
  "error": "unauthorized"
}
```

### 404 Not Found
Resource not found.

```json
{
  "ok": false,
  "error": "resource not found"
}
```

### 500 Internal Server Error
Server error.

```json
{
  "ok": false,
  "error": "internal server error"
}
```

## Rate Limiting

All API endpoints are rate-limited. Check response headers for rate limit information.

## Examples

### Upload Image and Get URL

```bash
#!/bin/bash
API_KEY="your-api-key"
API_URL="https://assets.afrisinc.com"

# Upload
RESPONSE=$(curl -X POST "$API_URL/api/v1/assets" \
  -H "X-API-Key: $API_KEY" \
  -F "file=@image.jpg" \
  -F "tags=profile,avatar")

# Extract URL
URL=$(echo $RESPONSE | jq -r '.url')
echo "Asset uploaded: $URL"
```

### Search by Tags

```bash
curl "https://assets.afrisinc.com/api/v1/assets?tags=important,client-abc" \
  -H "X-API-Key: your-api-key"
```

### Filter by Folder and Type

```bash
curl "https://assets.afrisinc.com/api/v1/assets?folder_id=123e4567-e89b-12d3-a456-426614174000&type=image" \
  -H "X-API-Key: your-api-key"
```

## SDK Support

Use the OpenAPI spec with code generators:

```bash
# Generate Go client
openapi-generator-cli generate \
  -i docs/openapi.yaml \
  -g go \
  -o generated/go-client
```

Supported generators: Go, Python, JavaScript/TypeScript, Java, C#, PHP, Ruby, and more.
