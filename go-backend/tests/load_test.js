import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

// Test configuration
export const options = {
  stages: [
    { duration: '30s', target: 50 },    // Ramp up to 50 users
    { duration: '1m', target: 100 },     // Ramp up to 100 users
    { duration: '2m', target: 200 },     // Ramp up to 200 users
    { duration: '1m', target: 100 },     // Ramp down to 100 users
    { duration: '30s', target: 0 },      // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests should be below 500ms
    http_req_failed: ['rate<0.01'],   // Error rate should be less than 1%
    errors: ['rate<0.05'],            // Custom error rate should be less than 5%
  },
};

// Base URL - update this to your deployment URL
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8081';

// Test data
const API_KEY = __ENV.API_KEY || 'pulse_key_test123';
const API_SECRET = __ENV.API_SECRET || 'pulse_secret_test456';

export default function () {
  // Test 1: Health check
  let healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, {
    'health check status is 200': (r) => r.status === 200,
    'health check returns healthy': (r) => JSON.parse(r.body).status === 'healthy',
  }) || errorRate.add(1);

  sleep(1);

  // Test 2: System status (no auth required)
  let statusRes = http.get(`${BASE_URL}/api/v1/status`);
  check(statusRes, {
    'status endpoint returns 200': (r) => r.status === 200,
  }) || errorRate.add(1);

  sleep(1);

  // Test 3: Create token (requires auth)
  const tokenPayload = JSON.stringify({
    identity: `user_${Math.floor(Math.random() * 10000)}`,
    room_name: `room_${Math.floor(Math.random() * 100)}`,
  });

  const tokenHeaders = {
    'Content-Type': 'application/json',
    'X-Pulse-Key': API_KEY,
  };

  let tokenRes = http.post(
    `${BASE_URL}/api/v1/tokens/create`,
    tokenPayload,
    { headers: tokenHeaders }
  );

  check(tokenRes, {
    'token creation status is 200 or 401': (r) => r.status === 200 || r.status === 401,
  });

  sleep(1);

  // Test 4: List organizations (no auth required for testing)
  let orgsRes = http.get(`${BASE_URL}/api/v1/organizations`);
  check(orgsRes, {
    'organizations endpoint accessible': (r) => r.status === 200,
  }) || errorRate.add(1);

  sleep(2);
}

// Scenario for testing specific endpoints
export function tokenCreationLoad() {
  const payload = JSON.stringify({
    identity: `load_test_user_${__VU}_${__ITER}`,
    room_name: `load_test_room_${Math.floor(__VU / 10)}`,
  });

  const headers = {
    'Content-Type': 'application/json',
    'X-Pulse-Key': API_KEY,
  };

  let res = http.post(`${BASE_URL}/api/v1/tokens/create`, payload, { headers });
  
  check(res, {
    'token created successfully': (r) => r.status === 200,
    'response has token': (r) => JSON.parse(r.body).token !== undefined,
  });
}

// Scenario for stress testing
export function stressTest() {
  // Rapid requests to test rate limiting
  for (let i = 0; i < 10; i++) {
    http.get(`${BASE_URL}/health`);
  }
  
  sleep(0.1);
}
