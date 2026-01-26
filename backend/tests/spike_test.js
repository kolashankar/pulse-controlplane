import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

// Spike test configuration - sudden traffic surge
export const options = {
  stages: [
    { duration: '10s', target: 10 },     // Start with 10 users
    { duration: '10s', target: 500 },    // Sudden spike to 500 users
    { duration: '30s', target: 500 },    // Stay at 500 users
    { duration: '10s', target: 10 },     // Drop back to 10 users
    { duration: '10s', target: 0 },      // Ramp down to 0
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'],   // 95% of requests under 1s during spike
    http_req_failed: ['rate<0.05'],      // Less than 5% errors
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8081';

export default function () {
  // Test concurrent requests during spike
  let responses = http.batch([
    ['GET', `${BASE_URL}/health`],
    ['GET', `${BASE_URL}/api/v1/status`],
    ['GET', `${BASE_URL}/api/v1/organizations`],
    ['GET', `${BASE_URL}/api/v1/status/regions`],
  ]);

  responses.forEach((res) => {
    check(res, {
      'status is 200': (r) => r.status === 200,
    }) || errorRate.add(1);
  });

  sleep(0.5);
}
