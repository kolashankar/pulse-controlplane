# Webhooks Guide - Pulse Control Plane

Complete guide to receiving and sending webhooks with Pulse.

---

## Table of Contents

1. [Overview](#overview)
2. [Setting Up Webhooks](#setting-up-webhooks)
3. [Webhook Events](#webhook-events)
4. [Webhook Security](#webhook-security)
5. [Handling Webhooks](#handling-webhooks)
6. [Retry Logic](#retry-logic)
7. [Testing Webhooks](#testing-webhooks)
8. [Troubleshooting](#troubleshooting)

---

## Overview

Webhooks allow Pulse to notify your application about events in real-time.

### Why Use Webhooks?

- **Real-time notifications**: Know when rooms start/end
- **Usage tracking**: Monitor participant activity
- **Billing accuracy**: Track exact usage for billing
- **Analytics**: Collect data for insights
- **Automation**: Trigger workflows based on events

### How Webhooks Work

```
Event Occurs    Pulse Sends      Your Server     Your Server
  in Pulse   Ąę> HTTP POST Ąę>  Receives   Ąę>  Processes
                   Webhook         Webhook         Event
```

---

## Setting Up Webhooks

### Step 1: Create Webhook Endpoint

Create an endpoint in your application:

```javascript
// Node.js/Express
app.post('/webhooks/pulse', (req, res) => {
  const event = req.body;
  console.log('Received webhook:', event.event);
  
  // Process the event
  processWebhook(event);
  
  // Acknowledge receipt
  res.status(200).send('OK');
});
```

### Step 2: Configure Webhook URL

Set your webhook URL when creating/updating a project:

```bash
curl -X PUT https://api.pulse.io/api/v1/projects/{PROJECT_ID} \
  -H "X-Pulse-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "webhook_url": "https://yourapp.com/webhooks/pulse"
  }'
```

### Step 3: Verify Webhook Signatures

Always verify webhook signatures for security:

```javascript
const crypto = require('crypto');

function verifyWebhook(payload, signature, secret) {
  const expected = 'sha256=' + crypto
    .createHmac('sha256', secret)
    .update(payload)
    .digest('hex');
  return crypto.timingSafeEqual(
    Buffer.from(signature),
    Buffer.from(expected)
  );
}

app.post('/webhooks/pulse', (req, res) => {
  const signature = req.headers['x-webhook-signature'];
  const payload = JSON.stringify(req.body);
  
  if (!verifyWebhook(payload, signature, process.env.WEBHOOK_SECRET)) {
    return res.status(401).send('Invalid signature');
  }
  
  // Process webhook...
});
```

---

## Webhook Events

### Room Events

#### `room.started`

Trigger when a room is created.

```json
{
  "event": "room.started",
  "project_id": "507f1f77bcf86cd799439012",
  "timestamp": "2025-01-19T10:30:00Z",
  "data": {
    "room_name": "meeting-room-456",
    "room_sid": "RM_abc123xyz789",
    "created_at": "2025-01-19T10:30:00Z"
  }
}
```

**Use Case**: Initialize room settings, create database records

#### `room.ended`

Triggered when a room is closed (all participants left).

```json
{
  "event": "room.ended",
  "project_id": "507f1f77bcf86cd799439012",
  "timestamp": "2025-01-19T11:30:00Z",
  "data": {
    "room_name": "meeting-room-456",
    "room_sid": "RM_abc123xyz789",
    "duration_seconds": 3600,
    "max_participants": 5,
    "total_participant_minutes": 300
  }
}
```

**Use Case**: Calculate charges, send summary emails, cleanup resources

### Participant Events

#### `participant.joined`

```json
{
  "event": "participant.joined",
  "project_id": "507f1f77bcf86cd799439012",
  "timestamp": "2025-01-19T10:31:00Z",
  "data": {
    "room_name": "meeting-room-456",
    "room_sid": "RM_abc123xyz789",
    "participant": {
      "identity": "user123",
      "sid": "PA_xyz789",
      "metadata": {
        "name": "John Doe",
        "email": "john@example.com"
      }
    }
  }
}
```

**Use Case**: Send notifications, update UI, track attendance

#### `participant.left`

```json
{
  "event": "participant.left",
  "project_id": "507f1f77bcf86cd799439012",
  "timestamp": "2025-01-19T11:15:00Z",
  "data": {
    "room_name": "meeting-room-456",
    "room_sid": "RM_abc123xyz789",
    "participant": {
      "identity": "user123",
      "sid": "PA_xyz789",
      "duration_seconds": 2640
    }
  }
}
```

**Use Case**: Track individual usage, update attendance records

### Track Events

#### `track.published`

```json
{
  "event": "track.published",
  "data": {
    "room_name": "meeting-room-456",
    "participant_identity": "user123",
    "track": {
      "sid": "TR_abc123",
      "type": "video",
      "source": "camera"
    }
  }
}
```

#### `track.unpublished`

```json
{
  "event": "track.unpublished",
  "data": {
    "room_name": "meeting-room-456",
    "participant_identity": "user123",
    "track": {
      "sid": "TR_abc123",
      "type": "video"
    }
  }
}
```

### Egress Events

#### `egress.started`

```json
{
  "event": "egress.started",
  "data": {
    "egress_id": "EG_abc123",
    "room_name": "meeting-room-456",
    "type": "RTMP",
    "started_at": "2025-01-19T10:35:00Z"
  }
}
```

#### `egress.ended`

```json
{
  "event": "egress.ended",
  "data": {
    "egress_id": "EG_abc123",
    "room_name": "meeting-room-456",
    "type": "RTMP",
    "duration_seconds": 1800,
    "ended_at": "2025-01-19T11:05:00Z"
  }
}
```

### Recording Events

#### `recording.completed`

```json
{
  "event": "recording.completed",
  "data": {
    "recording_id": "REC_abc123",
    "room_name": "meeting-room-456",
    "file_url": "https://storage.pulse.io/recordings/abc123.mp4",
    "duration_seconds": 3600,
    "file_size_bytes": 524288000
  }
}
```

**Use Case**: Notify users, process recordings, upload to your storage

---

## Webhook Security

### Signature Verification

Pulse signs all webhooks with HMAC SHA256.

**Header**: `X-Webhook-Signature: sha256=abc123...`

#### Node.js Verification

```javascript
const crypto = require('crypto');

function verifyWebhook(payload, signature, secret) {
  const hmac = crypto.createHmac('sha256', secret);
  hmac.update(payload);
  const expected = 'sha256=' + hmac.digest('hex');
  
  return crypto.timingSafeEqual(
    Buffer.from(signature),
    Buffer.from(expected)
  );
}

app.post('/webhooks/pulse', express.raw({ type: 'application/json' }), (req, res) => {
  const signature = req.headers['x-webhook-signature'];
  const payload = req.body.toString();
  
  if (!verifyWebhook(payload, signature, process.env.WEBHOOK_SECRET)) {
    return res.status(401).json({ error: 'Invalid signature' });
  }
  
  const event = JSON.parse(payload);
  processWebhook(event);
  res.status(200).send('OK');
});
```

#### Python Verification

```python
import hmac
import hashlib

def verify_webhook(payload: bytes, signature: str, secret: str) -> bool:
    expected = 'sha256=' + hmac.new(
        secret.encode(),
        payload,
        hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(signature, expected)

@app.route('/webhooks/pulse', methods=['POST'])
def handle_webhook():
    signature = request.headers.get('X-Webhook-Signature')
    payload = request.get_data()
    
    if not verify_webhook(payload, signature, os.environ['WEBHOOK_SECRET']):
        return jsonify({'error': 'Invalid signature'}), 401
    
    event = request.json
    process_webhook(event)
    return 'OK', 200
```

#### Go Verification

```go
import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func verifyWebhook(payload []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expected))
}
```

### Best Practices

1. **Always verify signatures**: Never process unverified webhooks
2. **Use HTTPS**: Webhooks should only be sent to HTTPS endpoints
3. **Keep secrets secure**: Store webhook secret in environment variables
4. **Implement replay protection**: Check timestamp to prevent replay attacks
5. **Use idempotency**: Process each webhook only once

---

## Handling Webhooks

### Complete Handler Example

```javascript
const express = require('express');
const crypto = require('crypto');
const { MongoClient } = require('mongodb');

const app = express();
const processedWebhooks = new Set();

// Verify webhook signature
function verifyWebhook(payload, signature, secret) {
  const expected = 'sha256=' + crypto
    .createHmac('sha256', secret)
    .update(payload)
    .digest('hex');
  return crypto.timingSafeEqual(
    Buffer.from(signature),
    Buffer.from(expected)
  );
}

// Check if webhook was already processed (idempotency)
function isProcessed(webhookId) {
  return processedWebhooks.has(webhookId);
}

function markAsProcessed(webhookId) {
  processedWebhooks.add(webhookId);
  // Clean up old entries after 24 hours
  setTimeout(() => processedWebhooks.delete(webhookId), 24 * 60 * 60 * 1000);
}

// Process different event types
async function processWebhook(event) {
  switch (event.event) {
    case 'room.started':
      await handleRoomStarted(event.data);
      break;
    case 'room.ended':
      await handleRoomEnded(event.data);
      break;
    case 'participant.joined':
      await handleParticipantJoined(event.data);
      break;
    case 'participant.left':
      await handleParticipantLeft(event.data);
      break;
    case 'recording.completed':
      await handleRecordingCompleted(event.data);
      break;
    default:
      console.log('Unknown event type:', event.event);
  }
}

// Event handlers
async function handleRoomStarted(data) {
  console.log('Room started:', data.room_name);
  // Save to database
  await db.collection('rooms').insertOne({
    room_sid: data.room_sid,
    room_name: data.room_name,
    started_at: new Date(data.created_at),
    status: 'active'
  });
}

async function handleRoomEnded(data) {
  console.log('Room ended:', data.room_name);
  // Update database
  await db.collection('rooms').updateOne(
    { room_sid: data.room_sid },
    {
      $set: {
        status: 'ended',
        ended_at: new Date(),
        duration_seconds: data.duration_seconds,
        participant_minutes: data.total_participant_minutes
      }
    }
  );
  
  // Calculate and save charges
  const cost = data.total_participant_minutes * 0.004; // $0.004 per minute
  await db.collection('billing').insertOne({
    room_sid: data.room_sid,
    participant_minutes: data.total_participant_minutes,
    cost: cost,
    date: new Date()
  });
  
  // Send notification email
  await sendEmail({
    to: 'host@example.com',
    subject: 'Meeting Ended',
    body: `Your meeting "${data.room_name}" has ended. Duration: ${data.duration_seconds / 60} minutes.`
  });
}

async function handleParticipantJoined(data) {
  console.log('Participant joined:', data.participant.identity);
  // Track attendance
  await db.collection('attendance').insertOne({
    room_sid: data.room_sid,
    participant_sid: data.participant.sid,
    identity: data.participant.identity,
    joined_at: new Date()
  });
}

async function handleParticipantLeft(data) {
  console.log('Participant left:', data.participant.identity);
  // Update attendance
  await db.collection('attendance').updateOne(
    {
      room_sid: data.room_sid,
      participant_sid: data.participant.sid
    },
    {
      $set: {
        left_at: new Date(),
        duration_seconds: data.participant.duration_seconds
      }
    }
  );
}

async function handleRecordingCompleted(data) {
  console.log('Recording completed:', data.recording_id);
  // Download and process recording
  await downloadRecording(data.file_url, data.recording_id);
  
  // Notify users
  await sendEmail({
    to: 'host@example.com',
    subject: 'Recording Ready',
    body: `Your recording is ready: ${data.file_url}`
  });
}

// Webhook endpoint
app.post('/webhooks/pulse',
  express.raw({ type: 'application/json' }),
  async (req, res) => {
    try {
      // Verify signature
      const signature = req.headers['x-webhook-signature'];
      const payload = req.body.toString();
      
      if (!verifyWebhook(payload, signature, process.env.WEBHOOK_SECRET)) {
        console.error('Invalid webhook signature');
        return res.status(401).send('Invalid signature');
      }
      
      const event = JSON.parse(payload);
      
      // Check for replay (idempotency)
      const webhookId = `${event.event}_${event.timestamp}_${event.data.room_sid || event.data.recording_id}`;
      if (isProcessed(webhookId)) {
        console.log('Webhook already processed:', webhookId);
        return res.status(200).send('Already processed');
      }
      
      // Process webhook asynchronously
      processWebhook(event)
        .then(() => {
          markAsProcessed(webhookId);
          console.log('Webhook processed successfully');
        })
        .catch(err => {
          console.error('Error processing webhook:', err);
        });
      
      // Acknowledge immediately (don't wait for processing)
      res.status(200).send('OK');
    } catch (error) {
      console.error('Webhook handler error:', error);
      res.status(500).send('Internal error');
    }
  }
);

app.listen(3000, () => {
  console.log('Webhook server running on port 3000');
});
```

---

## Retry Logic

Pulse automatically retries failed webhook deliveries.

### Retry Schedule

- **Attempt 1**: Immediately
- **Attempt 2**: After 5 seconds
- **Attempt 3**: After 30 seconds
- **Attempt 4**: After 2 minutes
- **Attempt 5**: After 10 minutes
- **Attempt 6**: After 1 hour

### Success Criteria

Webhook is considered successful if:
- HTTP status code is 200-299
- Response received within 30 seconds

### Failure Handling

```javascript
// Your endpoint should return 200 for success
app.post('/webhooks/pulse', async (req, res) => {
  try {
    await processWebhook(req.body);
    res.status(200).send('OK'); // Success
  } catch (error) {
    console.error('Error:', error);
    // Return 500 to trigger retry
    res.status(500).send('Error processing webhook');
  }
});
```

---

## Testing Webhooks

### Local Testing with ngrok

1. Install ngrok:
```bash
npm install -g ngrok
```

2. Start your webhook server:
```bash
node webhook-server.js
```

3. Create ngrok tunnel:
```bash
ngrok http 3000
```

4. Use ngrok URL as webhook URL:
```bash
curl -X PUT https://api.pulse.io/api/v1/projects/PROJECT_ID \
  -H "X-Pulse-Key: YOUR_KEY" \
  -d '{"webhook_url": "https://abc123.ngrok.io/webhooks/pulse"}'
```

### Manual Testing

Send test webhook:

```bash
curl -X POST http://localhost:3000/webhooks/pulse \
  -H "Content-Type: application/json" \
  -H "X-Webhook-Signature: sha256=$(echo -n '{"event":"room.started"}' | openssl dgst -sha256 -hmac 'your-secret' | cut -d' ' -f2)" \
  -d '{
    "event": "room.started",
    "timestamp": "2025-01-19T10:30:00Z",
    "data": {
      "room_name": "test-room",
      "room_sid": "RM_test123"
    }
  }'
```

### Unit Testing

```javascript
const request = require('supertest');
const app = require('./webhook-server');

describe('Webhook Handler', () => {
  it('should process room.started event', async () => {
    const payload = {
      event: 'room.started',
      data: { room_name: 'test-room' }
    };
    
    const response = await request(app)
      .post('/webhooks/pulse')
      .send(payload)
      .set('X-Webhook-Signature', generateSignature(payload));
    
    expect(response.status).toBe(200);
  });
  
  it('should reject invalid signature', async () => {
    const response = await request(app)
      .post('/webhooks/pulse')
      .send({ event: 'test' })
      .set('X-Webhook-Signature', 'invalid');
    
    expect(response.status).toBe(401);
  });
});
```

---

## Troubleshooting

### Webhooks Not Received

**Checklist**:
1. āļø Webhook URL is publicly accessible
2. āļø Endpoint returns 200 status
3. āļø HTTPS is used (required)
4. āļø Firewall allows incoming requests
5. āļø Response time < 30 seconds

### Check Webhook Logs

```bash
curl https://api.pulse.io/api/v1/webhooks/logs \
  -H "X-Pulse-Key: YOUR_KEY"
```

### Common Errors

#### "Connection timeout"
**Cause**: Server took too long to respond
**Solution**: Return 200 immediately, process async

```javascript
// BAD: Processing before response
app.post('/webhooks/pulse', async (req, res) => {
  await longRunningProcess(req.body); // Takes 60 seconds
  res.send('OK'); // Timeout!
});

// GOOD: Respond immediately
app.post('/webhooks/pulse', async (req, res) => {
  res.send('OK'); // Respond immediately
  await longRunningProcess(req.body); // Process async
});
```

#### "SSL certificate error"
**Cause**: Invalid SSL certificate
**Solution**: Use valid SSL certificate (Let's Encrypt, etc.)

#### "Invalid signature"
**Cause**: Wrong webhook secret or payload mismatch
**Solution**: Verify secret, use raw body for verification

---

## Resources

- [API Reference](./API.md)
- [Authentication Guide](./AUTHENTICATION.md)
- [Quick Start Guide](./QUICKSTART.md)

---

**Last Updated**: 2025-01-19
