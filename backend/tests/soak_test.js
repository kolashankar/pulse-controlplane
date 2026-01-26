import http from 'k6/http';
import { check, sleep } from 'k6';

// Soak test - sustained load over extended period
export const options = {
  stages: [
    { duration: '2m', target: 50 },      // Ramp up to 50 users
    { duration: '10m', target: 50 },     // Sustain 50 users for 10 minutes
    { duration: '2m', target: 0 },       // Ramp down to 0
  ],
  thresholds: {
    http_req_duration: ['p(99)<800'],    // 99% under 800ms
    http_req_failed: ['rate<0.01'],      // Less than 1% errors
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8081';

export default function () {
  // Simulate realistic user behavior
  
  // User visits health page
  http.get(`${BASE_URL}/health`);
  sleep(2);

  // User checks system status
  http.get(`${BASE_URL}/api/v1/status`);
  sleep(3);

  // User views organizations
  http.get(`${BASE_URL}/api/v1/organizations`);
  sleep(2);

  // User checks region availability
  http.get(`${BASE_URL}/api/v1/status/regions`);
  sleep(5);

  // User views projects
  http.get(`${BASE_URL}/api/v1/projects`);
  sleep(3);
}
