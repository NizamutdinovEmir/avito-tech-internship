import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export let options = {
  stages: [
    { duration: '30s', target: 1 },
    { duration: '1m', target: 1 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(99)<300'],
    'http_req_failed{name:!setup_create_team}': ['rate<0.001'],
    errors: ['rate<0.001'],
    http_reqs: ['rate>=4.8', 'rate<=5.2'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export function setup() {
  const teamPayload = JSON.stringify({
    team_name: 'load-test-team',
    members: [
      { user_id: 'lt-u1', username: 'LoadTestUser1', is_active: true },
      { user_id: 'lt-u2', username: 'LoadTestUser2', is_active: true },
      { user_id: 'lt-u3', username: 'LoadTestUser3', is_active: true },
    ],
  });

  const teamRes = http.post(`${BASE_URL}/team/add`, teamPayload, {
    headers: { 'Content-Type': 'application/json' },
    tags: { name: 'setup_create_team' },
  });

  // Team might already exist from previous test run - that's OK
  // Accept both 201 (created) and 400 (already exists) as success
  const teamCreated = teamRes.status === 201 || teamRes.status === 400;
  
  // Don't log error if team already exists (400) - it's expected
  if (teamRes.status !== 201 && teamRes.status !== 400) {
    console.error('Failed to create test team:', teamRes.status, teamRes.body);
  }
  
  return { teamCreated: teamCreated };
}

export default function (data) {
  // Test 1: Health check
  let res = http.get(`${BASE_URL}/health`);
  let success = check(res, {
    'health check status is 200': (r) => r.status === 200,
    'health check response time < 300ms': (r) => r.timings.duration < 300,
  });
  errorRate.add(!success);
  
  // Test 2: Get team
  res = http.get(`${BASE_URL}/team/get?team_name=load-test-team`);
  success = check(res, {
    'get team status is 200': (r) => r.status === 200,
    'get team response time < 300ms': (r) => r.timings.duration < 300,
  });
  errorRate.add(!success);

  // Test 3: Create PR (less frequent - ~10% of iterations)
  // Use timestamp + iteration to ensure unique PR IDs
  if (Math.random() < 0.1) {
    const timestamp = Date.now();
    const prPayload = JSON.stringify({
      pull_request_id: `pr-load-${__VU}-${__ITER}-${timestamp}`,
      pull_request_name: 'Load Test PR',
      author_id: 'lt-u1',
    });

    res = http.post(`${BASE_URL}/pullRequest/create`, prPayload, {
      headers: { 'Content-Type': 'application/json' },
    });
    // Accept both 201 (created) and 400 (already exists) as success for load testing
    success = check(res, {
      'create PR status is 201 or 400': (r) => r.status === 201 || r.status === 400,
      'create PR response time < 300ms': (r) => r.timings.duration < 300,
    });
    errorRate.add(!success);
  }
  sleep(0.405);
}
