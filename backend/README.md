# Pulse Control Plane - Go Backend

## Overview
Pulse Control Plane is a GetStream.io competitor built with Go. This backend service manages multi-tenancy, API security, and integration between users and LiveKit media engines.

## Tech Stack
- **Language**: Go 1.19+
- **Web Framework**: Gin
- **Database**: MongoDB
- **Media Engine**: LiveKit (via SDK)
- **Storage**: Cloudflare R2 / AWS S3
- **Caching/Queue**: Redis

## Project Structure
```
go-backend/
├── main.go              # Application entry point
├── config/              # Configuration management
├── models/              # Data models
├── database/            # Database connection & indexes
├── middleware/          # HTTP middleware (auth, cors)
├── routes/              # Route definitions
├── handlers/            # HTTP request handlers
├── services/            # Business logic
├── utils/               # Utility functions
├── workers/             # Background workers
├── queue/               # Queue management
└── tests/               # Test files
```

## Prerequisites
- Go 1.19 or higher
- MongoDB 5.0+
- Redis 6.0+ (for webhook queue)
- LiveKit server (or use hosted)

## Environment Variables
Copy `.env.example` to `.env` and configure:

```bash
# Server
PORT=8080
GIN_MODE=debug
ENVIRONMENT=development

# Database
MONGO_URI=mongodb://localhost:27017
MONGO_DB_NAME=pulse_development

# LiveKit
LIVEKIT_HOST=wss://your-livekit-host
LIVEKIT_API_KEY=your_api_key
LIVEKIT_API_SECRET=your_api_secret

# Security
JWT_SECRET=your-secret-key
API_KEY_PEPPER=random-pepper-string
```

## Installation

1. **Install Dependencies**
```bash
go mod download
```

2. **Build the Application**
```bash
go build -o pulse-control-plane .
```

3. **Run the Application**
```bash
./pulse-control-plane
```

Or run directly:
```bash
go run main.go
```

## Development

### Running in Development Mode
```bash
go run main.go
```

### Running Tests
```bash
go test ./...
```

### Code Linting
```bash
golangci-lint run
```

## API Endpoints

### Health Check
```
GET /health
```

### API v1
```
GET  /v1/status
```

### Phase 2 (Organizations & Projects) - Coming Soon
```
POST   /v1/organizations
GET    /v1/organizations
GET    /v1/organizations/:id
PUT    /v1/organizations/:id
DELETE /v1/organizations/:id

POST   /v1/projects
GET    /v1/projects
GET    /v1/projects/:id
PUT    /v1/projects/:id
DELETE /v1/projects/:id
POST   /v1/projects/:id/regenerate-keys
```

### Phase 2 (Token Management) - Coming Soon
```
POST /v1/tokens/create    # Requires X-Pulse-Key header
POST /v1/tokens/validate
```

### Phase 3 (Media Control) - Coming Soon
```
POST /v1/media/egress/start
POST /v1/media/egress/stop
GET  /v1/media/egress/:id

POST   /v1/media/ingress/create
DELETE /v1/media/ingress/:id
```

## Authentication

Pulse uses API key-based authentication:

**Headers:**
```
X-Pulse-Key: pulse_key_xxxxx
X-Pulse-Secret: pulse_secret_xxxxx  (for sensitive operations)
```

Or Bearer token:
```
Authorization: Bearer pulse_key_xxxxx
```

## MongoDB Indexes

Indexes are automatically created on startup:

- **organizations**: admin_email (unique), is_deleted
- **projects**: pulse_api_key (unique), org_id, is_deleted
- **users**: email (unique), org_id
- **usage_metrics**: project_id, timestamp (TTL 90 days)

## Logging

Pulse uses structured logging with `zerolog`:

- Development: Pretty console output
- Production: JSON structured logs

Log levels: `debug`, `info`, `warn`, `error`

## Phase 1 Completion Status ✅

- [x] Project setup and structure
- [x] Go module initialization
- [x] Environment configuration system
- [x] MongoDB connection with indexes
- [x] Database models (Organization, Project, User, UsageMetrics)
- [x] Authentication middleware
- [x] CORS middleware
- [x] Cryptographic utilities (key generation, hashing)
- [x] Structured logging
- [x] Route setup (foundation)
- [x] Main application entry point
- [x] Compilation successful

## Next Steps (Phase 2)

1. Implement Organization handlers
2. Implement Project handlers with API key generation
3. Implement Token service for LiveKit JWT generation
4. Add rate limiting middleware
5. Create integration tests

## License
MIT License
