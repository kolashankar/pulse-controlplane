# Pulse Control Plane - Implementation Plan

## Project Overview
**Pulse** is a GetStream.io competitor focusing on the **Control Plane** that orchestrates multi-tenancy, API security, and integration between users and underlying media engines (stream-go2, stream-cli).

### Tech Stack
- **Backend**: Go (Gin framework)
- **Frontend**: React (existing setup with Radix UI + Tailwind)
- **Database**: MongoDB
- **Media Engines**: LiveKit (via stream-go2, stream-cli)
- **Storage**: Cloudflare R2 / AWS S3
- **CDN**: Cloudflare

---

## Logo Concept
**Pulse Logo**: A simple, modern design featuring a waveform pulse icon similar to GetStream.io's style
- Icon: Simplified pulse/waveform in blue gradient (#0066FF to #00A3FF)
- Typography: "Pulse" in modern sans-serif (similar to Inter/Geist)
- Style: Clean, minimal, tech-forward

---

## Phase 1: Foundation & Core Infrastructure ✅ COMPLETED
**Duration**: Week 1-2  
**Status**: ✅ **COMPLETED** on 2026-01-19

### 1.1 Project Setup ✅
**Files Created:**
```
/app/go-backend/
├── main.go                      ✅ Created
├── go.mod                       ✅ Created
├── go.sum                       ✅ Generated
├── .env                         ✅ Created
├── README.md                    ✅ Created
├── pulse-control-plane          ✅ Binary compiled (15MB)
├── config/
│   └── config.go                ✅ Created
├── models/
│   ├── organization.go          ✅ Created
│   ├── project.go               ✅ Created
│   ├── user.go                  ✅ Created
│   └── usage_metrics.go         ✅ Created
├── database/
│   └── mongodb.go               ✅ Created
├── middleware/
│   ├── auth.go                  ✅ Created
│   └── cors.go                  ✅ Created
├── routes/
│   └── routes.go                ✅ Created
├── utils/
│   ├── crypto.go                ✅ Created
│   └── logger.go                ✅ Created
├── handlers/
│   └── health_handler.go        ✅ Created
└── services/
    ├── organization_service.go  ✅ Placeholder
    └── project_service.go       ✅ Placeholder
```

**Tasks:**
- [x] ✅ Initialize Go module with dependencies
- [x] ✅ Setup MongoDB connection with proper indexes
- [x] ✅ Create environment configuration system
- [x] ✅ Implement structured logging (zerolog)
- [x] ✅ Setup CORS middleware for React frontend
- [x] ✅ Create database models with validation
- [x] ✅ Compile and test Go backend
- [x] ✅ Configure supervisor for process management
- [x] ✅ Test endpoints (/health, /v1/status)

**Dependencies Installed:**
```go
github.com/gin-gonic/gin v1.10.0              ✅
go.mongodb.org/mongo-driver v1.13.1           ✅
github.com/joho/godotenv v1.5.1               ✅
github.com/golang-jwt/jwt/v5 v5.2.0           ✅
golang.org/x/crypto v0.18.0                   ✅
github.com/rs/zerolog v1.31.0                 ✅
github.com/go-playground/validator/v10 v10.16.0 ✅
```

### 1.2 Database Schema Implementation ✅
**Collections:**

**organizations** ✅
```go
type Organization struct {
    ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Name       string            `bson:"name" json:"name" binding:"required"`
    AdminEmail string            `bson:"admin_email" json:"admin_email" binding:"required,email"`
    Plan       string            `bson:"plan" json:"plan"` // Free, Pro, Enterprise
    CreatedAt  time.Time         `bson:"created_at" json:"created_at"`
    UpdatedAt  time.Time         `bson:"updated_at" json:"updated_at"`
}
```

**projects** ✅
```go
type Project struct {
    ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    OrgID           primitive.ObjectID `bson:"org_id" json:"org_id"`
    Name            string            `bson:"name" json:"name"`
    PulseAPIKey     string            `bson:"pulse_api_key" json:"pulse_api_key"`
    PulseAPISecret  string            `bson:"pulse_api_secret" json:"-"` // Never expose
    WebhookURL      string            `bson:"webhook_url" json:"webhook_url"`
    StorageConfig   StorageConfig     `bson:"storage_config" json:"storage_config"`
    LiveKitURL      string            `bson:"livekit_url" json:"livekit_url"`
    Region          string            `bson:"region" json:"region"` // us-east, eu-west, asia-south
    CreatedAt       time.Time         `bson:"created_at" json:"created_at"`
    UpdatedAt       time.Time         `bson:"updated_at" json:"updated_at"`
}
```

**usage_metrics**
```go
type UsageMetric struct {
    ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    ProjectID  primitive.ObjectID `bson:"project_id" json:"project_id"`
    EventType  string            `bson:"event_type" json:"event_type"`
    Value      float64           `bson:"value" json:"value"`
    Timestamp  time.Time         `bson:"timestamp" json:"timestamp"`
    Metadata   map[string]interface{} `bson:"metadata" json:"metadata"`
}
```

**Tasks:**
- [ ] Create MongoDB indexes (unique on pulse_api_key, org_id+project_id)
- [ ] Implement TTL index on usage_metrics for auto-cleanup (90 days)
- [ ] Add compound indexes for efficient querying
- [ ] Implement data validation at DB layer

---

## Phase 2: Core Control Plane APIs
**Duration**: Week 3-4

### 2.1 Organization & Project Management
**Files to Create:**
```
/app/go-backend/
├── handlers/
│   ├── organization_handler.go
│   ├── project_handler.go
│   └── health_handler.go
└── services/
    ├── organization_service.go
    └── project_service.go
```

**API Endpoints:**

**Organization Management:**
```
POST   /v1/organizations              # Create organization
GET    /v1/organizations              # List organizations
GET    /v1/organizations/:id          # Get organization details
PUT    /v1/organizations/:id          # Update organization
DELETE /v1/organizations/:id          # Delete organization
```

**Project Management:**
```
POST   /v1/projects                   # Create project (generates API keys)
GET    /v1/projects                   # List projects
GET    /v1/projects/:id               # Get project details
PUT    /v1/projects/:id               # Update project
DELETE /v1/projects/:id               # Delete project
POST   /v1/projects/:id/regenerate-keys # Regenerate API keys
```

**Tasks:**
- [ ] Implement organization CRUD operations
- [ ] Implement project CRUD operations
- [ ] Generate secure Pulse API keys (pulse_key_*, pulse_secret_*)
- [ ] Hash and store API secrets securely (bcrypt)
- [ ] Implement soft delete for data retention
- [ ] Add pagination and filtering

### 2.2 Authentication & Token Management
**Files to Create:**
```
/app/go-backend/
├── handlers/
│   └── token_handler.go
├── services/
│   └── token_service.go
└── middleware/
    └── project_auth.go
```

**API Endpoints:**
```
POST   /v1/tokens/create              # Exchange Pulse Key for Media Token
POST   /v1/tokens/validate            # Validate existing token
```

**Token Service Features:**
- [ ] Validate Pulse API key from request header
- [ ] Generate LiveKit JWT tokens with scoped permissions
- [ ] Support room-level permissions (join, publish, subscribe)
- [ ] Attach project_id to token metadata for tracking
- [ ] Set configurable token expiry (default 4 hours)
- [ ] Implement token refresh mechanism

**Tasks:**
- [ ] Implement middleware to authenticate requests via X-Pulse-Key header
- [ ] Create token generation service using LiveKit SDK
- [ ] Add support for custom token claims
- [ ] Implement rate limiting per project

---

## Phase 3: Media Control & Scaling
**Duration**: Week 5-6

### 3.1 Egress & HLS Distribution
**Files to Create:**
```
/app/go-backend/
├── handlers/
│   ├── egress_handler.go
│   └── ingress_handler.go
└── services/
    ├── egress_service.go
    ├── ingress_service.go
    └── cdn_service.go
```

**API Endpoints:**
```
POST   /v1/media/egress/start         # Start HLS stream for lakhs of viewers
POST   /v1/media/egress/stop          # Stop egress
GET    /v1/media/egress/:id           # Get egress status
POST   /v1/media/ingress/create       # Create ingress endpoint
DELETE /v1/media/ingress/:id          # Delete ingress
```

**Egress Features:**
- [ ] Convert WebRTC to LL-HLS for CDN distribution
- [ ] Support multiple output formats (HLS, RTMP, file)
- [ ] Implement room composite (speaker layout, grid layout)
- [ ] Push HLS segments to Cloudflare R2/S3
- [ ] Generate CDN playback URLs
- [ ] Handle egress lifecycle (started, ended, failed)

**Tasks:**
- [ ] Integrate LiveKit Egress SDK
- [ ] Implement HLS streaming to CDN
- [ ] Configure Cloudflare R2 for storage
- [ ] Add support for recording to cloud storage
- [ ] Implement egress status webhooks

### 3.2 Webhook System
**Files to Create:**
```
/app/go-backend/
├── handlers/
│   └── webhook_handler.go
├── services/
│   └── webhook_service.go
└── queue/
    └── retry_queue.go
```

**Webhook Events:**
- participant_joined
- participant_left
- room_started
- room_ended
- egress_started
- egress_ended
- recording_available

**Tasks:**
- [ ] Internal webhook listener for LiveKit events
- [ ] Forward events to customer webhook URLs
- [ ] Implement retry logic with exponential backoff (5, 10, 30 mins)
- [ ] Use Redis-backed queue for reliability
- [ ] Sign webhooks with HMAC for security
- [ ] Track webhook delivery success/failure
- [ ] Store webhook logs for debugging

---

## Phase 4: Usage Tracking & Billing
**Duration**: Week 7

### 4.1 Usage Metrics Collection
**Files to Create:**
```
/app/go-backend/
├── handlers/
│   └── usage_handler.go
├── services/
│   ├── usage_service.go
│   └── aggregator_service.go
└── workers/
    └── usage_aggregator.go
```

**API Endpoints:**
```
GET    /v1/usage/:project_id          # Get usage metrics
GET    /v1/usage/:project_id/summary  # Get aggregated summary
```

**Metrics to Track:**
- [ ] Participant minutes (per room)
- [ ] Egress minutes (streaming/recording)
- [ ] Storage usage (GB)
- [ ] Bandwidth usage (GB)
- [ ] API requests count

**Tasks:**
- [ ] Implement real-time usage tracking from webhooks
- [ ] Create background worker for hourly aggregation
- [ ] Calculate billing totals per project
- [ ] Store aggregated metrics for reporting
- [ ] Implement usage limits per plan (Free/Pro/Enterprise)
- [ ] Send alerts when approaching limits

### 4.2 Billing Integration (Placeholder)
**Files to Create:**
```
/app/go-backend/
├── handlers/
│   └── billing_handler.go
└── services/
    └── billing_service.go
```

**Tasks:**
- [ ] Design billing model (per-minute pricing)
- [ ] Create invoice generation system
- [ ] Add Stripe integration placeholder
- [ ] Implement billing dashboard API

---

## Phase 5: Admin Dashboard Features
**Duration**: Week 8-9

### 5.1 Team Management
**Files to Create:**
```
/app/go-backend/
├── models/
│   ├── team_member.go
│   └── invitation.go
├── handlers/
│   └── team_handler.go
└── services/
    └── team_service.go
```

**API Endpoints:**
```
GET    /v1/organizations/:id/members   # List team members
POST   /v1/organizations/:id/members   # Invite member
DELETE /v1/organizations/:id/members/:user_id # Remove member
PUT    /v1/organizations/:id/members/:user_id/role # Update role
```

**Tasks:**
- [ ] Implement team member management
- [ ] Add role-based access control (Owner, Admin, Developer, Viewer)
- [ ] Create invitation system with email tokens
- [ ] Implement member permissions matrix

### 5.2 Audit Logs
**Files to Create:**
```
/app/go-backend/
├── models/
│   └── audit_log.go
├── handlers/
│   └── audit_handler.go
├── services/
│   └── audit_service.go
└── middleware/
    └── audit_middleware.go
```

**API Endpoints:**
```
GET    /v1/audit-logs                 # Get audit logs
GET    /v1/audit-logs/export          # Export logs (CSV)
```

**Events to Log:**
- [ ] Project created/updated/deleted
- [ ] API key regenerated
- [ ] Team member added/removed
- [ ] Settings changed
- [ ] Webhook configuration changed

**Tasks:**
- [ ] Implement audit logging middleware
- [ ] Store user IP, timestamp, action, resource
- [ ] Add filtering by date, user, action type
- [ ] Implement log retention policy (1 year)

### 5.3 Status & Monitoring
**Files to Create:**
```
/app/go-backend/
├── handlers/
│   └── status_handler.go
└── services/
    └── status_service.go
```

**API Endpoints:**
```
GET    /v1/status                     # System status
GET    /v1/status/projects/:id        # Project health
GET    /v1/status/regions             # Region availability
```

**Tasks:**
- [ ] Implement health check endpoints
- [ ] Monitor LiveKit server status
- [ ] Check database connectivity
- [ ] Track API response times
- [ ] Display service status on dashboard

---

## Phase 6: Frontend Dashboard (React)
**Duration**: Week 10-11

### 6.1 Update Frontend for Go Backend
**Files to Create/Modify:**
```
/app/frontend/src/
├── api/
│   ├── client.js                     # Axios client with Pulse Key auth
│   ├── organizations.js              # Organization API calls
│   ├── projects.js                   # Project API calls
│   ├── tokens.js                     # Token API calls
│   └── usage.js                      # Usage API calls
├── pages/
│   ├── Dashboard.js                  # Main dashboard
│   ├── Organizations.js              # Org management
│   ├── Projects.js                   # Project management
│   ├── ProjectDetails.js             # Single project view
│   ├── Billing.js                    # Billing & usage
│   ├── Team.js                       # Team management
│   ├── AuditLogs.js                  # Audit logs viewer
│   └── Status.js                     # System status
├── components/
│   ├── Sidebar.js                    # Navigation sidebar
│   ├── ProjectCard.js                # Project card component
│   ├── APIKeyDisplay.js              # Secure key display
│   ├── UsageChart.js                 # Usage visualization
│   └── Logo.js                       # Pulse logo component
└── contexts/
    └── AuthContext.js                # Authentication context
```

**Pages to Build:**

**1. Dashboard** (Main Landing)
- Welcome message
- Quick stats (projects, usage, team size)
- Recent activity feed
- Quick actions (Create Project, View Docs)

**2. Apps/Projects Page** (like GetStream screenshot)
- List all projects with cards
- Display project ID, name, region
- Show enabled features (Chat, Video, Activity Feeds, Moderation)
- Create new project button
- Search and filter projects

**3. Project Details Page**
- API keys section (show/hide, regenerate)
- Configuration settings
- Region selector
- Webhook URL configuration
- Storage settings (R2/S3 credentials)
- Delete project option

**4. Chat Messaging** (Feature Panel)
- Enable/disable chat feature
- Configure chat settings
- View chat usage metrics
- Link to documentation

**5. Video & Audio** (Feature Panel)
- Enable/disable video/audio
- Configure room settings
- Egress configuration
- View streaming analytics

**6. Activity Feeds** (Feature Panel)
- Enable/disable activity feeds
- Configure feed types
- View feed activity

**7. Moderation** (Feature Panel)
- Enable/disable moderation
- Configure moderation rules
- View moderation logs

**8. Billing Page**
- Current plan display
- Usage breakdown (participant minutes, egress, storage)
- Usage charts (daily/monthly)
- Invoice history
- Upgrade/downgrade plan

**9. Team Page**
- List team members with roles
- Invite new members
- Manage permissions
- Pending invitations

**10. Audit Logs Page**
- Filterable log table
- Search by user, action, date
- Export logs button

**11. Status Page**
- System health indicators
- Region status
- Recent incidents
- API status

**Tasks:**
- [ ] Update API client to use Go backend URL
- [ ] Implement authentication flow
- [ ] Create all dashboard pages
- [ ] Add routing with react-router-dom
- [ ] Implement responsive design
- [ ] Add loading states and error handling
- [ ] Create reusable components

### 6.2 Logo Implementation
**Tasks:**
- [ ] Create SVG logo component
- [ ] Add logo to header/sidebar
- [ ] Create favicon
- [ ] Add logo to login/signup pages

---

## Phase 7: Security & Production Readiness
**Duration**: Week 12

### 7.1 Security Hardening
**Tasks:**
- [ ] Implement rate limiting (per IP, per API key)
- [ ] Add request validation middleware
- [ ] Implement CORS properly for frontend
- [ ] Hash sensitive data (API secrets)
- [ ] Add HTTPS enforcement
- [ ] Implement webhook signature verification
- [ ] Add SQL injection protection (MongoDB)
- [ ] Implement XSS protection

### 7.2 Testing
**Files to Create:**
```
/app/go-backend/
├── tests/
│   ├── handlers_test.go
│   ├── services_test.go
│   └── integration_test.go
```

**Tasks:**
- [ ] Write unit tests for all handlers
- [ ] Write integration tests for API endpoints
- [ ] Test webhook delivery and retries
- [ ] Load test with k6 (simulate lakhs of requests)
- [ ] Test MongoDB connection pooling
- [ ] Test token generation and validation

### 7.3 Documentation
**Files to Create:**
```
/app/
├── docs/
│   ├── API.md                        # API reference
│   ├── QUICKSTART.md                 # Quick start guide
│   ├── AUTHENTICATION.md             # Auth guide
│   ├── WEBHOOKS.md                   # Webhook guide
│   └── SCALING.md                    # Scaling guide
```

**Tasks:**
- [ ] Write API documentation
- [ ] Create integration examples
- [ ] Document environment variables
- [ ] Create deployment guide
- [ ] Add code examples (Go, JavaScript, Python)

### 7.4 Deployment
**Tasks:**
- [ ] Create Dockerfile for Go backend
- [ ] Setup supervisor for Go process
- [ ] Configure environment variables
- [ ] Setup MongoDB indexes
- [ ] Configure logging
- [ ] Setup monitoring (optional: Prometheus/Grafana)

---

## Phase 8: Advanced Features (Post-MVP)
**Duration**: Week 13+

### 8.1 Multi-Region Support
- [ ] Implement region-aware token generation
- [ ] Route users to nearest LiveKit server
- [ ] Add region failover logic
- [ ] Display region latency in dashboard

### 8.2 Advanced Analytics
- [ ] Real-time analytics dashboard
- [ ] Custom metrics and alerts
- [ ] Export analytics data
- [ ] Predictive usage forecasting

### 8.3 Developer Tools
- [ ] API playground
- [ ] SDK generation (Go, JavaScript, Python)
- [ ] Postman collection
- [ ] Interactive API docs (Swagger)

### 8.4 Enterprise Features
- [ ] SSO integration (SAML, OAuth)
- [ ] Custom SLAs
- [ ] Dedicated support
- [ ] Private cloud deployment

---

## Dependencies Summary

### Go Backend Dependencies
```go
// Core
github.com/gin-gonic/gin                    // Web framework
go.mongodb.org/mongo-driver/mongo           // MongoDB driver

// LiveKit Integration
github.com/livekit/protocol                 // LiveKit protocol
github.com/livekit/server-sdk-go/v2         // LiveKit SDK

// Authentication & Security
github.com/golang-jwt/jwt/v5                // JWT tokens
golang.org/x/crypto/bcrypt                  // Password hashing

// Configuration & Utils
github.com/joho/godotenv                    // Environment variables
github.com/rs/zerolog                       // Structured logging
github.com/go-playground/validator/v10      // Input validation

// Queue & Caching (for webhooks)
github.com/go-redis/redis/v8                // Redis client

// Storage (for egress)
github.com/aws/aws-sdk-go/service/s3        // S3/R2 client
```

### Frontend Dependencies (Already Available)
- React 19
- Radix UI (for components)
- Tailwind CSS (for styling)
- Axios (for API calls)
- React Router DOM (for navigation)
- Recharts (for analytics charts)
- Lucide React (for icons)

---

## Environment Variables

### Go Backend (.env)
```bash
# Server
PORT=8080
GIN_MODE=release

# Database
MONGO_URI=mongodb://localhost:27017
MONGO_DB_NAME=pulse_production

# LiveKit
LIVEKIT_HOST=wss://livekit.pulse.io
LIVEKIT_API_KEY=APIxxx
LIVEKIT_API_SECRET=SECRETxxx

# CDN & Storage
CDN_INGEST_URL=rtmp://stream.pulse.io/live
CDN_PLAYBACK_URL=https://cdn.pulse.io/hls
R2_ACCOUNT_ID=xxx
R2_ACCESS_KEY_ID=xxx
R2_SECRET_ACCESS_KEY=xxx
R2_BUCKET_NAME=pulse-recordings

# Redis (for webhook queue)
REDIS_URL=redis://localhost:6379

# CORS
CORS_ORIGINS=http://localhost:3000,https://app.pulse.io

# Security
JWT_SECRET=your-secret-key
API_KEY_PEPPER=random-pepper-string
```

### Frontend (.env)
```bash
REACT_APP_BACKEND_URL=http://localhost:8080
REACT_APP_API_VERSION=v1
```

---

## Success Metrics

### MVP Success Criteria
- [ ] Can create organizations and projects via dashboard
- [ ] Can generate and manage Pulse API keys
- [ ] Can issue LiveKit tokens for WebRTC connections
- [ ] Can start egress for HLS streaming
- [ ] Can track usage metrics per project
- [ ] Webhooks are delivered reliably with retries
- [ ] Dashboard displays all key features
- [ ] System can handle 100+ concurrent projects

### Production Readiness Checklist
- [ ] All API endpoints have proper error handling
- [ ] All inputs are validated
- [ ] All secrets are hashed/encrypted
- [ ] Rate limiting is implemented
- [ ] Audit logs are working
- [ ] Monitoring is setup
- [ ] Documentation is complete
- [ ] Load testing passed (1000+ req/s)

---

## Timeline Summary

| Phase | Duration | Deliverables |
|-------|----------|-------------|
| Phase 1: Foundation | 2 weeks | Go project setup, DB models, MongoDB connection |
| Phase 2: Core APIs | 2 weeks | Organization/Project CRUD, Token service, Auth middleware |
| Phase 3: Media Control | 2 weeks | Egress/Ingress APIs, Webhook system, CDN integration |
| Phase 4: Usage & Billing | 1 week | Usage tracking, Metrics aggregation, Billing placeholder |
| Phase 5: Admin Features | 2 weeks | Team management, Audit logs, Status monitoring |
| Phase 6: Frontend | 2 weeks | React dashboard, All pages, API integration |
| Phase 7: Production | 1 week | Security, Testing, Documentation, Deployment |
| Phase 8: Advanced | Ongoing | Multi-region, Analytics, SDKs, Enterprise features |

**Total MVP Timeline: 12 weeks**

---

## Next Steps

1. **Review and Approve** this implementation plan
2. **Setup Development Environment**
   - Install Go 1.21+
   - Setup MongoDB
   - Setup Redis (for webhook queue)
   - Clone LiveKit repos for reference
3. **Start Phase 1** - Foundation & Core Infrastructure
4. **Iterative Development** - Build and test each phase incrementally

---

## Notes

- **Focus on Control Plane Only**: We are NOT building media engines. We integrate with existing LiveKit engines.
- **Go Language Strictly**: All backend code will be in Go, no Python/Node.js.
- **Scalability**: Architecture designed to handle "lakhs" (100,000+) of concurrent viewers via CDN handover.
- **GetStream.io Parity**: Implementing all features visible in provided screenshots (Apps, Chat, Video, Feeds, Moderation, Billing, Team, Audit Logs).

