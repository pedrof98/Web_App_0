import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.1.0/index.js';
import { Trend, Rate } from 'k6/metrics';

// Custom metrics
const ingestTrend = new Trend('ingest_time');
const successRate = new Rate('success_rate');

// Test configuration
export const options = {
  stages: [
    { duration: '1m', target: 10 }, // Ramp up to 10 users over 1 minute
    { duration: '3m', target: 10 }, // Stay at 10 users for 3 minutes
    { duration: '1m', target: 50 }, // Ramp up to 50 users over 1 minute
    { duration: '3m', target: 50 }, // Stay at 50 users for 3 minutes
    { duration: '1m', target: 100 }, // Ramp up to 100 users over 1 minute
    { duration: '5m', target: 100 }, // Stay at 100 users for 5 minutes
    { duration: '1m', target: 0 }, // Ramp down to 0 users over 1 minute
  ],
  thresholds: {
    'ingest_time': ['p(95)<500'], // 95% of requests should be below 500ms
    'success_rate': ['rate>0.95'], // 95% success rate
    'http_req_duration': ['p(95)<1000'], // 95% of requests should be below 1000ms
  },
};

// Helper to generate random IP addresses
function randomIP() {
  return `${randomIntBetween(1, 255)}.${randomIntBetween(0, 255)}.${randomIntBetween(0, 255)}.${randomIntBetween(0, 255)}`;
}

// Helper to generate random security events
function generateSecurityEvent() {
  const sourceTypes = ['system', 'network', 'application', 'vehicle', 'v2x', 'sensor'];
  const categories = ['authentication', 'authorization', 'network', 'malware', 'system', 'vehicle', 'v2x'];
  const severities = ['info', 'low', 'medium', 'high', 'critical'];
  
  // Weight probabilities toward less severe events
  let severityIndex = 0;
  const r = Math.random();
  if (r < 0.4) {
    severityIndex = 0; // 40% info
  } else if (r < 0.7) {
    severityIndex = 1; // 30% low
  } else if (r < 0.85) {
    severityIndex = 2; // 15% medium
  } else if (r < 0.95) {
    severityIndex = 3; // 10% high
  } else {
    severityIndex = 4; // 5% critical
  }
  
  const sourceIndex = randomIntBetween(0, sourceTypes.length - 1);
  const categoryIndex = randomIntBetween(0, categories.length - 1);
  
  return {
    source_name: `test_source_${sourceIndex}`,
    source_type: sourceTypes[sourceIndex],
    timestamp: new Date().toISOString(),
    severity: severities[severityIndex],
    category: categories[categoryIndex],
    message: `Load test event ${Date.now()}`,
    details: {
      source_ip: randomIP(),
      destination_ip: randomIP(),
      device_id: `device-${randomIntBetween(1, 1000)}`,
      protocol: Math.random() > 0.5 ? 'TCP' : 'UDP',
      status: Math.random() > 0.7 ? 'blocked' : 'allowed',
    }
  };
}

// Default function that is executed for each virtual user
export default function() {
  const baseUrl = __ENV.API_URL || 'http://localhost:8080';
  
  // Generate a random security event
  const payload = JSON.stringify(generateSecurityEvent());
  
  // Configure request
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };
  
  // Send event to ingestion endpoint
  const startTime = new Date();
  const res = http.post(`${baseUrl}/ingest`, payload, params);
  const duration = new Date() - startTime;
  
  // Record the time taken
  ingestTrend.add(duration);
  
  // Check if the request was successful
  const success = check(res, {
    'status is 200': (r) => r.status === 200,
    'has event_id': (r) => JSON.parse(r.body).event_id !== undefined,
  });
  
  successRate.add(success);
  
  // Random sleep between 0.5 and 2 seconds to simulate real users
  sleep(randomIntBetween(0.5, 2));
}
