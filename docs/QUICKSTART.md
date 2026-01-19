# Quick Start Guide - Pulse Control Plane

Get up and running with Pulse in under 10 minutes.

## Prerequisites

- A Pulse account (sign up at https://pulse.io)
- Basic knowledge of REST APIs
- Development environment with HTTP client (cURL, Postman, or code)

---

## Step 1: Create an Organization

### Using cURL

```bash
curl -X POST https://api.pulse.io/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Company",
    "admin_email": "admin@mycompany.com",
    "plan": "Free"
  }'
```

### Response

```json
{
  "id": "507f1f77bcf86cd799439011",
  "name": "My Company",
  "admin_email": "admin@mycompany.com",
  "plan": "Free",
  "created_at": "2025-01-19T10:30:00Z"
}
```

š **Save the `id`** - you'll need it for the next step!

---

## Step 2: Create a Project

A project represents your application and contains API credentials.

```bash
curl -X POST https://api.pulse.io/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Video App",
    "org_id": "507f1f77bcf86cd799439011",
    "region": "US_EAST"
  }'
```

### Response

```json
{
  "id": "507f1f77bcf86cd799439012",
  "name": "My Video App",
  "pulse_api_key": "pulse_key_abc123xyz789",
  "pulse_api_secret": "pulse_secret_def456uvw123",
  "region": "US_EAST"
}
```

š **Important**: Save your API credentials!
- `pulse_api_key`: Use for all API requests
- `pulse_api_secret`: Keep this secure! Only shown once.

---

## Step 3: Create Your First Token

Tokens allow users to join video/audio rooms.

```bash
curl -X POST https://api.pulse.io/api/v1/tokens/create \
  -H "X-Pulse-Key: pulse_key_abc123xyz789" \
  -H "Content-Type: application/json" \
  -d '{
    "identity": "user123",
    "room_name": "my-first-room",
    "metadata": {
      "name": "John Doe"
    }
  }'
```

### Response

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDU2NzA0MDAsImlzcyI6IlBVTFNFIiwibmJmIjoxNzA1NjY2ODAwLCJzdWIiOiJ1c2VyMTIzIiwidmlkZW8iOnsicm9vbSI6Im15LWZpcnN0LXJvb20iLCJyb29tSm9pbiI6dHJ1ZX19.abc123xyz789",
  "identity": "user123",
  "room_name": "my-first-room",
  "expires_at": "2025-01-19T11:30:00Z"
}
```

š Use this `token` in your client application to connect to the room!

---

## Step 4: Integrate Client SDK

Now use the token in your client application.

### Web (JavaScript)

```html
<!DOCTYPE html>
<html>
<head>
  <title>My Video App</title>
  <script src="https://cdn.jsdelivr.net/npm/livekit-client/dist/livekit-client.umd.min.js"></script>
</head>
<body>
  <div id="video-container"></div>
  
  <script>
    const token = 'YOUR_TOKEN_FROM_STEP_3';
    const wsUrl = 'wss://livekit.pulse.io'; // Your LiveKit server URL
    
    async function connectToRoom() {
      const room = new LivekitClient.Room({
        adaptiveStream: true,
        dynacast: true,
      });
      
      room.on('participantConnected', (participant) => {
        console.log('Participant connected:', participant.identity);
      });
      
      room.on('trackSubscribed', (track, publication, participant) => {
        if (track.kind === 'video') {
          const element = track.attach();
          document.getElementById('video-container').appendChild(element);
        }
      });
      
      await room.connect(wsUrl, token);
      console.log('Connected to room:', room.name);
      
      // Enable camera and microphone
      await room.localParticipant.enableCameraAndMicrophone();
    }
    
    connectToRoom();
  </script>
</body>
</html>
```

### React

```javascript
import { useEffect, useState } from 'react';
import { Room, RoomEvent } from 'livekit-client';

function VideoRoom({ token }) {
  const [room, setRoom] = useState(null);
  
  useEffect(() => {
    const connectToRoom = async () => {
      const newRoom = new Room({
        adaptiveStream: true,
        dynacast: true,
      });
      
      newRoom.on(RoomEvent.ParticipantConnected, (participant) => {
        console.log('Participant connected:', participant.identity);
      });
      
      await newRoom.connect('wss://livekit.pulse.io', token);
      await newRoom.localParticipant.enableCameraAndMicrophone();
      
      setRoom(newRoom);
    };
    
    connectToRoom();
    
    return () => {
      room?.disconnect();
    };
  }, [token]);
  
  return (
    <div>
      <h1>Video Room</h1>
      <div id="video-grid"></div>
    </div>
  );
}
```

### Mobile (React Native)

```javascript
import { useRoom } from '@livekit/react-native';

function VideoCall() {
  const { connect, room } = useRoom();
  
  useEffect(() => {
    connect('wss://livekit.pulse.io', 'YOUR_TOKEN');
  }, []);
  
  return (
    <View>
      <Text>Video Call</Text>
      {/* Add video tracks here */}
    </View>
  );
}
```

---

## Step 5: Set Up Webhooks (Optional)

Receive notifications when events happen in your rooms.

### Update Project with Webhook URL

```bash
curl -X PUT https://api.pulse.io/api/v1/projects/507f1f77bcf86cd799439012 \
  -H "Content-Type: application/json" \
  -d '{
    "webhook_url": "https://myapp.com/webhooks/pulse"
  }'
```

### Handle Webhook in Your Server

**Node.js/Express**:
```javascript
const express = require('express');
const crypto = require('crypto');

const app = express();

function verifyWebhook(payload, signature, secret) {
  const expected = 'sha256=' + crypto
    .createHmac('sha256', secret)
    .update(payload)
    .digest('hex');
  return crypto.timingSafeEqual(Buffer.from(signature), Buffer.from(expected));
}

app.post('/webhooks/pulse', express.raw({ type: 'application/json' }), (req, res) => {
  const signature = req.headers['x-webhook-signature'];
  const payload = req.body.toString();
  
  if (!verifyWebhook(payload, signature, process.env.WEBHOOK_SECRET)) {
    return res.status(401).send('Invalid signature');
  }
  
  const event = JSON.parse(payload);
  console.log('Received webhook:', event.event);
  
  // Handle events
  switch (event.event) {
    case 'room.started':
      console.log('Room started:', event.data.room_name);
      break;
    case 'participant.joined':
      console.log('Participant joined:', event.data.identity);
      break;
  }
  
  res.status(200).send('OK');
});

app.listen(3000);
```

---

## Next Steps

Congratulations! š You've successfully:
- āļø Created an organization and project
- āļø Generated your first access token
- āļø Integrated Pulse into a client application
- āļø Set up webhooks (optional)

### What's Next?

1. **Explore Features**
   - [Chat Messaging](./docs/CHAT.md)
   - [Screen Sharing](./docs/SCREEN_SHARING.md)
   - [Recording & Streaming](./docs/RECORDING.md)

2. **Monitor Usage**
   - View usage metrics in the dashboard
   - Set up billing alerts
   - Check audit logs

3. **Production Readiness**
   - [Authentication Guide](./AUTHENTICATION.md)
   - [Scaling Guide](./SCALING.md)
   - [Security Best Practices](./SECURITY.md)

4. **Advanced Features**
   - Egress (streaming to RTMP, HLS)
   - Ingress (streaming from external sources)
   - Custom layouts and recordings
   - AI/ML integrations

---

## Common Patterns

### Server-Side Token Generation

Never expose API keys in client code! Always generate tokens server-side.

**Node.js/Express Example**:
```javascript
const axios = require('axios');

app.post('/api/get-token', async (req, res) => {
  const { userId, roomName } = req.body;
  
  try {
    const response = await axios.post(
      'https://api.pulse.io/api/v1/tokens/create',
      {
        identity: userId,
        room_name: roomName,
        metadata: {
          name: req.user.name,
          email: req.user.email
        }
      },
      {
        headers: {
          'X-Pulse-Key': process.env.PULSE_API_KEY,
          'Content-Type': 'application/json'
        }
      }
    );
    
    res.json({ token: response.data.token });
  } catch (error) {
    res.status(500).json({ error: 'Failed to generate token' });
  }
});
```

**Python/Flask Example**:
```python
import os
import requests
from flask import Flask, request, jsonify

app = Flask(__name__)

@app.route('/api/get-token', methods=['POST'])
def get_token():
    data = request.json
    user_id = data['userId']
    room_name = data['roomName']
    
    response = requests.post(
        'https://api.pulse.io/api/v1/tokens/create',
        json={
            'identity': user_id,
            'room_name': room_name,
            'metadata': {
                'name': request.user.name,
                'email': request.user.email
            }
        },
        headers={
            'X-Pulse-Key': os.environ['PULSE_API_KEY'],
            'Content-Type': 'application/json'
        }
    )
    
    return jsonify({'token': response.json()['token']})
```

---

## Troubleshooting

### Issue: "Invalid API key"

**Solution**: Verify your API key is correct:
```bash
# Test your API key
curl -X GET https://api.pulse.io/api/v1/projects \
  -H "X-Pulse-Key: YOUR_API_KEY"
```

### Issue: "Rate limit exceeded"

**Solution**: You're making too many requests. Current limits:
- 100 requests/minute per IP
- 1,000 requests/minute per project

### Issue: "Token expired"

**Solution**: Tokens have a default TTL of 1 hour. Generate a new token:
```javascript
if (error.message === 'Token expired') {
  const newToken = await fetchNewToken();
  await room.connect(wsUrl, newToken);
}
```

### Issue: Video not showing

**Checklist**:
- āļø Token is valid and not expired
- āļø WebSocket URL is correct
- āļø Camera/microphone permissions granted
- āļø HTTPS is used (required for WebRTC)
- āļø Firewall allows WebRTC traffic

---

## Resources

- š [Full API Reference](./API.md)
- š [Authentication Guide](./AUTHENTICATION.md)
- š [Webhook Guide](./WEBHOOKS.md)
- š [Scaling Guide](./SCALING.md)
- š¬ [Discord Community](https://discord.gg/pulse)
- š§ [Support Email](mailto:support@pulse.io)

---

## Example Projects

Check out our example implementations:

- [Pulse Video Chat (React)](https://github.com/pulse-io/example-react-video)
- [Pulse Webinar (Next.js)](https://github.com/pulse-io/example-webinar)
- [Pulse Live Stream (React Native)](https://github.com/pulse-io/example-mobile)
- [Pulse API Server (Node.js)](https://github.com/pulse-io/example-node-server)

---

**Need Help?** Join our [Discord](https://discord.gg/pulse) or email [support@pulse.io](mailto:support@pulse.io)
