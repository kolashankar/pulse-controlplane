# Authentication Guide - Pulse Control Plane

Complete guide to authentication and authorization in Pulse.

---

## Table of Contents

1. [Overview](#overview)
2. [API Keys](#api-keys)
3. [Token Generation](#token-generation)
4. [Client Authentication](#client-authentication)
5. [Security Best Practices](#security-best-practices)
6. [Team & Role-Based Access](#team--role-based-access)
7. [Troubleshooting](#troubleshooting)

---

## Overview

Pulse uses a two-tier authentication system:

1. **API Keys** (Server-Side): Authenticate your backend server with Pulse API
2. **Access Tokens** (Client-Side): Short-lived JWT tokens for end users

```
Your Server  Ąęęęę[API Key]Ąęęęę>  Pulse API  Ąęęęę[Access Token]Ąęęęę>  End Users
```

### Why Two Levels?

- **Security**: API keys never exposed to clients
- **Control**: You decide who gets tokens and what permissions
- **Flexibility**: Custom business logic before granting access

---

## API Keys

### Getting Your API Keys

1. Create a project in the Pulse dashboard
2. You'll receive:
   - **API Key**: `pulse_key_*` (use for authentication)
   - **API Secret**: `pulse_secret_*` (keep secure, use for sensitive operations)

### Using API Keys

#### Method 1: Header (Recommended)

```bash
curl -X POST https://api.pulse.io/api/v1/tokens/create \
  -H "X-Pulse-Key: pulse_key_abc123xyz789" \
  -H "Content-Type: application/json"
```

#### Method 2: Bearer Token

```bash
curl -X POST https://api.pulse.io/api/v1/tokens/create \
  -H "Authorization: Bearer pulse_key_abc123xyz789" \
  -H "Content-Type: application/json"
```

### API Secret (For Sensitive Operations)

Some operations require both key and secret:

```bash
curl -X POST https://api.pulse.io/api/v1/projects/PROJECT_ID/regenerate-keys \
  -H "X-Pulse-Key: pulse_key_abc123xyz789" \
  -H "X-Pulse-Secret: pulse_secret_def456uvw123"
```

**Requires Both**:
- Regenerating API keys
- Deleting projects
- Accessing billing information
- Modifying team permissions

### Regenerating API Keys

If your keys are compromised:

```bash
curl -X POST https://api.pulse.io/api/v1/projects/{PROJECT_ID}/regenerate-keys \
  -H "X-Pulse-Key: OLD_API_KEY" \
  -H "X-Pulse-Secret: OLD_API_SECRET"
```

**Response**:
```json
{
  "pulse_api_key": "pulse_key_NEW123",
  "pulse_api_secret": "pulse_secret_NEW456",
  "warning": "Old credentials are immediately invalidated"
}
```

š️ **Important**: Old keys stop working immediately!

---

## Token Generation

Access tokens allow end users to connect to rooms.

### Basic Token Creation

```javascript
// Node.js Example
const axios = require('axios');

async function createToken(userId, roomName) {
  const response = await axios.post(
    'https://api.pulse.io/api/v1/tokens/create',
    {
      identity: userId,
      room_name: roomName,
      ttl: 3600 // 1 hour
    },
    {
      headers: {
        'X-Pulse-Key': process.env.PULSE_API_KEY,
        'Content-Type': 'application/json'
      }
    }
  );
  return response.data.token;
}
```

### Token with Metadata

Attach user information to tokens:

```javascript
const response = await axios.post(
  'https://api.pulse.io/api/v1/tokens/create',
  {
    identity: 'user123',
    room_name: 'meeting-456',
    metadata: {
      name: 'John Doe',
      email: 'john@example.com',
      avatar: 'https://example.com/avatar.jpg',
      role: 'host'
    },
    ttl: 7200 // 2 hours
  },
  { headers: { 'X-Pulse-Key': API_KEY } }
);
```

### Token with Permissions

Control what users can do:

```javascript
const response = await axios.post(
  'https://api.pulse.io/api/v1/tokens/create',
  {
    identity: 'user123',
    room_name: 'meeting-456',
    permissions: {
      can_publish: true,      // Can share video/audio
      can_subscribe: true,    // Can see/hear others
      can_publish_data: true, // Can send chat messages
      hidden: false,          // Visible to others
      recorder: false         // Not a recording bot
    }
  },
  { headers: { 'X-Pulse-Key': API_KEY } }
);
```

### Read-Only Tokens (Viewers)

```javascript
// For viewers/audience members
const viewerToken = await createToken({
  identity: 'viewer123',
  room_name: 'webinar-789',
  permissions: {
    can_publish: false,     // Cannot share video/audio
    can_subscribe: true,    // Can see/hear
    can_publish_data: false // Cannot send messages
  }
});
```

---

## Client Authentication

### Web Client (JavaScript)

```javascript
import { Room } from 'livekit-client';

class VideoService {
  constructor() {
    this.room = null;
  }
  
  async connect(roomName) {
    // Get token from your backend
    const token = await this.getTokenFromBackend(roomName);
    
    // Connect to room
    this.room = new Room();
    await this.room.connect('wss://livekit.pulse.io', token);
  }
  
  async getTokenFromBackend(roomName) {
    const response = await fetch('/api/get-token', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ roomName })
    });
    const data = await response.json();
    return data.token;
  }
}
```

### React Client

```javascript
import { useState, useEffect } from 'react';
import { useRoom } from '@livekit/react-core';

function VideoCall({ roomName }) {
  const [token, setToken] = useState(null);
  const { connect, room } = useRoom();
  
  useEffect(() => {
    // Fetch token from your backend
    fetch('/api/get-token', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ roomName })
    })
      .then(res => res.json())
      .then(data => {
        setToken(data.token);
        connect('wss://livekit.pulse.io', data.token);
      });
  }, [roomName]);
  
  return <div>Video Call Component</div>;
}
```

### Mobile Client (React Native)

```javascript
import { useRoom } from '@livekit/react-native';

function VideoScreen({ route }) {
  const { roomName } = route.params;
  const { connect, room } = useRoom();
  
  useEffect(() => {
    async function initialize() {
      // Get token from API
      const response = await fetch('https://api.myapp.com/get-token', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${await getAuthToken()}`
        },
        body: JSON.stringify({ roomName })
      });
      const { token } = await response.json();
      
      // Connect to room
      await connect('wss://livekit.pulse.io', token);
    }
    
    initialize();
  }, [roomName]);
  
  return <View>{/* Video UI */}</View>;
}
```

---

## Security Best Practices

### 1. Never Expose API Keys

š× **DON'T** do this:
```javascript
// In client-side code
const API_KEY = 'pulse_key_abc123'; // NEVER!
```

šü **DO** this:
```javascript
// In server-side code
const API_KEY = process.env.PULSE_API_KEY; // Good!
```

### 2. Use Environment Variables

```bash
# .env file
PULSE_API_KEY=pulse_key_abc123xyz789
PULSE_API_SECRET=pulse_secret_def456uvw123
```

```javascript
// Load from environment
require('dotenv').config();
const API_KEY = process.env.PULSE_API_KEY;
```

### 3. Validate Users Before Creating Tokens

```javascript
app.post('/api/get-token', authenticateUser, async (req, res) => {
  // Only authenticated users can get tokens
  if (!req.user) {
    return res.status(401).json({ error: 'Unauthorized' });
  }
  
  const { roomName } = req.body;
  
  // Check if user has access to this room
  const hasAccess = await checkRoomAccess(req.user.id, roomName);
  if (!hasAccess) {
    return res.status(403).json({ error: 'Access denied' });
  }
  
  // Generate token
  const token = await createPulseToken(req.user.id, roomName);
  res.json({ token });
});
```

### 4. Use Short-Lived Tokens

```javascript
// Short TTL for security
const token = await createToken({
  identity: userId,
  room_name: roomName,
  ttl: 3600 // 1 hour
});

// Implement token refresh
room.on('tokenExpired', async () => {
  const newToken = await fetchNewToken();
  await room.reconnect(newToken);
});
```

### 5. Implement Rate Limiting

```javascript
const rateLimit = require('express-rate-limit');

const tokenLimiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 100, // Limit each IP to 100 requests per windowMs
  message: 'Too many token requests'
});

app.post('/api/get-token', tokenLimiter, async (req, res) => {
  // Token generation logic
});
```

### 6. Log Access Attempts

```javascript
app.post('/api/get-token', async (req, res) => {
  const { roomName } = req.body;
  
  // Log the access attempt
  await logAudit({
    user: req.user.id,
    action: 'token.created',
    room: roomName,
    ip: req.ip,
    timestamp: new Date()
  });
  
  const token = await createToken(req.user.id, roomName);
  res.json({ token });
});
```

### 7. Rotate Keys Regularly

```bash
# Rotate keys every 90 days
0 0 1 */3 * /path/to/rotate-keys.sh
```

### 8. Use HTTPS Only

```javascript
// Enforce HTTPS
app.use((req, res, next) => {
  if (req.header('x-forwarded-proto') !== 'https' && process.env.NODE_ENV === 'production') {
    return res.redirect(`https://${req.header('host')}${req.url}`);
  }
  next();
});
```

---

## Team & Role-Based Access

### Roles in Pulse

| Role | Permissions |
|------|-------------|
| **Owner** | Full access, billing, delete org |
| **Admin** | Manage team, projects, API keys |
| **Developer** | Manage projects, view usage |
| **Viewer** | Read-only access |

### Checking User Role

```javascript
app.post('/api/regenerate-keys', async (req, res) => {
  const userRole = await getUserRole(req.user.id, req.body.projectId);
  
  if (!['Owner', 'Admin'].includes(userRole)) {
    return res.status(403).json({ error: 'Insufficient permissions' });
  }
  
  // Proceed with key regeneration
});
```

### Inviting Team Members

```javascript
const response = await axios.post(
  `https://api.pulse.io/api/v1/organizations/${orgId}/members`,
  {
    email: 'newmember@example.com',
    role: 'Developer'
  },
  {
    headers: { 'X-Pulse-Key': API_KEY }
  }
);
```

---

## Token Validation

### Validate Token Before Use

```javascript
const response = await axios.post(
  'https://api.pulse.io/api/v1/tokens/validate',
  { token: userToken },
  { headers: { 'X-Pulse-Key': API_KEY } }
);

if (!response.data.valid) {
  throw new Error('Invalid token');
}
```

### Client-Side Token Refresh

```javascript
import { Room, RoomEvent } from 'livekit-client';

const room = new Room();

room.on(RoomEvent.TokenExpiring, async () => {
  console.log('Token expiring soon, fetching new token...');
  const newToken = await fetchNewTokenFromBackend();
  await room.setToken(newToken);
});

room.on(RoomEvent.Disconnected, (reason) => {
  if (reason === 'token expired') {
    // Reconnect with new token
    const newToken = await fetchNewTokenFromBackend();
    await room.connect(wsUrl, newToken);
  }
});
```

---

## Troubleshooting

### Error: "Invalid API key"

**Cause**: API key is wrong or project was deleted

**Solution**:
1. Verify key in dashboard
2. Check environment variables
3. Ensure key starts with `pulse_key_`

```bash
# Test your API key
curl -X GET https://api.pulse.io/api/v1/projects \
  -H "X-Pulse-Key: YOUR_KEY" -v
```

### Error: "Token expired"

**Cause**: Token TTL exceeded

**Solution**: Generate a new token
```javascript
try {
  await room.connect(wsUrl, token);
} catch (error) {
  if (error.message.includes('expired')) {
    const newToken = await fetchNewToken();
    await room.connect(wsUrl, newToken);
  }
}
```

### Error: "Rate limit exceeded"

**Cause**: Too many requests

**Solution**: Implement backoff
```javascript
async function createTokenWithRetry(userId, roomName, retries = 3) {
  for (let i = 0; i < retries; i++) {
    try {
      return await createToken(userId, roomName);
    } catch (error) {
      if (error.response?.status === 429 && i < retries - 1) {
        await new Promise(resolve => setTimeout(resolve, Math.pow(2, i) * 1000));
        continue;
      }
      throw error;
    }
  }
}
```

### Error: "Insufficient permissions"

**Cause**: User role doesn't allow this action

**Solution**: Check user role before attempting action

---

## Complete Example: Secure Token Server

```javascript
const express = require('express');
const axios = require('axios');
const rateLimit = require('express-rate-limit');
require('dotenv').config();

const app = express();
app.use(express.json());

// Rate limiting
const tokenLimiter = rateLimit({
  windowMs: 15 * 60 * 1000,
  max: 100
});

// Authentication middleware
function authenticateUser(req, res, next) {
  const authHeader = req.headers.authorization;
  if (!authHeader) {
    return res.status(401).json({ error: 'Unauthorized' });
  }
  // Verify JWT or session
  // req.user = decoded user
  next();
}

// Check room access
async function checkRoomAccess(userId, roomName) {
  // Implement your business logic
  return true;
}

// Create token endpoint
app.post('/api/get-token', 
  authenticateUser,
  tokenLimiter,
  async (req, res) => {
    try {
      const { roomName } = req.body;
      
      // Validate request
      if (!roomName) {
        return res.status(400).json({ error: 'Room name required' });
      }
      
      // Check access
      const hasAccess = await checkRoomAccess(req.user.id, roomName);
      if (!hasAccess) {
        return res.status(403).json({ error: 'Access denied' });
      }
      
      // Create token
      const response = await axios.post(
        'https://api.pulse.io/api/v1/tokens/create',
        {
          identity: req.user.id,
          room_name: roomName,
          metadata: {
            name: req.user.name,
            email: req.user.email
          },
          ttl: 3600
        },
        {
          headers: {
            'X-Pulse-Key': process.env.PULSE_API_KEY,
            'Content-Type': 'application/json'
          }
        }
      );
      
      // Log access
      console.log(`Token created for user ${req.user.id} in room ${roomName}`);
      
      res.json({ token: response.data.token });
    } catch (error) {
      console.error('Token creation failed:', error.message);
      res.status(500).json({ error: 'Failed to create token' });
    }
  }
);

app.listen(3000, () => {
  console.log('Token server running on port 3000');
});
```

---

## Resources

- [API Reference](./API.md)
- [Quick Start Guide](./QUICKSTART.md)
- [Security Best Practices](./SECURITY.md)
- [Webhook Guide](./WEBHOOKS.md)

---

**Last Updated**: 2025-01-19
