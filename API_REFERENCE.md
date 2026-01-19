# Pulse Control Plane - API Reference (Phase 2)

## Base URL
```
http://localhost:8081
```

---

## Authentication

### API Key Authentication
For token endpoints, include the Pulse API Key in the request header:

```
X-Pulse-Key: pulse_key_xxxxxxxxxxxxx
```

Or use Bearer token format:
```
Authorization: Bearer pulse_key_xxxxxxxxxxxxx
```

---

## Health & Status

### Health Check
```http
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "service": "pulse-control-plane",
  "version": "1.0.0"
}
```

### System Status
```http
GET /v1/status
```

**Response:**
```json
{
  "status": "operational",
  "message": "Pulse Control Plane is running"
}
```

---

## Organizations

### Create Organization
```http
POST /v1/organizations
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "Acme Corporation",
  "admin_email": "admin@acme.com",
  "plan": "Pro"
}
```

**Validation:**
- `name`: required, 3-100 characters
- `admin_email`: required, valid email format
- `plan`: optional, one of: `Free`, `Pro`, `Enterprise` (default: `Free`)

**Response (201 Created):**
```json
{
  "id": "60d5ec49f1b2c8b4f8a1b2c3",
  "name": "Acme Corporation",
  "admin_email": "admin@acme.com",
  "plan": "Pro",
  "created_at": "2025-01-19T10:30:00Z",
  "updated_at": "2025-01-19T10:30:00Z"
}
```

---

### List Organizations
```http
GET /v1/organizations?page=1&limit=10&search=acme
```

**Query Parameters:**
- `page`: page number (default: 1)
- `limit`: items per page, max 100 (default: 10)
- `search`: search by name or email (optional)

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": "60d5ec49f1b2c8b4f8a1b2c3",
      "name": "Acme Corporation",
      "admin_email": "admin@acme.com",
      "plan": "Pro",
      "created_at": "2025-01-19T10:30:00Z",
      "updated_at": "2025-01-19T10:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 1,
    "total_pages": 1
  }
}
```

---

### Get Organization
```http
GET /v1/organizations/:id
```

**Response (200 OK):**
```json
{
  "id": "60d5ec49f1b2c8b4f8a1b2c3",
  "name": "Acme Corporation",
  "admin_email": "admin@acme.com",
  "plan": "Pro",
  "created_at": "2025-01-19T10:30:00Z",
  "updated_at": "2025-01-19T10:30:00Z"
}
```

---

### Update Organization
```http
PUT /v1/organizations/:id
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "Acme Corp Updated",
  "plan": "Enterprise"
}
```

**Note:** All fields are optional. Only provided fields will be updated.

**Response (200 OK):**
```json
{
  "id": "60d5ec49f1b2c8b4f8a1b2c3",
  "name": "Acme Corp Updated",
  "admin_email": "admin@acme.com",
  "plan": "Enterprise",
  "created_at": "2025-01-19T10:30:00Z",
  "updated_at": "2025-01-19T11:00:00Z"
}
```

---

### Delete Organization
```http
DELETE /v1/organizations/:id
```

**Note:** This is a soft delete. The organization is marked as deleted but not removed from the database.

**Response (200 OK):**
```json
{
  "message": "Organization deleted successfully"
}
```

---

## Projects

### Create Project
```http
POST /v1/projects?org_id=60d5ec49f1b2c8b4f8a1b2c3
Content-Type: application/json
```

**Query Parameters:**
- `org_id`: Organization ID (required)

**Request Body:**
```json
{
  "name": "My Video App",
  "region": "us-east",
  "webhook_url": "https://myapp.com/webhooks",
  "storage_config": {
    "provider": "r2",
    "bucket": "my-recordings",
    "access_key_id": "your_key",
    "secret_access_key": "your_secret",
    "region": "auto"
  }
}
```

**Validation:**
- `name`: required, 3-100 characters
- `region`: required, one of: `us-east`, `us-west`, `eu-west`, `eu-central`, `asia-south`, `asia-east`
- `webhook_url`: optional, valid URL format
- `storage_config`: optional

**Response (201 Created):**
```json
{
  "project": {
    "id": "60d5ec49f1b2c8b4f8a1b2c4",
    "org_id": "60d5ec49f1b2c8b4f8a1b2c3",
    "name": "My Video App",
    "pulse_api_key": "pulse_key_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
    "webhook_url": "https://myapp.com/webhooks",
    "livekit_url": "wss://livekit-us-east.pulse.io",
    "region": "us-east",
    "chat_enabled": false,
    "video_enabled": true,
    "activity_feed_enabled": false,
    "moderation_enabled": false,
    "created_at": "2025-01-19T10:35:00Z",
    "updated_at": "2025-01-19T10:35:00Z"
  },
  "api_secret": "pulse_secret_x1y2z3a4b5c6d7e8f9g0h1i2j3k4l5m6n7o8p9q0r1s2t3u4v5w6x7y8z9a0b1c2",
  "message": "⚠️ IMPORTANT: Save your API secret now. It won't be shown again."
}
```

**⚠️ Important:** The `api_secret` is returned only once. Save it securely.

---

### List Projects
```http
GET /v1/projects?org_id=60d5ec49f1b2c8b4f8a1b2c3&page=1&limit=10&search=video
```

**Query Parameters:**
- `org_id`: filter by organization (optional)
- `page`: page number (default: 1)
- `limit`: items per page, max 100 (default: 10)
- `search`: search by project name (optional)

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": "60d5ec49f1b2c8b4f8a1b2c4",
      "org_id": "60d5ec49f1b2c8b4f8a1b2c3",
      "name": "My Video App",
      "pulse_api_key": "pulse_key_a1b2c3d4...",
      "webhook_url": "https://myapp.com/webhooks",
      "livekit_url": "wss://livekit-us-east.pulse.io",
      "region": "us-east",
      "video_enabled": true,
      "created_at": "2025-01-19T10:35:00Z",
      "updated_at": "2025-01-19T10:35:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 1,
    "total_pages": 1
  }
}
```

**Note:** API secrets are never returned in list/get responses for security.

---

### Get Project
```http
GET /v1/projects/:id
```

**Response (200 OK):**
```json
{
  "id": "60d5ec49f1b2c8b4f8a1b2c4",
  "org_id": "60d5ec49f1b2c8b4f8a1b2c3",
  "name": "My Video App",
  "pulse_api_key": "pulse_key_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
  "webhook_url": "https://myapp.com/webhooks",
  "livekit_url": "wss://livekit-us-east.pulse.io",
  "region": "us-east",
  "chat_enabled": false,
  "video_enabled": true,
  "activity_feed_enabled": false,
  "moderation_enabled": false,
  "created_at": "2025-01-19T10:35:00Z",
  "updated_at": "2025-01-19T10:35:00Z"
}
```

---

### Update Project
```http
PUT /v1/projects/:id
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "My Updated Video App",
  "webhook_url": "https://myapp.com/new-webhooks",
  "storage_config": {
    "provider": "s3",
    "bucket": "new-bucket"
  }
}
```

**Note:** All fields are optional. Only provided fields will be updated.

**Response (200 OK):**
```json
{
  "id": "60d5ec49f1b2c8b4f8a1b2c4",
  "name": "My Updated Video App",
  "webhook_url": "https://myapp.com/new-webhooks",
  ...
}
```

---

### Delete Project
```http
DELETE /v1/projects/:id
```

**Note:** This is a soft delete. The project is marked as deleted but not removed.

**Response (200 OK):**
```json
{
  "message": "Project deleted successfully"
}
```

---

### Regenerate API Keys
```http
POST /v1/projects/:id/regenerate-keys
```

**⚠️ Warning:** This invalidates the old API keys immediately.

**Response (200 OK):**
```json
{
  "pulse_api_key": "pulse_key_n1e2w3k4e5y6...",
  "pulse_api_secret": "pulse_secret_n1e2w3s4e5c6...",
  "message": "⚠️ IMPORTANT: Save your new API secret now. It won't be shown again. Your old keys are now invalid."
}
```

---

## Tokens

**Authentication Required:** All token endpoints require `X-Pulse-Key` header.

### Create Token
```http
POST /v1/tokens/create
Content-Type: application/json
X-Pulse-Key: pulse_key_xxxxxxxxxxxxx
```

**Request Body:**
```json
{
  "room_name": "live-stream-room",
  "participant_name": "john_doe",
  "can_publish": true,
  "can_subscribe": true,
  "metadata": {
    "user_id": "12345",
    "role": "host"
  }
}
```

**Validation:**
- `room_name`: required
- `participant_name`: required
- `can_publish`: optional (default: true)
- `can_subscribe`: optional (default: true)
- `metadata`: optional, key-value pairs

**Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJqb2huX2RvZSIsImlzcyI6InB1bHNlLWNvbnRyb2wtcGxhbmUiLCJleHAiOjE3MDU2NzY0MDAsImlhdCI6MTcwNTY2MjAwMCwibmJmIjoxNzA1NjYyMDAwLCJ2aWRlbyI6eyJyb29tSm9pbiI6dHJ1ZSwicm9vbU5hbWUiOiJsaXZlLXN0cmVhbS1yb29tIiwiY2FuUHVibGlzaCI6dHJ1ZSwiY2FuU3Vic2NyaWJlIjp0cnVlfSwibWV0YWRhdGEiOnsicHJvamVjdF9pZCI6IjYwZDVlYzQ5ZjFiMmM4YjRmOGExYjJjNCIsIm9yZ19pZCI6IjYwZDVlYzQ5ZjFiMmM4YjRmOGExYjJjMyIsInVzZXJfaWQiOiIxMjM0NSIsInJvbGUiOiJob3N0In19.abc123xyz456",
  "server_url": "wss://livekit-us-east.pulse.io",
  "expires_at": "2025-01-19T14:40:00Z",
  "project_id": "60d5ec49f1b2c8b4f8a1b2c4",
  "room_name": "live-stream-room",
  "participant_name": "john_doe"
}
```

**Token Claims:**
- `sub`: Participant name
- `iss`: "pulse-control-plane"
- `exp`: Expiration time (4 hours from creation)
- `iat`: Issued at time
- `nbf`: Not before time
- `video`: Room permissions
- `metadata`: Project and custom metadata

---

### Validate Token
```http
POST /v1/tokens/validate
Content-Type: application/json
X-Pulse-Key: pulse_key_xxxxxxxxxxxxx
```

**Request Body:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (200 OK) - Valid Token:**
```json
{
  "valid": true,
  "info": {
    "subject": "john_doe",
    "issuer": "pulse-control-plane",
    "expires_at": "2025-01-19T14:40:00Z",
    "metadata": {
      "project_id": "60d5ec49f1b2c8b4f8a1b2c4",
      "org_id": "60d5ec49f1b2c8b4f8a1b2c3",
      "user_id": "12345",
      "role": "host"
    },
    "video_grant": {
      "roomJoin": true,
      "roomName": "live-stream-room",
      "canPublish": true,
      "canSubscribe": true
    }
  }
}
```

**Response (401 Unauthorized) - Invalid Token:**
```json
{
  "valid": false,
  "error": "Invalid or expired token"
}
```

---

## Rate Limits

### Global Rate Limit (IP-based)
- **Limit:** 100 requests per minute per IP address
- **Applies to:** All endpoints

### Project Rate Limit (API Key-based)
- **Limit:** 1,000 requests per minute per project
- **Applies to:** Token endpoints (authenticated routes)

### Response when limit exceeded:
```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/json
```

```json
{
  "error": "Rate limit exceeded. Please try again later."
}
```

---

## Error Responses

### 400 Bad Request
```json
{
  "error": "Invalid input: name is required"
}
```

### 401 Unauthorized
```json
{
  "error": "Missing API key. Please provide X-Pulse-Key header"
}
```

### 404 Not Found
```json
{
  "error": "Organization not found"
}
```

### 429 Too Many Requests
```json
{
  "error": "Rate limit exceeded for this project"
}
```

### 500 Internal Server Error
```json
{
  "error": "Failed to retrieve organizations"
}
```

---

## Complete Example Workflow

### 1. Create an Organization
```bash
curl -X POST http://localhost:8081/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corp",
    "admin_email": "admin@acme.com",
    "plan": "Pro"
  }'
```

**Save the org `id` from response.**

---

### 2. Create a Project
```bash
curl -X POST "http://localhost:8081/v1/projects?org_id=YOUR_ORG_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Video Conferencing App",
    "region": "us-east",
    "webhook_url": "https://myapp.com/webhooks"
  }'
```

**⚠️ Save the `pulse_api_key` and `api_secret` from response!**

---

### 3. Create a Token for a User
```bash
curl -X POST http://localhost:8081/v1/tokens/create \
  -H "Content-Type: application/json" \
  -H "X-Pulse-Key: YOUR_PULSE_API_KEY" \
  -d '{
    "room_name": "meeting-room-123",
    "participant_name": "alice",
    "can_publish": true,
    "can_subscribe": true,
    "metadata": {
      "user_id": "user_123",
      "role": "host"
    }
  }'
```

**Use the returned token to connect to LiveKit.**

---

### 4. Validate a Token
```bash
curl -X POST http://localhost:8081/v1/tokens/validate \
  -H "Content-Type: application/json" \
  -H "X-Pulse-Key: YOUR_PULSE_API_KEY" \
  -d '{
    "token": "YOUR_JWT_TOKEN"
  }'
```

---

## Support

For issues or questions:
- Check logs: `tail -f /var/log/supervisor/go-backend.*.log`
- Health check: `curl http://localhost:8081/health`
- MongoDB status: `sudo supervisorctl status mongodb`

---

*API Version: 1.0.0*  
*Last Updated: 2025-01-19*
