# Pulse Control Plane - API Reference

## Overview

Pulse Control Plane is a comprehensive API for managing real-time communication infrastructure. This document provides detailed information about all available endpoints, request/response formats, and authentication methods.

**Base URL**: `https://your-domain.com/api/v1`

**Version**: 1.0.0

---

## Table of Contents

1. [Authentication](#authentication)
2. [Rate Limiting](#rate-limiting)
3. [Error Handling](#error-handling)
4. [Organizations](#organizations)
5. [Projects](#projects)
6. [Tokens](#tokens)
7. [Media](#media)
   - [Egress](#egress)
   - [Ingress](#ingress)
8. [Webhooks](#webhooks)
9. [Usage & Billing](#usage--billing)
10. [Team Management](#team-management)
11. [Audit Logs](#audit-logs)
12. [Status & Monitoring](#status--monitoring)

---

## Authentication

### API Key Authentication

Most endpoints require authentication using a Pulse API key.

**Method 1: Header Authentication**
```http
GET /api/v1/tokens/create
X-Pulse-Key: pulse_key_abc123xyz789
```

**Method 2: Bearer Token**
```http
GET /api/v1/tokens/create
Authorization: Bearer pulse_key_abc123xyz789
```

### API Key + Secret Authentication

Sensitive operations require both API key and secret:

```http
POST /api/v1/projects/:id/regenerate-keys
X-Pulse-Key: pulse_key_abc123xyz789
X-Pulse-Secret: pulse_secret_def456uvw123
```

---

## Rate Limiting

### Limits

- **Per IP**: 100 requests per minute
- **Per Project**: 1,000 requests per minute

### Rate Limit Headers

```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 998
X-RateLimit-Reset: 1640000000
```

### Exceeded Rate Limit Response

```json
{
  "error": "Rate limit exceeded. Please try again later."
}
```

**Status Code**: `429 Too Many Requests`

---

## Error Handling

### Error Response Format

```json
{
  "error": "Error message describing what went wrong",
  "code": "ERROR_CODE",
  "details": {}
}
```

### HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 409 | Conflict |
| 429 | Too Many Requests |
| 500 | Internal Server Error |
| 503 | Service Unavailable |

### Common Error Codes

- `INVALID_API_KEY`: API key is invalid or missing
- `RATE_LIMIT_EXCEEDED`: Too many requests
- `RESOURCE_NOT_FOUND`: Requested resource doesn't exist
- `VALIDATION_ERROR`: Request validation failed
- `INSUFFICIENT_PERMISSIONS`: User lacks required permissions

---

## Organizations

### Create Organization

**Endpoint**: `POST /api/v1/organizations`

**Request Body**:
```json
{
  "name": "Acme Corp",
  "admin_email": "admin@acme.com",
  "plan": "Pro"
}
```

**Response**: `201 Created`
```json
{
  "id": "507f1f77bcf86cd799439011",
  "name": "Acme Corp",
  "admin_email": "admin@acme.com",
  "plan": "Pro",
  "created_at": "2025-01-19T10:30:00Z",
  "updated_at": "2025-01-19T10:30:00Z"
}
```

### List Organizations

**Endpoint**: `GET /api/v1/organizations`

**Query Parameters**:
- `limit` (optional): Number of results (default: 20, max: 100)
- `offset` (optional): Pagination offset (default: 0)
- `plan` (optional): Filter by plan (Free, Pro, Enterprise)

**Response**: `200 OK`
```json
{
  "organizations": [
    {
      "id": "507f1f77bcf86cd799439011",
      "name": "Acme Corp",
      "admin_email": "admin@acme.com",
      "plan": "Pro",
      "created_at": "2025-01-19T10:30:00Z"
    }
  ],
  "total": 1,
  "limit": 20,
  "offset": 0
}
```

### Get Organization

**Endpoint**: `GET /api/v1/organizations/:id`

**Response**: `200 OK`
```json
{
  "id": "507f1f77bcf86cd799439011",
  "name": "Acme Corp",
  "admin_email": "admin@acme.com",
  "plan": "Pro",
  "created_at": "2025-01-19T10:30:00Z",
  "updated_at": "2025-01-19T10:30:00Z"
}
```

### Update Organization

**Endpoint**: `PUT /api/v1/organizations/:id`

**Request Body**:
```json
{
  "name": "Acme Corporation",
  "plan": "Enterprise"
}
```

**Response**: `200 OK`

### Delete Organization

**Endpoint**: `DELETE /api/v1/organizations/:id`

**Response**: `200 OK`
```json
{
  "message": "Organization deleted successfully"
}
```

---

## Projects

### Create Project

**Endpoint**: `POST /api/v1/projects`

**Request Body**:
```json
{
  "name": "My Video App",
  "org_id": "507f1f77bcf86cd799439011",
  "region": "US_EAST",
  "webhook_url": "https://myapp.com/webhooks/pulse"
}
```

**Response**: `201 Created`
```json
{
  "id": "507f1f77bcf86cd799439012",
  "name": "My Video App",
  "org_id": "507f1f77bcf86cd799439011",
  "region": "US_EAST",
  "pulse_api_key": "pulse_key_abc123xyz789",
  "pulse_api_secret": "pulse_secret_def456uvw123",
  "webhook_url": "https://myapp.com/webhooks/pulse",
  "created_at": "2025-01-19T10:30:00Z"
}
```

### List Projects

**Endpoint**: `GET /api/v1/projects`

**Query Parameters**:
- `org_id` (optional): Filter by organization
- `region` (optional): Filter by region
- `limit` (optional): Number of results
- `offset` (optional): Pagination offset

**Response**: `200 OK`
```json
{
  "projects": [
    {
      "id": "507f1f77bcf86cd799439012",
      "name": "My Video App",
      "org_id": "507f1f77bcf86cd799439011",
      "region": "US_EAST",
      "pulse_api_key": "pulse_key_abc123xyz789",
      "created_at": "2025-01-19T10:30:00Z"
    }
  ],
  "total": 1
}
```

### Get Project

**Endpoint**: `GET /api/v1/projects/:id`

**Response**: `200 OK`

### Update Project

**Endpoint**: `PUT /api/v1/projects/:id`

**Request Body**:
```json
{
  "name": "My Updated App",
  "webhook_url": "https://myapp.com/new-webhook"
}
```

**Response**: `200 OK`

### Regenerate API Keys

**Endpoint**: `POST /api/v1/projects/:id/regenerate-keys`

**Authentication**: Requires both API key and secret

**Response**: `200 OK`
```json
{
  "pulse_api_key": "pulse_key_new123xyz789",
  "pulse_api_secret": "pulse_secret_new456uvw123",
  "warning": "Old credentials are immediately invalidated"
}
```

### Delete Project

**Endpoint**: `DELETE /api/v1/projects/:id`

**Response**: `200 OK`

---

## Tokens

### Create Token

**Endpoint**: `POST /api/v1/tokens/create`

**Authentication**: Required (API Key)

**Request Body**:
```json
{
  "identity": "user123",
  "room_name": "meeting-room-456",
  "metadata": {
    "name": "John Doe",
    "email": "john@example.com"
  },
  "ttl": 3600
}
```

**Response**: `200 OK`
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "identity": "user123",
  "room_name": "meeting-room-456",
  "expires_at": "2025-01-19T11:30:00Z"
}
```

### Validate Token

**Endpoint**: `POST /api/v1/tokens/validate`

**Authentication**: Required (API Key)

**Request Body**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response**: `200 OK`
```json
{
  "valid": true,
  "identity": "user123",
  "room_name": "meeting-room-456",
  "expires_at": "2025-01-19T11:30:00Z"
}
```

---

## Media

### Egress

#### Start Egress

**Endpoint**: `POST /api/v1/media/egress/start`

**Authentication**: Required (API Key)

**Request Body**:
```json
{
  "room_name": "meeting-room-456",
  "type": "RTMP",
  "output": {
    "rtmp_url": "rtmp://stream.example.com/live",
    "stream_key": "secret_key_123"
  }
}
```

**Response**: `200 OK`
```json
{
  "egress_id": "EG_abc123",
  "status": "ACTIVE",
  "started_at": "2025-01-19T10:30:00Z"
}
```

#### Stop Egress

**Endpoint**: `POST /api/v1/media/egress/stop`

**Request Body**:
```json
{
  "egress_id": "EG_abc123"
}
```

**Response**: `200 OK`

#### Get Egress

**Endpoint**: `GET /api/v1/media/egress/:id`

**Response**: `200 OK`

#### List Egresses

**Endpoint**: `GET /api/v1/media/egress`

**Query Parameters**:
- `room_name` (optional): Filter by room
- `status` (optional): Filter by status

**Response**: `200 OK`

### Ingress

#### Create Ingress

**Endpoint**: `POST /api/v1/media/ingress/create`

**Authentication**: Required (API Key)

**Request Body**:
```json
{
  "name": "RTMP Stream",
  "type": "RTMP",
  "room_name": "meeting-room-456"
}
```

**Response**: `200 OK`
```json
{
  "ingress_id": "IN_abc123",
  "url": "rtmp://ingress.pulse.io/live",
  "stream_key": "sk_xyz789",
  "status": "READY"
}
```

#### Get Ingress

**Endpoint**: `GET /api/v1/media/ingress/:id`

**Response**: `200 OK`

#### List Ingresses

**Endpoint**: `GET /api/v1/media/ingress`

**Response**: `200 OK`

#### Delete Ingress

**Endpoint**: `DELETE /api/v1/media/ingress/:id`

**Response**: `200 OK`

---

## Webhooks

### Receive LiveKit Webhook

**Endpoint**: `POST /api/v1/webhooks/livekit`

**Authentication**: Webhook signature verification

**Request Headers**:
```
X-Webhook-Signature: sha256=abc123...
Content-Type: application/json
```

**Request Body**:
```json
{
  "event": "room_started",
  "room": {
    "name": "meeting-room-456",
    "sid": "RM_abc123"
  },
  "created_at": 1640000000
}
```

**Response**: `200 OK`

### Get Webhook Logs

**Endpoint**: `GET /api/v1/webhooks/logs`

**Authentication**: Required (API Key)

**Response**: `200 OK`
```json
{
  "logs": [
    {
      "id": "507f1f77bcf86cd799439013",
      "event": "room_started",
      "room_name": "meeting-room-456",
      "status": "delivered",
      "timestamp": "2025-01-19T10:30:00Z"
    }
  ]
}
```

---

## Usage & Billing

### Get Usage Metrics

**Endpoint**: `GET /api/v1/usage/:project_id`

**Authentication**: Required (API Key)

**Query Parameters**:
- `start_date`: Start date (ISO 8601)
- `end_date`: End date (ISO 8601)

**Response**: `200 OK`
```json
{
  "project_id": "507f1f77bcf86cd799439012",
  "period": {
    "start": "2025-01-01T00:00:00Z",
    "end": "2025-01-31T23:59:59Z"
  },
  "metrics": {
    "participant_minutes": 12450,
    "egress_minutes": 320,
    "storage_gb": 12,
    "bandwidth_gb": 45
  }
}
```

### Get Usage Summary

**Endpoint**: `GET /api/v1/usage/:project_id/summary`

**Response**: `200 OK`
```json
{
  "current_month": {
    "participant_minutes": 1250,
    "cost": 5.0
  },
  "previous_month": {
    "participant_minutes": 890,
    "cost": 3.56
  },
  "change_percent": 40.4
}
```

### Get Billing Dashboard

**Endpoint**: `GET /api/v1/billing/:project_id/dashboard`

**Response**: `200 OK`
```json
{
  "current_charges": 5.0,
  "usage_breakdown": {
    "participant_minutes": 1250,
    "egress_minutes": 32,
    "storage_gb": 1.2,
    "bandwidth_gb": 4.5
  },
  "invoices": [
    {
      "id": "INV-2025-01",
      "amount": 3.56,
      "status": "Paid",
      "period": "2025-01"
    }
  ]
}
```

### Generate Invoice

**Endpoint**: `POST /api/v1/billing/:project_id/invoice`

**Response**: `201 Created`

### List Invoices

**Endpoint**: `GET /api/v1/billing/:project_id/invoices`

**Response**: `200 OK`

---

## Team Management

### List Team Members

**Endpoint**: `GET /api/v1/organizations/:id/members`

**Response**: `200 OK`
```json
{
  "members": [
    {
      "user_id": "507f1f77bcf86cd799439014",
      "email": "john@acme.com",
      "role": "Admin",
      "joined_at": "2025-01-15T10:00:00Z"
    }
  ]
}
```

### Invite Team Member

**Endpoint**: `POST /api/v1/organizations/:id/members`

**Request Body**:
```json
{
  "email": "jane@acme.com",
  "role": "Developer"
}
```

**Response**: `201 Created`
```json
{
  "invitation_id": "507f1f77bcf86cd799439015",
  "email": "jane@acme.com",
  "role": "Developer",
  "expires_at": "2025-01-26T10:00:00Z"
}
```

### Remove Team Member

**Endpoint**: `DELETE /api/v1/organizations/:id/members/:user_id`

**Response**: `200 OK`

### Update Team Member Role

**Endpoint**: `PUT /api/v1/organizations/:id/members/:user_id/role`

**Request Body**:
```json
{
  "role": "Admin"
}
```

**Response**: `200 OK`

---

## Audit Logs

### Get Audit Logs

**Endpoint**: `GET /api/v1/audit-logs`

**Query Parameters**:
- `email` (optional): Filter by user email
- `action` (optional): Filter by action type
- `status` (optional): Filter by status (Success/Failed)
- `start_date` (optional): Start date
- `end_date` (optional): End date
- `limit` (optional): Number of results
- `offset` (optional): Pagination offset

**Response**: `200 OK`
```json
{
  "logs": [
    {
      "id": "507f1f77bcf86cd799439016",
      "user_email": "admin@acme.com",
      "action": "project.created",
      "resource": "My Video App",
      "ip_address": "192.168.1.1",
      "status": "Success",
      "timestamp": "2025-01-19T10:30:00Z"
    }
  ],
  "total": 1
}
```

### Export Audit Logs (CSV)

**Endpoint**: `GET /api/v1/audit-logs/export`

**Query Parameters**: Same as Get Audit Logs

**Response**: `200 OK`
```
Content-Type: text/csv
Content-Disposition: attachment; filename="audit_logs_2025-01-19.csv"

Timestamp,User,Action,Resource,IP Address,Status
2025-01-19T10:30:00Z,admin@acme.com,project.created,My Video App,192.168.1.1,Success
```

### Get Audit Stats

**Endpoint**: `GET /api/v1/audit-logs/stats`

**Response**: `200 OK`
```json
{
  "total_actions": 1250,
  "success_rate": 98.5,
  "failed_actions": 18,
  "top_actions": [
    {
      "action": "token.created",
      "count": 450
    },
    {
      "action": "project.updated",
      "count": 120
    }
  ]
}
```

---

## Status & Monitoring

### Get System Status

**Endpoint**: `GET /api/v1/status`

**Response**: `200 OK`
```json
{
  "status": "Operational",
  "uptime": 99.99,
  "version": "1.0.0",
  "services": {
    "database": {
      "status": "Healthy",
      "response_time_ms": 2
    },
    "api": {
      "status": "Healthy",
      "response_time_ms": 1
    },
    "livekit": {
      "status": "Healthy",
      "response_time_ms": 5
    }
  },
  "active_projects": 145,
  "last_checked": "2025-01-19T10:30:00Z"
}
```

### Get Project Health

**Endpoint**: `GET /api/v1/status/projects/:id`

**Response**: `200 OK`
```json
{
  "project_id": "507f1f77bcf86cd799439012",
  "status": "Healthy",
  "active_rooms": 5,
  "total_participants": 23,
  "last_activity": "2025-01-19T10:28:00Z"
}
```

### Get Region Availability

**Endpoint**: `GET /api/v1/status/regions`

**Response**: `200 OK`
```json
{
  "regions": [
    {
      "name": "US East",
      "status": "Operational",
      "latency_ms": 12,
      "active_rooms": 145
    },
    {
      "name": "EU West",
      "status": "Operational",
      "latency_ms": 28,
      "active_rooms": 87
    }
  ]
}
```

---

## Webhooks (Outgoing)

Pulse can send webhooks to your application for important events.

### Webhook Configuration

Set webhook URL when creating/updating a project:

```json
{
  "webhook_url": "https://myapp.com/webhooks/pulse"
}
```

### Webhook Events

- `room.started`: Room was created
- `room.ended`: Room was closed
- `participant.joined`: User joined a room
- `participant.left`: User left a room
- `egress.started`: Egress streaming started
- `egress.ended`: Egress streaming ended
- `recording.completed`: Recording is ready

### Webhook Payload

```json
{
  "event": "room.started",
  "project_id": "507f1f77bcf86cd799439012",
  "timestamp": "2025-01-19T10:30:00Z",
  "data": {
    "room_name": "meeting-room-456",
    "room_sid": "RM_abc123"
  }
}
```

### Webhook Signature

Verify webhook authenticity using the signature header:

```
X-Webhook-Signature: sha256=abc123...
```

**Verification** (Python):
```python
import hmac
import hashlib

def verify_webhook(payload, signature, secret):
    expected = 'sha256=' + hmac.new(
        secret.encode(),
        payload.encode(),
        hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(signature, expected)
```

---

## Code Examples

### JavaScript/Node.js

```javascript
const axios = require('axios');

const API_KEY = 'pulse_key_abc123xyz789';
const BASE_URL = 'https://api.pulse.io/api/v1';

// Create a token
async function createToken(identity, roomName) {
  try {
    const response = await axios.post(
      `${BASE_URL}/tokens/create`,
      {
        identity,
        room_name: roomName,
        ttl: 3600
      },
      {
        headers: {
          'X-Pulse-Key': API_KEY,
          'Content-Type': 'application/json'
        }
      }
    );
    return response.data.token;
  } catch (error) {
    console.error('Error creating token:', error.response.data);
    throw error;
  }
}

// Get usage metrics
async function getUsage(projectId) {
  const response = await axios.get(
    `${BASE_URL}/usage/${projectId}/summary`,
    {
      headers: { 'X-Pulse-Key': API_KEY }
    }
  );
  return response.data;
}
```

### Python

```python
import requests

API_KEY = 'pulse_key_abc123xyz789'
BASE_URL = 'https://api.pulse.io/api/v1'

def create_token(identity, room_name):
    response = requests.post(
        f'{BASE_URL}/tokens/create',
        json={
            'identity': identity,
            'room_name': room_name,
            'ttl': 3600
        },
        headers={
            'X-Pulse-Key': API_KEY,
            'Content-Type': 'application/json'
        }
    )
    response.raise_for_status()
    return response.json()['token']

def get_usage(project_id):
    response = requests.get(
        f'{BASE_URL}/usage/{project_id}/summary',
        headers={'X-Pulse-Key': API_KEY}
    )
    return response.json()
```

### Go

```go
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

const (
	APIKey  = "pulse_key_abc123xyz789"
	BaseURL = "https://api.pulse.io/api/v1"
)

type TokenRequest struct {
	Identity string `json:"identity"`
	RoomName string `json:"room_name"`
	TTL      int    `json:"ttl"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

func createToken(identity, roomName string) (string, error) {
	reqBody := TokenRequest{
		Identity: identity,
		RoomName: roomName,
		TTL:      3600,
	}
	body, _ := json.Marshal(reqBody)
	
	req, _ := http.NewRequest("POST", BaseURL+"/tokens/create", bytes.NewBuffer(body))
	req.Header.Set("X-Pulse-Key", APIKey)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	var tokenResp TokenResponse
	json.NewDecoder(resp.Body).Decode(&tokenResp)
	return tokenResp.Token, nil
}
```

---

## Support

- **Documentation**: https://docs.pulse.io
- **Email**: support@pulse.io
- **Discord**: https://discord.gg/pulse
- **GitHub**: https://github.com/pulse-io/pulse

---

**Last Updated**: 2025-01-19
